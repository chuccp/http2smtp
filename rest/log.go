package rest

import (
	"strconv"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
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
	one, err := l.logModel.FindById(uint(atoi))
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
	searchKey := req.GetFormParam("searchKey")
	return l.logModel.PageBySearch(page, searchKey)
}
func (l *Log) downLoad(req *web.Request) (any, error) {
	rFilePath := req.GetFormParam("file")
	log.Info("downLoad", zap.String("filePath", rFilePath))
	return &web.File{Path: rFilePath}, nil
}

func (l *Log) Init(context *core.Context) error {
	l.logModel = wf.GetModel[*model.LogModel](context)
	context.Get("/log/:id", l.getOne).WithMeta(auth2.WithLogin())
	context.Get("/log", l.getPage).WithMeta(auth2.WithLogin())
	context.Get("/download", l.downLoad).WithMeta(auth2.WithLogin())
	return nil

}
