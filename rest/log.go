package rest

import (
	"strconv"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/model"
	"go.uber.org/zap"
)

type Log struct {
	context  *core.Context
	logModel *model.LogModel
}

func (l *Log) getOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, l.context)
	if user == nil {
		return nil, err
	}
	one, err := l.logModel.Query().Where("id = ? AND user_id = ?", uint(atoi), user.Id).One()
	if err != nil {
		return nil, err
	}
	return one, nil
}

func (l *Log) getPage(req *web.Request) (any, error) {
	page, err := req.Page()
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, l.context)
	if user == nil {
		return nil, err
	}
	searchKey := req.GetFormParam("searchKey")
	query := l.logModel.Query().Where("user_id = ?", user.Id)
	if searchKey != "" {
		query = query.Where("(name LIKE ? OR mail LIKE ? OR subject LIKE ?)", "%"+searchKey+"%", "%"+searchKey+"%", "%"+searchKey+"%")
	}
	return query.Order("id desc").PageForWeb(page)
}
func (l *Log) downLoad(req *web.Request) (any, error) {
	rFilePath := req.GetFormParam("file")
	log.Info("downLoad", zap.String("filePath", rFilePath))
	return &web.File{Path: rFilePath}, nil
}

func (l *Log) Init(context *core.Context) error {
	l.context = context
	l.logModel = wf.GetModel[*model.LogModel](context)
	context.Get("/log/:id", l.getOne).WithMeta(auth2.WithLogin())
	context.Get("/log", l.getPage).WithMeta(auth2.WithLogin())
	context.Get("/download", l.downLoad).WithMeta(auth2.WithLogin())
	return nil

}
