package main

import (
	"flag"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/component/cors"
	"github.com/chuccp/go-web-frame/component/schedule"
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/http2smtp/db"
	"github.com/chuccp/http2smtp/runner"

	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/rest"
	"github.com/chuccp/http2smtp/service"
	"go.uber.org/zap"
)

func createAPP() (*wf.WebFrame, error) {

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
		return nil, err
	}
	init := fileConfig.GetBoolOrDefault("core.init", false)
	if webPort > 0 || apiPort > 0 {
		if apiPort == 0 {
			apiPort = webPort
		}
		if webPort == 0 {
			webPort = apiPort
		}
		fileConfig.Put("core.isdocker", "true")
		fileConfig.Put("manage.port", webPort)
		fileConfig.Put("api.port", apiPort)
	}
	if len(storageRoot) > 0 {
		fileConfig.Put("core.cachepath", storageRoot)
	}
	builder := wf.NewBuilder(fileConfig)
	if webPort == 0 {
		webPort = model.ManagePort
	}
	restGroupBuilder := core.NewRestGroupBuilder()
	restGroupBuilder.Rest(&rest.Set{}, &rest.User{}, &rest.Token{}, &rest.Mail{}, &rest.Smtp{}, &rest.Schedule{}, &rest.Log{})
	restGroupBuilder.Port(webPort)
	restGroupBuilder.ContextPath("/api")
	restGroupBuilder.Filter(auth2.NewAuthenticationFilter(&auth.Authentication{}), cors.NewCrosFilter())
	restGroup := restGroupBuilder.Build()

	builder.RestGroup(restGroup)

	apiRestGroupBuilder := core.NewRestGroupBuilder()
	apiRestGroupBuilder.Rest(&rest.API{})
	if apiPort == 0 {
		apiPort = model.ApiPort
	}
	apiRestGroupBuilder.Port(apiPort)
	builder.RestGroup(apiRestGroupBuilder.Build())

	manageModelGroupBuilder := core.NewModelGroupBuilder()
	manageModelGroupBuilder.AutoCreateTable(true)
	manageModelGroupBuilder.Model(
		&model.MailModel{},
		&model.SMTPModel{},
		&model.TokenModel{},
		&model.ScheduleModel{},
		&model.LogModel{},
		&model.UserModel{},
	)

	if init || fileConfig.GetBoolOrDefault("core.dbinit", false) {
		connection, err := db.GetDb(fileConfig)
		if err != nil {
			return nil, err
		}
		manageModelGroupBuilder.DB(connection)
	}
	builder.Service(&service.TokenService{}, &service.ScheduleService{}, &service.LogService{}, &service.SmtpService{}, &service.UserService{})

	builder.Runner(schedule.NewScheduleWithSeconds(), &runner.ScheduleRunner{})

	builder.ModelGroup(manageModelGroupBuilder.Build())
	return builder.Build(), nil
}
func main() {
	app, err := createAPP()
	if err != nil {
		log.Panic("启动失败", zap.Error(err))
		return
	}
	err = app.Start()
	if err != nil {
		log.Panic("启动失败", zap.Error(err))
		return
	}
}
