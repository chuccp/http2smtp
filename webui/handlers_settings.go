package webui

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/configapply"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
)

// SettingsData is the system settings page (admin only).
type SettingsData struct {
	Base
	IsDocker      bool
	WebPort       int
	ApiPort       int
	MaxBodySize   string
	DbType        string
	SqliteFile    string
	MysqlHost     string
	MysqlPort     int
	MysqlDbname   string
	MysqlUsername string
	MysqlCharset  string
}

func (p *Pages) settingsGet(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireAdmin(view)
	if !ok {
		return nil
	}
	data := &SettingsData{Base: p.base(req, user, "settings.systemSettings", true)}
	cfg, err := model.GetConfig(p.ctx.GetConfig())
	if err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	data.IsDocker = cfg.Core.IsDocker
	data.MaxBodySize = cfg.Core.MaxBodySize.String()
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
	view.View("settings").Layout(layoutFor("settings")).Data(data)
	return nil
}

// settingsFromForm reads the database and general section of the settings form. A
// blank MySQL password means "keep the stored one" — MaskSecrets never echoes it.
func settingsFromForm(req *web.Request) (*model.Config, error) {
	setInfo := model.DefaultConfig()
	// A missing field (an old page left open across an upgrade) keeps the current value;
	// a present but invalid one is a real error.
	if raw := strings.TrimSpace(req.GetFormParam("maxBodySize")); raw != "" {
		maxBodySize, err := model.ParseByteSize(raw)
		if err != nil {
			return nil, fmt.Errorf("maxBodySize: %w", err)
		}
		setInfo.Core.MaxBodySize = maxBodySize
	}
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
	return setInfo, nil
}

// settingsSave persists without restarting; the ports only take effect through the
// save-and-restart button.
func (p *Pages) settingsSave(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireAdmin(view)
	if !ok {
		return nil
	}
	setInfo, err := settingsFromForm(req)
	if err != nil {
		return p.renderErr(view, http.StatusBadRequest, "common.failed", err.Error())
	}
	if err := configapply.ApplyReSet(p.ctx, setInfo); err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	p.recordOp(req, user.Id, entity.OpUpdate, entity.TargetSettings, 0, "")
	return redirect(view, backTo("/settings", "settings.settingsSaved", "", ""))
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

func (p *Pages) settingsRestart(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireAdmin(view)
	if !ok {
		return nil
	}
	// The port this page was served from — captured before ApplyReSet rewrites the
	// config, which is what tells the interstitial whether reloading can work at all.
	current := 0
	if cfg, err := model.GetConfig(p.ctx.GetConfig()); err == nil && cfg.Manage != nil {
		current = cfg.Manage.Port
	}
	setInfo, err := settingsFromForm(req)
	if err != nil {
		return p.renderErr(view, http.StatusBadRequest, "common.failed", err.Error())
	}
	if err := configapply.ApplyReSet(p.ctx, setInfo); err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	managePort, apiPort, err := wf.GetService[*service.RestartService](p.ctx).Restart()
	if err != nil {
		// The settings were saved before the restart was attempted, so audit the save
		// even though no restart is going to happen.
		p.recordOp(req, user.Id, entity.OpUpdate, entity.TargetSettings, 0, "")
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	p.recordOp(req, user.Id, entity.OpRestart, entity.TargetSettings, 0, "")
	data := &RestartingData{
		Base:        p.base(req, user, "settings.restarting", false),
		ManagePort:  managePort,
		ApiPort:     apiPort,
		CurrentPort: current,
		PortChanged: managePort != current,
	}
	data.Message = translatorFor(data.Lang).T("settings.restartPortChanged", strconv.Itoa(managePort), strconv.Itoa(apiPort))
	view.View("restarting").Layout(layoutFor("restarting")).Data(data)
	return nil
}
