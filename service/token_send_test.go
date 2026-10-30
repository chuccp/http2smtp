package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
)

// tokenSendFixture wires a TokenService against a throwaway sqlite file and a throwaway
// cache directory. The SMTP rows point at 127.0.0.1:1, so every send fails fast at the
// dial and what is under test is the behaviour around the send: recipient merging and
// filtering, the owner checks, the log entry and the attachment cleanup.
type tokenSendFixture struct {
	ts       *TokenService
	users    *model.UserModel
	smtps    *model.SMTPModel
	mails    *model.MailModel
	tokens   *model.TokenModel
	logs     *model.LogModel
	attachs  *model.AttachmentModel
	cacheDir string
}

func newTokenSendFixture(t *testing.T) *tokenSendFixture {
	t.Helper()
	conn, err := db.ConnectionSQLite(filepath.Join(t.TempDir(), "send.db"))
	if err != nil {
		t.Fatalf("opening sqlite: %v", err)
	}
	sqlDB, err := conn.GetGorm().DB()
	if err != nil {
		t.Fatalf("getting sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	initModel := func(m interface {
		Init(*db.DB, *core.Context) error
		CreateTable() error
	}) {
		t.Helper()
		if err := m.Init(conn, nil); err != nil {
			t.Fatalf("init model: %v", err)
		}
		if err := m.CreateTable(); err != nil {
			t.Fatalf("create table: %v", err)
		}
	}
	users := &model.UserModel{}
	smtps := &model.SMTPModel{}
	mails := &model.MailModel{}
	tokens := &model.TokenModel{}
	logs := &model.LogModel{}
	attachs := &model.AttachmentModel{}
	initModel(users)
	initModel(smtps)
	initModel(mails)
	initModel(tokens)
	initModel(logs)
	initModel(attachs)

	coreCtx := core.NewContext(config.NewConfig(), context.Background())
	coreCtx.AddModel(users, smtps, mails, tokens, logs, attachs)

	cacheDir := t.TempDir()
	ts := &TokenService{
		context:     coreCtx,
		mailModel:   mails,
		smtpModel:   smtps,
		tokenModel:  tokens,
		logService:  &LogService{logModel: logs, attachmentModel: attachs},
		lock:        new(sync.RWMutex),
		cachePath:   cacheDir,
		maxBodySize: model.DefaultMaxBodySize,
	}
	return &tokenSendFixture{ts: ts, users: users, smtps: smtps, mails: mails, tokens: tokens, logs: logs, attachs: attachs, cacheDir: cacheDir}
}

func (f *tokenSendFixture) user(t *testing.T, name string, isUse bool) *model.User {
	t.Helper()
	u := &model.User{Name: name, Password: "x", Salt: "salt-" + name, IsUse: isUse}
	if err := f.users.Save(u); err != nil {
		t.Fatalf("save user %s: %v", name, err)
	}
	return u
}

func (f *tokenSendFixture) smtpFor(t *testing.T, u *model.User) *model.SMTP {
	t.Helper()
	// Port 1 on loopback: the dial is refused instantly, so the send fails fast and
	// deterministically without a network.
	s := &model.SMTP{UserId: u.Id, Name: "smtp-" + u.Name, Host: "127.0.0.1", Port: 1,
		Mail: "sender@example.com", Username: "sender", Password: "pw"}
	if err := f.smtps.Save(s); err != nil {
		t.Fatalf("save smtp: %v", err)
	}
	return s
}

func (f *tokenSendFixture) mailFor(t *testing.T, u *model.User, name, addr string) *model.Mail {
	t.Helper()
	m := &model.Mail{UserId: u.Id, Name: name, Mail: addr}
	if err := f.mails.Save(m); err != nil {
		t.Fatalf("save mail: %v", err)
	}
	return m
}

func (f *tokenSendFixture) tokenFor(t *testing.T, userId uint, smtpId uint, recvIds, subject, token string, state uint8) *model.Token {
	t.Helper()
	tk := &model.Token{UserId: userId, SMTPId: smtpId, ReceiveEmailIds: recvIds,
		Subject: subject, Token: token, Name: "tok-" + token, State: state}
	if err := f.tokens.Save(tk); err != nil {
		t.Fatalf("save token: %v", err)
	}
	return tk
}

func (f *tokenSendFixture) logRows(t *testing.T) []*model.Log {
	t.Helper()
	rows, err := f.logs.Query().Order("id asc").All()
	if err != nil {
		t.Fatalf("read logs: %v", err)
	}
	return rows
}

// cacheFileCount counts the files left in the cache directory tree.
func (f *tokenSendFixture) cacheFileCount(t *testing.T) int {
	t.Helper()
	n := 0
	err := filepath.WalkDir(f.cacheDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			n++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk cache: %v", err)
	}
	return n
}

// publicAPI serves the public /sendMail entry point over real HTTP: no session, the
// token value is the only credential.
func (f *tokenSendFixture) publicAPI(t *testing.T) *httptest.Server {
	t.Helper()
	servers := web.NewServers()
	server, err := servers.CreateServer(&web.ServerConfig{Port: 0})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	server.Post("/sendMail", func(req *web.Request) (any, error) {
		return f.ts.SendMailByToken(req, nil)
	})
	srv := httptest.NewServer(server.GetHandler())
	t.Cleanup(srv.Close)
	return srv
}

// idList formats ids for the comma-separated ReceiveEmailIds column.
func idList(ids ...uint) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.Itoa(int(id))
	}
	return strings.Join(parts, ",")
}

// The configured list is the default and the request adds to it; a malformed address is
// dropped rather than failing the send, a duplicate (case-insensitively) is sent once,
// and the log records exactly the addresses that were attempted — under the token
// owner. The send itself fails at the dial, and the log entry is written anyway.
func TestSendMailByTokenIdMergesFiltersAndLogs(t *testing.T) {
	f := newTokenSendFixture(t)
	alice := f.user(t, "alice", true)
	s := f.smtpFor(t, alice)
	good := f.mailFor(t, alice, "A", "a@example.com")
	bad := f.mailFor(t, alice, "bad", "not-an-email")
	tk := f.tokenFor(t, alice.Id, s.Id, idList(good.Id, bad.Id), "tok-subject", "tok-1", entity.TokenStateInUse)

	res, err := f.ts.SendMailByTokenId(alice, &entity.SendMailByTokenId{
		TokenId:    tk.Id,
		Recipients: []string{"b@example.com", "A@example.com", "bogus"},
		Subject:    "",
		Content:    "hello",
	})
	if err == nil {
		t.Fatal("expected the dial to fail against 127.0.0.1:1")
	}
	if res != "error" {
		t.Errorf("result = %v, want \"error\"", res)
	}

	rows := f.logRows(t)
	if len(rows) != 1 {
		t.Fatalf("log rows = %d, want 1", len(rows))
	}
	row := rows[0]
	if row.UserId != alice.Id {
		t.Errorf("log owner = %d, want the token owner %d", row.UserId, alice.Id)
	}
	// a@example.com keeps its configured name, the malformed rows are gone, the
	// requested address is added bare, and A@example.com was merged into a@example.com.
	if want := `"A" <a@example.com>,b@example.com`; row.Mail != want {
		t.Errorf("log mail = %q, want %q", row.Mail, want)
	}
	// An empty request subject falls back to the token's subject.
	if row.Subject != "tok-subject" {
		t.Errorf("log subject = %q, want the token subject", row.Subject)
	}
	if row.Status != entity.ERROR || row.Result == "success" {
		t.Errorf("log status/result = %d/%q, want ERROR with the dial error", row.Status, row.Result)
	}
}

// A non-admin may only use their own tokens and gets the same "not found" answer a
// nonexistent id gets, so existence does not leak. An admin may use any token, and the
// send is logged under the token's owner — not the admin who ran it.
func TestSendMailByTokenIdOwnership(t *testing.T) {
	f := newTokenSendFixture(t)
	alice := f.user(t, "alice", true)
	bob := f.user(t, "bob", true)
	admin := f.user(t, "admin", true)
	admin.IsAdmin = true
	if err := f.users.Save(admin); err != nil {
		t.Fatalf("promote admin: %v", err)
	}
	s := f.smtpFor(t, alice)
	m := f.mailFor(t, alice, "A", "a@example.com")
	tk := f.tokenFor(t, alice.Id, s.Id, idList(m.Id), "sub", "tok-2", entity.TokenStateInUse)

	_, err := f.ts.SendMailByTokenId(bob, &entity.SendMailByTokenId{TokenId: tk.Id, Content: "x"})
	if err == nil || err.Error() != "token not found" {
		t.Fatalf("foreign token as regular user: err = %v, want \"token not found\"", err)
	}
	if rows := f.logRows(t); len(rows) != 0 {
		t.Fatalf("a refused send wrote %d log rows", len(rows))
	}

	_, err = f.ts.SendMailByTokenId(admin, &entity.SendMailByTokenId{TokenId: tk.Id, Content: "x"})
	if err == nil {
		t.Fatal("expected the dial to fail for the admin send too")
	}
	rows := f.logRows(t)
	if len(rows) != 1 {
		t.Fatalf("log rows after admin send = %d, want 1", len(rows))
	}
	if rows[0].UserId != alice.Id {
		t.Errorf("log owner = %d, want the token owner %d (the admin ran the send)", rows[0].UserId, alice.Id)
	}
}

// The public API has no session, so the token service re-reads the owner on every send:
// a disabled account stops every token it holds, and the attempt is not even logged —
// no send happened. Any cached attachment is removed again.
func TestSendMailByTokenIdOwnerDisabled(t *testing.T) {
	f := newTokenSendFixture(t)
	alice := f.user(t, "alice", false)
	s := f.smtpFor(t, alice)
	m := f.mailFor(t, alice, "A", "a@example.com")
	tk := f.tokenFor(t, alice.Id, s.Id, idList(m.Id), "sub", "tok-3", entity.TokenStateInUse)

	_, err := f.ts.SendMailByTokenId(alice, &entity.SendMailByTokenId{
		TokenId: tk.Id,
		Content: "x",
		Files:   []entity.File{{Name: "a.txt", Data: base64.StdEncoding.EncodeToString([]byte("hi"))}},
	})
	if err == nil || err.Error() != "token owner is disabled" {
		t.Fatalf("err = %v, want \"token owner is disabled\"", err)
	}
	if rows := f.logRows(t); len(rows) != 0 {
		t.Fatalf("a blocked send wrote %d log rows", len(rows))
	}
	if n := f.cacheFileCount(t); n != 0 {
		t.Fatalf("cache holds %d files after a blocked send, want 0", n)
	}
}

// Rows with user_id = 0 predate token ownership: there is no account to consult, so the
// public API — where the token value is the only credential — still sends with them.
// The management send-by-id path does not: the row belongs to nobody and is invisible
// to every list, so a regular caller gets "token not found" just like for a missing id.
func TestLegacyOwnerZeroTokenSendsOnPublicAPI(t *testing.T) {
	f := newTokenSendFixture(t)
	s := f.smtpFor(t, f.user(t, "keeper", true))
	m := f.mailFor(t, f.user(t, "keeper2", true), "K", "k@example.com")
	tk := f.tokenFor(t, 0, s.Id, idList(m.Id), "sub", "tok-4", entity.TokenStateInUse)

	srv := f.publicAPI(t)
	resp, err := http.Post(srv.URL+"/sendMail", "application/json",
		strings.NewReader(`{"token":"tok-4","content":"x"}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	payload := make([]byte, 4096)
	n, _ := resp.Body.Read(payload)
	resp.Body.Close()
	if strings.Contains(string(payload[:n]), "token owner is disabled") {
		t.Fatal("a user_id = 0 token was blocked by the owner check")
	}
	rows := f.logRows(t)
	if len(rows) != 1 {
		t.Fatalf("log rows = %d, want 1 (the send was attempted and logged)", len(rows))
	}
	if rows[0].UserId != 0 {
		t.Errorf("log owner = %d, want 0", rows[0].UserId)
	}

	// Management path: nobody owns the row, so it is not the caller's own token.
	_, err = f.ts.SendMailByTokenId(f.user(t, "caller", true), &entity.SendMailByTokenId{TokenId: tk.Id, Content: "x"})
	if err == nil || err.Error() != "token not found" {
		t.Fatalf("management send with a legacy token: err = %v, want \"token not found\"", err)
	}
}

// Nothing to send is reported before anything is attempted, and in particular before a
// log entry is written: the log records sends, not rejected requests.
func TestSendMailByTokenIdNoValidRecipients(t *testing.T) {
	f := newTokenSendFixture(t)
	alice := f.user(t, "alice", true)
	s := f.smtpFor(t, alice)
	tk := f.tokenFor(t, alice.Id, s.Id, "", "sub", "tok-5", entity.TokenStateInUse)

	_, err := f.ts.SendMailByTokenId(alice, &entity.SendMailByTokenId{TokenId: tk.Id, Content: "x"})
	if err == nil || !strings.Contains(err.Error(), "no recipients") {
		t.Fatalf("err = %v, want the \"no recipients\" refusal", err)
	}

	bad := f.mailFor(t, alice, "bad", "not-an-email")
	tk2 := f.tokenFor(t, alice.Id, s.Id, idList(bad.Id), "sub", "tok-6", entity.TokenStateInUse)
	_, err = f.ts.SendMailByTokenId(alice, &entity.SendMailByTokenId{TokenId: tk2.Id, Recipients: []string{"also bogus"}, Content: "x"})
	if err == nil || err.Error() != "no valid recipients" {
		t.Fatalf("err = %v, want \"no valid recipients\"", err)
	}
	if rows := f.logRows(t); len(rows) != 0 {
		t.Fatalf("rejected sends wrote %d log rows", len(rows))
	}
}

// A disabled token does not send, whoever asks.
func TestSendMailByTokenIdDisabledToken(t *testing.T) {
	f := newTokenSendFixture(t)
	alice := f.user(t, "alice", true)
	s := f.smtpFor(t, alice)
	m := f.mailFor(t, alice, "A", "a@example.com")

	for _, state := range []uint8{entity.TokenStateUserDisabled, entity.TokenStateAdminDisabled} {
		tk := f.tokenFor(t, alice.Id, s.Id, idList(m.Id), "sub", fmt.Sprintf("tok-off-%d", state), state)
		_, err := f.ts.SendMailByTokenId(alice, &entity.SendMailByTokenId{TokenId: tk.Id, Content: "x"})
		if err == nil || err.Error() != "token is not use" {
			t.Errorf("state %d: err = %v, want \"token is not use\"", state, err)
		}
	}
	if rows := f.logRows(t); len(rows) != 0 {
		t.Fatalf("disabled-token sends wrote %d log rows", len(rows))
	}
}

// A send that reached the log keeps its cached attachment: from the moment the log
// records the paths, its cleanup owns them. The row also links an attachment entry and
// the legacy files JSON.
func TestSendMailByTokenIdRecordsAttachments(t *testing.T) {
	f := newTokenSendFixture(t)
	alice := f.user(t, "alice", true)
	s := f.smtpFor(t, alice)
	m := f.mailFor(t, alice, "A", "a@example.com")
	tk := f.tokenFor(t, alice.Id, s.Id, idList(m.Id), "sub", "tok-7", entity.TokenStateInUse)

	_, err := f.ts.SendMailByTokenId(alice, &entity.SendMailByTokenId{
		TokenId: tk.Id,
		Content: "x",
		Files:   []entity.File{{Name: "report.txt", Data: base64.StdEncoding.EncodeToString([]byte("payload"))}},
	})
	if err == nil {
		t.Fatal("expected the dial to fail")
	}
	rows := f.logRows(t)
	if len(rows) != 1 {
		t.Fatalf("log rows = %d, want 1", len(rows))
	}
	atts, err := f.attachs.FindByLogId(rows[0].Id)
	if err != nil {
		t.Fatalf("read attachments: %v", err)
	}
	if len(atts) != 1 {
		t.Fatalf("attachment rows = %d, want 1", len(atts))
	}
	if atts[0].Name != "report.txt" {
		t.Errorf("attachment name = %q, want report.txt", atts[0].Name)
	}
	data, err := os.ReadFile(atts[0].FilePath)
	if err != nil || string(data) != "payload" {
		t.Errorf("cached copy = %q (%v), want the uploaded payload", data, err)
	}
	if rows[0].Files == "" {
		t.Error("the log's files JSON is empty; legacy downloads would break")
	}
}

// A scheduled call is logged under whoever owns the schedule — not the owner of the
// token it fires through — because the schedule owner is who sees it in their log list.
func TestSendApiCallMailLogsUnderScheduleOwner(t *testing.T) {
	f := newTokenSendFixture(t)
	alice := f.user(t, "alice", true)
	bob := f.user(t, "bob", true)
	s := f.smtpFor(t, alice)
	m := f.mailFor(t, alice, "A", "a@example.com")
	tk := f.tokenFor(t, alice.Id, s.Id, idList(m.Id), "sub", "tok-8", entity.TokenStateInUse)

	schedule := &model.Schedule{
		UserId:  bob.Id,
		TokenId: tk.Id,
		Name:    "nightly",
		Url:     "http://127.0.0.1:1/hook",
		Method:  "GET",
	}
	if err := f.ts.SendApiCallMail(schedule); err == nil {
		t.Fatal("expected the API call and the alert mail to fail")
	}
	rows := f.logRows(t)
	if len(rows) != 1 {
		t.Fatalf("log rows = %d, want 1", len(rows))
	}
	if rows[0].UserId != bob.Id {
		t.Errorf("log owner = %d, want the schedule owner %d (token owner is %d)", rows[0].UserId, bob.Id, alice.Id)
	}
	if rows[0].Subject != "nightly" {
		t.Errorf("log subject = %q, want the schedule name", rows[0].Subject)
	}
	if rows[0].Status != entity.ERROR {
		t.Errorf("log status = %d, want ERROR: a failed schedule must not be logged as success", rows[0].Status)
	}
}

// Same owner rule for the scheduled path: a disabled account's tokens stop firing.
func TestSendApiCallMailOwnerDisabled(t *testing.T) {
	f := newTokenSendFixture(t)
	alice := f.user(t, "alice", false)
	s := f.smtpFor(t, alice)
	m := f.mailFor(t, alice, "A", "a@example.com")
	tk := f.tokenFor(t, alice.Id, s.Id, idList(m.Id), "sub", "tok-9", entity.TokenStateInUse)

	err := f.ts.SendApiCallMail(&model.Schedule{UserId: alice.Id, TokenId: tk.Id, Name: "s", Url: "http://127.0.0.1:1/hook", Method: "GET"})
	if err == nil || err.Error() != "token owner is disabled" {
		t.Fatalf("err = %v, want \"token owner is disabled\"", err)
	}
	if rows := f.logRows(t); len(rows) != 0 {
		t.Fatalf("a blocked schedule wrote %d log rows", len(rows))
	}
}

// The multipart form path carries attachments as file parts instead of base64 fields;
// the same rules apply: they are cached, recorded on the log and owned by its cleanup.
func TestSendMailByTokenMultipartUpload(t *testing.T) {
	f := newTokenSendFixture(t)
	alice := f.user(t, "alice", true)
	s := f.smtpFor(t, alice)
	m := f.mailFor(t, alice, "A", "a@example.com")
	tk := f.tokenFor(t, alice.Id, s.Id, idList(m.Id), "sub", "tok-mp", entity.TokenStateInUse)

	srv := f.publicAPI(t)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("token", tk.Token)
	_ = writer.WriteField("content", "hello")
	_ = writer.WriteField("recipients", "extra@example.com")
	part, err := writer.CreateFormFile("files", "upload.txt")
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := part.Write([]byte("multipart payload")); err != nil {
		t.Fatalf("write file part: %v", err)
	}
	_ = writer.Close()

	resp, err := http.Post(srv.URL+"/sendMail", writer.FormDataContentType(), &body)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	resp.Body.Close()

	rows := f.logRows(t)
	if len(rows) != 1 {
		t.Fatalf("log rows = %d, want 1", len(rows))
	}
	// The recipient added by the request is in the log alongside the configured one.
	if !strings.Contains(rows[0].Mail, "extra@example.com") || !strings.Contains(rows[0].Mail, "a@example.com") {
		t.Errorf("log mail = %q, want both recipients", rows[0].Mail)
	}
	atts, err := f.attachs.FindByLogId(rows[0].Id)
	if err != nil {
		t.Fatalf("read attachments: %v", err)
	}
	if len(atts) != 1 || atts[0].Name != "upload.txt" {
		t.Fatalf("attachments = %+v, want one upload.txt", atts)
	}
	data, err := os.ReadFile(atts[0].FilePath)
	if err != nil || string(data) != "multipart payload" {
		t.Errorf("cached copy = %q (%v), want the uploaded bytes", data, err)
	}
}

// limitSendBody runs before anything parses the body: a request over the cap is refused
// with a clear message, 0 means the 10 MB default, and a negative value is unlimited.
func TestLimitSendBody(t *testing.T) {
	f := newTokenSendFixture(t)
	srv := f.publicAPI(t)

	post := func(body string) string {
		t.Helper()
		resp, err := http.Post(srv.URL+"/sendMail", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer resp.Body.Close()
		payload := make([]byte, 4096)
		n, _ := resp.Body.Read(payload)
		return string(payload[:n])
	}
	body := `{"token":"nope","content":"` + strings.Repeat("x", 500) + `"}`

	f.ts.maxBodySize = 64
	if got := post(body); !strings.Contains(got, "too large") {
		t.Errorf("over the cap: got %q, want the size refusal", got)
	}

	f.ts.maxBodySize = -1
	got := post(body)
	if strings.Contains(got, "too large") {
		t.Errorf("unlimited (-1): got the size refusal %q", got)
	}
	if !strings.Contains(got, "token not found") {
		t.Errorf("unlimited (-1): got %q, want the request to reach the token lookup", got)
	}

	f.ts.maxBodySize = 0
	got = post(body)
	if strings.Contains(got, "too large") {
		t.Errorf("default (0): got the size refusal %q", got)
	}
	if !strings.Contains(got, "token not found") {
		t.Errorf("default (0): got %q, want the request to reach the token lookup", got)
	}
}
