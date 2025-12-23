package rest

import (
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service2"
)

type Token struct {
	context *core.Context
}

func (token *Token) getOne(req *web.Request) (any, error) {
	id := req.ParamUint("id")
	tokenService := core.GetService[*service2.TokenService](token.context)
	return tokenService.GetOne(id)
}

func (token *Token) deleteOne(req *web.Request) (any, error) {
	id := req.ParamUint("id")
	tokenModel := core.GetModel[*model.TokenModel](token.context)
	err := tokenModel.DeleteOne(id)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (token *Token) getPage(req *web.Request) (any, error) {
	//page := req.GetPage()
	//return token.context.GetTokenService().GetPage(page)
	return "ok", nil
}

func (token *Token) postOne(req *web.Request) (any, error) {
	//var st db.Token
	//err := req.ShouldBindBodyWithJSON(&st)
	//if err != nil {
	//	return nil, err
	//}
	//err = token.context.GetDb().GetTokenModel().Save(&st)
	//if err != nil {
	//	return nil, err
	//}
	return "ok", nil
}
func (token *Token) putOne(req *web.Request) (any, error) {
	//var st db.Token
	//err := req.ShouldBindBodyWithJSON(&st)
	//if err != nil {
	//	return nil, err
	//}
	//st.ReceiveEmailIds = util.DeduplicateIds(st.ReceiveEmailIds)
	//err = token.context.GetDb().GetTokenModel().Edit(&st)
	//if err != nil {
	//	return nil, err
	//}
	return "ok", nil
}

func (token *Token) sendMail(req *web.Request) (any, error) {
	//return token.context.GetTokenService().SendMailByToken(req)
	return "ok", nil
}

func (token *Token) Init(context *core.Context) {
	token.context = context
	token.context.GetAuth("/token/:id", token.getOne)
	token.context.DeleteAuth("/token/:id", token.deleteOne)
	token.context.GetAuth("/token", token.getPage)
	token.context.PostAuth("/token", token.postOne)
	token.context.PutAuth("/token", token.putOne)
	token.context.PostAuth("/sendMailByToken", token.sendMail)
}
func (token *Token) Name() string {
	return entity.TokenRest
}
