package manage

import (
	"github.com/chuccp/http2smtp/core"
	"github.com/chuccp/http2smtp/web"
	"go.uber.org/zap"
	"strconv"
)

type Log struct {
	context *core.Context
}

func (log *Log) getOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	one, err := log.context.GetDb().GetLogModel().GetOne(uint(atoi))
	if err != nil {
		return nil, err
	}
	return one, nil
}

func (log *Log) getPage(req *web.Request) (any, error) {
	page := req.GetPage()
	searchKey := req.GetSearchKey()
	p, err := log.context.GetDb().GetLogModel().PageBySearch(page, searchKey)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (log *Log) downLoad(req *web.Request) (any, error) {
	rFilePath := req.FormValue("file")
	log.context.GetLog().Info("downLoad", zap.String("filePath", rFilePath))
	return &web.File{Path: rFilePath}, nil
}
func (log *Log) Init(context *core.Context, server core.IHttpServer) {
	log.context = context
	server.GETAuth("/log/:id", log.getOne)
	server.GETAuth("/log", log.getPage)
	server.GETAuth("/download", log.downLoad)
}
