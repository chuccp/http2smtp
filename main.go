package main

import (
	"flag"

	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/http2smtp/auth"
	config2 "github.com/chuccp/http2smtp/config"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/rest"
	"github.com/chuccp/http2smtp/service2"
	"go.uber.org/zap"
)

func changeConfig(fileConfig config.IConfig, webPort int, apiPort int, storageRoot string) {
	if webPort > 0 {
		fileConfig.Put("web.server.port", webPort)
	} else {
		fileConfig.ReplaceKey("manage.port", "web.server.port")
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

	changeConfig(fileConfig, webPort, apiPort, storageRoot)

	frame := core.New(fileConfig)
	frame.AddRest(&rest.User{}, &rest.Token{})
	frame.AddModel(&model.MailModel{}, &model.SMTPModel{}, &model.TokenModel{}, &model.ScheduleModel{}, &model.LogModel{})
	frame.RegisterConfig(&config2.Manage{})
	frame.AddService(&service2.TokenService{}, &service2.ScheduleService{}, &service2.LogService{})
	frame.Authentication(&auth.Authentication{})
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
