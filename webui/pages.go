package webui

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
)

type authFilter = auth2.AuthenticationFilter[*model.User]

// Pages is the server-rendered management UI. It implements core.IRest and rides the
// same port and auth filter as the JSON API, but its routes live at the site root.
type Pages struct {
	ctx            *core.Context
	userService    *service.UserService
	smtpService    *service.SmtpService
	tokenService   *service.TokenService
	scheduleSvc    *service.ScheduleService
	oplogService   *service.OperationLogService
	captchaService *service.CaptchaService
	authFilter     *authFilter

	// Models are captured once: SwitchDB re-runs Init on the same instances, so the
	// pointers stay valid across a settings save.
	smtpModel       *model.SMTPModel
	mailModel       *model.MailModel
	tokenModel      *model.TokenModel
	scheduleModel   *model.ScheduleModel
	logModel        *model.LogModel
	attachmentModel *model.AttachmentModel
	userModel       *model.UserModel
	oplogModel      *model.OperationLogModel

	// Locale parsing is process-wide; the template set has its own process-wide cache.
	once    sync.Once
	loadErr error
}

// NewPages returns the UI rest component for main.go to register.
func NewPages() *Pages { return &Pages{} }

func (p *Pages) Init(ctx *core.Context) error {
	p.ctx = ctx
	p.userService = wf.GetService[*service.UserService](ctx)
	p.smtpService = wf.GetService[*service.SmtpService](ctx)
	p.tokenService = wf.GetService[*service.TokenService](ctx)
	p.scheduleSvc = wf.GetService[*service.ScheduleService](ctx)
	p.oplogService = wf.GetService[*service.OperationLogService](ctx)
	p.captchaService = wf.GetService[*service.CaptchaService](ctx)
	p.authFilter = wf.GetFilter[*authFilter](ctx)
	p.smtpModel = core.GetModel[*model.SMTPModel](ctx)
	p.mailModel = core.GetModel[*model.MailModel](ctx)
	p.tokenModel = core.GetModel[*model.TokenModel](ctx)
	p.scheduleModel = core.GetModel[*model.ScheduleModel](ctx)
	p.logModel = core.GetModel[*model.LogModel](ctx)
	p.attachmentModel = core.GetModel[*model.AttachmentModel](ctx)
	p.userModel = core.GetModel[*model.UserModel](ctx)
	p.oplogModel = core.GetModel[*model.OperationLogModel](ctx)

	p.once.Do(func() {
		p.loadErr = loadI18n()
	})
	if p.loadErr != nil {
		return p.loadErr
	}
	templateOption, err := web.LoadTemplateOption(
		web.EmbedFs("templates", templatesFS).
			Reload(devMode).
			ReloadPath(devTemplatesDir),
		".",
		web.WithTemplateFuncs(templateFuncs),
		web.WithLayouts("layout", "bare"),
		web.WithPartials("partials"),
	)
	if err != nil {
		return fmt.Errorf("load web templates: %w", err)
	}

	// Static assets live in the binary and take part in the normal route pipeline.
	// The framework chooses the embedded copy or the development directory.
	ctx.AddEmbedFs(
		"/static",
		web.EmbedFs("static", staticFS).
			Reload(devMode).
			ReloadPath(devStaticDir),
	)
	ctx.Get("/favicon.svg", p.staticFile("favicon.svg", "image/svg+xml"))

	ctx.GetTemplate("/", p.home, templateOption)
	ctx.GetTemplate("/login", p.loginGet, templateOption)
	ctx.PostTemplate("/login", p.loginPost, templateOption)
	ctx.Post("/login/captcha", p.loginCaptcha)
	ctx.PostTemplate("/logout", p.logout, templateOption)
	ctx.Any("/lang", p.lang)

	p.registerPageRoutes(ctx, templateOption)
	return nil
}

// staticFile serves one asset by name with an explicit content type — from disk in
// dev mode, from the embedded FS otherwise.
func (p *Pages) staticFile(name, contentType string) web.HandlerFunc {
	return func(req *web.Request) (any, error) {
		var data []byte
		var err error
		if devMode {
			data, err = os.ReadFile(filepath.Join(devStaticDir, name))
		} else {
			data, err = staticFS.ReadFile("static/" + name)
		}
		if err != nil {
			return nil, err
		}
		resp := req.Response()
		resp.Header().Set("Content-Type", contentType)
		_, _ = resp.Write(data)
		return nil, nil
	}
}
