package rest

import (
	"errors"
	"strconv"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
	"github.com/chuccp/http2smtp/util"
	"go.uber.org/zap"
)

type authFilter = auth2.AuthenticationFilter[*model.User]

type User struct {
	context              *core.Context
	authenticationFilter *authFilter
	userService          *service.UserService
	captchaService       *service.CaptchaService
}

// signInRequest carries the click captcha alongside the credentials. The page login and
// this endpoint share the same challenge service; the browser page is not the only way
// in, so skipping the captcha here would make it decorative.
type signInRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	CaptchaId string `json:"captchaId"`
	CaptchaXs string `json:"captchaXs"`
	CaptchaYs string `json:"captchaYs"`
}

func (l *User) signIn(request *web.Request) (any, error) {
	var req signInRequest
	err := request.BindJSON(&req)
	if err != nil {
		return nil, err
	}
	if !l.captchaService.VerifyForm(req.CaptchaId, req.Username, req.Password, req.CaptchaXs, req.CaptchaYs) {
		return nil, web.NewErrorCode(web.CodeBadRequest, "captcha verification failed")
	}

	dbUser, err := l.userService.ValidateLogin(req.Username, req.Password)
	if err != nil {
		// Failed attempts are not audited: the endpoint is unauthenticated, so recording
		// them would let anyone write rows into the audit table for free.
		return nil, err
	}

	u := &entity.LoginUser{
		Id:       dbUser.Id,
		Username: dbUser.Name,
		IsAdmin:  dbUser.IsAdmin,
		Salt:     dbUser.Salt,
	}
	// The session travels only in the HttpOnly cookie. Echoing the encrypted token in
	// the response body as well would hand it to any script that can read a fetch
	// result, defeating the point of HttpOnly.
	if _, err := l.authenticationFilter.SignIn(u, request); err != nil {
		return nil, err
	}
	recordOp(l.context, request, dbUser.Id, entity.OpLogin, entity.TargetUser, dbUser.Id, dbUser.Name)
	return map[string]any{
		"isAdmin": dbUser.IsAdmin,
	}, nil
}

func (l *User) logout(request *web.Request) (any, error) {
	// Clearing the cookie only stops this browser. Rotating the session salt also makes
	// every token already issued to this user unusable, so a copied token cannot be
	// replayed. A failure here must not block signing out.
	if user, err := auth.User(request, l.context); err == nil && user != nil {
		if err := l.userService.RotateSalt(user.Id); err != nil {
			log.Error("rotate session salt failed", zap.Uint("userId", user.Id), zap.Error(err))
		}
		recordOp(l.context, request, user.Id, entity.OpLogout, entity.TargetUser, user.Id, user.Name)
	}
	return l.authenticationFilter.SignOut(request)
}

// changePassword lets a signed-in user replace their own password; the account page's
// reset is admin-only and asks for no old password, so without this a normal user had no
// way at all to change theirs. The old password is required so a stolen session cannot
// lock the owner out, and the salt rotation that follows ends every session of that user
// — the caller's included, which is why the client signs out afterwards.
func (l *User) changePassword(request *web.Request) (any, error) {
	user, err := auth.User(request, l.context)
	if user == nil {
		return nil, err
	}
	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := request.BindJSON(&req); err != nil {
		return nil, err
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		return nil, errors.New("the current password and the new password are required")
	}
	if err := util.ValidatePassword(req.NewPassword); err != nil {
		return nil, err
	}
	if !util.CheckPasswordHash(req.OldPassword, user.Password) {
		return nil, errors.New("the current password is incorrect")
	}
	if err := l.userService.ChangeOwnPassword(user.Id, req.NewPassword); err != nil {
		return nil, err
	}
	recordOp(l.context, request, user.Id, entity.OpUpdate, entity.TargetUser, user.Id, user.Name)
	return web.Ok("ok"), nil
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
	page = capPage(page)
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
	if err := util.ValidatePassword(req.Password); err != nil {
		return nil, err
	}
	if err := l.userService.CreateUser(req.Name, req.Password, req.IsAdmin, req.IsUse); err != nil {
		return nil, err
	}
	recordOp(l.context, request, user.Id, entity.OpCreate, entity.TargetUser, 0, req.Name)
	return web.Ok("ok"), nil
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
	// A blank password means "keep the current one"
	if req.Password != "" {
		if err := util.ValidatePassword(req.Password); err != nil {
			return nil, err
		}
	}
	if err := l.userService.UpdateUser(req.Id, req.Name, req.Password, req.IsAdmin, req.IsUse); err != nil {
		return nil, err
	}
	recordOp(l.context, request, user.Id, entity.OpUpdate, entity.TargetUser, req.Id, req.Name)
	return web.Ok("ok"), nil
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
	target, _ := l.userService.GetOne(uint(id))
	if err := l.userService.DeleteUser(uint(id)); err != nil {
		return nil, err
	}
	// DeleteUser disables the account (and ends its sessions); the audit entry says so.
	name := ""
	if target != nil {
		name = target.Name
	}
	recordOp(l.context, request, user.Id, entity.OpDisable, entity.TargetUser, uint(id), name)
	return web.Ok("ok"), nil
}

func (l *User) Init(context *core.Context) error {
	l.context = context
	l.authenticationFilter = wf.GetFilter[*authFilter](context)
	l.userService = wf.GetService[*service.UserService](context)
	l.captchaService = wf.GetService[*service.CaptchaService](context)
	context.Post("/api/signIn", l.signIn)
	context.Post("/api/logout", l.logout)
	context.Post("/api/changePassword", l.changePassword).WithMeta(auth2.WithLogin())
	context.Get("/api/user", l.getUserPage).WithMeta(auth2.WithLogin())
	context.Get("/api/user/:id", l.getUserOne).WithMeta(auth2.WithLogin())
	context.Post("/api/user", l.createUser).WithMeta(auth2.WithLogin())
	context.Put("/api/user", l.updateUser).WithMeta(auth2.WithLogin())
	context.Delete("/api/user/:id", l.deleteUser).WithMeta(auth2.WithLogin())
	return nil
}
