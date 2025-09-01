package manage

import (
	"github.com/chuccp/http2smtp/core"
	"github.com/chuccp/http2smtp/db"
	"github.com/chuccp/http2smtp/service"
	stmp2 "github.com/chuccp/http2smtp/smtp"
	"github.com/chuccp/http2smtp/web"
	"net/mail"
	"strconv"
)

type Smtp struct {
	context  *core.Context
	schedule *service.Schedule
}

func (smtp *Smtp) getOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	one, err := smtp.context.GetDb().GetSMTPModel().GetOne(uint(atoi))
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
	err = smtp.context.GetDb().GetSMTPModel().DeleteOne(uint(atoi))
	if err != nil {
		return nil, err
	}
	return "ok", nil

}
func (smtp *Smtp) postOne(req *web.Request) (any, error) {
	var st db.SMTP
	err := req.ShouldBindBodyWithJSON(&st)
	if err != nil {
		return nil, err
	}

	_, err = mail.ParseAddress(st.Mail)
	if err != nil {
		return nil, err
	}

	err = smtp.context.GetDb().GetSMTPModel().Save(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil

}
func (smtp *Smtp) putOne(req *web.Request) (any, error) {
	var st db.SMTP
	err := req.ShouldBindBodyWithJSON(&st)
	if err != nil {
		return nil, err
	}
	_, err = mail.ParseAddress(st.Mail)
	if err != nil {
		return nil, err
	}
	err = smtp.context.GetDb().GetSMTPModel().Edit(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil

}
func (smtp *Smtp) test(req *web.Request) (any, error) {
	var st db.SMTP
	err := req.ShouldBindBodyWithJSON(&st)
	if err != nil {
		return nil, err
	}
	_, err = mail.ParseAddress(st.Mail)
	if err != nil {
		return nil, err
	}
	err = stmp2.SendTestMsg(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil

}
func (smtp *Smtp) getPage(req *web.Request) (any, error) {
	page := req.GetPage()
	p, err := smtp.context.GetDb().GetSMTPModel().Page(page)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (smtp *Smtp) Init(context *core.Context, server core.IHttpServer) {
	smtp.context = context
	server.GETAuth("/smtp/:id", smtp.getOne)
	server.DELETEAuth("/smtp/:id", smtp.deleteOne)
	server.GETAuth("/smtp", smtp.getPage)
	server.POSTAuth("/smtp", smtp.postOne)
	server.POSTAuth("/test", smtp.test)
	server.PUTAuth("/smtp", smtp.putOne)
}
