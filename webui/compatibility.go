package webui

import (
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/compatibility"
)

// CompatibilityPage is the management-only UI shown when the persisted data belongs
// to an incompatible release. It deliberately has no database-backed routes.
type CompatibilityPage struct {
	issue *compatibility.IncompatibleError
}

// NewCompatibilityPage returns the standalone page registered by main.go when the
// database must not be attached.
func NewCompatibilityPage(issue *compatibility.IncompatibleError) *CompatibilityPage {
	return &CompatibilityPage{issue: issue}
}

func (p *CompatibilityPage) Init(ctx *core.Context) error {
	if err := loadI18n(); err != nil {
		return err
	}
	templateOption, err := loadPageTemplates()
	if err != nil {
		return err
	}
	ctx.AddEmbedFs(
		"/static",
		web.EmbedFs("static", staticFS).
			Reload(devMode).
			ReloadPath(devStaticDir),
	)
	ctx.Get("/favicon.svg", staticFileHandler("favicon.svg", "image/svg+xml"))

	// Cover the landing page and every top-level management entry. /settings matters
	// most here: a bookmarked settings URL must show the compatibility page rather than
	// a login redirect or a database-backed error.
	for _, path := range []string{
		"/",
		"/login",
		"/setup",
		"/dashboard",
		"/smtp",
		"/mail",
		"/tokens",
		"/schedule",
		"/logs",
		"/oplogs",
		"/users",
		"/settings",
	} {
		ctx.GetTemplate(path, p.render, templateOption)
	}
	ctx.Any("/lang", func(req *web.Request) (any, error) {
		switchLanguage(req)
		return nil, nil
	})
	return nil
}

// CompatibilityData is the recovery page's template data.
type CompatibilityData struct {
	Base
	ProblemCode     compatibility.ProblemCode
	RawVersion      string
	FoundVersion    int
	RequiredVersion int
	ConfigPath      string
	SQLitePath      string
	StorageRoot     string
	DatabaseType    string
	MySQLDatabase   string
}

func (p *CompatibilityPage) render(view *web.Template) error {
	if p.issue == nil {
		return nil
	}
	req := view.Request()
	lang := langOf(req)
	tr := translatorFor(lang)
	data := &CompatibilityData{
		Base: Base{
			Lang:      lang,
			L:         tr,
			Title:     tr.T("compatibility.pageTitle"),
			Path:      req.URL().Path,
			Languages: Languages,
			LangNames: langNames,
		},
		ProblemCode:     p.issue.Code,
		RawVersion:      p.issue.RawVersion,
		FoundVersion:    p.issue.FoundVersion,
		RequiredVersion: p.issue.RequiredVersion,
		ConfigPath:      p.issue.ConfigPath,
		SQLitePath:      p.issue.SQLitePath,
		StorageRoot:     p.issue.StorageRoot,
		DatabaseType:    p.issue.DatabaseType,
		MySQLDatabase:   p.issue.MySQLDatabase,
	}
	view.View("compatibility").Layout("bare").Data(data)
	return nil
}
