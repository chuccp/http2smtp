package rest

import (
	"errors"
	"net/mail"
	"os"
	"strconv"
	"time"

	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	wfutil "github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
	stmp2 "github.com/chuccp/http2smtp/smtp"
	"github.com/chuccp/http2smtp/util"
	"go.uber.org/zap"
)

type Smtp struct {
	context     *core.Context
	schedule    *service.ScheduleService
	smtpModel   *model.SMTPModel
	smtpService *service.SmtpService
	logService  *service.LogService
	cachePath   string
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
	recordOp(smtp.context, req, user.Id, entity.OpDelete, entity.TargetSMTP, uint(atoi), exist.Name)
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
	recordOp(smtp.context, req, user.Id, entity.OpCreate, entity.TargetSMTP, st.Id, st.Name)
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
	recordOp(smtp.context, req, user.Id, entity.OpUpdate, entity.TargetSMTP, exist.Id, exist.Name)
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
	// Admins may send through any user's server (the list offers them every row); a
	// regular caller is limited to servers they own.
	query := smtp.smtpModel.Query().Where("id = ?", sendMail.SMTPId)
	if !user.IsAdmin {
		query = query.Where("user_id = ?", user.Id)
	}
	smtpServer, err := query.One()
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
	// The attachments are written into the cache before sending. From the moment the log
	// entry records them its cleanup owns them; until then a failure must not leave
	// copies behind.
	files := make([]*stmp2.File, 0)
	logged := false
	defer func() {
		if !logged {
			for _, f := range files {
				if f != nil && f.FilePath != "" {
					_ = os.Remove(f.FilePath)
				}
			}
		}
	}()
	for _, file := range sendMail.Files {
		if file.Data == "" {
			continue
		}
		data, err := wfutil.DecodeFileBase64(file.Data)
		if err != nil {
			return nil, err
		}
		name := file.Name
		if name == "" {
			name, err = wfutil.CalculateMD5(data)
			if err != nil {
				return nil, err
			}
		}
		filePath := wfutil.GetCachePath(smtp.cachePath, name)
		if err := wfutil.WriteFile(data, filePath); err != nil {
			return nil, err
		}
		vFile, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		files = append(files, &stmp2.File{File: vFile, Name: name, FilePath: filePath})
	}
	// go-mail reads these readers but never closes them.
	defer func() {
		for _, f := range files {
			if f != nil && f.File != nil {
				_ = f.File.Close()
			}
		}
	}()
	// 发送邮件
	err = stmp2.SendAllMsg(smtpServer, recipients, files, sendMail.Subject, sendMail.Content)
	logged = true
	// The log is the instance's send record, so a test send belongs there too, under the
	// user who ran the test — including an admin testing another user's server.
	if logErr := smtp.logService.Log(user.Id, smtpServer, recipients, files, "", "", sendMail.Subject, sendMail.Content, err); logErr != nil {
		log.Error("sendMailBySMTP log error", zap.Error(logErr))
	}
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
	smtp.logService = core.GetService[*service.LogService](context)
	smtp.cachePath = context.GetConfig().GetStringOrDefault("core.cachePath", "cache")
	context.Get("/api/smtp/:id", smtp.getOne).WithMeta(auth2.WithLogin())
	context.Delete("/api/smtp/:id", smtp.deleteOne).WithMeta(auth2.WithLogin())
	context.Get("/api/smtp", smtp.getPage).WithMeta(auth2.WithLogin())
	context.Post("/api/smtp", smtp.postOne).WithMeta(auth2.WithLogin())
	context.Post("/api/test", smtp.test).WithMeta(auth2.WithLogin())
	context.Put("/api/smtp", smtp.putOne).WithMeta(auth2.WithLogin())
	context.Post("/api/sendMailBySMTP", smtp.sendMail).WithMeta(auth2.WithLogin())
	return nil
}
