package webui

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
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
}

// layoutFor maps a page template onto the skeleton it renders inside: the full app
// chrome for signed-in pages, the bare card for login/setup/error. The name is relative
// to whichever FS is in play — the embedded "templates/" prefix or the dev directory
// root — so callers add the prefix their FS needs.
func layoutFor(name string) string {
	switch name {
	case "login", "setup", "error", "notadmin":
		return "bare.html"
	default:
		return "layout.html"
	}
}

// parseTemplates parses each page together with its skeleton into its own set, so
// every page can define "content" without colliding with the others' blocks. Parsed
// once per process: restart re-runs Init and the templates do not change.
func parseTemplates() (map[string]*template.Template, error) {
	pages := []string{
		"login", "setup", "error", "notadmin",
		"dashboard",
		"smtp_list", "smtp_form", "smtp_send",
		"mail_list", "mail_form",
		"token_list", "token_form", "token_send",
		"schedule_list", "schedule_form",
		"logs_list", "logs_detail",
		"users_list", "users_form",
		"settings", "password", "restarting",
	}
	tmpls := make(map[string]*template.Template, len(pages))
	for _, name := range pages {
		t, err := template.New(name).Funcs(templateFuncs).
			ParseFS(templatesFS, "templates/"+layoutFor(name), "templates/"+name+".html")
		if err != nil {
			return nil, fmt.Errorf("parsing template %s: %w", name, err)
		}
		tmpls[name] = t
	}
	return tmpls, nil
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

// templateFor returns the template set for a page: the once-parsed embedded copy, or
// in dev mode a fresh parse from disk so edits show up on the next request.
func (p *Pages) templateFor(name string) (*template.Template, error) {
	if devMode {
		return template.New(name).Funcs(templateFuncs).
			ParseFS(os.DirFS(devTemplatesDir), layoutFor(name), name+".html")
	}
	tmpl, ok := p.tmpls[name]
	if !ok {
		return nil, fmt.Errorf("template %q is not registered", name)
	}
	return tmpl, nil
}

// render executes the named page and writes it as UTF-8 HTML. Returning nil, nil lets
// the framework's converter leave the response alone — the bytes are already written.
func (p *Pages) render(req *web.Request, name string, data any) (any, error) {
	tmpl, err := p.templateFor(name)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		return nil, err
	}
	resp := req.Response()
	resp.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := resp.Write(buf.Bytes()); err != nil {
		return nil, err
	}
	return nil, nil
}

// renderErr shows the bare error page with a translated headline and the raw detail.
func (p *Pages) renderErr(req *web.Request, status int, titleKey, detail string) (any, error) {
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
	// The status has to be set before the write; the converter sees nil, nil and stops.
	req.Response().WriteStatus(status)
	return p.render(req, "error", data)
}

// redirect answers with a 302. next must be a site-relative path: the caller has
// validated it, this just refuses anything that drifts off-site in transit.
func (p *Pages) redirect(req *web.Request, to string) (any, error) {
	if !strings.HasPrefix(to, "/") {
		to = "/"
	}
	req.Response().Redirect(http.StatusFound, to)
	return nil, nil
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
