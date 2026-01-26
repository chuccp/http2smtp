package auth

import (
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
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

func (authentication *Authentication) User(request *web.Request) (any, error) {
	return User(request)
}

func (authentication *Authentication) NewUser() any {
	return &entity.LoginUser{}
}

func User(request *web.Request) (*entity.LoginUser, error) {
	token := request.Cookie().Get(entity.UserToken)
	if token == "" {
		return nil, auth2.NoLogin
	}
	user := &entity.LoginUser{}
	err := Decrypt(token, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
