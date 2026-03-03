package main

import (
	"flag"
	"path"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/rest"
	"github.com/chuccp/http2smtp/service2"
	"go.uber.org/zap"
)

func converter(value any, err error, ctx *web.Request, response web.Response) {
	if err != nil {
		response.WriteStatus(500)
		response.Write([]byte(err.Error()))
		response.Abort()
	} else {
		if value != nil {
			switch t := value.(type) {
			case string:
				response.Write([]byte(t))
			case *web.File:
				if len(t.FileName) == 0 {
					_, filename := path.Split(t.Path)
					t.FileName = filename
				}
				response.FileAttachment(t.Path, t.FileName)
			default:
				response.AbortWithStatusJSON(200, t)
			}
		}
	}
}

func main() {

	var webPort int
	var apiPort int
	var storageRoot string
	flag.IntVar(&webPort, "web_port", 0, "web port")
	flag.IntVar(&apiPort, "api_port", 0, "api port")
	flag.StringVar(&storageRoot, "storage_root", "./", "storage root")
	flag.Parse()
	webPort = util.GetEnvIntOrDefault("web_port", webPort)
	apiPort = util.GetEnvIntOrDefault("api_port", apiPort)
	storageRoot = util.GetEnvOrDefault("storage_root", storageRoot)
	fileConfig, err := config.LoadSingleFileConfig("config.ini")
	if err != nil {
		log.Panic("加载配置文件失败", zap.Error(err))
		return
	}
	if webPort > 0 || apiPort > 0 {
		if apiPort == 0 {
			apiPort = webPort
		}
		if webPort == 0 {
			webPort = apiPort
		}
		fileConfig.Put("core.isDocker", "true")
		fileConfig.Put("manage.port", webPort)
		fileConfig.Put("api.port", apiPort)
	}
	if len(storageRoot) > 0 {
		fileConfig.Put("core.cachePath", storageRoot)
	}

	var cfg = model.DefaultConfig()

	err = fileConfig.Unmarshal(cfg)
	if err != nil {
		log.Panic("加载配置文件失败", zap.Error(err))
		return
	}
	frame := wf.New(fileConfig)
	var serverConfig = web.DefaultServerConfig()
	serverConfig.Port = cfg.Manage.Port
	authFilter := auth2.NewAuthenticationFilter(&auth.Authentication{})
	frame.NewRestGroup(serverConfig).AddRest(
		&rest.Set{},
		&rest.User{},
		&rest.Token{},
		&rest.Mail{},
		&rest.Smtp{},
	).AddFilter(authFilter)
	frame.AddModel(
		&model.MailModel{},
		&model.SMTPModel{},
		&model.TokenModel{},
		&model.ScheduleModel{},
		&model.LogModel{},
	)
	var apiServerConfig = web.DefaultServerConfig()
	apiServerConfig.Port = cfg.Api.Port
	frame.NewRestGroup(apiServerConfig).AddRest(&rest.API{})
	frame.AddService(
		&service2.TokenService{},
		&service2.ScheduleService{},
		&service2.LogService{},
	)
	err = frame.Start()
	if err != nil {
		log.Panic("启动失败", zap.Error(err))
		return
	}

	//smtp2Http := core.Create()
	//smtp2Http.AddServer(manage.NewServer())
	//smtp2Http.AddServer(api.NewServer())
	//smtp2Http.AddServer(schedule.NewServer())
	//smtp2Http.Start(webPort, apiPort, storageRoot)
}
