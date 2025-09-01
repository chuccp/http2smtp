package api

import (
	"github.com/chuccp/http2smtp/core"
	"github.com/chuccp/http2smtp/web"
)

type Server struct {
	context *core.Context
	core.IHttpServer
}

func NewServer() *Server {
	server := &Server{}
	httpServer := core.NewHttpServer(server.Name())
	server.IHttpServer = httpServer
	return server
}
func (s *Server) Name() string {
	return "api"
}
func (s *Server) Start() {}

func (s *Server) SendMail(req *web.Request) (any, error) {
	return s.context.GetTokenService().SendMailByToken(req)
}

func (s *Server) Init(context *core.Context) {
	s.context = context
	s.POST("/sendMail", s.SendMail)
	s.GET("/sendMail", s.SendMail)
}
