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
func (p *Pages) registerPageRoutes(ctx *core.Context, templateOption web.TemplateOption) {
	get := func(relativePath string, handler web.TemplateHandler) {
		ctx.GetTemplate(relativePath, handler, templateOption)
	}
	post := func(relativePath string, handler web.TemplateHandler) {
		ctx.PostTemplate(relativePath, handler, templateOption)
	}

	get("/setup", p.setupGet)
	post("/setup/db", p.setupDb)
	post("/setup/test", p.setupTest)
	post("/setup/admin", p.setupAdmin)
	post("/setup/skip", p.setupSkip)

	get("/dashboard", p.dashboard)

	get("/smtp", p.smtpList)
	get("/smtp/form", p.smtpForm)
	get("/smtp/form/:id", p.smtpForm)
	post("/smtp/save", p.smtpSave)
	get("/smtp/send/:id", p.smtpSendForm)
	ctx.Post("/smtp/test", p.smtpTest)

	get("/mail", p.mailList)
	get("/mail/form", p.mailForm)
	get("/mail/form/:id", p.mailForm)
	post("/mail/save", p.mailSave)

	get("/tokens", p.tokenList)
	get("/tokens/form", p.tokenForm)
	get("/tokens/form/:id", p.tokenForm)
	post("/tokens/save", p.tokenSave)
	get("/tokens/send/:id", p.tokenSendForm)

	get("/schedule", p.scheduleList)
	get("/schedule/form", p.scheduleForm)
	get("/schedule/form/:id", p.scheduleForm)
	post("/schedule/save", p.scheduleSave)

	get("/logs", p.logList)
	get("/logs/:id", p.logDetail)

	get("/users", p.userList)
	get("/users/form", p.userForm)
	get("/users/form/:id", p.userForm)
	post("/users/save", p.userSave)

	get("/settings", p.settingsGet)
	post("/settings/save", p.settingsSave)
	post("/settings/restart", p.settingsRestart)

	get("/password", p.passwordGet)
	post("/password", p.passwordPost)
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
