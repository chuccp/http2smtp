package webui

import (
	"html/template"
	"io/fs"
	"net/http"
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
	ctx         *core.Context
	userService *service.UserService
	authFilter  *authFilter

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
	p.authFilter = wf.GetFilter[*authFilter](ctx)

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

	// Static assets live in the binary and are served through the route table (not the
	// NoRoute fallback), so they take part in the same pipeline as everything else.
	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return err
	}
	ctx.StaticFs("/static", http.FS(staticSub))
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

// staticFile serves one embedded asset by name with an explicit content type.
func (p *Pages) staticFile(name, contentType string) web.HandlerFunc {
	return func(req *web.Request) (any, error) {
		data, err := staticFS.ReadFile("static/" + name)
		if err != nil {
			return nil, err
		}
		resp := req.Response()
		resp.Header().Set("Content-Type", contentType)
		_, _ = resp.Write(data)
		return nil, nil
	}
}
