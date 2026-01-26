package rest

import (
	"strings"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/config"
	"github.com/chuccp/http2smtp/entity"
)

type User struct {
	context              *core.Context
	authenticationFilter *auth.AuthenticationFilter
}

func (l *User) signIn(request *web.Request) (any, error) {
	var u entity.LoginUser
	err := request.BindJSON(&u)
	if err != nil {
		return nil, err
	}
	manage := wf.UnmarshalKeyConfig[*config.Manage]("manage", l.context)
	key := util.MD5Str(util.MD5Str(manage.Password) + manage.Username)
	sign := util.MD5Str(key + u.Nonce)
	if strings.EqualFold(sign, u.Response) {
		return l.authenticationFilter.SignIn(&u, request)
	}
	return nil, nil
}
func (l *User) logout(request *web.Request) (any, error) {
	return l.authenticationFilter.SignOut(request)
}
func (l *User) Init(context *core.Context) error {
	l.context = context
	l.authenticationFilter = wf.GetFilter[*auth.AuthenticationFilter](context)
	context.Post("/signIn", l.signIn)
	context.Post("/logout", l.logout)
	return nil
}
