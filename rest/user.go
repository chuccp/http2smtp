package rest

import (
	"errors"
	"strconv"
	"strings"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/config"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
	localutil "github.com/chuccp/http2smtp/util"
	"go.uber.org/zap"
)

type User struct {
	context              *core.Context
	authenticationFilter *auth2.AuthenticationFilter
	userService          *service.UserService
}

func (l *User) signIn(request *web.Request) (any, error) {
	var u entity.LoginUser
	err := request.BindJSON(&u)
	if err != nil {
		return nil, err
	}

	// First try database authentication
	userModel := core.GetModel[*model.UserModel](l.context)
	if userModel != nil {
		dbUser, err := userModel.FindOneByName(u.Username)
		if err == nil && dbUser != nil && dbUser.IsUse {
			// Verify password using bcrypt
			if localutil.CheckPasswordHash(u.Password, dbUser.Password) {
				// Set user ID and admin status before signing in
				u.Id = dbUser.Id
				u.IsAdmin = dbUser.IsAdmin
				log.Debug("signIn via database",
					zap.String("username", u.Username),
					zap.Uint("userId", dbUser.Id),
				)
				return l.authenticationFilter.SignIn(&u, request)
			}
		}
	}

	// Fallback to config file authentication (legacy challenge-response)
	manage, err := wf.UnmarshalKeyConfig[*config.Manage]("manage", l.context)
	if err != nil {
		return nil, err
	}

	log.Debug("signIn",
		zap.String("username_from_request", u.Username),
		zap.String("username_from_config", manage.Username),
		zap.String("nonce", u.Nonce),
		zap.String("response_from_client", u.Response),
		zap.String("password_from_config", manage.Password),
	)

	// The password in config is stored as plain text
	// Calculate: key = MD5(MD5(password) + username)
	key := util.MD5Str(util.MD5Str(manage.Password) + manage.Username)
	sign := util.MD5Str(key + u.Nonce)

	log.Debug("signIn calculation",
		zap.String("calculated_key", key),
		zap.String("calculated_sign", sign),
	)

	if strings.EqualFold(sign, u.Response) {
		log.Debug("signIn success")
		return l.authenticationFilter.SignIn(&u, request)
	}
	log.Debug("signIn failed - invalid credentials")
	return nil, errors.New("invalid credentials")
}

func (l *User) logout(request *web.Request) (any, error) {
	return l.authenticationFilter.SignOut(request)
}

func (l *User) getUserPage(request *web.Request) (any, error) {
	page, err := request.Page()
	if err != nil {
		return nil, err
	}
	return l.userService.GetPage(page)
}

func (l *User) getUserOne(request *web.Request) (any, error) {
	id, err := strconv.Atoi(request.Param("id"))
	if err != nil {
		return nil, err
	}
	return l.userService.GetOne(uint(id))
}

func (l *User) createUser(request *web.Request) (any, error) {
	var req struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"isAdmin"`
		IsUse    bool   `json:"isUse"`
	}
	if err := request.BindJSON(&req); err != nil {
		return nil, err
	}
	if req.Name == "" || req.Password == "" {
		return nil, errors.New("name and password are required")
	}
	return web.Ok("ok"), l.userService.CreateUser(req.Name, req.Password, req.IsAdmin, req.IsUse)
}

func (l *User) updateUser(request *web.Request) (any, error) {
	var req struct {
		Id       uint   `json:"id"`
		Name     string `json:"name"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"isAdmin"`
		IsUse    bool   `json:"isUse"`
	}
	if err := request.BindJSON(&req); err != nil {
		return nil, err
	}
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	return web.Ok("ok"), l.userService.UpdateUser(req.Id, req.Name, req.Password, req.IsAdmin, req.IsUse)
}

func (l *User) deleteUser(request *web.Request) (any, error) {
	id, err := strconv.Atoi(request.Param("id"))
	if err != nil {
		return nil, err
	}
	// Soft delete: set is_use = false instead of hard delete
	return web.Ok("ok"), l.userService.DeleteUser(uint(id))
}

func (l *User) Init(context *core.Context) error {
	l.context = context
	l.authenticationFilter = wf.GetFilter[*auth2.AuthenticationFilter](context)
	l.userService = wf.GetService[*service.UserService](context)
	context.Post("/signIn", l.signIn)
	context.Post("/logout", l.logout)
	context.Get("/user", l.getUserPage).WithMeta(auth2.WithLogin())
	context.Get("/user/:id", l.getUserOne).WithMeta(auth2.WithLogin())
	context.Post("/user", l.createUser).WithMeta(auth2.WithLogin())
	context.Put("/user", l.updateUser).WithMeta(auth2.WithLogin())
	context.Delete("/user/:id", l.deleteUser).WithMeta(auth2.WithLogin())
	return nil
}
