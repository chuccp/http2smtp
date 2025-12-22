package main

import (
	"flag"

	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/http2smtp/util"
	"go.uber.org/zap"
)

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
	frame := core.New(fileConfig)
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
