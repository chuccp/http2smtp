package rest

import (
	"errors"
	"net/mail"
	"strconv"
	"time"

	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
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
	user, err := auth.User(req, smtp.context)
	if user == nil {
		return nil, err
	}
	one, err := smtp.smtpModel.Query().Where("id = ? AND user_id = ?", uint(atoi), user.Id).One()
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
	user, err := auth.User(req, smtp.context)
	if user == nil {
		return nil, err
	}
	exist, err := smtp.smtpModel.Query().Where("id = ? AND user_id = ?", uint(atoi), user.Id).One()
	if err != nil {
		return nil, err
	}
	if exist == nil {
		return nil, errors.New("smtp not found")
	}
	err = smtp.smtpModel.DeleteByPK(uint(atoi))
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil

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
	user, err := auth.User(req, smtp.context)
	if user == nil {
		return nil, err
	}
	// Create is create: a client-supplied id must not turn this into an UPDATE of an
	// existing row (gorm Save updates when the primary key is set), which would let any
	// signed-in user overwrite another account's SMTP credentials and take ownership.
	st.Id = 0
	st.UserId = user.Id

	err = smtp.smtpModel.Save(&st)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil

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
	user, err := auth.User(req, smtp.context)
	if user == nil {
		return nil, err
	}
	exist, err := smtp.smtpModel.FindByPK(st.Id)
	if err != nil {
		return nil, err
	}
	// Report "not found" rather than "forbidden" so we don't leak existence
	if exist == nil || (exist.UserId != user.Id && !user.IsAdmin) {
		return nil, errors.New("smtp not found")
	}
	// Keep the original owner: an admin editing someone else's row must not take it over
	exist.Host = st.Host
	exist.Port = st.Port
	exist.Mail = st.Mail
	exist.Username = st.Username
	exist.Name = st.Name
	// Keep the stored password when the request omits it, the same rule the token PUT
	// uses: the UI never echoes a password back, so an unconditional write turned a
	// form that simply did not carry one into silently broken deliveries.
	if st.Password != "" {
		exist.Password = st.Password
	}
	exist.UpdateTime = time.Now()
	err = smtp.smtpModel.Save(exist)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil

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
	err = stmp2.SendTestMsg(&st)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil

}

func (smtp *Smtp) sendMail(req *web.Request) (any, error) {
	var sendMail entity.SendMail
	err := req.BindJSON(&sendMail)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, smtp.context)
	if user == nil {
		return nil, err
	}
	smtpServer, err := smtp.smtpModel.Query().Where("id = ? AND user_id = ?", sendMail.SMTPId, user.Id).One()
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
	err = stmp2.SendAllMsg(smtpServer, recipients, nil, sendMail.Subject, sendMail.Content)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}
func (smtp *Smtp) getPage(req *web.Request) (any, error) {
	page, err := req.Page()
	if err != nil {
		return nil, err
	}
	page = capPage(page)
	user, err := auth.User(req, smtp.context)
	if user == nil {
		return nil, err
	}
	adminOnly := req.Query("adminOnly") == "true"
	name := req.Query("name")
	host := req.Query("host")
	username := req.Query("username")
	p, err := smtp.smtpService.PageForWeb(page, user.Id, user.IsAdmin, adminOnly, name, host, username)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (smtp *Smtp) Init(context *core.Context) error {
	smtp.context = context
	smtp.smtpModel = core.GetModel[*model.SMTPModel](context)
	smtp.smtpService = core.GetService[*service.SmtpService](context)
	context.Get("/api/smtp/:id", smtp.getOne).WithMeta(auth2.WithLogin())
	context.Delete("/api/smtp/:id", smtp.deleteOne).WithMeta(auth2.WithLogin())
	context.Get("/api/smtp", smtp.getPage).WithMeta(auth2.WithLogin())
	context.Post("/api/smtp", smtp.postOne).WithMeta(auth2.WithLogin())
	context.Post("/api/test", smtp.test).WithMeta(auth2.WithLogin())
	context.Put("/api/smtp", smtp.putOne).WithMeta(auth2.WithLogin())
	context.Post("/api/sendMailBySMTP", smtp.sendMail).WithMeta(auth2.WithLogin())
	return nil
}
