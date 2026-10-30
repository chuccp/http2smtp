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

// tokenOwnershipFixture serves the token create/update/delete endpoints and the other
// row deletes over real HTTP with real sessions, so the reference-ownership checks run
// exactly as they do in the app: a token may only point at its owner's own SMTP server
// and its own contacts, and delete stays with the owner except for admins.
type tokenOwnershipFixture struct {
	ts         *httptest.Server
	tokenModel *model.TokenModel
	smtpModel  *model.SMTPModel
	mailModel  *model.MailModel
	schedModel *model.ScheduleModel
	alice      *model.User
	bob        *model.User
	admin      *model.User
	aliceSMTP  *model.SMTP
	bobSMTP    *model.SMTP
	aliceMail  *model.Mail
	bobMail    *model.Mail
}

func newTokenOwnershipFixture(t *testing.T) *tokenOwnershipFixture {
	t.Helper()
	conn, err := db.ConnectionSQLite(filepath.Join(t.TempDir(), "token.db"))
	if err != nil {
		t.Fatalf("opening sqlite: %v", err)
	}
	sqlDB, err := conn.GetGorm().DB()
	if err != nil {
		t.Fatalf("getting sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	tokenModel := &model.TokenModel{}
	smtpModel := &model.SMTPModel{}
	mailModel := &model.MailModel{}
	schedModel := &model.ScheduleModel{}
	userModel := &model.UserModel{}
	oplogModel := &model.OperationLogModel{}
	for _, m := range []interface {
		Init(*db.DB, *core.Context) error
		CreateTable() error
	}{tokenModel, smtpModel, mailModel, schedModel, userModel, oplogModel} {
		if err := m.Init(conn, nil); err != nil {
			t.Fatalf("init model: %v", err)
		}
		if err := m.CreateTable(); err != nil {
			t.Fatalf("create table: %v", err)
		}
	}

	coreCtx := core.NewContext(config.NewConfig(), context.Background())
	coreCtx.AddModel(tokenModel, smtpModel, mailModel, schedModel, userModel, oplogModel)
	oplogService := &service.OperationLogService{}
	if err := oplogService.Init(coreCtx); err != nil {
		t.Fatalf("init oplog service: %v", err)
	}
	coreCtx.AddService(oplogService)

	f := &tokenOwnershipFixture{tokenModel: tokenModel, smtpModel: smtpModel, mailModel: mailModel, schedModel: schedModel}
	f.alice = &model.User{Name: "alice", Password: "x", Salt: "salt-alice", IsUse: true}
	f.bob = &model.User{Name: "bob", Password: "x", Salt: "salt-bob", IsUse: true}
	f.admin = &model.User{Name: "admin", Password: "x", Salt: "salt-admin", IsUse: true, IsAdmin: true}
	for _, u := range []*model.User{f.alice, f.bob, f.admin} {
		if err := userModel.Save(u); err != nil {
			t.Fatalf("save user %s: %v", u.Name, err)
		}
	}
	f.aliceSMTP = &model.SMTP{UserId: f.alice.Id, Name: "alice-smtp", Host: "127.0.0.1", Port: 1, Mail: "a@example.com"}
	f.bobSMTP = &model.SMTP{UserId: f.bob.Id, Name: "bob-smtp", Host: "127.0.0.1", Port: 1, Mail: "b@example.com"}
	for _, s := range []*model.SMTP{f.aliceSMTP, f.bobSMTP} {
		if err := smtpModel.Save(s); err != nil {
			t.Fatalf("save smtp: %v", err)
		}
	}
	f.aliceMail = &model.Mail{UserId: f.alice.Id, Name: "alice-contact", Mail: "ac@example.com"}
	f.bobMail = &model.Mail{UserId: f.bob.Id, Name: "bob-contact", Mail: "bc@example.com"}
	for _, m := range []*model.Mail{f.aliceMail, f.bobMail} {
		if err := mailModel.Save(m); err != nil {
			t.Fatalf("save mail: %v", err)
		}
	}

	tokenRest := &Token{context: coreCtx, tokenModel: tokenModel}
	smtpRest := &Smtp{context: coreCtx, smtpModel: smtpModel}
	mailRest := &Mail{context: coreCtx, mailModel: mailModel}
	schedRest := &Schedule{context: coreCtx, scheduleModel: schedModel}
	servers := web.NewServers()
	server, err := servers.CreateServer(&web.ServerConfig{Port: 0})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	server.Post("/api/token", tokenRest.postOne)
	server.Put("/api/token", tokenRest.putOne)
	server.Delete("/api/token/:id", tokenRest.deleteOne)
	server.Delete("/api/smtp/:id", smtpRest.deleteOne)
	server.Delete("/api/mail/:id", mailRest.deleteOne)
	server.Delete("/api/schedule/:id", schedRest.deleteOne)
	f.ts = httptest.NewServer(server.GetHandler())
	t.Cleanup(f.ts.Close)
	return f
}

func (f *tokenOwnershipFixture) do(t *testing.T, method, body string, user *model.User) string {
	t.Helper()
	return f.doPath(t, method, "/api/token", body, user)
}

func (f *tokenOwnershipFixture) doPath(t *testing.T, method, path, body string, user *model.User) string {
	t.Helper()
	req, err := http.NewRequest(method, f.ts.URL+path, strings.NewReader(body))
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

func (f *tokenOwnershipFixture) saveToken(t *testing.T, u *model.User, name, tokenValue string) *model.Token {
	t.Helper()
	row := &model.Token{UserId: u.Id, Name: name, Token: tokenValue, SMTPId: f.bobSMTP.Id}
	if err := f.tokenModel.Save(row); err != nil {
		t.Fatalf("save token: %v", err)
	}
	return row
}

// Creating a token that points at another account's SMTP server would mint a credential
// for that account's mailbox; a recipient list holding another account's contacts is the
// same leak. Both are refused on create.
func TestTokenCreateRejectsForeignReferences(t *testing.T) {
	f := newTokenOwnershipFixture(t)

	body := fmt.Sprintf(`{"name":"t1","subject":"s","token":"tk-1","SMTPId":%d,"receiveEmailIds":"%d","state":0}`,
		f.bobSMTP.Id, f.aliceMail.Id)
	out := f.do(t, http.MethodPost, body, f.alice)
	if !strings.Contains(out, "SMTP not found") {
		t.Errorf("foreign SMTP: got %q, want the SMTP refusal", out)
	}

	body = fmt.Sprintf(`{"name":"t1","subject":"s","token":"tk-1","SMTPId":%d,"receiveEmailIds":"%d","state":0}`,
		f.aliceSMTP.Id, f.bobMail.Id)
	out = f.do(t, http.MethodPost, body, f.alice)
	if !strings.Contains(out, "mail not found") {
		t.Errorf("foreign contact: got %q, want the contact refusal", out)
	}
	if rows, _ := f.tokenModel.Query().All(); len(rows) != 0 {
		t.Fatalf("a refused create left %d token rows behind", len(rows))
	}

	// The owner's own references pass.
	body = fmt.Sprintf(`{"name":"t1","subject":"s","token":"tk-1","SMTPId":%d,"receiveEmailIds":"%d","state":0}`,
		f.aliceSMTP.Id, f.aliceMail.Id)
	out = f.do(t, http.MethodPost, body, f.alice)
	if !strings.Contains(out, "ok") {
		t.Errorf("own references: got %q, want ok", out)
	}
	rows, _ := f.tokenModel.Query().All()
	if len(rows) != 1 || rows[0].UserId != f.alice.Id {
		t.Fatalf("created rows = %+v, want one row owned by alice", rows)
	}
}

// Create is create: a client-supplied id must not turn the insert into an UPDATE of an
// existing row, and a client-supplied userId must not decide the owner.
func TestTokenCreateIgnoresClientSuppliedId(t *testing.T) {
	f := newTokenOwnershipFixture(t)
	bobToken := f.saveToken(t, f.bob, "bob-token", "bob-secret")

	body := fmt.Sprintf(`{"id":%d,"userId":%d,"name":"hijack","subject":"s","token":"hijack-secret","SMTPId":%d,"state":0}`,
		bobToken.Id, f.bob.Id, f.aliceSMTP.Id)
	out := f.do(t, http.MethodPost, body, f.alice)
	if !strings.Contains(out, "ok") {
		t.Fatalf("create with a client id: got %q, want ok", out)
	}

	if got, _ := f.tokenModel.FindByPK(bobToken.Id); got == nil || got.Token != "bob-secret" || got.UserId != f.bob.Id {
		t.Fatalf("bob's row was overwritten: %+v", got)
	}
	rows, _ := f.tokenModel.Query().Where("token = ?", "hijack-secret").All()
	if len(rows) != 1 {
		t.Fatalf("rows with the new token = %d, want 1", len(rows))
	}
	if rows[0].UserId != f.alice.Id {
		t.Errorf("created row owner = %d, want the caller %d", rows[0].UserId, f.alice.Id)
	}
	if rows[0].Id == bobToken.Id {
		t.Error("the create reused bob's primary key")
	}
}

// An update keeps the row with its original owner and checks the referenced resources
// against that owner, not against the caller — an admin editing someone else's row must
// not silently re-point it at the admin's own credentials.
func TestTokenUpdatePreservesOwnerAndChecksReferences(t *testing.T) {
	f := newTokenOwnershipFixture(t)
	bobToken := f.saveToken(t, f.bob, "bob-token", "bob-secret")

	// A regular user cannot even see the row.
	out := f.do(t, http.MethodPut,
		fmt.Sprintf(`{"id":%d,"name":"x","subject":"s","SMTPId":%d,"state":0}`, bobToken.Id, f.aliceSMTP.Id),
		f.alice)
	if !strings.Contains(out, "token not found") {
		t.Errorf("foreign update: got %q, want \"token not found\"", out)
	}

	// An admin reaches the row, but its references are still checked against its owner.
	out = f.do(t, http.MethodPut,
		fmt.Sprintf(`{"id":%d,"name":"renamed","subject":"s","SMTPId":%d,"state":0}`, bobToken.Id, f.aliceSMTP.Id),
		f.admin)
	if !strings.Contains(out, "SMTP not found") {
		t.Errorf("admin re-point at their own SMTP: got %q, want the refusal", out)
	}

	out = f.do(t, http.MethodPut,
		fmt.Sprintf(`{"id":%d,"name":"renamed","subject":"s","SMTPId":%d,"state":0}`, bobToken.Id, f.bobSMTP.Id),
		f.admin)
	if !strings.Contains(out, "ok") {
		t.Fatalf("admin edit with the owner's references: got %q, want ok", out)
	}
	got, _ := f.tokenModel.FindByPK(bobToken.Id)
	if got == nil {
		t.Fatal("the row is gone after the admin edit")
	}
	if got.UserId != f.bob.Id {
		t.Errorf("row owner = %d, want the original owner %d", got.UserId, f.bob.Id)
	}
	if got.Name != "renamed" {
		t.Errorf("row name = %q, want the edited name", got.Name)
	}
}

// Delete reaches every row for an admin — the lists put the delete button on all of
// them — while a regular user may only delete their own and gets "not found" for the
// rest, so existence does not leak.
func TestRowDeleteAllowsAdminAnyRow(t *testing.T) {
	f := newTokenOwnershipFixture(t)
	bobToken := f.saveToken(t, f.bob, "bob-token", "bob-secret")
	aliceToken := f.saveToken(t, f.alice, "alice-token", "alice-secret")

	// A regular user cannot delete a foreign row and it stays intact.
	out := f.doPath(t, http.MethodDelete, fmt.Sprintf("/api/token/%d", bobToken.Id), "", f.alice)
	if !strings.Contains(out, "token not found") {
		t.Errorf("regular user deleting a foreign token: got %q, want \"token not found\"", out)
	}
	if got, _ := f.tokenModel.FindByPK(bobToken.Id); got == nil {
		t.Fatal("a regular user deleted someone else's token")
	}

	// An admin deletes that same row.
	out = f.doPath(t, http.MethodDelete, fmt.Sprintf("/api/token/%d", bobToken.Id), "", f.admin)
	if !strings.Contains(out, "ok") {
		t.Errorf("admin deleting a foreign token: got %q, want ok", out)
	}
	if got, _ := f.tokenModel.FindByPK(bobToken.Id); got != nil {
		t.Error("the admin's delete left the row behind")
	}

	// The owner's own delete still works.
	out = f.doPath(t, http.MethodDelete, fmt.Sprintf("/api/token/%d", aliceToken.Id), "", f.alice)
	if !strings.Contains(out, "ok") {
		t.Errorf("owner deleting their own token: got %q, want ok", out)
	}
	if got, _ := f.tokenModel.FindByPK(aliceToken.Id); got != nil {
		t.Error("the owner's delete left the row behind")
	}
}

// The same delete rule holds for the other row resources: the button is on every list
// row, so admin delete must work on all four.
func TestRowDeleteAdminCoversAllResources(t *testing.T) {
	f := newTokenOwnershipFixture(t)

	bobSMTP := &model.SMTP{UserId: f.bob.Id, Name: "bob-smtp-2", Host: "127.0.0.1", Port: 1, Mail: "b2@example.com"}
	if err := f.smtpModel.Save(bobSMTP); err != nil {
		t.Fatalf("save smtp: %v", err)
	}
	bobMail := &model.Mail{UserId: f.bob.Id, Name: "bob-contact-2", Mail: "bc2@example.com"}
	if err := f.mailModel.Save(bobMail); err != nil {
		t.Fatalf("save mail: %v", err)
	}
	bobSched := &model.Schedule{UserId: f.bob.Id, Name: "bob-sched", TokenId: 1}
	if err := f.schedModel.Save(bobSched); err != nil {
		t.Fatalf("save schedule: %v", err)
	}

	for _, path := range []string{
		fmt.Sprintf("/api/smtp/%d", bobSMTP.Id),
		fmt.Sprintf("/api/mail/%d", bobMail.Id),
		fmt.Sprintf("/api/schedule/%d", bobSched.Id),
	} {
		if out := f.doPath(t, http.MethodDelete, path, "", f.admin); !strings.Contains(out, "ok") {
			t.Errorf("admin DELETE %s: got %q, want ok", path, out)
		}
	}
	if got, _ := f.smtpModel.FindByPK(bobSMTP.Id); got != nil {
		t.Error("the smtp row survived the admin delete")
	}
	if got, _ := f.mailModel.FindByPK(bobMail.Id); got != nil {
		t.Error("the mail row survived the admin delete")
	}
	if got, _ := f.schedModel.FindByPK(bobSched.Id); got != nil {
		t.Error("the schedule row survived the admin delete")
	}
}
