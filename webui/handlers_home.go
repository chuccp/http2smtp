package webui

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"go.uber.org/zap"
)

// registerPageRoutes wires the rest of the server-rendered pages. It grows with each
// migration step; home/login/logout/lang are wired directly in Init.
func (p *Pages) registerPageRoutes(ctx *core.Context) {
	ctx.Get("/setup", p.setupGet)
	ctx.Post("/setup/db", p.setupDb)
	ctx.Post("/setup/test", p.setupTest)
	ctx.Post("/setup/admin", p.setupAdmin)
	ctx.Post("/setup/skip", p.setupSkip)
}

// home routes a first visit: the setup wizard before init, the login page without a
// session, the dashboard otherwise.
func (p *Pages) home(req *web.Request) (any, error) {
	st := model.CoreState(p.ctx.GetConfig())
	if !st.Init {
		return p.redirect(req, "/setup")
	}
	if p.currentUser(req) == nil {
		return p.redirect(req, "/login")
	}
	return p.redirect(req, "/dashboard")
}

func (p *Pages) loginData(req *web.Request, next string, flash Flash) *LoginData {
	lang := langOf(req)
	b := p.newBase(lang, "auth.login")
	b.Flash = flash
	return &LoginData{Base: b, Next: next}
}

// LoginData is the login page's template data.
type LoginData struct {
	Base
	Next string
}

func (p *Pages) loginGet(req *web.Request) (any, error) {
	if p.currentUser(req) != nil {
		return p.redirect(req, "/dashboard")
	}
	next := safeNext(req.Query("next"))
	return p.render(req, "login", p.loginData(req, next, Flash{}))
}

func (p *Pages) loginPost(req *web.Request) (any, error) {
	username := req.GetFormParam("username")
	password := req.GetFormParam("password")
	next := safeNext(req.GetFormParam("next"))

	dbUser, err := p.userService.ValidateLogin(username, password)
	if err != nil {
		data := p.loginData(req, next, Flash{Err: err.Error()})
		req.Response().WriteStatus(http.StatusUnauthorized)
		return p.render(req, "login", data)
	}
	lu := &entity.LoginUser{
		Id:       dbUser.Id,
		Username: dbUser.Name,
		Salt:     dbUser.Salt,
		IsAdmin:  dbUser.IsAdmin,
	}
	if _, err := p.authFilter.SignIn(lu, req); err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	return p.redirect(req, next)
}

// logout ends this session everywhere: the salt rotation invalidates every token
// already issued to the user, so a copied session stops working immediately.
func (p *Pages) logout(req *web.Request) (any, error) {
	if user := p.currentUser(req); user != nil {
		if err := p.userService.RotateSalt(user.Id); err != nil {
			// Signing out must not be blocked by the rotation failing
			log.Error("rotate session salt failed", zap.Uint("userId", user.Id), zap.Error(err))
		}
	}
	if _, err := p.authFilter.SignOut(req); err != nil {
		return nil, err
	}
	return p.redirect(req, "/login")
}

// lang switches the interface language and returns to where the visitor came from.
// The return target is restricted to same-site paths.
func (p *Pages) lang(req *web.Request) (any, error) {
	if req.IsPost() {
		lang := req.GetFormParam("lang")
		if isSupportedLang(lang) {
			setLangCookie(req, lang)
		}
	}
	back := req.GetFormParam("back")
	if back == "" {
		// The browser sends a full URL as Referer; keep only its site path.
		if ref := req.GetHeader("Referer"); ref != "" {
			if u, err := url.Parse(ref); err == nil {
				back = u.RequestURI()
			}
		}
	}
	// Only a bare same-site path may come back — never //host or an absolute URL.
	if !strings.HasPrefix(back, "/") || strings.HasPrefix(back, "//") {
		back = "/"
	}
	return p.redirect(req, back)
}
