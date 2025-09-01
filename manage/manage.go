package manage

import "github.com/chuccp/http2smtp/core"

type manage interface {
	Init(context *core.Context, server core.IHttpServer)
}
type Server struct {
	manages []manage
	core.IHttpServer
	context *core.Context
}

func NewServer() *Server {
	server := &Server{}
	httpServer := core.NewHttpServer(server.Name())
	server.IHttpServer = httpServer
	return server
}

func (s *Server) Start() {}

func (s *Server) Name() string {
	return "manage"
}
func (s *Server) addManage(manage manage) {
	s.manages = append(s.manages, manage)
}
func (s *Server) Init(context *core.Context) {
	s.context = context
	s.manages = make([]manage, 0)
	s.addManage(&Smtp{})
	s.addManage(&Mail{})
	s.addManage(&Set{})
	s.addManage(&Token{})
	s.addManage(&Log{})
	s.addManage(&User{})
	s.addManage(&Schedule{})
	for _, a := range s.manages {
		a.Init(s.context, s)
	}
	webPath := context.GetConfig().GetString("manage", "webPath")
	s.StaticHandle("/", webPath)
}
