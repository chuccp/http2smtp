package service

import (
	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
)

type SmtpService struct {
	context     *core.Context
	smtpModel   *model.SMTPModel
	userService *UserService
}

func (s *SmtpService) PageForWeb(page *web.Page, userId uint, isAdmin bool, adminOnly bool, name string, host string, username string) (*web.PageAble[*model.SMTP], error) {
	var result *web.PageAble[*model.SMTP]
	var err error
	if isAdmin {
		query := s.smtpModel.Query()
		if adminOnly {
			query = query.Where("user_id IN (SELECT id FROM t_user WHERE is_admin = ?)", true)
		}
		if name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		if host != "" {
			query = query.Where("host LIKE ?", "%"+host+"%")
		}
		if username != "" {
			query = query.Where("username LIKE ?", "%"+username+"%")
		}
		result, err = query.Order("id desc").PageForWeb(page)
	} else {
		query := s.smtpModel.Query().Where("user_id = ?", userId)
		if name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		if host != "" {
			query = query.Where("host LIKE ?", "%"+host+"%")
		}
		if username != "" {
			query = query.Where("username LIKE ?", "%"+username+"%")
		}
		result, err = query.Order("id desc").PageForWeb(page)
	}
	if err != nil {
		return nil, err
	}
	if isAdmin {
		userIds := make([]uint, 0)
		for _, item := range result.List {
			if item.UserId > 0 {
				userIds = append(userIds, item.UserId)
			}
		}
		s.userService.FillUserNames(userIds, func(uid uint, name string) {
			for _, item := range result.List {
				if item.UserId == uid {
					item.UserName = name
				}
			}
		})
	}
	return result, nil
}

func (s *SmtpService) Init(context *core.Context) error {
	s.context = context
	s.smtpModel = wf.GetModel[*model.SMTPModel](context)
	s.userService = wf.GetService[*UserService](context)
	return nil
}
