package rest

import (
	"errors"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/config"
	"github.com/chuccp/http2smtp/model"
	"go.uber.org/zap"
)

type Set struct {
	context *core.Context
}

func (set *Set) putSet(req *web.Request) (any, error) {
	//if set.context.IsInit() {
	//	req.Status(http.StatusMethodNotAllowed)
	//	return nil, errors.New("has init")
	//}
	return set.putReSet(req)
}
func (set *Set) putReSet(req *web.Request) (any, error) {
	var setInfo config.SetInfo
	setInfo.HasInit = true
	err := req.BindJSON(&setInfo)
	if err != nil {
		return nil, err
	}

	if len(setInfo.Manage.Username) == 0 || len(setInfo.Manage.Password) == 0 {
		return nil, errors.New("username or password is blank")
	}
	log.Debug("putSet", zap.Any("setInfo", &setInfo))
	//err := set.context.UpdateSetInfo(&setInfo)
	//if err != nil {
	//	return nil, err
	//}
	return "ok", nil
}

func (set *Set) getSet(req *web.Request) (any, error) {
	u, err := req.User()
	init := set.context.GetConfig().GetBoolOrDefault("core.init", false)
	isDocker := set.context.GetConfig().GetBoolOrDefault("core.isDocker", false)
	return &config.System{HasInit: init, HasLogin: err != nil && u != nil, IsDocker: isDocker}, nil

}
func (set *Set) defaultSet(req *web.Request) (any, error) {
	var cfg = model.DefaultConfig()
	err := set.context.GetConfig().Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (set *Set) testConnection(req *web.Request) (any, error) {
	//if set.context.IsInit() {
	//	req.Status(http.StatusMethodNotAllowed)
	//	return nil, errors.New("has init")
	//}
	//var setInfo config.SetInfo
	//err := req.ShouldBindBodyWithJSON(&setInfo)
	//if err != nil {
	//	return nil, err
	//} else {
	//	mysql := setInfo.Mysql
	//	_, err := db.CreateMysqlConnection(mysql.Username, mysql.Password, mysql.Host, mysql.Port, mysql.Dbname, "UTF8")
	//	if err != nil {
	//		return nil, err
	//	}
	//	return "ok", nil
	//}
	return nil, nil
}
func (set *Set) readSet(req *web.Request) (any, error) {
	//return set.context.GetDefaultSetInfo(), nil
	return nil, nil
}

func (set *Set) reStart(req *web.Request) (any, error) {
	//set.context.ReStart()
	return "ok", nil
}
func (set *Set) Init(context *core.Context) error {
	set.context = context
	context.Get("/set", set.getSet)
	context.Get("/defaultSet", set.defaultSet)
	context.Put("/set", set.putSet)
	context.GetAuth("/readSet", set.readSet)
	context.PutAuth("/reSet", set.putReSet)
	context.PostAuth("/reStart", set.reStart)
	context.Post("/testConnection", set.testConnection)
	return nil
}
