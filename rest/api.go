package rest

import (
	"errors"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
)

type API struct {
	context      *core.Context
	tokenService *service.TokenService
}

func (s *API) SendMail(req *web.Request) (any, error) {
	// The token service caches its models at Init; before the wizard has attached a
	// database they hold a nil connection, and the first send on a fresh instance
	// would panic dereferencing it instead of answering.
	if !model.CoreState(s.context.GetConfig()).DbInit {
		return nil, errors.New("database is not initialized, please complete setup first")
	}
	// nil user: the API port has no session, the token itself is the credential
	return s.tokenService.SendMailByToken(req, nil)
}

func (s *API) Init(context *core.Context) error {
	s.context = context
	s.tokenService = wf.GetService[*service.TokenService](context)
	context.Post("/sendMail", s.SendMail)
	context.Get("/sendMail", s.SendMail)
	return nil
}
