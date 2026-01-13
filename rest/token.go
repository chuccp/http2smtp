package rest

import (
	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service2"
	"github.com/chuccp/http2smtp/util"
)

type Token struct {
	context      *core.Context
	tokenService *service2.TokenService
	tokenModel   *model.TokenModel
}

func (token *Token) getOne(req *web.Request) (any, error) {
	id := req.ParamUint("id")
	return token.tokenService.GetOne(id)
}

func (token *Token) deleteOne(req *web.Request) (any, error) {
	id := req.ParamUint("id")
	tokenModel := core.GetModel[*model.TokenModel](token.context)
	err := tokenModel.DeleteById(id)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (token *Token) getPage(req *web.Request) (any, error) {

	page, err := req.Page()
	if err != nil {
		return nil, err
	}
	return token.tokenService.GetPage(page)
}

func (token *Token) postOne(req *web.Request) (any, error) {

	var st model.Token
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	err = token.tokenModel.Save(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (token *Token) putOne(req *web.Request) (any, error) {
	var st model.Token
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	st.ReceiveEmailIds = util.DeduplicateIds(st.ReceiveEmailIds)
	err = token.tokenModel.UpdateById(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}

func (token *Token) sendMail(req *web.Request) (any, error) {
	return token.tokenService.SendMailByToken(req)
}

func (token *Token) Init(context *core.Context) error {
	token.context = context
	token.tokenService = wf.GetService[*service2.TokenService](token.context)
	token.tokenModel = wf.GetModel[*model.TokenModel](token.context)
	token.context.GetAuth("/token/:id", token.getOne)
	token.context.DeleteAuth("/token/:id", token.deleteOne)
	token.context.GetAuth("/token", token.getPage)
	token.context.PostAuth("/token", token.postOne)
	token.context.PutAuth("/token", token.putOne)
	token.context.PostAuth("/sendMailByToken", token.sendMail)
	return nil
}
func (token *Token) Name() string {
	return entity.TokenRest
}
