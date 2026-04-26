package rest

import (
	"errors"
	"strconv"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/service"
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

	dbUser, err := l.userService.ValidateLogin(u.Username, u.Password)
	if err != nil {
		return nil, err
	}

	u.Id = dbUser.Id
	u.IsAdmin = dbUser.IsAdmin
	u.Salt = dbUser.Salt
	_, err = l.authenticationFilter.SignIn(&u, request)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"isAdmin": dbUser.IsAdmin,
	}, nil
}

func (l *User) logout(request *web.Request) (any, error) {
	return l.authenticationFilter.SignOut(request)
}

func (l *User) getUserPage(request *web.Request) (any, error) {
	user, err := auth.User(request, l.context)
	if user == nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New("admin access required")
	}
	page, err := request.Page()
	if err != nil {
		return nil, err
	}
	return l.userService.GetPage(page)
}

func (l *User) getUserOne(request *web.Request) (any, error) {
	user, err := auth.User(request, l.context)
	if user == nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New("admin access required")
	}
	id, err := strconv.Atoi(request.Param("id"))
	if err != nil {
		return nil, err
	}
	return l.userService.GetOne(uint(id))
}

func (l *User) createUser(request *web.Request) (any, error) {
	user, err := auth.User(request, l.context)
	if user == nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New("admin access required")
	}
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
	user, err := auth.User(request, l.context)
	if user == nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New("admin access required")
	}
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
	user, err := auth.User(request, l.context)
	if user == nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New("admin access required")
	}
	id, err := strconv.Atoi(request.Param("id"))
	if err != nil {
		return nil, err
	}
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
