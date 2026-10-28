package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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
	"github.com/chuccp/http2smtp/smtp"
)

// maintenanceFixture serves the operation-log and mail-log maintenance endpoints over
// real HTTP, so the session lookup and the admin gate run exactly as they do in the app.
type maintenanceFixture struct {
	ts              *httptest.Server
	oplogModel      *model.OperationLogModel
	logModel        *model.LogModel
	attachmentModel *model.AttachmentModel
	admin           *model.User
	regular         *model.User
	other           *model.User
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
	attachmentModel := &model.AttachmentModel{}
	if err := attachmentModel.Init(conn, nil); err != nil {
		t.Fatalf("init attachment model: %v", err)
	}
	if err := attachmentModel.CreateTable(); err != nil {
		t.Fatalf("create t_attachment: %v", err)
	}
	userModel := &model.UserModel{}
	if err := userModel.Init(conn, nil); err != nil {
		t.Fatalf("init user model: %v", err)
	}
	if err := userModel.CreateTable(); err != nil {
		t.Fatalf("create t_user: %v", err)
	}

	coreCtx := core.NewContext(config.NewConfig(), context.Background())
	coreCtx.AddModel(oplogModel, logModel, attachmentModel, userModel)
	oplogService := &service.OperationLogService{}
	if err := oplogService.Init(coreCtx); err != nil {
		t.Fatalf("init oplog service: %v", err)
	}
	coreCtx.AddService(oplogService)

	admin := &model.User{Name: "admin", Password: "x", Salt: "salt-admin", IsUse: true, IsAdmin: true}
	regular := &model.User{Name: "alice", Password: "x", Salt: "salt-alice", IsUse: true}
	other := &model.User{Name: "bob", Password: "x", Salt: "salt-bob", IsUse: true}
	for _, u := range []*model.User{admin, regular, other} {
		if err := userModel.Save(u); err != nil {
			t.Fatalf("save user %s: %v", u.Name, err)
		}
	}

	oplog := &Oplog{context: coreCtx, oplogModel: oplogModel}
	logs := &Log{context: coreCtx, logModel: logModel, attachmentModel: attachmentModel}
	servers := web.NewServers()
	server, err := servers.CreateServer(&web.ServerConfig{Port: 0})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	server.Delete("/api/oplog/:id", oplog.deleteOne)
	server.Post("/api/oplog/cleanup", oplog.cleanUp)
	server.Delete("/api/log/:id", logs.deleteOne)
	server.Post("/api/log/cleanup", logs.cleanUp)
	server.Get("/api/download", logs.downLoad)

	ts := httptest.NewServer(server.GetHandler())
	t.Cleanup(ts.Close)
	return &maintenanceFixture{
		ts:              ts,
		oplogModel:      oplogModel,
		logModel:        logModel,
		attachmentModel: attachmentModel,
		admin:           admin,
		regular:         regular,
		other:           other,
	}
}

// doResponse sends a request with the given user's session cookie (nil for anonymous)
// and returns the response and body.
func (f *maintenanceFixture) doResponse(t *testing.T, method, path string, user *model.User, body string) (*http.Response, string) {
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
	return resp, string(payload)
}

func (f *maintenanceFixture) do(t *testing.T, method, path string, user *model.User, body string) (int, string) {
	t.Helper()
	resp, payload := f.doResponse(t, method, path, user, body)
	return resp.StatusCode, payload
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

func (f *maintenanceFixture) saveAttachment(t *testing.T, logId uint, name, path string) *model.Attachment {
	t.Helper()
	row := &model.Attachment{LogId: logId, Name: name, FilePath: path}
	if err := f.attachmentModel.Save(row); err != nil {
		t.Fatalf("save attachment: %v", err)
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
	path := filepath.Join(t.TempDir(), "stored.bin")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	f.saveAttachment(t, row.Id, "report.txt", path)

	status, body := f.do(t, http.MethodDelete, fmt.Sprintf("/api/log/%d", row.Id), f.regular, "")
	if status != http.StatusInternalServerError || !strings.Contains(body, "admin access required") {
		t.Fatalf("regular user: status %d body %q, want the admin refusal", status, body)
	}
	if one, _ := f.logModel.FindByPK(row.Id); one == nil {
		t.Fatal("a regular user deleted a mail-log entry")
	}
	if rows, _ := f.attachmentModel.FindByLogId(row.Id); len(rows) != 1 {
		t.Fatalf("attachments after a regular delete = %d, want 1", len(rows))
	}

	status, body = f.do(t, http.MethodDelete, fmt.Sprintf("/api/log/%d", row.Id), f.admin, "")
	if status != http.StatusOK {
		t.Fatalf("admin delete: status %d body %q, want 200", status, body)
	}
	if one, _ := f.logModel.FindByPK(row.Id); one != nil {
		t.Fatal("the mail-log entry is still present after the admin delete")
	}
	if rows, _ := f.attachmentModel.FindByLogId(row.Id); len(rows) != 0 {
		t.Fatalf("attachments after the admin delete = %d, want 0", len(rows))
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("the cached attachment should be gone, stat err = %v", err)
	}
}

// Two attachments sent under one filename are stored under distinct cache paths, so
// the download must address the position rather than collapse both onto the first copy.
func TestDownloadAttachmentByIndex(t *testing.T) {
	f := newMaintenanceFixture(t)
	dir := t.TempDir()
	first := filepath.Join(dir, "one.txt")
	second := filepath.Join(dir, "two.txt")
	if err := os.WriteFile(first, []byte("first"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("second"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := json.Marshal([]*smtp.File{
		{Name: "report.txt", FilePath: first},
		{Name: "report.txt", FilePath: second},
	})
	if err != nil {
		t.Fatal(err)
	}
	row := &model.Log{UserId: f.regular.Id, Subject: "attachments", Files: string(files), Status: entity.SUCCESS}
	if err := f.logModel.Save(row); err != nil {
		t.Fatalf("save log: %v", err)
	}
	base := fmt.Sprintf("/api/download?logId=%d", row.Id)

	resp, body := f.doResponse(t, http.MethodGet, base+"&index=1", f.regular, "")
	if resp.StatusCode != http.StatusOK || body != "second" {
		t.Fatalf("index 1: status %d body %q, want 200/second", resp.StatusCode, body)
	}
	if cd := resp.Header.Get("Content-Disposition"); !strings.Contains(cd, "report.txt") {
		t.Errorf("Content-Disposition = %q, want the original filename", cd)
	}

	resp, body = f.doResponse(t, http.MethodGet, base+"&index=0", f.regular, "")
	if resp.StatusCode != http.StatusOK || body != "first" {
		t.Fatalf("index 0: status %d body %q, want 200/first", resp.StatusCode, body)
	}

	// The name-only form keeps old links working, resolving to the first copy.
	resp, body = f.doResponse(t, http.MethodGet, base+"&file=report.txt", f.regular, "")
	if resp.StatusCode != http.StatusOK || body != "first" {
		t.Fatalf("name lookup: status %d body %q, want 200/first", resp.StatusCode, body)
	}

	status, _ := f.do(t, http.MethodGet, base+"&index=2", f.regular, "")
	if status == http.StatusOK {
		t.Error("an out-of-range index should not download anything")
	}
	status, _ = f.do(t, http.MethodGet, base+"&index=0", nil, "")
	if status == http.StatusOK {
		t.Error("an anonymous request should not download an attachment")
	}
}

func TestMailLogCleanupRequiresAdmin(t *testing.T) {
	f := newMaintenanceFixture(t)
	one := f.saveLog(t, "one")
	two := f.saveLog(t, "two")
	dir := t.TempDir()
	first := filepath.Join(dir, "one.bin")
	second := filepath.Join(dir, "two.bin")
	for _, path := range []string{first, second} {
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	f.saveAttachment(t, one.Id, "one.bin", first)
	f.saveAttachment(t, two.Id, "two.bin", second)

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
	if rows, _ := f.attachmentModel.Query().All(); len(rows) != 0 {
		t.Fatalf("attachment rows after the admin cleanup = %d, want 0", len(rows))
	}
	for _, path := range []string{first, second} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("cached attachment %s should be gone, stat err = %v", path, err)
		}
	}
}

// A new-style link carries the attachment id; the parent log gates who may fetch it.
func TestDownloadAttachmentById(t *testing.T) {
	f := newMaintenanceFixture(t)
	entry := f.saveLog(t, "with attachment")
	path := filepath.Join(t.TempDir(), "stored.bin")
	if err := os.WriteFile(path, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	attachment := f.saveAttachment(t, entry.Id, "report.txt", path)
	downloadURL := fmt.Sprintf("/api/download?id=%d", attachment.Id)

	resp, body := f.doResponse(t, http.MethodGet, downloadURL, f.regular, "")
	if resp.StatusCode != http.StatusOK || body != "content" {
		t.Fatalf("owner download: status %d body %q, want 200/content", resp.StatusCode, body)
	}
	if cd := resp.Header.Get("Content-Disposition"); !strings.Contains(cd, "report.txt") {
		t.Errorf("Content-Disposition = %q, want the original filename", cd)
	}

	status, body := f.do(t, http.MethodGet, downloadURL, f.admin, "")
	if status != http.StatusOK || body != "content" {
		t.Fatalf("admin download: status %d body %q, want 200/content", status, body)
	}

	status, _ = f.do(t, http.MethodGet, downloadURL, f.other, "")
	if status == http.StatusOK {
		t.Error("another user's session should not download the attachment")
	}
	status, _ = f.do(t, http.MethodGet, downloadURL, nil, "")
	if status == http.StatusOK {
		t.Error("an anonymous request should not download the attachment")
	}

	// An attachment row whose log is gone is unreachable, even for its former owner.
	orphanLog := f.saveLog(t, "gone")
	orphan := f.saveAttachment(t, orphanLog.Id, "orphan.txt", path)
	if err := f.logModel.DeleteByPK(orphanLog.Id); err != nil {
		t.Fatalf("deleting the orphan's log: %v", err)
	}
	status, _ = f.do(t, http.MethodGet, fmt.Sprintf("/api/download?id=%d", orphan.Id), f.regular, "")
	if status == http.StatusOK {
		t.Error("an attachment whose log is gone should not download")
	}
}
