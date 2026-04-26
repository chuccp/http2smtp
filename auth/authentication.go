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

func User(request *web.Request, ctx *core.Context) (*model.User, error) {
	token := request.Cookie().Get(entity.UserToken)
	if token == "" {
		return nil, auth2.NoLogin
	}
	loginUser := &entity.LoginUser{}
	err := Decrypt(token, loginUser)
	if err != nil {
		return nil, err
	}
	userModel := core.GetModel[*model.UserModel](ctx)
	var dbUser *model.User
	var dbErr error
	if loginUser.Id != 0 {
		dbUser, dbErr = userModel.FindById(loginUser.Id)
	} else if loginUser.Username != "" {
		dbUser, dbErr = userModel.FindOneByName(loginUser.Username)
	}
	if dbUser != nil && dbErr == nil {
		if dbUser.Salt == loginUser.Salt {
			return dbUser, nil
		}
	}
	return nil, auth2.NoLogin
}
