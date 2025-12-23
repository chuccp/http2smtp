package rest

import (
	"strings"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/config"
	"github.com/chuccp/http2smtp/entity"
)

type User struct {
	context *core.Context
}

func (l *User) signIn(request *web.Request) (any, error) {
	var u entity.LoginUser
	err := request.BindJSON(&u)
	if err != nil {
		return nil, err
	}
	manage := core.GetValueConfig[*config.Manage](l.context)
	key := util.MD5Str(util.MD5Str(manage.Password) + manage.Username)
	sign := util.MD5Str(key + u.Nonce)
	if strings.EqualFold(sign, u.Response) {
		return request.SignIn(u)
	}
	return nil, nil
}
func (l *User) logout(request *web.Request) (any, error) {
	return request.SignOut()
}
func (l *User) Init(context *core.Context) {
	l.context = context
	context.Post("/signIn", l.signIn)
	context.Post("/logout", l.logout)
}
func (l *User) Name() string {
	return entity.UserRest
}
