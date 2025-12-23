package service

import (
	"encoding/json"
	"errors"

	"github.com/chuccp/http2smtp/db"
	"github.com/chuccp/http2smtp/smtp"
	"github.com/chuccp/http2smtp/util"
	"go.uber.org/zap/buffer"
)

type Log struct {
	db *db.DB
}

func NewLog(db *db.DB) *Log {
	return &Log{db: db}
}

func (a *Log) log(st *db.SMTP, mails []*db.Mail, token string, subject, bodyString string, files []*smtp.File, err error) error {
	var lg db.Log
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
	return a.db.GetLogModel().Save(&lg)
}

func (a *Log) Log(smtp *db.SMTP, mails []*db.Mail, files []*smtp.File, token string, subject, bodyString string, err error) error {
	return a.log(smtp, mails, token, subject, bodyString, files, err)
}
