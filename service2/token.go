package service2

import (
	"bytes"
	"errors"
	"sync"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/smtp"
	"go.uber.org/zap"
)

type TokenService struct {
	context    *core.Context
	mailModel  *model.MailModel
	smtpModel  *model.SMTPModel
	tokenModel *model.TokenModel
	logService *LogService
	lock       *sync.RWMutex
}

func (l *TokenService) GetOne(id uint) (*model.Token, error) {
	tokenModel := core.GetModel[*model.TokenModel](l.context)
	token, err := tokenModel.FindById(id)
	if err != nil {
		return nil, err
	}
	return token, nil
}
func (l *TokenService) GetPage(page *web.Page) (any, error) {

	tokens, i, err := l.tokenModel.Page(page)
	if err != nil {
		return nil, err
	}
	l.supplementToken(tokens...)
	return web.ToPage[*model.Token](int64(i), tokens), nil
}

func (l *TokenService) SendApiCallMail(schedule *model.Schedule) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	byToken, err := l.tokenModel.Query().Where("token=?", schedule.Token).One()
	if err != nil {
		return err
	}
	if byToken == nil {
		return errors.New("token not found")
	}
	if byToken.IsUse {
		body, err := smtp.SendAPIMail2(schedule, byToken.SMTP, byToken.ReceiveEmails)
		if err != nil {
			log.Error("SendAPIMail log error", zap.Error(err))
		}
		err2 := l.logService.Log(byToken.SMTP, byToken.ReceiveEmails, nil, schedule.Token, schedule.Name, body, err)
		if err2 != nil {
			log.Error("SendAPIMail log error", zap.Error(err2))
		}
		return err
	}
	return errors.New("token is not use")
}

func (l *TokenService) supplementToken(st ...*model.Token) {
	mailIds := make([]uint, 0)
	stmpIds := make([]uint, 0)
	for _, d := range st {
		d.Name = d.Token
		mailIds = append(mailIds, util.StringToUintIds(d.ReceiveEmailIds)...)
		stmpIds = append(stmpIds, d.SMTPId)
	}
	mailMap, err := l.mailModel.FindMapByIds(mailIds)
	if err == nil {
		for _, d := range st {
			mailIds := util.StringToUintIds(d.ReceiveEmailIds)
			d.ReceiveEmails = GetMails(mailIds, mailMap)
			d.ReceiveEmailsStr = GetMailsStr(d.ReceiveEmails)
		}
	}
	idsMap, err := l.smtpModel.FindMapByIds(stmpIds)
	if err == nil {
		for _, d := range st {
			d.SMTP = idsMap[d.SMTPId]
			if d.SMTP != nil {
				d.SMTPStr = d.SMTP.Name
			}
		}
	}
}

func (l *TokenService) Name() string {
	return "LogService"
}

func (l *TokenService) Init(context *core.Context) {
	l.context = context
	l.lock = new(sync.RWMutex)
	l.mailModel = core.GetModel[*model.MailModel](l.context)
	l.smtpModel = core.GetModel[*model.SMTPModel](l.context)
	l.tokenModel = core.GetModel[*model.TokenModel](l.context)
	l.logService = core.GetService[*LogService](l.context)
}

func GetMails(ids []uint, mailMap map[uint]*model.Mail) []*model.Mail {
	mails := make([]*model.Mail, 0)
	for _, id := range ids {
		v, ok := mailMap[id]
		if ok {
			mails = append(mails, v)
		}
	}
	return mails
}
func GetMailsStr(mails []*model.Mail) string {
	buffer := new(bytes.Buffer)
	for _, mail := range mails {
		buffer.WriteString("," + util.FormatMail(mail.Name, mail.Mail))
	}
	if buffer.Len() == 0 {
		return ""
	}
	return buffer.String()[1:]
}
