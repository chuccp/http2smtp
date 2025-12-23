package service2

import (
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/http2smtp/db"
	"github.com/chuccp/http2smtp/model"
)

type TokenService struct {
	context *core.Context
}

func (l *TokenService) GetOne(id uint) (*model.Token, error) {
	tokenModel := core.GetModel[*model.TokenModel](l.context)
	token, err := tokenModel.FindById(id)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (token *TokenService) supplementToken(st ...*db.Token) {
	mailIds := make([]uint, 0)
	stmpIds := make([]uint, 0)
	for _, d := range st {
		d.Name = d.Token
		mailIds = append(mailIds, util.StringToUintIds(d.ReceiveEmailIds)...)
		stmpIds = append(stmpIds, d.SMTPId)
	}
	mailMap, err := token.db.GetMailModel().GetMapByIds(mailIds)
	if err == nil {
		for _, d := range st {
			mailIds := util.StringToUintIds(d.ReceiveEmailIds)
			d.ReceiveEmails = db.GetMails(mailIds, mailMap)
			d.ReceiveEmailsStr = db.GetMailsStr(d.ReceiveEmails)
		}
	}
	idsMap, err := token.db.GetSMTPModel().GetMapByIds(stmpIds)
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
}
