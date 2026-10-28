package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/smtp"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
)

type LogService struct {
	context         *core.Context
	logModel        *model.LogModel
	attachmentModel *model.AttachmentModel
}

func (l *LogService) log(userId uint, st *model.SMTP, mails []*model.Mail, tokenName string, token string, subject, bodyString string, files []*smtp.File, err error) error {
	var lg model.Log
	// The log list and the attachment download are both filtered by user_id, so the
	// owner must be recorded here — nothing fills it in later.
	lg.UserId = userId
	lg.Token = token
	lg.Name = tokenName
	// st may be nil when the referenced SMTP server was deleted
	if st != nil {
		lg.SMTP = util.FormatMail(st.Username, st.Mail)
	}
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
	status := entity.SUCCESS
	if err != nil {
		status = entity.ERROR
	}
	if status == entity.SUCCESS {
		lg.Result = "success"
		lg.Status = status
	} else {
		var ee *smtp.UserNotFoundError
		ok := errors.As(err, &ee)
		if ok {
			lg.Status = entity.WARM
			lg.Result = err.Error()
		} else {
			lg.Result = err.Error()
			lg.Status = status
		}
	}

	if err := l.logModel.Save(&lg); err != nil {
		return err
	}
	l.recordAttachments(&lg, files)
	return nil
}

// recordAttachments stores one row per file, linking it to the mail log it was sent
// with. Best effort: the log's files JSON still records the paths, so a failure here
// leaves downloads working through the legacy lookup.
func (l *LogService) recordAttachments(logEntry *model.Log, files []*smtp.File) {
	if l.attachmentModel == nil || len(files) == 0 {
		return
	}
	now := time.Now()
	rows := make([]*model.Attachment, 0, len(files))
	for _, file := range files {
		if file == nil || file.FilePath == "" {
			continue
		}
		rows = append(rows, &model.Attachment{
			LogId:      logEntry.Id,
			Name:       file.Name,
			FilePath:   file.FilePath,
			CreateTime: now,
			UpdateTime: now,
		})
	}
	if len(rows) == 0 {
		return
	}
	if err := l.attachmentModel.Saves(rows); err != nil {
		log.Error("attachment rows write failed", zap.Uint("logId", logEntry.Id), zap.Error(err))
	}
}

// Log records one send attempt. userId is the owner the entry belongs to: the token
// owner for a token send, the schedule owner for a scheduled call.
func (l *LogService) Log(userId uint, smtp *model.SMTP, mails []*model.Mail, files []*smtp.File, tokenName string, token string, subject, bodyString string, err error) error {
	return l.log(userId, smtp, mails, tokenName, token, subject, bodyString, files, err)
}

func (l *LogService) Name() string {
	return "LogService"
}
func (l *LogService) Init(context *core.Context) error {
	l.context = context
	l.logModel = core.GetModel[*model.LogModel](context)
	l.attachmentModel = core.GetModel[*model.AttachmentModel](context)
	return nil
}
