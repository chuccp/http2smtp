package webui

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
)

// devMode reads templates and static assets from disk on every request instead of the
// embedded copies, so a template edit only needs a browser refresh — no rebuild, no
// restart. It defaults to on for local builds and off for release builds
// (`go build -tags release`, which the release workflow uses); WEBUI_DEV=0 or 1
// overrides either default. When on, run the binary from the repository root, where
// the webui/ directory lives. The locale JSON stays embedded either way:
// translations change far less often than templates, and the key-sync test guards
// them.
var devMode = resolveDevMode()

func resolveDevMode() bool {
	switch os.Getenv("WEBUI_DEV") {
	case "1", "true", "yes":
		return true
	case "0", "false", "no":
		return false
	}
	return devDefault
}

const devTemplatesDir = "webui/templates"
const devStaticDir = "webui/static"

// NavUser is the slice of the signed-in user the layout may show. Nothing sensitive —
// the session itself stays in the HttpOnly cookie.
type NavUser struct {
	Id      uint
	Name    string
	IsAdmin bool
}

// Flash carries a post-redirect message into the next page render. OK is an i18n key
// (optionally with one raw argument for {count}-style placeholders); Err is server
// error text shown as-is — the services report in English, same as the JSON API.
type Flash struct {
	OK  string
	OKN string
	Err string
}

func (f Flash) empty() bool { return f.OK == "" && f.Err == "" }

// Base is embedded in every page's data struct. Title is already translated.
type Base struct {
	Lang  string
	L     *Translator
	User  *NavUser
	Title string
	Flash Flash
	// Path is the request path — the tab bar and the nav highlight are drawn client-side
	// from location, but the server needs it for the language form's return target.
	Path string
	// Tab says whether this page belongs in the tab bar; list/detail pages do,
	// login/setup/form pages do not.
	Tab bool
	// Languages and LangNames back the language switcher in the layout.
	Languages []string
	LangNames map[string]string
}

// pageData is implemented by every data struct via embedding Base.
type pageData interface{ base() *Base }

func (b *Base) base() *Base { return b }

// templateFuncs are the only helpers bound at parse time. Everything request-specific
// (language above all) travels through the data, never through a func closure.
var templateFuncs = template.FuncMap{
	"fmtTime": func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02 15:04:05")
	},
	"shortToken": func(s string) string {
		if len(s) <= 12 {
			return s
		}
		return s[:12] + "..."
	},
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	"seq": func(from, to int) []int {
		if to < from {
			return nil
		}
		s := make([]int, 0, to-from+1)
		for i := from; i <= to; i++ {
			s = append(s, i)
		}
		return s
	},
	"flashText": flashText,
	// dict builds the option map the form partials take. It only assembles data;
	// labels are translated inside the partial with the request's .L.T, so this is
	// not the parse-time translation the templates must avoid.
	"dict": func(values ...any) map[string]any {
		m := make(map[string]any, len(values)/2)
		for i := 0; i+1 < len(values); i += 2 {
			m[fmt.Sprint(values[i])] = values[i+1]
		}
		return m
	},
}

// layoutFor maps a page template onto its skeleton: the full app chrome for signed-in
// pages, the bare card for login/setup/error.
func layoutFor(name string) string {
	switch name {
	case "login", "setup", "error", "notadmin":
		return "bare"
	default:
		return "layout"
	}
}

var (
	templateSetOnce sync.Once
	loadedTemplates *web.TemplateSet
	templateSetErr  error
)

// loadTemplateSet discovers and compiles every page/layout combination once. The
// returned set is immutable and can be reused when the framework restarts.
func loadTemplateSet() (*web.TemplateSet, error) {
	templateSetOnce.Do(func() {
		var fsys fs.FS = templatesFS
		root := "templates"
		if devMode {
			fsys = os.DirFS(devTemplatesDir)
			root = "."
		}
		loadedTemplates, templateSetErr = loadTemplates(fsys, root, devMode)
		if templateSetErr != nil && devMode {
			templateSetErr = fmt.Errorf("development mode serves templates from disk but %s could not be loaded: %w — run the binary from the repository root, or build with -tags release for a self-contained binary (WEBUI_DEV=0 also forces the embedded copies in a dev build)", devTemplatesDir, templateSetErr)
		}
	})
	return loadedTemplates, templateSetErr
}

func loadTemplates(fsys fs.FS, root string, devReload bool) (*web.TemplateSet, error) {
	options := []web.TemplateOption{
		web.WithTemplateFuncs(templateFuncs),
		web.WithLayouts("layout", "bare"),
		web.WithPartials("partials"),
	}
	if devReload {
		options = append(options, web.WithDevReload(true))
	}
	return web.LoadTemplates(fsys, root, options...)
}

// newBase builds a Base for pages that construct their data by hand (login, errors):
// active language, translated title and the language switcher's options.
func (p *Pages) newBase(lang, titleKey string) Base {
	tr := translatorFor(lang)
	return Base{
		Lang:      lang,
		L:         tr,
		Title:     tr.T(titleKey),
		Languages: Languages,
		LangNames: langNames,
	}
}

// base builds the Base for a page render: active language, translated title, nav user
// and any flash carried on the URL.
func (p *Pages) base(req *web.Request, user *model.User, titleKey string, tab bool) Base {
	lang := langOf(req)
	b := p.newBase(lang, titleKey)
	b.Path = req.URL().Path
	b.Tab = tab
	if user != nil {
		b.User = &NavUser{Id: user.Id, Name: user.Name, IsAdmin: user.IsAdmin}
	}
	b.Flash = flashOf(req)
	return b
}

func renderPageStatus(view *web.Template, name string, data any, status int) error {
	view.View(name).Layout(layoutFor(name)).Status(status).Data(data)
	return nil
}

// renderErr shows the bare error page with a translated headline and the raw detail.
func (p *Pages) renderErr(view *web.Template, status int, titleKey, detail string) error {
	req := view.Request()
	lang := langOf(req)
	tr := translatorFor(lang)
	b := p.newBase(lang, titleKey)
	b.Path = req.URL().Path
	data := struct {
		Base
		Message string
		Detail  string
	}{
		Base:    b,
		Message: tr.T(titleKey),
		Detail:  detail,
	}
	if status == 0 {
		status = http.StatusInternalServerError
	}
	return renderPageStatus(view, "error", data, status)
}

// redirect answers with a 302 after validating that the target is site-relative.
func redirect(view *web.Template, to string) error {
	if !strings.HasPrefix(to, "/") {
		to = "/"
	}
	view.Redirect(http.StatusFound, to)
	return nil
}

// safeNext validates a ?next= return target: same-site absolute path only, so a
// crafted login link cannot bounce the browser off to another origin.
func safeNext(next string) string {
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		return "/dashboard"
	}
	if _, err := url.ParseRequestURI(next); err != nil {
		return "/dashboard"
	}
	return next
}

// langOf resolves the request's language: the lang cookie first, then Accept-Language.
func langOf(req *web.Request) string {
	return pickLang(req.Cookie().Get("lang"), req.GetHeader("Accept-Language"))
}
