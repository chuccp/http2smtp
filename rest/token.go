package rest

import (
	"errors"
	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/entity"
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
	user, err := auth.User(req, token.context)
	if user == nil {
		return nil, err
	}
	return token.tokenService.GetOne(id, user.Id)
}

func (token *Token) deleteOne(req *web.Request) (any, error) {
	id := req.ParamUint("id")
	user, err := auth.User(req, token.context)
	if user == nil {
		return nil, err
	}
	tokenModel := core.GetModel[*model.TokenModel](token.context)
	exist, err := tokenModel.Query().Where("id = ? AND user_id = ?", id, user.Id).One()
	if err != nil {
		return nil, err
	}
	if exist == nil {
		return nil, errors.New("token not found")
	}
	err = tokenModel.DeleteById(id)
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
	user, err := auth.User(req, token.context)
	if user == nil {
		return nil, err
	}
	return token.tokenService.GetPage(page, user.Id)
}

func (token *Token) postOne(req *web.Request) (any, error) {
	var st model.Token
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, token.context)
	if user == nil {
		return nil, err
	}
	st.UserId = user.Id
	st.State = token.resolveState(st.State, user.IsAdmin)
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
	user, err := auth.User(req, token.context)
	if user == nil {
		return nil, err
	}
	st.UserId = user.Id
	st.ReceiveEmailIds = util.DeduplicateIds(st.ReceiveEmailIds)
	// 保留原有 token
	exist, _ := token.tokenModel.FindById(st.Id)
	if exist != nil && exist.Token != "" {
		st.Token = exist.Token
	}
	st.State = token.resolveState(st.State, user.IsAdmin)
	err = token.tokenModel.UpdateById(&st)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}

func (token *Token) resolveState(state uint8, isAdmin bool) uint8 {
	if state == entity.TokenStateInUse {
		return entity.TokenStateInUse
	}
	// 前端传来非 0 值，后端根据角色决定具体状态
	if isAdmin {
		return entity.TokenStateAdminDisabled
	}
	return entity.TokenStateUserDisabled
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
