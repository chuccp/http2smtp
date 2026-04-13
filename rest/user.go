package rest

import (
	"errors"
	"strings"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/config"
	"github.com/chuccp/http2smtp/entity"
	"go.uber.org/zap"
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
func (l *User) Init(context *core.Context) error {
	l.context = context
	l.authenticationFilter = wf.GetFilter[*auth.AuthenticationFilter](context)
	context.Post("/signIn", l.signIn)
	context.Post("/logout", l.logout)
	return nil
}
