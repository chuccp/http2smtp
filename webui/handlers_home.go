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

// registerPageRoutes wires the server-rendered pages. Home, logout and language
// switching redirect without rendering and are registered directly in Init.
func (p *Pages) registerPageRoutes(ctx *core.Context) {
	ctx.GetTemplate("/setup", p.setupGet)
	ctx.PostTemplate("/setup/db", p.setupDb)
	ctx.PostTemplate("/setup/test", p.setupTest)
	ctx.PostTemplate("/setup/admin", p.setupAdmin)
	ctx.PostTemplate("/setup/skip", p.setupSkip)

	ctx.GetTemplate("/dashboard", p.dashboard)

	ctx.GetTemplate("/smtp", p.smtpList)
	ctx.GetTemplate("/smtp/form", p.smtpForm)
	ctx.GetTemplate("/smtp/form/:id", p.smtpForm)
	ctx.PostTemplate("/smtp/save", p.smtpSave)
	ctx.GetTemplate("/smtp/send/:id", p.smtpSendForm)
	ctx.Post("/smtp/test", p.smtpTest)

	ctx.GetTemplate("/mail", p.mailList)
	ctx.GetTemplate("/mail/form", p.mailForm)
	ctx.GetTemplate("/mail/form/:id", p.mailForm)
	ctx.PostTemplate("/mail/save", p.mailSave)

	ctx.GetTemplate("/tokens", p.tokenList)
	ctx.GetTemplate("/tokens/form", p.tokenForm)
	ctx.GetTemplate("/tokens/form/:id", p.tokenForm)
	ctx.PostTemplate("/tokens/save", p.tokenSave)
	ctx.GetTemplate("/tokens/send/:id", p.tokenSendForm)

	ctx.GetTemplate("/schedule", p.scheduleList)
	ctx.GetTemplate("/schedule/form", p.scheduleForm)
	ctx.GetTemplate("/schedule/form/:id", p.scheduleForm)
	ctx.PostTemplate("/schedule/save", p.scheduleSave)

	ctx.GetTemplate("/logs", p.logList)
	ctx.GetTemplate("/logs/:id", p.logDetail)

	ctx.GetTemplate("/users", p.userList)
	ctx.GetTemplate("/users/form", p.userForm)
	ctx.GetTemplate("/users/form/:id", p.userForm)
	ctx.PostTemplate("/users/save", p.userSave)

	ctx.GetTemplate("/settings", p.settingsGet)
	ctx.PostTemplate("/settings/save", p.settingsSave)
	ctx.PostTemplate("/settings/restart", p.settingsRestart)

	ctx.GetTemplate("/password", p.passwordGet)
	ctx.PostTemplate("/password", p.passwordPost)
}

// home routes a first visit: the setup wizard before init, the login page without a
// session, the dashboard otherwise.
func (p *Pages) home(view *web.Template) error {
	req := view.Request()
	st := model.CoreState(p.ctx.GetConfig())
	if !st.Init {
		return redirect(view, "/setup")
	}
	if p.currentUser(req) == nil {
		return redirect(view, "/login")
	}
	return redirect(view, "/dashboard")
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

func (p *Pages) loginGet(view *web.Template) error {
	req := view.Request()
	if p.currentUser(req) != nil {
		return redirect(view, "/dashboard")
	}
	next := safeNext(req.Query("next"))
	view.View("login").Layout(layoutFor("login")).Data(p.loginData(req, next, Flash{}))
	return nil
}

func (p *Pages) loginPost(view *web.Template) error {
	req := view.Request()
	username := req.GetFormParam("username")
	password := req.GetFormParam("password")
	next := safeNext(req.GetFormParam("next"))

	dbUser, err := p.userService.ValidateLogin(username, password)
	if err != nil {
		data := p.loginData(req, next, Flash{Err: err.Error()})
		return renderPageStatus(view, "login", data, http.StatusUnauthorized)
	}
	lu := &entity.LoginUser{
		Id:       dbUser.Id,
		Username: dbUser.Name,
		Salt:     dbUser.Salt,
		IsAdmin:  dbUser.IsAdmin,
	}
	if _, err := p.authFilter.SignIn(lu, req); err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	return redirect(view, next)
}

// logout ends this session everywhere: the salt rotation invalidates every token
// already issued to the user, so a copied session stops working immediately.
func (p *Pages) logout(view *web.Template) error {
	req := view.Request()
	if user := p.currentUser(req); user != nil {
		if err := p.userService.RotateSalt(user.Id); err != nil {
			// Signing out must not be blocked by the rotation failing
			log.Error("rotate session salt failed", zap.Uint("userId", user.Id), zap.Error(err))
		}
	}
	if _, err := p.authFilter.SignOut(req); err != nil {
		return err
	}
	return redirect(view, "/login")
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
	req.Response().Redirect(http.StatusFound, back)
	return nil, nil
}
