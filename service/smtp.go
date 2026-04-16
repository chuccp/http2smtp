package service

import (
	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
)

type SmtpService struct {
	context   *core.Context
	smtpModel *model.SMTPModel
}

func (s *SmtpService) PageForWeb(page *web.Page, userId uint) (*web.PageAble[*model.SMTP], error) {
	result, err := s.smtpModel.Query().Where("user_id = ?", userId).Order("id desc").PageForWeb(page)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *SmtpService) Init(context *core.Context) error {
	s.context = context
	s.smtpModel = wf.GetModel[*model.SMTPModel](context)
	return nil
}
