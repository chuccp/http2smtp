package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
)

// maintenanceFixture serves the operation-log and mail-log maintenance endpoints over
// real HTTP, so the session lookup and the admin gate run exactly as they do in the app.
type maintenanceFixture struct {
	ts         *httptest.Server
	oplogModel *model.OperationLogModel
	logModel   *model.LogModel
	admin      *model.User
	regular    *model.User
}

func newMaintenanceFixture(t *testing.T) *maintenanceFixture {
	t.Helper()
	conn, err := db.ConnectionSQLite(filepath.Join(t.TempDir(), "maintenance.db"))
	if err != nil {
		t.Fatalf("opening sqlite: %v", err)
	}
	sqlDB, err := conn.GetGorm().DB()
	if err != nil {
		t.Fatalf("getting sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	oplogModel := &model.OperationLogModel{}
	if err := oplogModel.Init(conn, nil); err != nil {
		t.Fatalf("init oplog model: %v", err)
	}
	if err := oplogModel.CreateTable(); err != nil {
		t.Fatalf("create t_operation_log: %v", err)
	}
	logModel := &model.LogModel{}
	if err := logModel.Init(conn, nil); err != nil {
		t.Fatalf("init log model: %v", err)
	}
	if err := logModel.CreateTable(); err != nil {
		t.Fatalf("create t_log: %v", err)
	}
	userModel := &model.UserModel{}
	if err := userModel.Init(conn, nil); err != nil {
		t.Fatalf("init user model: %v", err)
	}
	if err := userModel.CreateTable(); err != nil {
		t.Fatalf("create t_user: %v", err)
	}

	coreCtx := core.NewContext(config.NewConfig(), context.Background())
	coreCtx.AddModel(oplogModel, logModel, userModel)
	oplogService := &service.OperationLogService{}
	if err := oplogService.Init(coreCtx); err != nil {
		t.Fatalf("init oplog service: %v", err)
	}
	coreCtx.AddService(oplogService)

	admin := &model.User{Name: "admin", Password: "x", Salt: "salt-admin", IsUse: true, IsAdmin: true}
	regular := &model.User{Name: "alice", Password: "x", Salt: "salt-alice", IsUse: true}
	for _, u := range []*model.User{admin, regular} {
		if err := userModel.Save(u); err != nil {
			t.Fatalf("save user %s: %v", u.Name, err)
		}
	}

	oplog := &Oplog{context: coreCtx, oplogModel: oplogModel}
	logs := &Log{context: coreCtx, logModel: logModel}
	servers := web.NewServers()
	server, err := servers.CreateServer(&web.ServerConfig{Port: 0})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	server.Delete("/api/oplog/:id", oplog.deleteOne)
	server.Post("/api/oplog/cleanup", oplog.cleanUp)
	server.Delete("/api/log/:id", logs.deleteOne)
	server.Post("/api/log/cleanup", logs.cleanUp)

	ts := httptest.NewServer(server.GetHandler())
	t.Cleanup(ts.Close)
	return &maintenanceFixture{ts: ts, oplogModel: oplogModel, logModel: logModel, admin: admin, regular: regular}
}

// do sends a request with the given user's session cookie (nil for anonymous) and
// returns the status and body.
func (f *maintenanceFixture) do(t *testing.T, method, path string, user *model.User, body string) (int, string) {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, f.ts.URL+path, reader)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if user != nil {
		token, err := auth.Encrypt(&entity.LoginUser{Id: user.Id, Salt: user.Salt})
		if err != nil {
			t.Fatalf("encrypting session: %v", err)
		}
		req.AddCookie(&http.Cookie{Name: entity.UserToken, Value: token})
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(payload)
}

func (f *maintenanceFixture) saveOplog(t *testing.T, name string) *model.OperationLog {
	t.Helper()
	row := &model.OperationLog{UserId: f.regular.Id, Action: entity.OpCreate, TargetType: entity.TargetSMTP, TargetName: name}
	if err := f.oplogModel.Save(row); err != nil {
		t.Fatalf("save oplog: %v", err)
	}
	return row
}

func (f *maintenanceFixture) saveLog(t *testing.T, subject string) *model.Log {
	t.Helper()
	row := &model.Log{UserId: f.regular.Id, Subject: subject, Status: entity.SUCCESS}
	if err := f.logModel.Save(row); err != nil {
		t.Fatalf("save log: %v", err)
	}
	return row
}

func TestOplogDeleteRequiresAdmin(t *testing.T) {
	f := newMaintenanceFixture(t)
	row := f.saveOplog(t, "demo")

	status, body := f.do(t, http.MethodDelete, fmt.Sprintf("/api/oplog/%d", row.Id), f.regular, "")
	if status != http.StatusInternalServerError || !strings.Contains(body, "admin access required") {
		t.Fatalf("regular user: status %d body %q, want the admin refusal", status, body)
	}
	if one, _ := f.oplogModel.FindByPK(row.Id); one == nil {
		t.Fatal("a regular user deleted an operation-log entry")
	}

	status, body = f.do(t, http.MethodDelete, fmt.Sprintf("/api/oplog/%d", row.Id), f.admin, "")
	if status != http.StatusOK {
		t.Fatalf("admin delete: status %d body %q, want 200", status, body)
	}
	if one, _ := f.oplogModel.FindByPK(row.Id); one != nil {
		t.Fatal("the entry is still present after the admin delete")
	}
}

func TestOplogDeleteRejectsAnonymous(t *testing.T) {
	f := newMaintenanceFixture(t)
	row := f.saveOplog(t, "demo")

	status, _ := f.do(t, http.MethodDelete, fmt.Sprintf("/api/oplog/%d", row.Id), nil, "")
	if status == http.StatusOK {
		t.Fatal("an anonymous request deleted an operation-log entry")
	}
	if one, _ := f.oplogModel.FindByPK(row.Id); one == nil {
		t.Fatal("the entry is gone after an anonymous delete")
	}
}

func TestOplogCleanupRequiresAdmin(t *testing.T) {
	f := newMaintenanceFixture(t)
	f.saveOplog(t, "one")
	f.saveOplog(t, "two")

	status, body := f.do(t, http.MethodPost, "/api/oplog/cleanup", f.regular, `{"days":0}`)
	if status != http.StatusInternalServerError || !strings.Contains(body, "admin access required") {
		t.Fatalf("regular user: status %d body %q, want the admin refusal", status, body)
	}
	if rows, _ := f.oplogModel.Query().All(); len(rows) != 2 {
		t.Fatalf("rows after a regular cleanup = %d, want 2", len(rows))
	}

	status, body = f.do(t, http.MethodPost, "/api/oplog/cleanup", f.admin, `{"days":0}`)
	if status != http.StatusOK {
		t.Fatalf("admin cleanup: status %d body %q, want 200", status, body)
	}
	// The cleanup itself is audited, so exactly one entry remains.
	rows, err := f.oplogModel.Query().All()
	if err != nil {
		t.Fatalf("reading rows: %v", err)
	}
	if len(rows) != 1 || rows[0].Action != entity.OpCleanup {
		t.Fatalf("rows after the admin cleanup = %+v, want one OpCleanup entry", rows)
	}
}

func TestMailLogDeleteRequiresAdmin(t *testing.T) {
	f := newMaintenanceFixture(t)
	row := f.saveLog(t, "hello")

	status, body := f.do(t, http.MethodDelete, fmt.Sprintf("/api/log/%d", row.Id), f.regular, "")
	if status != http.StatusInternalServerError || !strings.Contains(body, "admin access required") {
		t.Fatalf("regular user: status %d body %q, want the admin refusal", status, body)
	}
	if one, _ := f.logModel.FindByPK(row.Id); one == nil {
		t.Fatal("a regular user deleted a mail-log entry")
	}

	status, body = f.do(t, http.MethodDelete, fmt.Sprintf("/api/log/%d", row.Id), f.admin, "")
	if status != http.StatusOK {
		t.Fatalf("admin delete: status %d body %q, want 200", status, body)
	}
	if one, _ := f.logModel.FindByPK(row.Id); one != nil {
		t.Fatal("the mail-log entry is still present after the admin delete")
	}
}

func TestMailLogCleanupRequiresAdmin(t *testing.T) {
	f := newMaintenanceFixture(t)
	f.saveLog(t, "one")
	f.saveLog(t, "two")

	status, body := f.do(t, http.MethodPost, "/api/log/cleanup", f.regular, `{"days":0}`)
	if status != http.StatusInternalServerError || !strings.Contains(body, "admin access required") {
		t.Fatalf("regular user: status %d body %q, want the admin refusal", status, body)
	}
	if rows, _ := f.logModel.Query().All(); len(rows) != 2 {
		t.Fatalf("mail logs after a regular cleanup = %d, want 2", len(rows))
	}

	status, body = f.do(t, http.MethodPost, "/api/log/cleanup", f.admin, `{"days":0}`)
	if status != http.StatusOK {
		t.Fatalf("admin cleanup: status %d body %q, want 200", status, body)
	}
	if rows, _ := f.logModel.Query().All(); len(rows) != 0 {
		t.Fatalf("mail logs after the admin cleanup = %d, want 0", len(rows))
	}
}
