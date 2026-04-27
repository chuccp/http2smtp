package rest

import (
	"errors"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/http2smtp/db"

	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
	"go.uber.org/zap"
)

type Set struct {
	context *core.Context
}

// putSet is the original one-step init (now blocked if already init).
func (set *Set) putSet(req *web.Request) (any, error) {
	init := set.context.GetConfig().GetBoolOrDefault("core.init", false)
	if init {
		req.Response().WriteStatus(405)
		return nil, errors.New("has init")
	}
	return set.putReSet(req)
}

// putDbInit handles Step 1 of setup: save database config and test connection.
func (set *Set) putDbInit(req *web.Request) (any, error) {
	init := set.context.GetConfig().GetBoolOrDefault("core.init", false)
	if init {
		req.Response().WriteStatus(405)
		return nil, errors.New("has init")
	}

	setInfo := model.DefaultConfig()
	err := req.BindJSON(&setInfo)
	if err != nil {
		return nil, err
	}

	// Save database configuration
	if setInfo.Core.DbType != "" {
		set.context.GetConfig().Put("core.dbtype", setInfo.Core.DbType)
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

	// Save manage port config
	if setInfo.Manage != nil {
		if setInfo.Manage.Port > 0 {
			set.context.GetConfig().Put("manage.port", setInfo.Manage.Port)
		}
		if setInfo.Manage.WebPath != "" {
			set.context.GetConfig().Put("manage.webPath", setInfo.Manage.WebPath)
		}
	}

	// Save API port
	if setInfo.Api != nil && setInfo.Api.Port > 0 {
		set.context.GetConfig().Put("api.port", setInfo.Api.Port)
	}

	// Mark database as initialized
	set.context.GetConfig().Put("core.dbinit", "true")

	// Create the database connection and switch models to it
	createDB, err := db.GetDb(set.context.GetConfig())
	if err != nil {
		return nil, err
	}
	err = set.context.DefaultModelGroup().SwitchDB(createDB, set.context)
	if err != nil {
		return nil, err
	}

	// Save config to file
	err = set.context.GetConfig().WriteConfig()
	if err != nil {
		return nil, err
	}

	return web.Ok("ok"), nil
}

// putAdminInit handles Step 2 of setup: create admin account.
func (set *Set) putAdminInit(req *web.Request) (any, error) {
	init := set.context.GetConfig().GetBoolOrDefault("core.init", false)
	if init {
		req.Response().WriteStatus(405)
		return nil, errors.New("has init")
	}

	dbInit := set.context.GetConfig().GetBoolOrDefault("core.dbinit", false)
	if !dbInit {
		req.Response().WriteStatus(400)
		return nil, errors.New("database not initialized, please complete step 1 first")
	}

	setInfo := model.DefaultConfig()
	err := req.BindJSON(&setInfo)
	if err != nil {
		return nil, err
	}

	if setInfo.Manage == nil || len(setInfo.Manage.Username) == 0 || len(setInfo.Manage.Password) == 0 {
		return nil, errors.New("username or password is blank")
	}

	log.Debug("putAdminInit", zap.Any("setInfo", &setInfo))

	// Create admin user
	userService := wf.GetService[*service.UserService](set.context)
	if err := userService.CreateAdminUser(setInfo.Manage.Username, setInfo.Manage.Password); err != nil {
		return nil, err
	}

	// Mark system as fully initialized
	set.context.GetConfig().Put("core.init", "true")
	err = set.context.GetConfig().WriteConfig()
	if err != nil {
		return nil, err
	}

	return web.Ok("ok"), nil
}

// putAdminSkip skips admin creation if an admin already exists.
func (set *Set) putAdminSkip(req *web.Request) (any, error) {
	init := set.context.GetConfig().GetBoolOrDefault("core.init", false)
	if init {
		req.Response().WriteStatus(405)
		return nil, errors.New("has init")
	}

	dbInit := set.context.GetConfig().GetBoolOrDefault("core.dbinit", false)
	if !dbInit {
		req.Response().WriteStatus(400)
		return nil, errors.New("database not initialized, please complete step 1 first")
	}

	// Check if an admin user exists
	userService := wf.GetService[*service.UserService](set.context)
	hasAdmin, err := userService.HasAdminUser()
	if err != nil {
		return nil, err
	}
	if !hasAdmin {
		return nil, errors.New("no admin user exists, cannot skip")
	}

	// Mark system as fully initialized
	set.context.GetConfig().Put("core.init", "true")
	err = set.context.GetConfig().WriteConfig()
	if err != nil {
		return nil, err
	}

	return web.Ok("ok"), nil
}

// getAdminExists checks if an admin user already exists in the database.
func (set *Set) getAdminExists(req *web.Request) (any, error) {
	userService := wf.GetService[*service.UserService](set.context)
	hasAdmin, err := userService.HasAdminUser()
	if err != nil {
		return nil, err
	}
	return map[string]bool{"hasAdmin": hasAdmin}, nil
}

func (set *Set) getSet(req *web.Request) (any, error) {
	v, err := auth.User(req, set.context)
	hasLogin := false
	if err == nil && v != nil {
		hasLogin = true
	}
	cfg, err := model.GetConfig(set.context.GetConfig())
	if err != nil {
		return nil, err
	}

	// Check if admin user exists (only meaningful when db is initialized)
	hasAdmin := false
	if cfg.Core.DbInit {
		userService := wf.GetService[*service.UserService](set.context)
		hasAdmin, _ = userService.HasAdminUser()
	}

	return &model.System{HasInit: cfg.Core.Init, HasDbInit: cfg.Core.DbInit, HasAdmin: hasAdmin, HasLogin: hasLogin, IsDocker: cfg.Core.IsDocker}, nil
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
	db, err := db.GetDb(set.context.GetConfig())
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, errors.New("db is nil")
	}
	return "ok", nil
}
func (set *Set) readSet(req *web.Request) (any, error) {
	cfg, err := model.GetConfig(set.context.GetConfig())
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (set *Set) reStart(req *web.Request) (any, error) {
	return "ok", nil
}
func (set *Set) Init(context *core.Context) error {
	set.context = context
	context.Get("/set", set.getSet)
	context.Get("/defaultSet", set.defaultSet)
	context.Put("/set", set.putSet)
	context.Put("/dbInit", set.putDbInit)
	context.Put("/adminInit", set.putAdminInit)
	context.Put("/adminSkip", set.putAdminSkip)
	context.Get("/adminExists", set.getAdminExists)
	context.Get("/readSet", set.readSet)
	context.Put("/reSet", set.putReSet).WithMeta(auth2.WithLogin())
	context.Post("/reStart", set.reStart).WithMeta(auth2.WithLogin())
	context.Post("/testConnection", set.testConnection)
	return nil
}

// putReSet handles re-configuration of the system (requires login).
func (set *Set) putReSet(req *web.Request) (any, error) {
	setInfo := model.DefaultConfig()
	setInfo.Core.Init = true
	err := req.BindJSON(&setInfo)
	if err != nil {
		return nil, err
	}

	if len(setInfo.Manage.Username) == 0 || len(setInfo.Manage.Password) == 0 {
		return nil, errors.New("username or password is blank")
	}

	log.Debug("putSet", zap.Any("setInfo", &setInfo))

	// Update database configuration
	if setInfo.Core.DbType != "" {
		set.context.GetConfig().Put("core.dbtype", setInfo.Core.DbType)
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
		if setInfo.Manage.WebPath != "" {
			set.context.GetConfig().Put("manage.webpath", setInfo.Manage.WebPath)
		}
	}

	// Update API port
	if setInfo.Api != nil && setInfo.Api.Port > 0 {
		set.context.GetConfig().Put("api.port", setInfo.Api.Port)
	}

	set.context.GetConfig().Put("core.init", "true")

	createDB, err := db.GetDb(set.context.GetConfig())
	if err != nil {
		return nil, err
	}
	err = set.context.DefaultModelGroup().SwitchDB(createDB, set.context)
	if err != nil {
		return nil, err
	}

	// 添加管理员用户到数据库
	if setInfo.Manage != nil {
		userService := wf.GetService[*service.UserService](set.context)
		if err := userService.CreateAdminUser(setInfo.Manage.Username, setInfo.Manage.Password); err != nil {
			return nil, err
		}
	}
	// Save the config to file
	err = set.context.GetConfig().WriteConfig()
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}
