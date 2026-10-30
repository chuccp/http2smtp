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

// scheduleTokenFixture serves the schedule create/update endpoints over real HTTP. The
// rule under test: a schedule may only reference an in-use token — the schedule owner's
// own for a regular caller, any user's for an admin — which is the same rule the form's
// picker and its save enforce.
type scheduleTokenFixture struct {
	ts            *httptest.Server
	tokenModel    *model.TokenModel
	scheduleModel *model.ScheduleModel
	alice         *model.User
	bob           *model.User
	admin         *model.User
}

func newScheduleTokenFixture(t *testing.T) *scheduleTokenFixture {
	t.Helper()
	conn, err := db.ConnectionSQLite(filepath.Join(t.TempDir(), "schedtoken.db"))
	if err != nil {
		t.Fatalf("opening sqlite: %v", err)
	}
	sqlDB, err := conn.GetGorm().DB()
	if err != nil {
		t.Fatalf("getting sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	userModel := &model.UserModel{}
	smtpModel := &model.SMTPModel{}
	mailModel := &model.MailModel{}
	tokenModel := &model.TokenModel{}
	logModel := &model.LogModel{}
	attachModel := &model.AttachmentModel{}
	schedModel := &model.ScheduleModel{}
	oplogModel := &model.OperationLogModel{}
	for _, m := range []interface {
		Init(*db.DB, *core.Context) error
		CreateTable() error
	}{userModel, smtpModel, mailModel, tokenModel, logModel, attachModel, schedModel, oplogModel} {
		if err := m.Init(conn, nil); err != nil {
			t.Fatalf("init model: %v", err)
		}
		if err := m.CreateTable(); err != nil {
			t.Fatalf("create table: %v", err)
		}
	}

	coreCtx := core.NewContext(config.NewConfig(), context.Background())
	coreCtx.AddModel(userModel, smtpModel, mailModel, tokenModel, logModel, attachModel, schedModel, oplogModel)
	logService := &service.LogService{}
	if err := logService.Init(coreCtx); err != nil {
		t.Fatalf("init log service: %v", err)
	}
	coreCtx.AddService(logService)
	oplogService := &service.OperationLogService{}
	if err := oplogService.Init(coreCtx); err != nil {
		t.Fatalf("init oplog service: %v", err)
	}
	coreCtx.AddService(oplogService)
	tokenService := &service.TokenService{}
	if err := tokenService.Init(coreCtx); err != nil {
		t.Fatalf("init token service: %v", err)
	}
	coreCtx.AddService(tokenService)
	scheduleService := &service.ScheduleService{}
	if err := scheduleService.Init(coreCtx); err != nil {
		t.Fatalf("init schedule service: %v", err)
	}
	coreCtx.AddService(scheduleService)

	f := &scheduleTokenFixture{tokenModel: tokenModel, scheduleModel: schedModel}
	f.alice = &model.User{Name: "alice", Password: "x", Salt: "salt-alice", IsUse: true}
	f.bob = &model.User{Name: "bob", Password: "x", Salt: "salt-bob", IsUse: true}
	f.admin = &model.User{Name: "admin", Password: "x", Salt: "salt-admin", IsUse: true, IsAdmin: true}
	for _, u := range []*model.User{f.alice, f.bob, f.admin} {
		if err := userModel.Save(u); err != nil {
			t.Fatalf("save user %s: %v", u.Name, err)
		}
	}

	schedRest := &Schedule{context: coreCtx, scheduleService: scheduleService, tokenService: tokenService, scheduleModel: schedModel}
	servers := web.NewServers()
	server, err := servers.CreateServer(&web.ServerConfig{Port: 0})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	server.Post("/api/schedule", schedRest.postOne)
	server.Put("/api/schedule", schedRest.putOne)
	f.ts = httptest.NewServer(server.GetHandler())
	t.Cleanup(f.ts.Close)
	return f
}

func (f *scheduleTokenFixture) saveToken(t *testing.T, u *model.User, name string, state uint8) *model.Token {
	t.Helper()
	tk := &model.Token{UserId: u.Id, Name: name, Token: name + "-secret", SMTPId: 1, State: state}
	if err := f.tokenModel.Save(tk); err != nil {
		t.Fatalf("save token: %v", err)
	}
	return tk
}

func (f *scheduleTokenFixture) do(t *testing.T, method, body string, user *model.User) string {
	t.Helper()
	req, err := http.NewRequest(method, f.ts.URL+"/api/schedule", strings.NewReader(body))
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
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
	return string(payload)
}

func scheduleBody(tokenId uint, name string) string {
	return fmt.Sprintf(`{"name":%q,"tokenId":%d,"cron":"0 0 0/1 * * ?","url":"http://example.com/hook","method":"GET","body":"","state":0}`, name, tokenId)
}

// A schedule may only reference an in-use token: the picker only offers active tokens
// and the form save rejects anything else ("token not found"). The REST save applies
// the same rule — for admins too, who may point a schedule at any user's active token.
func TestScheduleCreateTokenRules(t *testing.T) {
	f := newScheduleTokenFixture(t)
	aliceActive := f.saveToken(t, f.alice, "alice-active", entity.TokenStateInUse)
	aliceDisabled := f.saveToken(t, f.alice, "alice-disabled", entity.TokenStateUserDisabled)
	bobActive := f.saveToken(t, f.bob, "bob-active", entity.TokenStateInUse)
	bobDisabled := f.saveToken(t, f.bob, "bob-disabled", entity.TokenStateAdminDisabled)

	// The owner's own active token passes.
	if out := f.do(t, http.MethodPost, scheduleBody(aliceActive.Id, "ok"), f.alice); !strings.Contains(out, "ok") {
		t.Errorf("own active token: got %q, want ok", out)
	}

	// A foreign token is not the caller's to reference.
	if out := f.do(t, http.MethodPost, scheduleBody(bobActive.Id, "foreign"), f.alice); !strings.Contains(out, "token not found") {
		t.Errorf("foreign token: got %q, want \"token not found\"", out)
	}

	// A disabled token is not usable — the same rule the form enforces.
	if out := f.do(t, http.MethodPost, scheduleBody(aliceDisabled.Id, "disabled"), f.alice); !strings.Contains(out, "token not found") {
		t.Errorf("disabled own token: got %q, want \"token not found\"", out)
	}

	// An admin may point a schedule at any user's active token...
	if out := f.do(t, http.MethodPost, scheduleBody(bobActive.Id, "admin-ok"), f.admin); !strings.Contains(out, "ok") {
		t.Errorf("admin with a foreign active token: got %q, want ok", out)
	}
	// ...but not at a disabled one.
	if out := f.do(t, http.MethodPost, scheduleBody(bobDisabled.Id, "admin-disabled"), f.admin); !strings.Contains(out, "token not found") {
		t.Errorf("admin with a disabled token: got %q, want \"token not found\"", out)
	}
}

// The update path applies the same rules, and keeps the row with its original owner.
func TestScheduleUpdateTokenRules(t *testing.T) {
	f := newScheduleTokenFixture(t)
	aliceActive := f.saveToken(t, f.alice, "alice-active", entity.TokenStateInUse)
	aliceDisabled := f.saveToken(t, f.alice, "alice-disabled", entity.TokenStateUserDisabled)
	bobActive := f.saveToken(t, f.bob, "bob-active", entity.TokenStateInUse)

	created := &model.Schedule{UserId: f.alice.Id, Name: "s", TokenId: aliceActive.Id, IsUse: true}
	if err := f.scheduleModel.Save(created); err != nil {
		t.Fatalf("seed schedule: %v", err)
	}

	body := func(tokenId uint) string {
		return fmt.Sprintf(`{"id":%d,"name":"s","tokenId":%d,"cron":"0 0 0/1 * * ?","url":"http://example.com/hook","method":"GET","body":""}`, created.Id, tokenId)
	}

	// Re-pointing at a disabled token is refused and the row keeps its token.
	if out := f.do(t, http.MethodPut, body(aliceDisabled.Id), f.alice); !strings.Contains(out, "token not found") {
		t.Errorf("re-point at a disabled token: got %q, want \"token not found\"", out)
	}
	if got, _ := f.scheduleModel.FindByPK(created.Id); got == nil || got.TokenId != aliceActive.Id {
		t.Fatalf("row after refused update = %+v, want the original token kept", got)
	}

	// A regular user cannot edit someone else's row.
	bobSchedule := &model.Schedule{UserId: f.bob.Id, Name: "bob-s", TokenId: bobActive.Id, IsUse: true}
	if err := f.scheduleModel.Save(bobSchedule); err != nil {
		t.Fatalf("seed bob's schedule: %v", err)
	}
	foreign := fmt.Sprintf(`{"id":%d,"name":"x","tokenId":%d,"cron":"0 0 0/1 * * ?","url":"http://example.com/hook","method":"GET","body":""}`, bobSchedule.Id, aliceActive.Id)
	if out := f.do(t, http.MethodPut, foreign, f.alice); !strings.Contains(out, "schedule not found") {
		t.Errorf("foreign schedule edit: got %q, want \"schedule not found\"", out)
	}

	// An admin may re-point the row at any user's active token; the owner stays.
	if out := f.do(t, http.MethodPut, body(bobActive.Id), f.admin); !strings.Contains(out, "ok") {
		t.Fatalf("admin re-point: got %q, want ok", out)
	}
	got, _ := f.scheduleModel.FindByPK(created.Id)
	if got == nil || got.TokenId != bobActive.Id {
		t.Errorf("row after admin update = %+v, want the new token", got)
	}
	if got.UserId != f.alice.Id {
		t.Errorf("row owner = %d, want the original owner %d", got.UserId, f.alice.Id)
	}
}

// Malformed schedules are refused before anything touches a token.
func TestScheduleValidateRejectsBadInput(t *testing.T) {
	f := newScheduleTokenFixture(t)
	tk := f.saveToken(t, f.alice, "alice-active", entity.TokenStateInUse)

	cases := []struct {
		name, body string
		want       string
	}{
		{"bad cron", fmt.Sprintf(`{"name":"s","tokenId":%d,"cron":"not a cron","url":"http://example.com/hook","method":"GET"}`, tk.Id), "cron"},
		{"missing token", `{"name":"s","tokenId":0,"cron":"0 0 0/1 * * ?","url":"http://example.com/hook","method":"GET"}`, "tokenId"},
		{"bad url", fmt.Sprintf(`{"name":"s","tokenId":%d,"cron":"0 0 0/1 * * ?","url":"ftp://example.com/hook","method":"GET"}`, tk.Id), "http"},
	}
	for _, c := range cases {
		out := f.do(t, http.MethodPost, c.body, f.alice)
		if strings.Contains(out, `"ok"`) || !strings.Contains(out, c.want) {
			t.Errorf("%s: got %q, want an error mentioning %q", c.name, out, c.want)
		}
	}
}
