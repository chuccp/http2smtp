package service2

import (
	"encoding/json"
	"errors"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/http2smtp/db"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/smtp"
	"go.uber.org/zap/buffer"
)

type LogService struct {
	context  *core.Context
	logModel *model.LogModel
}

func (l *LogService) log(st *model.SMTP, mails []*model.Mail, token string, subject, bodyString string, files []*smtp.File, err error) error {
	var lg model.Log
	lg.Token = token
	lg.SMTP = util.FormatMail(st.Username, st.Mail)
	b := new(buffer.Buffer)
	for _, mail := range mails {
		b.AppendString(",")
		b.AppendString(util.FormatMail(mail.Name, mail.Mail))
	}
	if b.Len() > 0 {
		lg.Mail = b.String()[1:]
	}
	lg.Subject = subject
	lg.Content = bodyString
	if files != nil && len(files) > 0 {
		marshal, err := json.Marshal(files)
		if err == nil {
			lg.Files = string(marshal)
		}
	}
	status := db.SUCCESS
	if err != nil {
		status = db.ERROR
	}
	if status == db.SUCCESS {
		lg.Result = "success"
		lg.Status = status
	} else {
		var ee *smtp.UserNotFoundError
		ok := errors.As(err, &ee)
		if ok {
			lg.Status = db.WARM
			lg.Result = err.Error()
		} else {
			lg.Result = err.Error()
			lg.Status = status
		}
	}

	return l.logModel.Save(&lg)
}
func (l *LogService) Log(smtp *model.SMTP, mails []*model.Mail, files []*smtp.File, token string, subject, bodyString string, err error) error {
	return l.log(smtp, mails, token, subject, bodyString, files, err)
}

func (l *LogService) Name() string {
	return "LogService"
}
func (l *LogService) Init(context *core.Context) {
	l.context = context
	l.logModel = core.GetModel[*model.LogModel](context)
}
