package rest

import (
	"net/mail"
	"strconv"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
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
	one, err := m.mailModel.FindById(uint(atoi))
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
	err = m.mailModel.DeleteById(uint(atoi))
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (m *Mail) getPage(req *web.Request) (any, error) {
	page, err := req.Page()
	if err != nil {
		return nil, err
	}
	return m.mailModel.PageForWeb(page)
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
	err = m.mailModel.Save(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil
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

	err = m.mailModel.UpdateById(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (m *Mail) Init(context *core.Context) error {
	m.context = context
	context.GetAuth("/mail/:id", m.getOne)
	context.DeleteAuth("/mail/:id", m.deleteOne)
	context.GetAuth("/mail", m.getPage)
	context.PostAuth("/mail", m.postOne)
	context.PutAuth("/mail", m.putOne)
	m.mailModel = core.GetModel[*model.MailModel](context)
	return nil
}
