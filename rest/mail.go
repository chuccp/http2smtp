package rest

import (
	"errors"
	"net/mail"
	"strconv"

	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/model"
)

type Mail struct {
	context   *core.Context
	mailModel *model.MailModel
}

func (m *Mail) getOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, m.context)
	if user == nil {
		return nil, err
	}
	one, err := m.mailModel.Query().Where("id = ? AND user_id = ?", uint(atoi), user.Id).One()
	if err != nil {
		return nil, err
	}
	return one, nil
}

func (m *Mail) deleteOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, m.context)
	if user == nil {
		return nil, err
	}
	exist, err := m.mailModel.Query().Where("id = ? AND user_id = ?", uint(atoi), user.Id).One()
	if err != nil {
		return nil, err
	}
	if exist == nil {
		return nil, errors.New("mail not found")
	}
	err = m.mailModel.DeleteById(uint(atoi))
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}
func (m *Mail) getPage(req *web.Request) (any, error) {
	page, err := req.Page()
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, m.context)
	if user == nil {
		return nil, err
	}
	return m.mailModel.Query().Where("user_id = ?", user.Id).Order("id desc").PageForWeb(page)
}
func (m *Mail) postOne(req *web.Request) (any, error) {
	var st model.Mail
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	_, err = mail.ParseAddress(st.Mail)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, m.context)
	if user == nil {
		return nil, err
	}
	st.UserId = user.Id
	err = m.mailModel.Save(&st)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}
func (m *Mail) putOne(req *web.Request) (any, error) {
	var st model.Mail
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}

	_, err = mail.ParseAddress(st.Mail)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, m.context)
	if user == nil {
		return nil, err
	}
	st.UserId = user.Id

	err = m.mailModel.UpdateById(&st)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}
func (m *Mail) Init(context *core.Context) error {
	m.context = context
	context.Get("/mail/:id", m.getOne).WithMeta(auth2.WithLogin())
	context.Delete("/mail/:id", m.deleteOne).WithMeta(auth2.WithLogin())
	context.Get("/mail", m.getPage).WithMeta(auth2.WithLogin())
	context.Post("/mail", m.postOne).WithMeta(auth2.WithLogin())
	context.Put("/mail", m.putOne).WithMeta(auth2.WithLogin())
	m.mailModel = core.GetModel[*model.MailModel](context)
	return nil
}
