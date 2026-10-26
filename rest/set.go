package rest

import (
	"errors"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/http2smtp/configapply"
	"github.com/chuccp/http2smtp/entity"

	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
)

type Set struct {
	context *core.Context
}

// putSet is the original one-step init (now blocked if already init). It runs before any
// account exists, so it cannot ask for a session.
func (set *Set) putSet(req *web.Request) (any, error) {
	if model.CoreState(set.context.GetConfig()).Init {
		return nil, web.NewErrorCode(web.CodeMethodNotAllowed, "has init")
	}
	return set.applyReSet(req)
}

// writeWebPathIfSent stores manage.webPath only when the request body actually carried
// one. The caller binds the body onto model.DefaultConfig(), so an omitted field arrives
// as the default "web" — writing that back would reset a custom webPath from config.ini
// every time the wizard or the settings page saves.
func (set *Set) writeWebPathIfSent(req *web.Request) {
	body, err := req.Json()
	if err != nil {
		return
	}
	v := body.LookupByPath("manage.webPath")
	if v == nil || v.IsNull() || v.String() == "" {
		return
	}
	set.context.GetConfig().Put("manage.webpath", v.String())
}

// putDbInit handles Step 1 of setup: save database config and test connection.
func (set *Set) putDbInit(req *web.Request) (any, error) {
	if model.CoreState(set.context.GetConfig()).Init {
		return nil, web.NewErrorCode(web.CodeMethodNotAllowed, "has init")
	}

	setInfo := model.DefaultConfig()
	err := req.BindJSON(&setInfo)
	if err != nil {
		return nil, err
	}
	set.writeWebPathIfSent(req)
	return web.Ok("ok"), configapply.ApplyDbInit(set.context, setInfo)
}

// putAdminInit handles Step 2 of setup: create or reset admin account.
func (set *Set) putAdminInit(req *web.Request) (any, error) {
	if model.CoreState(set.context.GetConfig()).Init {
		return nil, web.NewErrorCode(web.CodeMethodNotAllowed, "has init")
	}

	if !model.CoreState(set.context.GetConfig()).DbInit {
		return nil, web.NewErrorCode(web.CodeBadRequest, "database not initialized, please complete step 1 first")
	}

	var u entity.LoginUser
	err := req.BindJSON(&u)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), configapply.ApplyAdminInit(set.context, &u)
}

// putAdminSkip skips admin creation if an admin already exists.
func (set *Set) putAdminSkip(req *web.Request) (any, error) {
	if model.CoreState(set.context.GetConfig()).Init {
		return nil, web.NewErrorCode(web.CodeMethodNotAllowed, "has init")
	}

	if !model.CoreState(set.context.GetConfig()).DbInit {
		return nil, web.NewErrorCode(web.CodeBadRequest, "database not initialized, please complete step 1 first")
	}
	return web.Ok("ok"), configapply.ApplyAdminInitSkip(set.context)
}

// postRestart restarts the process so settings that only apply at startup take effect.
// It saves nothing itself — the settings page persists through /reSet first — and answers
// before the listeners close, so the caller sees the ports it will have to reconnect on.
func (set *Set) postRestart(req *web.Request) (any, error) {
	user, err := auth.User(req, set.context)
	if user == nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New("admin access required")
	}
	managePort, apiPort, err := wf.GetService[*service.RestartService](set.context).Restart()
	if err != nil {
		return nil, err
	}
	return web.Data(map[string]any{"managePort": managePort, "apiPort": apiPort}), nil
}

// requireAdminOnceInitialised guards the two setup endpoints the wizard calls before an
// account exists. Until core.init is set they must stay open; afterwards the same
// information is nobody else's business — /adminExists hands out the administrator's
// username and /readSet the database settings.
func (set *Set) requireAdminOnceInitialised(req *web.Request) error {
	if !model.CoreState(set.context.GetConfig()).Init {
		return nil
	}
	user, err := auth.User(req, set.context)
	if user == nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New("admin access required")
	}
	return nil
}

// getAdminExists checks if an admin user already exists in the database.
func (set *Set) getAdminExists(req *web.Request) (any, error) {
	if err := set.requireAdminOnceInitialised(req); err != nil {
		return nil, err
	}
	userService := wf.GetService[*service.UserService](set.context)
	hasAdmin, err := userService.HasAdminUser()
	if err != nil {
		return nil, err
	}
	adminName := ""
	if hasAdmin {
		adminName, _ = userService.GetAdminUsername()
	}
	return map[string]any{"hasAdmin": hasAdmin, "adminName": adminName}, nil
}

func (set *Set) getSet(req *web.Request) (any, error) {
	v, err := auth.User(req, set.context)
	hasLogin := false
	username := ""
	isAdmin := false
	if err == nil && v != nil {
		hasLogin = true
		username = v.Name
		isAdmin = v.IsAdmin
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

	return &model.System{
		HasInit:   cfg.Core.Init,
		HasDbInit: cfg.Core.DbInit,
		HasAdmin:  hasAdmin,
		HasLogin:  hasLogin,
		IsDocker:  cfg.Core.IsDocker,
		Username:  username,
		IsAdmin:   isAdmin,
	}, nil
}
func (set *Set) defaultSet(req *web.Request) (any, error) {
	cfg := model.DefaultConfig()
	err := set.context.GetConfig().Unmarshal(cfg)
	if err != nil {
		return nil, err
	}
	return cfg.MaskSecrets(), nil
}

// testConnection tries the database settings supplied in the request body without
// persisting them, so a typo in the form is caught before it is saved.
// Before setup completes it is reachable without a session (the wizard needs it);
// afterwards it is a reconfiguration helper and requires login.
func (set *Set) testConnection(req *web.Request) (any, error) {
	if model.CoreState(set.context.GetConfig()).Init {
		user, err := auth.User(req, set.context)
		if user == nil {
			return nil, err
		}
	}
	setInfo := model.DefaultConfig()
	if err := req.BindJSON(&setInfo); err != nil {
		return nil, err
	}
	return "ok", configapply.ProbeDb(setInfo)
}
func (set *Set) readSet(req *web.Request) (any, error) {
	if err := set.requireAdminOnceInitialised(req); err != nil {
		return nil, err
	}
	cfg, err := model.GetConfig(set.context.GetConfig())
	if err != nil {
		return nil, err
	}
	return cfg.MaskSecrets(), nil
}

func (set *Set) Init(context *core.Context) error {
	set.context = context
	context.Get("/api/set", set.getSet)
	context.Get("/api/defaultSet", set.defaultSet)
	context.Put("/api/set", set.putSet)
	context.Put("/api/dbInit", set.putDbInit)
	context.Put("/api/adminInit", set.putAdminInit)
	context.Put("/api/adminSkip", set.putAdminSkip)
	context.Get("/api/adminExists", set.getAdminExists)
	context.Get("/api/readSet", set.readSet)
	context.Put("/api/reSet", set.putReSet).WithMeta(auth2.WithLogin())
	context.Post("/api/restart", set.postRestart).WithMeta(auth2.WithLogin())
	context.Post("/api/testConnection", set.testConnection)
	return nil
}

// putReSet reconfigures a running instance: database, web path and ports. It swaps the
// database connection under the running models and rewrites config.ini, so it is admin
// only — the settings page is the only caller and the UI already hides it.
func (set *Set) putReSet(req *web.Request) (any, error) {
	user, err := auth.User(req, set.context)
	if user == nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New("admin access required")
	}
	return set.applyReSet(req)
}

// applyReSet is the shared body of putSet (the legacy one-step init, before any account
// exists) and putReSet (the authenticated reconfiguration).
func (set *Set) applyReSet(req *web.Request) (any, error) {
	setInfo := model.DefaultConfig()
	setInfo.Core.Init = true
	err := req.BindJSON(&setInfo)
	if err != nil {
		return nil, err
	}
	set.writeWebPathIfSent(req)
	return web.Ok("ok"), configapply.ApplyReSet(set.context, setInfo)
}
