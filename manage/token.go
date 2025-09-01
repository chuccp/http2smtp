package manage

import (
	"github.com/chuccp/http2smtp/core"
	"github.com/chuccp/http2smtp/db"
	"github.com/chuccp/http2smtp/util"
	"github.com/chuccp/http2smtp/web"
	"strconv"
)

type Token struct {
	context *core.Context
}

func (token *Token) getOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	return token.context.GetTokenService().GetOne(atoi)
}

func (token *Token) deleteOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	err = token.context.GetDb().GetTokenModel().DeleteOne(uint(atoi))
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (token *Token) getPage(req *web.Request) (any, error) {
	page := req.GetPage()
	return token.context.GetTokenService().GetPage(page)
}

func (token *Token) postOne(req *web.Request) (any, error) {
	var st db.Token
	err := req.ShouldBindBodyWithJSON(&st)
	if err != nil {
		return nil, err
	}
	err = token.context.GetDb().GetTokenModel().Save(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (token *Token) putOne(req *web.Request) (any, error) {
	var st db.Token
	err := req.ShouldBindBodyWithJSON(&st)
	if err != nil {
		return nil, err
	}
	st.ReceiveEmailIds = util.DeduplicateIds(st.ReceiveEmailIds)
	err = token.context.GetDb().GetTokenModel().Edit(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}

func (token *Token) sendMail(req *web.Request) (any, error) {
	return token.context.GetTokenService().SendMailByToken(req)
}

func (token *Token) Init(context *core.Context, server core.IHttpServer) {
	token.context = context
	server.GETAuth("/token/:id", token.getOne)
	server.DELETEAuth("/token/:id", token.deleteOne)
	server.GETAuth("/token", token.getPage)
	server.POSTAuth("/token", token.postOne)
	server.PUTAuth("/token", token.putOne)
	server.POSTAuth("/sendMailByToken", token.sendMail)

}
