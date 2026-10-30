package rest

import (
	"context"
	"encoding/json"
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

// userAdminGuardFixture serves the user management endpoints over real HTTP. The rule
// under test is the one the webui already enforces: an admin row keeps its flags and
// cannot be deleted, because demoting or disabling the last active admin would lock
// the instance with no way back in.
type userAdminGuardFixture struct {
	ts        *httptest.Server
	userModel *model.UserModel
	admin     *model.User
	regular   *model.User
}

func newUserAdminGuardFixture(t *testing.T) *userAdminGuardFixture {
	t.Helper()
	conn, err := db.ConnectionSQLite(filepath.Join(t.TempDir(), "userguard.db"))
	if err != nil {
		t.Fatalf("opening sqlite: %v", err)
	}
	sqlDB, err := conn.GetGorm().DB()
	if err != nil {
		t.Fatalf("getting sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	userModel := &model.UserModel{}
	oplogModel := &model.OperationLogModel{}
	for _, m := range []interface {
		Init(*db.DB, *core.Context) error
		CreateTable() error
	}{userModel, oplogModel} {
		if err := m.Init(conn, nil); err != nil {
			t.Fatalf("init model: %v", err)
		}
		if err := m.CreateTable(); err != nil {
			t.Fatalf("create table: %v", err)
		}
	}

	coreCtx := core.NewContext(config.NewConfig(), context.Background())
	coreCtx.AddModel(userModel, oplogModel)
	userService := &service.UserService{}
	if err := userService.Init(coreCtx); err != nil {
		t.Fatalf("init user service: %v", err)
	}
	coreCtx.AddService(userService)
	oplogService := &service.OperationLogService{}
	if err := oplogService.Init(coreCtx); err != nil {
		t.Fatalf("init oplog service: %v", err)
	}
	coreCtx.AddService(oplogService)

	f := &userAdminGuardFixture{userModel: userModel}
	f.admin = &model.User{Name: "admin", Password: "x", Salt: "salt-admin", IsUse: true, IsAdmin: true}
	f.regular = &model.User{Name: "alice", Password: "x", Salt: "salt-alice", IsUse: true}
	for _, u := range []*model.User{f.admin, f.regular} {
		if err := userModel.Save(u); err != nil {
			t.Fatalf("save user %s: %v", u.Name, err)
		}
	}

	userRest := &User{context: coreCtx, userService: userService}
	servers := web.NewServers()
	server, err := servers.CreateServer(&web.ServerConfig{Port: 0})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	server.Post("/api/user", userRest.createUser)
	server.Put("/api/user", userRest.updateUser)
	server.Delete("/api/user/:id", userRest.deleteUser)
	f.ts = httptest.NewServer(server.GetHandler())
	t.Cleanup(f.ts.Close)
	return f
}

func (f *userAdminGuardFixture) do(t *testing.T, method, path, body string) string {
	t.Helper()
	req, err := http.NewRequest(method, f.ts.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	token, err := auth.Encrypt(&entity.LoginUser{Id: f.admin.Id, Salt: f.admin.Salt})
	if err != nil {
		t.Fatalf("encrypting session: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: entity.UserToken, Value: token})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(resp.Body)
	return string(payload)
}

func (f *userAdminGuardFixture) reload(t *testing.T, id uint) *model.User {
	t.Helper()
	row, err := f.userModel.FindByPK(id)
	if err != nil || row == nil {
		t.Fatalf("reload user %d: %v", id, err)
	}
	return row
}

// An admin row keeps its flags. The webui silently keeps them because its form posts
// the checkboxes whatever their state; the API refuses outright so the caller notices.
// Demoting or disabling an admin would leave the instance without one.
func TestUpdateUserCannotDemoteOrDisableAdmin(t *testing.T) {
	f := newUserAdminGuardFixture(t)
	other := &model.User{Name: "admin2", Password: "x", Salt: "salt-admin2", IsUse: true, IsAdmin: true}
	if err := f.userModel.Save(other); err != nil {
		t.Fatalf("save second admin: %v", err)
	}

	body, _ := json.Marshal(map[string]any{"id": other.Id, "name": "admin2", "isAdmin": false, "isUse": false})
	out := f.do(t, http.MethodPut, "/api/user", string(body))
	if !strings.Contains(out, "cannot be demoted or disabled") {
		t.Errorf("demote+disable admin: got %q, want the refusal", out)
	}
	got := f.reload(t, other.Id)
	if !got.IsAdmin || !got.IsUse {
		t.Errorf("admin row after refused update: IsAdmin=%v IsUse=%v, want both true", got.IsAdmin, got.IsUse)
	}

	// Demote-only and disable-only attempts are refused too.
	body, _ = json.Marshal(map[string]any{"id": other.Id, "name": "admin2", "isAdmin": false, "isUse": true})
	if out := f.do(t, http.MethodPut, "/api/user", string(body)); !strings.Contains(out, "cannot be demoted or disabled") {
		t.Errorf("demote admin: got %q, want the refusal", out)
	}
	body, _ = json.Marshal(map[string]any{"id": other.Id, "name": "admin2", "isAdmin": true, "isUse": false})
	if out := f.do(t, http.MethodPut, "/api/user", string(body)); !strings.Contains(out, "cannot be demoted or disabled") {
		t.Errorf("disable admin: got %q, want the refusal", out)
	}

	// Renaming an admin (or resetting their password) is still allowed.
	body, _ = json.Marshal(map[string]any{"id": other.Id, "name": "renamed", "password": "password123", "isAdmin": true, "isUse": true})
	if out := f.do(t, http.MethodPut, "/api/user", string(body)); !strings.Contains(out, "ok") {
		t.Fatalf("rename admin: got %q, want ok", out)
	}
	got = f.reload(t, other.Id)
	if got.Name != "renamed" || !got.IsAdmin || !got.IsUse {
		t.Errorf("admin row after rename: %+v, want renamed and still admin+active", got)
	}

	// A regular account is still fully manageable.
	body, _ = json.Marshal(map[string]any{"id": f.regular.Id, "name": "alice", "isAdmin": false, "isUse": false})
	if out := f.do(t, http.MethodPut, "/api/user", string(body)); !strings.Contains(out, "ok") {
		t.Fatalf("disable regular user: got %q, want ok", out)
	}
	if got = f.reload(t, f.regular.Id); got.IsUse {
		t.Error("the regular row stayed enabled after being disabled")
	}
}

// The webui offers no delete button on admin rows at all; the API refuses the same
// request. Deleting (disabling) the last admin would lock the instance out of its own
// management surface.
func TestDeleteUserRefusesAdminRows(t *testing.T) {
	f := newUserAdminGuardFixture(t)

	out := f.do(t, http.MethodDelete, fmt.Sprintf("/api/user/%d", f.admin.Id), "")
	if !strings.Contains(out, "cannot be deleted") {
		t.Errorf("delete admin: got %q, want the refusal", out)
	}
	if got := f.reload(t, f.admin.Id); !got.IsUse || !got.IsAdmin {
		t.Errorf("admin row after refused delete: %+v, want intact", got)
	}

	out = f.do(t, http.MethodDelete, fmt.Sprintf("/api/user/%d", f.regular.Id), "")
	if !strings.Contains(out, "ok") {
		t.Fatalf("delete regular user: got %q, want ok", out)
	}
	if got := f.reload(t, f.regular.Id); got.IsUse {
		t.Error("the regular row is still enabled after the delete")
	}
}

// The create endpoint keeps its own rules while the guards above hold.
func TestCreateUserValidatesPassword(t *testing.T) {
	f := newUserAdminGuardFixture(t)
	body, _ := json.Marshal(map[string]any{"name": "bob", "password": "short", "isAdmin": false, "isUse": true})
	out := f.do(t, http.MethodPost, "/api/user", string(body))
	if !strings.Contains(out, "password") {
		t.Errorf("short password: got %q, want the password refusal", out)
	}
	body, _ = json.Marshal(map[string]any{"name": "bob", "password": "longenough", "isAdmin": false, "isUse": true})
	if out := f.do(t, http.MethodPost, "/api/user", string(body)); !strings.Contains(out, "ok") {
		t.Errorf("valid create: got %q, want ok", out)
	}
}
