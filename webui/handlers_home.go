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
	"github.com/chuccp/http2smtp/service"
	"go.uber.org/zap"
)

// registerPageRoutes wires the server-rendered pages. Home, logout and language
// switching redirect without rendering and are registered directly in Init.
func (p *Pages) registerPageRoutes(ctx *core.Context, templateOption web.TemplateOption) {
	ctx.GetTemplate("/setup", p.setupGet, templateOption)
	ctx.PostTemplate("/setup/db", p.setupDb, templateOption)
	ctx.Post("/setup/test", p.setupTest)
	ctx.PostTemplate("/setup/admin", p.setupAdmin, templateOption)
	ctx.PostTemplate("/setup/skip", p.setupSkip, templateOption)

	ctx.GetTemplate("/dashboard", p.dashboard, templateOption)

	ctx.GetTemplate("/smtp", p.smtpList, templateOption)
	ctx.GetTemplate("/smtp/form", p.smtpForm, templateOption)
	ctx.GetTemplate("/smtp/form/:id", p.smtpForm, templateOption)
	ctx.PostTemplate("/smtp/save", p.smtpSave, templateOption)
	ctx.GetTemplate("/smtp/send/:id", p.smtpSendForm, templateOption)
	ctx.Post("/smtp/test", p.smtpTest)

	ctx.GetTemplate("/mail", p.mailList, templateOption)
	ctx.GetTemplate("/mail/form", p.mailForm, templateOption)
	ctx.GetTemplate("/mail/form/:id", p.mailForm, templateOption)
	ctx.PostTemplate("/mail/save", p.mailSave, templateOption)

	ctx.GetTemplate("/tokens", p.tokenList, templateOption)
	ctx.GetTemplate("/tokens/form", p.tokenForm, templateOption)
	ctx.GetTemplate("/tokens/form/:id", p.tokenForm, templateOption)
	ctx.PostTemplate("/tokens/save", p.tokenSave, templateOption)
	ctx.GetTemplate("/tokens/send/:id", p.tokenSendForm, templateOption)

	ctx.GetTemplate("/schedule", p.scheduleList, templateOption)
	ctx.GetTemplate("/schedule/form", p.scheduleForm, templateOption)
	ctx.GetTemplate("/schedule/form/:id", p.scheduleForm, templateOption)
	ctx.PostTemplate("/schedule/save", p.scheduleSave, templateOption)

	ctx.GetTemplate("/logs", p.logList, templateOption)
	ctx.GetTemplate("/logs/:id", p.logDetail, templateOption)

	ctx.GetTemplate("/oplogs", p.oplogList, templateOption)
	ctx.GetTemplate("/oplogs/:id", p.oplogDetail, templateOption)

	ctx.GetTemplate("/users", p.userList, templateOption)
	ctx.GetTemplate("/users/form", p.userForm, templateOption)
	ctx.GetTemplate("/users/form/:id", p.userForm, templateOption)
	ctx.PostTemplate("/users/save", p.userSave, templateOption)

	ctx.GetTemplate("/settings", p.settingsGet, templateOption)
	ctx.PostTemplate("/settings/save", p.settingsSave, templateOption)
	ctx.PostTemplate("/settings/restart", p.settingsRestart, templateOption)

	ctx.GetTemplate("/password", p.passwordGet, templateOption)
	ctx.PostTemplate("/password", p.passwordPost, templateOption)
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

// CaptchaView is the login page's challenge: the rendered shapes, the translated click
// order and the id the answer is checked against. The image and prompt also travel
// through /login/captcha when the refresh button asks for a new one.
type CaptchaView struct {
	ID    string `json:"id"`
	Image string `json:"image"`
	// Prompt is the text fallback; the page renders Shapes as icons in order and uses
	// Labels for their accessible names, so the sequence itself needs no translation.
	Prompt string   `json:"prompt"`
	Shapes []string `json:"shapes"`
	Labels []string `json:"labels"`
	Count  int      `json:"count"`
	Width  int      `json:"width"`
	Height int      `json:"height"`
}

var captchaShapeKeys = map[string]string{
	"circle":   "captcha.circle",
	"triangle": "captcha.triangle",
	"square":   "captcha.square",
	"diamond":  "captcha.diamond",
	"star":     "captcha.star",
}

func (p *Pages) captchaView(lang string, data *service.CaptchaData) *CaptchaView {
	tr := translatorFor(lang)
	labels := make([]string, 0, len(data.Shapes))
	for _, shape := range data.Shapes {
		if key, ok := captchaShapeKeys[shape]; ok {
			labels = append(labels, tr.T(key))
		} else {
			labels = append(labels, shape)
		}
	}
	return &CaptchaView{
		ID:     data.ID,
		Image:  data.Image,
		Prompt: strings.Join(labels, " → "),
		Shapes: append([]string(nil), data.Shapes...),
		Labels: labels,
		Count:  len(data.Shapes),
		Width:  data.Width,
		Height: data.Height,
	}
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

// loginCaptcha issues a fresh challenge bound to the credentials the form is about to
// submit. Sending them here (and not only with the login POST) lets the service tie the
// challenge to this exact account and password, so a solved challenge cannot be pointed
// at another account or replayed.
func (p *Pages) loginCaptcha(req *web.Request) (any, error) {
	username := req.GetFormParam("username")
	password := req.GetFormParam("password")
	if username == "" || password == "" {
		return nil, web.NewErrorCode(web.CodeBadRequest, "username and password are required")
	}
	data, err := p.captchaService.New(username, password)
	if err != nil {
		return nil, err
	}
	return web.Data(p.captchaView(langOf(req), data)), nil
}

func (p *Pages) loginPost(view *web.Template) error {
	req := view.Request()
	username := req.GetFormParam("username")
	password := req.GetFormParam("password")
	next := safeNext(req.GetFormParam("next"))

	// Check the captcha before the password: an invalid answer must not become a
	// password oracle, and the challenge is consumed by the attempt either way.
	if !p.captchaService.VerifyForm(req.GetFormParam("captchaId"), username, password, req.GetFormParam("captchaXs"), req.GetFormParam("captchaYs")) {
		data := p.loginData(req, next, Flash{Err: translatorFor(langOf(req)).T("auth.captchaFailed")})
		return renderPageStatus(view, "login", data, http.StatusUnauthorized)
	}

	dbUser, err := p.userService.ValidateLogin(username, password)
	if err != nil {
		// Failed attempts are not audited: the endpoint is unauthenticated, so recording
		// them would let anyone write rows into the audit table for free.
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
	p.recordOp(req, dbUser.Id, entity.OpLogin, entity.TargetUser, dbUser.Id, dbUser.Name)
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
		p.recordOp(req, user.Id, entity.OpLogout, entity.TargetUser, user.Id, user.Name)
	}
	if _, err := p.authFilter.SignOut(req); err != nil {
		return err
	}
	return redirect(view, "/login")
}

// lang switches the interface language and returns to where the visitor came from.
// The return target is restricted to same-site paths.
func (p *Pages) lang(req *web.Request) (any, error) {
	switchLanguage(req)
	return nil, nil
}

func switchLanguage(req *web.Request) {
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
}
