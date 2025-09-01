package web

import (
	"github.com/chuccp/http2smtp/login"
	"github.com/chuccp/http2smtp/util"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

func NewServer(digestAuth *login.DigestAuth) *HttpServer {
	engine := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOriginFunc = func(origin string) bool {
		return true
	}
	config.AllowCredentials = true
	config.AllowHeaders = []string{"Nonce", "Content-Type"} // 允
	engine.Use(cors.New(config))
	return &HttpServer{engine: engine, digestAuth: digestAuth, paths: make(map[string]any)}
}

type HttpServer struct {
	isTls      bool
	engine     *gin.Engine
	digestAuth *login.DigestAuth
	paths      map[string]any
	httpServer *http.Server
}

func (hs *HttpServer) IsTls() bool {
	return hs.isTls
}
func (hs *HttpServer) DELETE(pattern string, handlers ...HandlerFunc) {
	hs.paths[pattern] = true
	hs.engine.DELETE(pattern, ToGinHandlerFuncs(handlers, hs.digestAuth)...)
}
func (hs *HttpServer) PUT(pattern string, handlers ...HandlerFunc) {
	hs.paths[pattern] = true
	hs.engine.PUT(pattern, ToGinHandlerFuncs(handlers, hs.digestAuth)...)
}
func (hs *HttpServer) Any(pattern string, handlers ...HandlerFunc) {
	hs.paths[pattern] = true
	hs.engine.Any(pattern, ToGinHandlerFuncs(handlers, hs.digestAuth)...)
}

func (hs *HttpServer) POST(pattern string, handlers ...HandlerFunc) {
	hs.paths[pattern] = true
	hs.engine.POST(pattern, ToGinHandlerFuncs(handlers, hs.digestAuth)...)
}
func (hs *HttpServer) HasPaths(queryPath string) bool {
	_, ok := hs.paths[queryPath]
	if ok {
		return ok
	}
	for k, _ := range hs.paths {
		h := util.IsMatchPath(queryPath, k)
		if h {
			return h
		}
	}
	return ok
}

func (hs *HttpServer) StaticHandle(relativePath string, filepath string) {
	hs.engine.Use(func(context *gin.Context) {
		path_ := context.Request.URL.Path
		if hs.HasPaths(path_) || context.Request.Method != "GET" {
			context.Next()
		} else {
			if strings.Contains(path_, "/manifest.json") {
				filePath := path.Join(filepath, "/manifest.json")
				context.File(filePath)
				context.Abort()
			} else {
				relativeFilePath := ""
				if path_ == relativePath {
					relativeFilePath = relativePath + "index.html"
				} else {
					relativeFilePath = path_
				}
				filePath := path.Join(filepath, relativeFilePath)
				context.File(filePath)
				context.Abort()
			}
		}
	})
}

func (hs *HttpServer) GET(pattern string, handlers ...HandlerFunc) {
	hs.paths[pattern] = true
	hs.engine.GET(pattern, ToGinHandlerFuncs(handlers, hs.digestAuth)...)
}
func (hs *HttpServer) SignIn(pattern string, handlers ...HandlerFunc) {
	hs.paths[pattern] = true
	hs.engine.GET(pattern, ToGinHandlerFuncs(handlers, hs.digestAuth)...)
}

const MaxHeaderBytes = 8192

const MaxReadHeaderTimeout = time.Second * 30

const MaxReadTimeout = time.Minute * 10

func (hs *HttpServer) Start(port int) error {
	hs.httpServer = &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           hs.engine,
		ReadHeaderTimeout: MaxReadHeaderTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ReadTimeout:       MaxReadTimeout,
	}
	hs.isTls = false
	error := hs.httpServer.ListenAndServe()
	return error
}
func (hs *HttpServer) StartTLS(port int, certFile, keyFile string) error {
	hs.httpServer = &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           hs.engine,
		ReadHeaderTimeout: MaxReadHeaderTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ReadTimeout:       MaxReadTimeout,
	}
	hs.isTls = true
	return hs.httpServer.ListenAndServeTLS(certFile, keyFile)
}

func (hs *HttpServer) Stop() {
	if hs.httpServer != nil {
		hs.httpServer.Close()
	}
}
func (hs *HttpServer) StartAutoTLS(port int, certFile, keyFile string) error {
	if len(certFile) > 0 && len(keyFile) > 0 {
		return hs.StartTLS(port, certFile, keyFile)
	} else {
		return hs.Start(port)
	}
}
