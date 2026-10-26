// Package configapply holds the configuration-writing bodies shared by the JSON API
// (rest.Set, which binds a request body) and the server-rendered setup/settings pages
// (webui, which binds form fields). Both must persist exactly the same way — the
// wizard used to exist twice.
package configapply

import (
	"errors"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	framedb "github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/log"
	h2sdb "github.com/chuccp/http2smtp/db"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
	"github.com/chuccp/http2smtp/util"
	"go.uber.org/zap"
)

// putDbSettings copies the non-empty database fields of setInfo onto the live config.
// Empty fields are left alone so a partial form never wipes what config.ini holds.
func putDbSettings(ctx *core.Context, setInfo *model.Config) {
	if setInfo.Core.DbType != "" {
		ctx.GetConfig().Put("core.dbtype", setInfo.Core.DbType)
	}
	if setInfo.Sqlite != nil && setInfo.Sqlite.Filename != "" {
		ctx.GetConfig().Put("sqlite.filename", setInfo.Sqlite.Filename)
	}
	if setInfo.Mysql != nil {
		m := setInfo.Mysql
		if m.Host != "" {
			ctx.GetConfig().Put("mysql.host", m.Host)
		}
		if m.Port > 0 {
			ctx.GetConfig().Put("mysql.port", m.Port)
		}
		if m.Dbname != "" {
			ctx.GetConfig().Put("mysql.dbname", m.Dbname)
		}
		if m.Username != "" {
			ctx.GetConfig().Put("mysql.username", m.Username)
		}
		if m.Password != "" {
			ctx.GetConfig().Put("mysql.password", m.Password)
		}
		if m.Charset != "" {
			ctx.GetConfig().Put("mysql.charset", m.Charset)
		}
	}
}

func putPortSettings(ctx *core.Context, setInfo *model.Config) {
	if setInfo.Manage != nil && setInfo.Manage.Port > 0 {
		ctx.GetConfig().Put("manage.port", setInfo.Manage.Port)
	}
	if setInfo.Api != nil && setInfo.Api.Port > 0 {
		ctx.GetConfig().Put("api.port", setInfo.Api.Port)
	}
}

// attachDb opens the database named by the live config and hot-swaps it under the
// running models, then persists the file.
func attachDb(ctx *core.Context) error {
	createDB, err := h2sdb.GetDb(ctx.GetConfig())
	if err != nil {
		return err
	}
	if err := ctx.DefaultModelGroup().SwitchDB(createDB, ctx); err != nil {
		return err
	}
	return ctx.GetConfig().WriteConfig()
}

// ApplyDbInit is setup step 1: persist the database settings, connect, and mark the
// database initialized. core.init stays false until an admin exists.
func ApplyDbInit(ctx *core.Context, setInfo *model.Config) error {
	putDbSettings(ctx, setInfo)
	putPortSettings(ctx, setInfo)
	ctx.GetConfig().Put("core.dbinit", "true")
	return attachDb(ctx)
}

// ApplyAdminInit is setup step 2: create the administrator, or reset the existing one's
// credentials, then mark the system initialized.
func ApplyAdminInit(ctx *core.Context, u *entity.LoginUser) error {
	if len(u.Username) == 0 || len(u.Password) == 0 {
		return errors.New("username or password is blank")
	}
	if err := util.ValidatePassword(u.Password); err != nil {
		return err
	}
	userService := wf.GetService[*service.UserService](ctx)
	hasAdmin, _ := userService.HasAdminUser()
	if hasAdmin {
		if err := userService.ResetAdminPassword(u.Username, u.Password); err != nil {
			return err
		}
	} else {
		if err := userService.CreateAdminUser(u.Username, u.Password); err != nil {
			return err
		}
	}
	ctx.GetConfig().Put("core.init", "true")
	return ctx.GetConfig().WriteConfig()
}

// ApplyAdminSkip finishes setup without touching accounts, for when an admin already
// exists and their credentials are known.
func ApplyAdminInitSkip(ctx *core.Context) error {
	userService := wf.GetService[*service.UserService](ctx)
	hasAdmin, err := userService.HasAdminUser()
	if err != nil {
		return err
	}
	if !hasAdmin {
		return errors.New("no admin user exists, cannot skip")
	}
	ctx.GetConfig().Put("core.init", "true")
	return ctx.GetConfig().WriteConfig()
}

// ApplyReSet reconfigures a running instance the way the settings page does: database,
// ports, then restart-or-not decided by the caller.
func ApplyReSet(ctx *core.Context, setInfo *model.Config) error {
	log.Debug("applyReSet", zap.Any("setInfo", setInfo))
	putDbSettings(ctx, setInfo)
	putPortSettings(ctx, setInfo)
	ctx.GetConfig().Put("core.init", "true")
	return attachDb(ctx)
}

// openProbeDB opens a throwaway connection from the supplied settings.
// gorm.Open pings by default, so wrong credentials surface here as an error.
func openProbeDB(setInfo *model.Config) (*framedb.DB, error) {
	switch setInfo.Core.DbType {
	case "mysql":
		m := setInfo.Mysql
		return framedb.ConnectionMysql(m.Host, m.Port, m.Username, m.Password, m.Dbname, m.Charset)
	case "sqlite":
		return framedb.ConnectionSQLite(setInfo.Sqlite.Filename)
	}
	return nil, nil
}

// ProbeDb tests the database settings without persisting them, so a typo in the form
// is caught before it is saved.
func ProbeDb(setInfo *model.Config) error {
	probe, err := openProbeDB(setInfo)
	if err != nil {
		return err
	}
	if probe == nil {
		return errors.New("unsupported dbType: " + setInfo.Core.DbType)
	}
	if sqlDB, err := probe.GetGorm().DB(); err == nil {
		defer func() { _ = sqlDB.Close() }()
	}
	return nil
}
