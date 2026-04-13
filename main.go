package main

import (
	"flag"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/rest"
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
	builder := wf.NewBuilder(fileConfig)

	restGroupBuilder := core.NewRestGroupBuilder()
	restGroupBuilder.Rest(&rest.Set{}, &rest.User{}, &rest.Token{}, &rest.Mail{}, &rest.Smtp{}, &rest.Schedule{}, &rest.Log{})
	restGroupBuilder.Port(webPort)
	restGroupBuilder.Filter(auth2.NewAuthenticationFilter(&auth.Authentication{}))
	restGroup := restGroupBuilder.Build()

	builder.RestGroup(restGroup)

	apiRestGroupBuilder := core.NewRestGroupBuilder()
	apiRestGroupBuilder.Rest(&rest.API{})
	apiRestGroupBuilder.Port(apiPort)
	builder.RestGroup(apiRestGroupBuilder.Build())

	manageModelGroupBuilder := core.NewModelGroupBuilder()

	manageModelGroupBuilder.Model(
		&model.MailModel{},
		&model.SMTPModel{},
		&model.TokenModel{},
		&model.ScheduleModel{},
		&model.LogModel{},
	)

	dbType := fileConfig.GetString("core.dbtype")
	if dbType == "sqlite" {
		var sqliteConfig = model.DefaultSqliteConfig()
		err := fileConfig.UnmarshalKey("sqlite", &sqliteConfig)
		if err != nil {
			return nil, err
		}
		connection, err := db.ConnectionSQLite(sqliteConfig.Filename)
		if err != nil {
			return nil, err
		}
		manageModelGroupBuilder.DB(connection)
	} else if dbType == "mysql" {
		var mysqlConfig = model.DefaultMysqlConfig()
		err := fileConfig.UnmarshalKey("mysql", &mysqlConfig)
		if err != nil {
			return nil, err
		}
		connection, err := db.ConnectionMysql(mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.Username, mysqlConfig.Password, mysqlConfig.Dbname, mysqlConfig.Charset)
		if err != nil {
			return nil, err
		}
		manageModelGroupBuilder.DB(connection)
	}
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
