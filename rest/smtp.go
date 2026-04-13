package rest

import (
	"errors"
	"net/mail"
	"strconv"

	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
	stmp2 "github.com/chuccp/http2smtp/smtp"
	"github.com/chuccp/http2smtp/util"
)

type Smtp struct {
	context     *core.Context
	schedule    *service.ScheduleService
	smtpModel   *model.SMTPModel
	smtpService *service.SmtpService
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

func (smtp *Smtp) sendMail(req *web.Request) (any, error) {
	var sendMail entity.SendMail
	err := req.BindJSON(&sendMail)
	if err != nil {
		return nil, err
	}
	smtpServer, err := smtp.smtpModel.FindById(sendMail.SMTPId)
	if err != nil {
		return nil, err
	}
	if smtpServer == nil {
		return nil, errors.New("SMTP server not found")
	}
	// 解析收件人邮箱
	var recipients []*model.Mail
	for _, recipient := range sendMail.Recipients {
		name, mailStr, err := util.ParseMail(recipient)
		if err == nil {
			recipients = append(recipients, &model.Mail{Name: name, Mail: mailStr})
		}
	}
	if len(recipients) == 0 {
		return nil, errors.New("no valid recipients")
	}
	// 发送邮件
	err = stmp2.SendAllMsg2(smtpServer, recipients, nil, sendMail.Subject, sendMail.Content)
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
	p, err := smtp.smtpService.PageForWeb(page)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (smtp *Smtp) Init(context *core.Context) error {
	smtp.context = context
	smtp.smtpModel = core.GetModel[*model.SMTPModel](context)
	smtp.smtpService = core.GetService[*service.SmtpService](context)
	context.Get("/smtp/:id", smtp.getOne).WithMeta(auth2.WithLogin())
	context.Delete("/smtp/:id", smtp.deleteOne).WithMeta(auth2.WithLogin())
	context.Get("/smtp", smtp.getPage).WithMeta(auth2.WithLogin())
	context.Post("/smtp", smtp.postOne).WithMeta(auth2.WithLogin())
	context.Post("/test", smtp.test).WithMeta(auth2.WithLogin())
	context.Put("/smtp", smtp.putOne).WithMeta(auth2.WithLogin())
	context.Post("/sendMailBySMTP", smtp.sendMail).WithMeta(auth2.WithLogin())
	return nil
}
