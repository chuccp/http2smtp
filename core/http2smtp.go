package core

import (
	"errors"
	"github.com/chuccp/http2smtp/config"
	"github.com/chuccp/http2smtp/login"
	"github.com/chuccp/http2smtp/util"
	"github.com/chuccp/http2smtp/web"
	"go.uber.org/zap"
	"log"
	"net/http"
	"path"
	"time"
)

type SMTP2Http struct {
	context     *Context
	httpServer  *web.HttpServer
	log         *zap.Logger
	config      *config.Config
	servers     []Server
	webPort     int
	apiPort     int
	iHttpServer []IHttpServer
	IsDocker    bool
	storageRoot string
}

func Create() *SMTP2Http {
	return &SMTP2Http{webPort: 0, apiPort: 0, servers: make([]Server, 0), config: config.NewConfig(), iHttpServer: make([]IHttpServer, 0)}
}
func (m *SMTP2Http) AddServer(server Server) {
	m.servers = append(m.servers, server)
}

func (m *SMTP2Http) startHttpServer() error {
	port := m.context.GetCfgInt("manage", "port")
	certFile := m.context.GetCfgString("manage", "certFile")
	keyFile := m.context.GetCfgString("manage", "keyFile")
	m.context.log.Info("startHttpServer", zap.String("name", "manage"), zap.Int("port", port))
	err := m.httpServer.StartAutoTLS(port, certFile, keyFile)
	if err != nil {
		m.context.log.Error("服务启动失败", zap.String("name", "http2smtp"), zap.Int("port", port), zap.Error(err))
		return err
	}
	return nil
}

func (m *SMTP2Http) Start(webPort int, apiPort int, storageRoot string) {
	m.storageRoot = storageRoot
	err := util.EnsureDir(m.storageRoot)
	if err != nil {
		log.Panic(err, m.storageRoot)
	}
	m.webPort = webPort
	m.apiPort = apiPort
	if m.webPort > 0 || m.apiPort > 0 {
		m.IsDocker = true
	}
	for {
		m.reStart()
	}
}
func (m *SMTP2Http) ReStart() {
	for _, server := range m.iHttpServer {
		server.Stop()
	}
	time.Sleep(1 * time.Second)
	m.httpServer.Stop()
	time.Sleep(1 * time.Second)
	m.reStart()
}
func (m *SMTP2Http) StartServer(name string) {
	for _, server := range m.servers {
		if server.Name() == name {
			server.Init(m.context)
			server.Start()
		}
	}
}
func (m *SMTP2Http) reStart() {
	m.iHttpServer = make([]IHttpServer, 0)
	err := m.config.Init(m.webPort, m.apiPort, m.storageRoot)
	if err != nil {
		log.Panic(err)
		return
	}
	logPath := m.config.GetStringOrDefault("log", "filename", "run.log")
	logger, err := initLogger(path.Join(m.storageRoot, logPath))
	if err != nil {
		log.Panic(err)
		return
	}
	m.log = logger
	m.context = &Context{log: m.log, config: m.config, reStart: m.ReStart, IsDocker: m.IsDocker, startServer: m.StartServer, storageRoot: m.storageRoot}
	digestAuth := login.NewDigestAuth(m.context.SecretProvider)
	m.context.digestAuth = digestAuth
	m.httpServer = web.NewServer(digestAuth)
	m.context.httpServer = m.httpServer
	isInit := m.config.GetBooleanOrDefault("core", "init", false)
	if isInit {
		err := m.context.initDb()
		if err != nil {
			m.log.Panic("initDb", zap.Error(err))
			return
		}
	}
	for _, server := range m.servers {
		if s, ok := server.(IHttpServer); ok {
			s.init(m.context)
		}
		server.Init(m.context)
		if s, ok := server.(IHttpServer); ok {
			if !s.useCorePort() {
				go func() {
					m.iHttpServer = append(m.iHttpServer, s)
					err := s.start()
					if !errors.Is(err, http.ErrServerClosed) && err != nil {
						log.Panic(err)
					}
				}()
			}
		}
	}
	for _, server := range m.servers {
		server.Start()
	}
	err = m.startHttpServer()
	if !errors.Is(err, http.ErrServerClosed) && err != nil {
		m.log.Panic("Start", zap.Error(err))
		return
	}
}
