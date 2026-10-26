package webui

import (
	"net/url"

	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/model"
)

// currentUser returns the signed-in user behind the session cookie, or nil.
func (p *Pages) currentUser(req *web.Request) *model.User {
	user, err := auth.User(req, p.ctx)
	if err != nil || user == nil {
		return nil
	}
	return user
}

// requireUser gates a page on being signed in, redirecting to the login page (with a
// return target) otherwise. Page routes deliberately do not carry the framework's
// WithLogin meta: its 403 JSON is the wrong answer for a browser navigation.
func (p *Pages) requireUser(req *web.Request) (*model.User, bool) {
	if user := p.currentUser(req); user != nil {
		return user, true
	}
	next := req.URL().RequestURI()
	login := "/login"
	if next != "" && next != "/" {
		login += "?next=" + url.QueryEscape(next)
	}
	_, _ = p.redirect(req, login)
	return nil, false
}

// requireAdmin gates a page on the admin flag; a signed-in non-admin gets the
// not-admin page rather than a silent bounce, so the refusal is visible.
func (p *Pages) requireAdmin(req *web.Request) (*model.User, bool) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, false
	}
	if !user.IsAdmin {
		lang := langOf(req)
		data := struct {
			Base
			Message string
		}{
			Base:    p.newBase(lang, "auth.noPermission"),
			Message: translatorFor(lang).T("auth.noPermission"),
		}
		_, _ = p.render(req, "notadmin", data)
		return nil, false
	}
	return user, true
}

// setLangCookie records the interface language. It is deliberately not HttpOnly —
// there is nothing secret in it — and lasts a year, matching the old localStorage
// behaviour across browser restarts.
func setLangCookie(req *web.Request, lang string) {
	req.Cookie().Set("lang", lang, web.WithHttpOnly(false))
}

// langNames are the switcher's labels: each language in its own name, as is customary.
var langNames = map[string]string{
	"zh-cn": "简体中文",
	"zh-tw": "繁體中文",
	"en":    "English",
	"ja":    "日本語",
}
