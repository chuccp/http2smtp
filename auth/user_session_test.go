package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
)

// sessionFixture serves a tiny endpoint that resolves the caller through auth.User, so
// the cookie and header lookup run against the real request plumbing.
type sessionFixture struct {
	ts        *httptest.Server
	userModel *model.UserModel
}

func newSessionFixture(t *testing.T) *sessionFixture {
	t.Helper()
	conn, err := db.ConnectionSQLite(filepath.Join(t.TempDir(), "session.db"))
	if err != nil {
		t.Fatalf("opening sqlite: %v", err)
	}
	sqlDB, err := conn.GetGorm().DB()
	if err != nil {
		t.Fatalf("getting sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	userModel := &model.UserModel{}
	if err := userModel.Init(conn, nil); err != nil {
		t.Fatalf("init user model: %v", err)
	}
	if err := userModel.CreateTable(); err != nil {
		t.Fatalf("create t_user: %v", err)
	}
	coreCtx := core.NewContext(config.NewConfig(), context.Background())
	coreCtx.AddModel(userModel)

	servers := web.NewServers()
	server, err := servers.CreateServer(&web.ServerConfig{Port: 0})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	server.Get("/who", func(req *web.Request) (any, error) {
		u, err := User(req, coreCtx)
		if err != nil {
			return nil, err
		}
		return u.Name, nil
	})
	ts := httptest.NewServer(server.GetHandler())
	t.Cleanup(ts.Close)
	return &sessionFixture{ts: ts, userModel: userModel}
}

// who sends one request with the given cookie value ("" for none) and optional
// Authorization header, and reports whether the session resolved.
func (f *sessionFixture) who(t *testing.T, cookie, authorization string) bool {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, f.ts.URL+"/who", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: entity.UserToken, Value: cookie})
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (f *sessionFixture) saveUser(t *testing.T, name, salt string) *model.User {
	t.Helper()
	u := &model.User{Name: name, Password: "x", Salt: salt, IsUse: true}
	if err := f.userModel.Save(u); err != nil {
		t.Fatalf("save user: %v", err)
	}
	return u
}

// The session token embeds the salt it was minted with. Rotating the stored salt — which
// sign-out, password change and user disable all do — must reject the old cookie at
// once, on every device.
func TestSessionInvalidatedBySaltRotation(t *testing.T) {
	f := newSessionFixture(t)
	u := f.saveUser(t, "alice", "salt-1")

	token, err := Encrypt(&entity.LoginUser{Id: u.Id, Salt: u.Salt})
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if !f.who(t, token, "") {
		t.Fatal("a freshly minted session was rejected")
	}

	u.Salt = "salt-2"
	if err := f.userModel.Save(u); err != nil {
		t.Fatalf("rotate salt: %v", err)
	}
	if f.who(t, token, "") {
		t.Fatal("a session minted before the salt rotation is still accepted")
	}
}

// The token also travels in an Authorization header for external callers; the browser
// never uses that path.
func TestSessionAuthorizationHeaderAccepted(t *testing.T) {
	f := newSessionFixture(t)
	u := f.saveUser(t, "alice", "salt-1")

	token, err := Encrypt(&entity.LoginUser{Id: u.Id, Salt: u.Salt})
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if !f.who(t, "", token) {
		t.Fatal("the Authorization header path rejected a valid session token")
	}
}

// A token may identify the user by name instead of id; both lookups must agree to the
// stored salt.
func TestSessionUsernameLookup(t *testing.T) {
	f := newSessionFixture(t)
	f.saveUser(t, "alice", "salt-1")

	token, err := Encrypt(&entity.LoginUser{Username: "alice", Salt: "salt-1"})
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if !f.who(t, token, "") {
		t.Fatal("a valid username-based session was rejected")
	}

	if token, err = Encrypt(&entity.LoginUser{Username: "alice", Salt: "wrong"}); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if f.who(t, token, "") {
		t.Fatal("a session with a stale salt was accepted")
	}
}

// A token for a user that no longer exists, or one that cannot be decrypted at all, is
// not a login.
func TestSessionRejectsUnknownAndGarbage(t *testing.T) {
	f := newSessionFixture(t)

	token, err := Encrypt(&entity.LoginUser{Id: 999, Salt: "salt-1"})
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if f.who(t, token, "") {
		t.Fatal("a token for a nonexistent user was accepted")
	}
	if f.who(t, "not-a-valid-token", "") {
		t.Fatal("a garbage cookie was accepted")
	}
	if f.who(t, "", "") {
		t.Fatal("a request without any credentials was accepted")
	}
}

// The session is delivered as an HttpOnly cookie, so script cannot read it.
func TestSignInSetsHttpOnlyCookie(t *testing.T) {
	a := &Authentication{}
	if err := a.Init(core.NewContext(config.NewConfig(), context.Background())); err != nil {
		t.Fatalf("init: %v", err)
	}

	server, err := web.NewServers().CreateServer(&web.ServerConfig{Port: 0})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	server.Post("/in", func(req *web.Request) (any, error) {
		return a.SignIn(&entity.LoginUser{Id: 1, Salt: "s"}, req)
	})
	srv := httptest.NewServer(server.GetHandler())
	t.Cleanup(srv.Close)
	resp, err := http.Post(srv.URL+"/in", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()

	var session *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == entity.UserToken {
			session = c
		}
	}
	if session == nil {
		t.Fatal("sign-in did not set the session cookie")
	}
	if !session.HttpOnly {
		t.Error("the session cookie is not HttpOnly; page script could steal it")
	}
}
