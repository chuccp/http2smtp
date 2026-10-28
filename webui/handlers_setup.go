package webui

import (
	"errors"
	"net/http"
	"strconv"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/configapply"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
	"github.com/chuccp/http2smtp/util"
)

// SetupData is the wizard's template data. Values are the raw form contents, echoed
// back on a failed test so nothing has to be retyped.
type SetupData struct {
	Base
	Step      int
	HasDbInit bool
	HasAdmin  bool
	AdminName string
	IsDocker  bool
	WebPort   int
	ApiPort   int
	// database form
	DbType        string
	SqliteFile    string
	MysqlHost     string
	MysqlPort     int
	MysqlDbname   string
	MysqlUsername string
	MysqlCharset  string
	// admin form
	AdminUsername string
}

func (p *Pages) setupGet(view *web.Template) error {
	req := view.Request()
	st := model.CoreState(p.ctx.GetConfig())
	if st.Init {
		return redirect(view, "/login")
	}
	step := 1
	if st.DbInit {
		step = 2
	}
	if s := req.Query("step"); s == "1" {
		step = 1
	}

	data := &SetupData{Base: p.newBase(langOf(req), "setup.setupWizard")}
	data.Step = step
	data.HasDbInit = st.DbInit

	cfg, err := model.GetConfig(p.ctx.GetConfig())
	if err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	data.IsDocker = cfg.Core.IsDocker
	if cfg.Manage != nil {
		data.WebPort = cfg.Manage.Port
	}
	if cfg.Api != nil {
		data.ApiPort = cfg.Api.Port
	}
	// Echo the stored database settings into the form
	if cfg.Core.DbType != "" {
		data.DbType = cfg.Core.DbType
	} else {
		data.DbType = "sqlite"
	}
	if cfg.Sqlite != nil {
		data.SqliteFile = cfg.Sqlite.Filename
	}
	if cfg.Mysql != nil {
		data.MysqlHost = cfg.Mysql.Host
		data.MysqlPort = cfg.Mysql.Port
		data.MysqlDbname = cfg.Mysql.Dbname
		data.MysqlUsername = cfg.Mysql.Username
		data.MysqlCharset = cfg.Mysql.Charset
	}

	if st.DbInit {
		userService := wf.GetService[*service.UserService](p.ctx)
		if hasAdmin, _ := userService.HasAdminUser(); hasAdmin {
			data.HasAdmin = true
			data.AdminName, _ = userService.GetAdminUsername()
		}
	}
	view.View("setup").Layout(layoutFor("setup")).Data(data)
	return nil
}

// setupDbFromForm reads the database section of the wizard form into a Config.
func (p *Pages) setupDbFromForm(req *web.Request) *model.Config {
	setInfo := model.DefaultConfig()
	setInfo.Core.DbType = req.GetFormParam("dbType")
	if setInfo.Core.DbType == "sqlite" {
		setInfo.Sqlite.Filename = req.GetFormParam("sqliteFile")
	} else {
		setInfo.Mysql.Host = req.GetFormParam("mysqlHost")
		setInfo.Mysql.Port, _ = strconv.Atoi(req.GetFormParam("mysqlPort"))
		setInfo.Mysql.Dbname = req.GetFormParam("mysqlDbname")
		setInfo.Mysql.Username = req.GetFormParam("mysqlUsername")
		// A blank password on the test button means "the stored one": the probe needs
		// whatever is already configured, exactly like the saved settings would use.
		setInfo.Mysql.Password = req.GetFormParam("mysqlPassword")
		setInfo.Mysql.Charset = req.GetFormParam("mysqlCharset")
		if setInfo.Mysql.Charset == "" {
			setInfo.Mysql.Charset = "utf8mb4"
		}
	}
	return setInfo
}

// setupRebuild regenerates the form data from the submitted values so a failed test
// connection keeps everything that was typed.
func (p *Pages) setupRebuild(req *web.Request, data *SetupData) {
	data.DbType = req.GetFormParam("dbType")
	data.SqliteFile = req.GetFormParam("sqliteFile")
	data.MysqlHost = req.GetFormParam("mysqlHost")
	data.MysqlPort, _ = strconv.Atoi(req.GetFormParam("mysqlPort"))
	data.MysqlDbname = req.GetFormParam("mysqlDbname")
	data.MysqlUsername = req.GetFormParam("mysqlUsername")
	data.MysqlCharset = req.GetFormParam("mysqlCharset")
}

// setupTest is the step-1 "test connection" button: probe the typed settings without
// saving anything and answer the fetch with the verdict, so the page keeps the values
// and shows the result in a dialog instead of re-rendering.
func (p *Pages) setupTest(req *web.Request) (any, error) {
	st := model.CoreState(p.ctx.GetConfig())
	if st.Init {
		return nil, errors.New("has init")
	}
	if err := configapply.ProbeDb(p.setupDbFromForm(req)); err != nil {
		return nil, err
	}
	return web.Ok(translatorFor(langOf(req)).T("setup.connectionSuccess")), nil
}

// setupDb is step 1's save: persist the database settings and move on to the admin step.
func (p *Pages) setupDb(view *web.Template) error {
	req := view.Request()
	st := model.CoreState(p.ctx.GetConfig())
	if st.Init {
		return redirect(view, "/login")
	}
	if err := configapply.ApplyDbInit(p.ctx, p.setupDbFromForm(req)); err != nil {
		data := &SetupData{Base: p.newBase(langOf(req), "setup.setupWizard"), Step: 1}
		data.DbType = "sqlite"
		p.setupRebuild(req, data)
		data.Flash = Flash{Err: err.Error()}
		return renderPageStatus(view, "setup", data, http.StatusBadRequest)
	}
	return redirect(view, "/setup?step=2")
}

// setupAdmin is step 2: create the administrator (or reset the existing one's
// credentials) and finish setup.
func (p *Pages) setupAdmin(view *web.Template) error {
	req := view.Request()
	st := model.CoreState(p.ctx.GetConfig())
	if st.Init {
		return redirect(view, "/login")
	}
	fail := func(msgKey string) error {
		data := p.setupStep2Data(req)
		data.Flash = Flash{Err: translatorFor(data.Lang).T(msgKey)}
		return renderPageStatus(view, "setup", data, http.StatusBadRequest)
	}
	username := req.GetFormParam("username")
	password := req.GetFormParam("password")
	confirm := req.GetFormParam("confirm")
	if username == "" {
		return fail("setup.adminUsernameRequired")
	}
	if password == "" {
		return fail("setup.adminPasswordRequired")
	}
	if confirm != password {
		return fail("setup.passwordsNotMatch")
	}
	if err := util.ValidatePassword(password); err != nil {
		return fail("auth.passwordMinLength")
	}
	u := &entity.LoginUser{Username: username, Password: password}
	if err := configapply.ApplyAdminInit(p.ctx, u); err != nil {
		data := p.setupStep2Data(req)
		data.AdminUsername = username
		data.Flash = Flash{Err: err.Error()}
		return renderPageStatus(view, "setup", data, http.StatusBadRequest)
	}
	return redirect(view, "/login?ok=setup.initializationSuccess")
}

// setupSkip finishes setup without touching the existing admin account.
func (p *Pages) setupSkip(view *web.Template) error {
	req := view.Request()
	st := model.CoreState(p.ctx.GetConfig())
	if st.Init {
		return redirect(view, "/login")
	}
	if err := configapply.ApplyAdminInitSkip(p.ctx); err != nil {
		data := p.setupStep2Data(req)
		data.Flash = Flash{Err: err.Error()}
		return renderPageStatus(view, "setup", data, http.StatusBadRequest)
	}
	return redirect(view, "/login?ok=setup.initializationSuccess")
}

func (p *Pages) setupStep2Data(req *web.Request) *SetupData {
	st := model.CoreState(p.ctx.GetConfig())
	data := &SetupData{Base: p.newBase(langOf(req), "setup.setupWizard"), Step: 2, HasDbInit: st.DbInit}
	if st.DbInit {
		userService := wf.GetService[*service.UserService](p.ctx)
		if hasAdmin, _ := userService.HasAdminUser(); hasAdmin {
			data.HasAdmin = true
			data.AdminName, _ = userService.GetAdminUsername()
		}
	}
	return data
}
