package webui

import (
	"strconv"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/configapply"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
)

// SettingsData is the system settings page (admin only).
type SettingsData struct {
	Base
	IsDocker      bool
	WebPort       int
	ApiPort       int
	DbType        string
	SqliteFile    string
	MysqlHost     string
	MysqlPort     int
	MysqlDbname   string
	MysqlUsername string
	MysqlCharset  string
}

func (p *Pages) settingsGet(req *web.Request) (any, error) {
	user, ok := p.requireAdmin(req)
	if !ok {
		return nil, nil
	}
	data := &SettingsData{Base: p.base(req, user, "settings.systemSettings", true)}
	cfg, err := model.GetConfig(p.ctx.GetConfig())
	if err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	data.IsDocker = cfg.Core.IsDocker
	data.DbType = cfg.Core.DbType
	if data.DbType == "" {
		data.DbType = "sqlite"
	}
	if cfg.Manage != nil {
		data.WebPort = cfg.Manage.Port
	}
	if cfg.Api != nil {
		data.ApiPort = cfg.Api.Port
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
	return p.render(req, "settings", data)
}

// settingsFromForm reads the database section of the settings form. A blank MySQL
// password means "keep the stored one" — MaskSecrets never echoes it.
func settingsFromForm(req *web.Request) *model.Config {
	setInfo := model.DefaultConfig()
	setInfo.Core.DbType = req.GetFormParam("dbType")
	if setInfo.Core.DbType == "sqlite" {
		setInfo.Sqlite.Filename = req.GetFormParam("sqliteFile")
	} else {
		setInfo.Mysql.Host = req.GetFormParam("mysqlHost")
		setInfo.Mysql.Port, _ = strconv.Atoi(req.GetFormParam("mysqlPort"))
		setInfo.Mysql.Dbname = req.GetFormParam("mysqlDbname")
		setInfo.Mysql.Username = req.GetFormParam("mysqlUsername")
		setInfo.Mysql.Password = req.GetFormParam("mysqlPassword")
		setInfo.Mysql.Charset = req.GetFormParam("mysqlCharset")
	}
	if req.GetFormParam("isDocker") == "" {
		// Under Docker the ports are the container's; a posted value must not move them.
		setInfo.Manage.Port, _ = strconv.Atoi(req.GetFormParam("webPort"))
		setInfo.Api.Port, _ = strconv.Atoi(req.GetFormParam("apiPort"))
	}
	return setInfo
}

// settingsSave persists without restarting; the ports only take effect through the
// save-and-restart button.
func (p *Pages) settingsSave(req *web.Request) (any, error) {
	if _, ok := p.requireAdmin(req); !ok {
		return nil, nil
	}
	if err := configapply.ApplyReSet(p.ctx, settingsFromForm(req)); err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	return p.redirect(req, backTo("/settings", "settings.settingsSaved", "", ""))
}

// RestartingData is the post-restart interstitial. The client's own origin may be gone
// by the time it renders, so reload only happens when the port did not move.
type RestartingData struct {
	Base
	Message     string
	ManagePort  int
	ApiPort     int
	CurrentPort int
	PortChanged bool
}

func (p *Pages) settingsRestart(req *web.Request) (any, error) {
	if _, ok := p.requireAdmin(req); !ok {
		return nil, nil
	}
	// The port this page was served from — captured before ApplyReSet rewrites the
	// config, which is what tells the interstitial whether reloading can work at all.
	current := 0
	if cfg, err := model.GetConfig(p.ctx.GetConfig()); err == nil && cfg.Manage != nil {
		current = cfg.Manage.Port
	}
	if err := configapply.ApplyReSet(p.ctx, settingsFromForm(req)); err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	managePort, apiPort, err := wf.GetService[*service.RestartService](p.ctx).Restart()
	if err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	data := &RestartingData{
		Base:        p.newBase(langOf(req), "settings.restarting"),
		ManagePort:  managePort,
		ApiPort:     apiPort,
		CurrentPort: current,
		PortChanged: managePort != current,
	}
	data.Message = translatorFor(data.Lang).T("settings.restartPortChanged", strconv.Itoa(managePort), strconv.Itoa(apiPort))
	return p.render(req, "restarting", data)
}
