package auth

import (
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
)

type Authentication struct {
}

func (authentication *Authentication) SignIn(user any, request *web.Request) (any, error) {
	encrypt, err := Encrypt(user)
	if err != nil {
		return nil, err
	}
	request.Cookie().Set(entity.UserToken, encrypt)
	return web.Ok("success"), nil
}
func (authentication *Authentication) SignOut(request *web.Request) (any, error) {
	request.Cookie().Delete(entity.UserToken)
	return web.Ok("success"), nil
}

func (authentication *Authentication) User(request *web.Request, ctx *core.Context) (any, error) {
	return User(request, ctx)
}

func (authentication *Authentication) NewUser() any {
	return &entity.LoginUser{}
}

func User(request *web.Request, ctx *core.Context) (*entity.LoginUser, error) {
	// 检查是否处于 debug 模式
	var isDebug bool
	if ctx.GetConfig() != nil {
		isDebug = ctx.GetConfig().GetBoolOrDefault("core.debug", false)
	}

	if isDebug {
		// debug 模式下返回虚拟用户
		return &entity.LoginUser{
			Username: "debug_user",
		}, nil
	}

	// 非 debug 模式下执行正常的认证逻辑
	token := request.Cookie().Get(entity.UserToken)
	if token == "" {
		return nil, auth2.NoLogin
	}
	user := &entity.LoginUser{}
	err := Decrypt(token, user)
	if err != nil {
		return nil, err
	}
	// 根据用户名获取用户ID
	if user.Id == 0 && user.Username != "" {
		userModel := core.GetModel[*model.UserModel](ctx)
		dbUser, dbErr := userModel.Query().Where("name = ?", user.Username).One()
		if dbErr == nil && dbUser != nil {
			user.Id = dbUser.Id
			user.IsAdmin = dbUser.IsAdmin
		}
	}
	return user, nil
}
