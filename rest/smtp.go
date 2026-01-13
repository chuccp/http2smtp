package rest

import (
	"net/mail"
	"strconv"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
	stmp2 "github.com/chuccp/http2smtp/smtp"
)

type Smtp struct {
	context   *core.Context
	schedule  *service.Schedule
	smtpModel *model.SMTPModel
}

func (smtp *Smtp) getOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	one, err := smtp.smtpModel.FindById(uint(atoi))
	if err != nil {
		return nil, err
	}
	return one, nil

}
func (smtp *Smtp) deleteOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	err = smtp.smtpModel.DeleteById(uint(atoi))
	if err != nil {
		return nil, err
	}
	return "ok", nil

}
func (smtp *Smtp) postOne(req *web.Request) (any, error) {
	var st model.SMTP
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}

	_, err = mail.ParseAddress(st.Mail)
	if err != nil {
		return nil, err
	}

	err = smtp.smtpModel.Save(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil

}
func (smtp *Smtp) putOne(req *web.Request) (any, error) {
	var st model.SMTP
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	_, err = mail.ParseAddress(st.Mail)
	if err != nil {
		return nil, err
	}
	err = smtp.smtpModel.UpdateById(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil

}
func (smtp *Smtp) test(req *web.Request) (any, error) {
	var st model.SMTP
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	_, err = mail.ParseAddress(st.Mail)
	if err != nil {
		return nil, err
	}
	err = stmp2.SendTestMsg2(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil

}
func (smtp *Smtp) getPage(req *web.Request) (any, error) {
	page, err := req.Page()
	if err != nil {
		return nil, err
	}
	p, err := smtp.smtpModel.PageForWeb(page)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (smtp *Smtp) Init(context *core.Context) error {
	smtp.context = context
	context.GetAuth("/smtp/:id", smtp.getOne)
	context.DeleteAuth("/smtp/:id", smtp.deleteOne)
	context.GetAuth("/smtp", smtp.getPage)
	context.PostAuth("/smtp", smtp.postOne)
	context.PostAuth("/test", smtp.test)
	context.PutAuth("/smtp", smtp.putOne)
	return nil
}
