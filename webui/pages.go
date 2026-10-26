package webui

import (
	"fmt"
	"html/template"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	ctx          *core.Context
	userService  *service.UserService
	smtpService  *service.SmtpService
	tokenService *service.TokenService
	scheduleSvc  *service.ScheduleService
	authFilter   *authFilter

	// Models are captured once: SwitchDB re-runs Init on the same instances, so the
	// pointers stay valid across a settings save.
	smtpModel     *model.SMTPModel
	mailModel     *model.MailModel
	tokenModel    *model.TokenModel
	scheduleModel *model.ScheduleModel
	logModel      *model.LogModel

	// Templates are parsed once per process: a restart re-runs Init on a fresh engine
	// and re-registering the routes is cheap, re-parsing is not.
	once    sync.Once
	loadErr error
	tmpls   map[string]*template.Template
}

// NewPages returns the UI rest component for main.go to register.
func NewPages() *Pages { return &Pages{} }

func (p *Pages) Init(ctx *core.Context) error {
	p.ctx = ctx
	p.userService = wf.GetService[*service.UserService](ctx)
	p.smtpService = wf.GetService[*service.SmtpService](ctx)
	p.tokenService = wf.GetService[*service.TokenService](ctx)
	p.scheduleSvc = wf.GetService[*service.ScheduleService](ctx)
	p.authFilter = wf.GetFilter[*authFilter](ctx)
	p.smtpModel = core.GetModel[*model.SMTPModel](ctx)
	p.mailModel = core.GetModel[*model.MailModel](ctx)
	p.tokenModel = core.GetModel[*model.TokenModel](ctx)
	p.scheduleModel = core.GetModel[*model.ScheduleModel](ctx)
	p.logModel = core.GetModel[*model.LogModel](ctx)

	p.once.Do(func() {
		if err := loadI18n(); err != nil {
			p.loadErr = err
			return
		}
		tmpls, err := parseTemplates()
		if err != nil {
			p.loadErr = err
			return
		}
		p.tmpls = tmpls
	})
	if p.loadErr != nil {
		return p.loadErr
	}
	// In dev mode the templates are re-read per request; fail early and loudly when
	// the directory is not where the docs say it is (binary started outside the repo).
	if devMode {
		if fi, err := os.Stat(devTemplatesDir); err != nil || !fi.IsDir() {
			return fmt.Errorf("WEBUI_DEV is set but %s is not a directory — run the binary from the repository root", devTemplatesDir)
		}
	}

	// Static assets live in the binary and are served through the route table (not the
	// NoRoute fallback), so they take part in the same pipeline as everything else.
	// In dev mode they are read from disk per request, like the templates.
	if devMode {
		ctx.Get("/static/*filepath", p.devStatic)
	} else {
		staticSub, err := fs.Sub(staticFS, "static")
		if err != nil {
			return err
		}
		ctx.StaticFs("/static", http.FS(staticSub))
	}
	ctx.Get("/favicon.svg", p.staticFile("favicon.svg", "image/svg+xml"))
	ctx.Get("/icons.svg", p.staticFile("icons.svg", "image/svg+xml"))

	ctx.Get("/", p.home)
	ctx.Get("/login", p.loginGet)
	ctx.Post("/login", p.loginPost)
	ctx.Post("/logout", p.logout)
	ctx.Any("/lang", p.lang)

	p.registerPageRoutes(ctx)
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

// devStatic serves /static/* from the working tree. The path is cleaned and confined
// to devStaticDir — a traversal attempt must not read the filesystem at large.
func (p *Pages) devStatic(req *web.Request) (any, error) {
	clean := path.Clean("/" + req.Param("filepath"))
	// Clean both sides so the comparison holds on Windows, where Clean rewrites the
	// forward slashes in devStaticDir to backslashes.
	base := filepath.Clean(devStaticDir)
	full := filepath.Join(base, clean)
	if !strings.HasPrefix(full, base+string(filepath.Separator)) {
		return nil, os.ErrNotExist
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return nil, err
	}
	contentType := mime.TypeByExtension(filepath.Ext(full))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	resp := req.Response()
	resp.Header().Set("Content-Type", contentType)
	_, _ = resp.Write(data)
	return nil, nil
}
