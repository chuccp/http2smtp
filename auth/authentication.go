package auth

import (
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
)

type Authentication struct {
}

func (authentication *Authentication) SignIn(user any, request *web.Request) (any, error) {
	request.Cookie().Set(entity.UserToken, Encrypt(user))
	return web.Ok("success"), nil
}
func (authentication *Authentication) SignOut(request *web.Request) (any, error) {
	request.Cookie().Delete(entity.UserToken)
	return web.Ok("success"), nil
}

func (authentication *Authentication) User(request *web.Request) (any, error) {
	user := authentication.NewUser()
	token := request.Cookie().Get(entity.UserToken)
	if token == "" {
		return user, web.NoLogin
	}
	err := Decrypt(token, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (authentication *Authentication) NewUser() any {
	return &entity.LoginUser{}
}

func User(request *web.Request) (*entity.LoginUser, error) {
	return web.User[*entity.LoginUser](request)
}
