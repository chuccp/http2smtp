package rest

import (
	"errors"

	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/http2smtp/db"

	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/config"
	"github.com/chuccp/http2smtp/model"
	"go.uber.org/zap"
)

type Set struct {
	context *core.Context
}

func (set *Set) putSet(req *web.Request) (any, error) {
	init := set.context.GetConfig().GetBoolOrDefault("core.init", false)
	if init {
		req.Response().WriteStatus(405)
		return nil, errors.New("has init")
	}
	return set.putReSet(req)
}
func (set *Set) putReSet(req *web.Request) (any, error) {
	var setInfo config.SetInfo
	setInfo.HasInit = true
	err := req.BindJSON(&setInfo)
	if err != nil {
		return nil, err
	}

	if len(setInfo.Manage.Username) == 0 || len(setInfo.Manage.Password) == 0 {
		return nil, errors.New("username or password is blank")
	}

	// Save original password (not encrypted)

	log.Debug("putSet", zap.Any("setInfo", &setInfo))

	// Update database configuration
	if setInfo.DbType != "" {
		set.context.GetConfig().Put("core.dbtype", setInfo.DbType)
	}
	if setInfo.Sqlite != nil {
		if setInfo.Sqlite.Filename != "" {
			set.context.GetConfig().Put("sqlite.filename", setInfo.Sqlite.Filename)
		}
	}
	if setInfo.Mysql != nil {
		if setInfo.Mysql.Host != "" {
			set.context.GetConfig().Put("mysql.host", setInfo.Mysql.Host)
		}
		if setInfo.Mysql.Port > 0 {
			set.context.GetConfig().Put("mysql.port", setInfo.Mysql.Port)
		}
		if setInfo.Mysql.Dbname != "" {
			set.context.GetConfig().Put("mysql.dbname", setInfo.Mysql.Dbname)
		}
		if setInfo.Mysql.Username != "" {
			set.context.GetConfig().Put("mysql.username", setInfo.Mysql.Username)
		}
		if setInfo.Mysql.Password != "" {
			set.context.GetConfig().Put("mysql.password", setInfo.Mysql.Password)
		}
		if setInfo.Mysql.Charset != "" {
			set.context.GetConfig().Put("mysql.charset", setInfo.Mysql.Charset)
		}
	}

	// Update manage configuration
	if setInfo.Manage != nil {
		if setInfo.Manage.Port > 0 {
			set.context.GetConfig().Put("manage.port", setInfo.Manage.Port)
		}
		if setInfo.Manage.Username != "" {
			set.context.GetConfig().Put("manage.username", setInfo.Manage.Username)
		}
		if setInfo.Manage.Password != "" {
			set.context.GetConfig().Put("manage.password", setInfo.Manage.Password)
		}
		if setInfo.Manage.WebPath != "" {
			set.context.GetConfig().Put("manage.webPath", setInfo.Manage.WebPath)
		}
	}

	// Update API port
	if setInfo.Api != nil && setInfo.Api.Port > 0 {
		set.context.GetConfig().Put("api.port", setInfo.Api.Port)
	}

	set.context.GetConfig().Put("core.init", "true")

	// Save the config to file
	err = set.context.GetConfig().WriteConfig()
	if err != nil {
		return nil, err
	}

	createDB, err := db.GetDb(set.context.GetConfig())
	if err != nil {
		return nil, err
	}
	err = set.context.DefaultModelGroup().SwitchDB(createDB, set.context)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}

func (set *Set) getSet(req *web.Request) (any, error) {
	var hasLogin bool
	_, err := auth.User(req, set.context)
	if err == nil {
		hasLogin = true
	}

	// Check if core.init exists and is true
	// First check if the key exists at all
	initStr := set.context.GetConfig().GetStringOrDefault("core.init", "")
	// If the value is empty, it means it hasn't been set yet - return false
	if initStr == "" {
		return &config.System{HasInit: false, HasLogin: hasLogin, IsDocker: set.context.GetConfig().GetBoolOrDefault("core.isDocker", false)}, nil
	}
	// Otherwise parse the boolean value
	init := set.context.GetConfig().GetBoolOrDefault("core.init", false)
	isDocker := set.context.GetConfig().GetBoolOrDefault("core.isDocker", false)
	return &config.System{HasInit: init, HasLogin: hasLogin, IsDocker: isDocker}, nil

}
func (set *Set) defaultSet(req *web.Request) (any, error) {
	var cfg = model.DefaultConfig()
	err := set.context.GetConfig().Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (set *Set) testConnection(req *web.Request) (any, error) {
	init := set.context.GetConfig().GetBoolOrDefault("core.init", false)
	if init {
		req.Response().WriteStatus(405)
		return nil, errors.New("has init")
	}

	var setInfo config.SetInfo
	err := req.BindJSON(&setInfo)
	if err != nil {
		return nil, err
	}

	// Test database connection
	if setInfo.DbType == "mysql" && setInfo.Mysql != nil {
		// Implement MySQL connection test
		return "mysql connection test not implemented", nil
	}

	return "ok", nil
}
func (set *Set) readSet(req *web.Request) (any, error) {
	cfg := config.DefaultSetInfo
	err := set.context.GetConfig().Unmarshal(cfg)
	if err != nil {
		return nil, err
	}
	// Don't return password
	cfg.Manage.Password = ""
	return cfg, nil
}

func (set *Set) reStart(req *web.Request) (any, error) {
	// Restart functionality will be implemented in the framework
	return "ok", nil
}
func (set *Set) Init(context *core.Context) error {
	set.context = context
	context.Get("/set", set.getSet)
	context.Get("/defaultSet", set.defaultSet)
	context.Put("/set", set.putSet)
	context.Get("/readSet", set.readSet)
	context.Put("/reSet", set.putReSet).WithMeta(auth2.WithLogin())
	context.Post("/reStart", set.reStart).WithMeta(auth2.WithLogin())
	context.Post("/testConnection", set.testConnection)
	return nil
}
