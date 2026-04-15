package rest

import (
	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
	"github.com/chuccp/http2smtp/util"
)

type Token struct {
	context      *core.Context
	tokenService *service.TokenService
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
	return web.Ok("ok"), nil
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
	return web.Ok("ok"), nil
}
func (token *Token) putOne(req *web.Request) (any, error) {
	var st model.Token
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	st.ReceiveEmailIds = util.DeduplicateIds(st.ReceiveEmailIds)
	// 保留原有 token
	exist, _ := token.tokenModel.FindById(st.Id)
	if exist != nil && exist.Token != "" {
		st.Token = exist.Token
	}
	err = token.tokenModel.UpdateById(&st)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}

func (token *Token) sendMail(req *web.Request) (any, error) {
	return token.tokenService.SendMailByToken(req)
}

func (token *Token) Init(context *core.Context) error {
	token.context = context
	token.tokenService = wf.GetService[*service.TokenService](token.context)
	token.tokenModel = wf.GetModel[*model.TokenModel](token.context)
	token.context.Get("/token/:id", token.getOne).WithMeta(auth2.WithLogin())
	token.context.Delete("/token/:id", token.deleteOne).WithMeta(auth2.WithLogin())
	token.context.Get("/token", token.getPage).WithMeta(auth2.WithLogin())
	token.context.Post("/token", token.postOne).WithMeta(auth2.WithLogin())
	token.context.Put("/token", token.putOne).WithMeta(auth2.WithLogin())
	token.context.Post("/sendMailByToken", token.sendMail).WithMeta(auth2.WithLogin())
	return nil
}
