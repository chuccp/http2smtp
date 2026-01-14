package rest

import (
	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/service2"
)

type API struct {
	context      *core.Context
	tokenService *service2.TokenService
}

func (s *API) SendMail(req *web.Request) (any, error) {
	return s.tokenService.SendMailByToken(req)
}

func (s *API) Init(context *core.Context) error {
	s.context = context
	s.tokenService = wf.GetService[*service2.TokenService](context)
	context.Post("/sendMail", s.SendMail)
	context.Get("/sendMail", s.SendMail)
	return nil
}
