package core

import (
	"github.com/chuccp/http2smtp/web"
	"go.uber.org/zap"
)

type Server interface {
	Init(context *Context)
	Name() string
	Start()
}
type IHttpServer interface {
	init(context *Context)
	start() error
	Stop()
	useCorePort() bool
	GET(relativePath string, handlers ...web.HandlerFunc)
	POST(relativePath string, handlers ...web.HandlerFunc)
	StaticHandle(relativePath string, filepath string)
	SignIn(relativePath string)
	Logout(relativePath string)
	DELETE(relativePath string, handlers ...web.HandlerFunc)
	PUT(relativePath string, handlers ...web.HandlerFunc)

	Any(relativePath string, handlers ...web.HandlerFunc)

	GETAuth(relativePath string, handlers ...web.HandlerFunc)
	POSTAuth(relativePath string, handlers ...web.HandlerFunc)
	DELETEAuth(relativePath string, handlers ...web.HandlerFunc)
	PUTAuth(relativePath string, handlers ...web.HandlerFunc)
	AnyAuth(relativePath string, handlers ...web.HandlerFunc)
}
type httpServer struct {
	context    *Context
	port       int
	usePort    int
	httpServer *web.HttpServer
	certFile   string
	keyFile    string
	name       string
}

func NewHttpServer(name string) IHttpServer {
	return &httpServer{name: name}
}
func (server *httpServer) PUT(pattern string, handlers ...web.HandlerFunc) {
	if server.port > 0 {
		server.httpServer.PUT(pattern, handlers...)
	} else {
		server.context.put(pattern, handlers...)
	}
}
func (server *httpServer) readStatic(req *web.Request) (any, error) {

	return nil, nil
}
func (server *httpServer) StaticHandle(relativePath string, filepath string) {
	if server.port > 0 {
		server.httpServer.StaticHandle(relativePath, filepath)
	} else {
		server.context.staticHandle(relativePath, filepath)
	}
}
func (server *httpServer) POST(pattern string, handlers ...web.HandlerFunc) {
	if server.port > 0 {
		server.httpServer.POST(pattern, handlers...)
	} else {
		server.context.post(pattern, handlers...)
	}
}
func (server *httpServer) sigIn(req *web.Request) (any, error) {
	return req.GetDigestAuth().CheckSign(req.GetContext())
}

func (server *httpServer) logout(req *web.Request) (any, error) {
	return req.GetDigestAuth().Logout(req.GetContext())
}

func (server *httpServer) SignIn(relativePath string) {
	if server.port > 0 {
		server.httpServer.Any(relativePath, server.sigIn)
	} else {
		server.context.any(relativePath, server.sigIn)
	}
}
func (server *httpServer) Logout(relativePath string) {
	if server.port > 0 {
		server.httpServer.Any(relativePath, server.logout)
	} else {
		server.context.any(relativePath, server.logout)
	}
}

func (server *httpServer) DELETE(pattern string, handlers ...web.HandlerFunc) {
	if server.port > 0 {
		server.httpServer.DELETE(pattern, handlers...)
	} else {
		server.context.delete(pattern, handlers...)
	}
}
func (server *httpServer) GET(pattern string, handlers ...web.HandlerFunc) {
	if server.port > 0 {
		server.httpServer.GET(pattern, handlers...)
	} else {
		server.context.get(pattern, handlers...)
	}
}

func (server *httpServer) Any(pattern string, handlers ...web.HandlerFunc) {
	if server.port > 0 {
		server.httpServer.Any(pattern, handlers...)
	} else {
		server.context.any(pattern, handlers...)
	}
}

func (server *httpServer) justChecks(handlers ...web.HandlerFunc) []web.HandlerFunc {
	var hs = make([]web.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		hs[i] = func(req *web.Request) (any, error) {
			check, err := req.GetDigestAuth().JustCheck(req.GetContext())
			if err != nil || check != nil {
				return nil, err
			}
			return handler(req)
		}
	}
	return hs
}
func (server *httpServer) GETAuth(relativePath string, handlers ...web.HandlerFunc) {
	server.GET(relativePath, server.justChecks(handlers...)...)
}
func (server *httpServer) POSTAuth(relativePath string, handlers ...web.HandlerFunc) {
	server.POST(relativePath, server.justChecks(handlers...)...)
}
func (server *httpServer) DELETEAuth(relativePath string, handlers ...web.HandlerFunc) {
	server.DELETE(relativePath, server.justChecks(handlers...)...)
}
func (server *httpServer) PUTAuth(relativePath string, handlers ...web.HandlerFunc) {
	server.PUT(relativePath, server.justChecks(handlers...)...)
}

func (server *httpServer) AnyAuth(relativePath string, handlers ...web.HandlerFunc) {
	server.Any(relativePath, server.justChecks(handlers...)...)
}

func (server *httpServer) IsTls() bool {
	if server.port > 0 {
		return server.httpServer.IsTls()
	} else {
		return server.context.httpServer.IsTls()
	}
}

func (server *httpServer) init(context *Context) {
	server.context = context
	port := context.GetCfgInt(server.name, "port")
	corePort := context.GetCfgInt("manage", "port")
	if port > 0 && corePort != port {
		context.log.Info("服务名称与端口", zap.String("name", server.name), zap.Int("port", port))
		server.certFile = context.GetCfgString(server.name, "certFile")
		server.keyFile = context.GetCfgString(server.name, "keyFile")
		server.port = port
		server.usePort = port
		server.httpServer = web.NewServer(context.GetDigestAuth())
	} else {
		server.usePort = corePort
		context.log.Info("服务名称与端口", zap.String("name", server.name), zap.Int("port", corePort))
	}

}
func (server *httpServer) useCorePort() bool {
	return server.port < 1
}
func (server *httpServer) Stop() {
	server.httpServer.Stop()
}
func (server *httpServer) start() error {
	if server.port > 0 {
		err := server.httpServer.StartAutoTLS(server.port, server.certFile, server.keyFile)
		if err != nil {
			server.context.log.Error("服务启动失败", zap.String("name", server.name), zap.Int("port", server.port), zap.Error(err))
			return err
		}
		return nil
	} else {
		return nil
	}
	return nil
}
