package service

import (
	"encoding/json"

	"emperror.dev/errors"
	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
)

type ScheduleService struct {
	context       *core.Context
	tokenService  *TokenService
	tokenModel    *model.TokenModel
	scheduleModel *model.ScheduleModel
}

func (l *ScheduleService) Init(context *core.Context) error {
	l.context = context
	l.tokenService = wf.GetService[*TokenService](context)
	l.tokenModel = wf.GetModel[*model.TokenModel](context)
	l.scheduleModel = wf.GetModel[*model.ScheduleModel](context)
	return nil
}

func (l *ScheduleService) FindAll() ([]*model.Schedule, error) {
	return l.scheduleModel.FindAll()
}

func (l *ScheduleService) GetPage(page *web.Page, userId uint) (any, error) {
	if userId > 0 {
		return l.scheduleModel.Query().Where("user_id = ?", userId).Order("id desc").PageForWeb(page)
	}
	return l.scheduleModel.PageForWeb(page)
}
func (l *ScheduleService) Edit(sd *model.Schedule) error {
	if sd.Headers != nil && len(sd.Headers) > 0 {
		jsonData, err := json.Marshal(sd.Headers)
		if err != nil {
			return err
		}
		sd.HeaderStr = string(jsonData)
	}

	v, err := l.tokenModel.FindById(sd.TokenId)
	if err != nil {
		return err
	}
	if v == nil {
		return errors.New("token not found")
	}
	return l.scheduleModel.UpdateById(sd)
}
func (l *ScheduleService) Save(sd *model.Schedule) error {
	v, err := l.tokenModel.FindById(sd.TokenId)
	if err != nil {
		return err
	}
	if v == nil {
		return errors.New("token not found")
	}
	return l.scheduleModel.Save(sd)

}

func (l *ScheduleService) GetOne(id int, userId uint) (*model.Schedule, error) {
	one, err := l.scheduleModel.Query().Where("id = ? AND user_id = ?", uint(id), userId).One()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New("schedule not found")
	}
	headerStr := one.HeaderStr
	if util.IsNotBlank(headerStr) {
		err := json.Unmarshal([]byte(headerStr), &one.Headers)
		if err != nil {
			one.Headers = []*model.Header{}
		}
	} else {
		one.Headers = []*model.Header{}
	}

	byToken, err := l.tokenModel.FindById(one.TokenId)
	if err != nil {
		return nil, err
	}
	one.TokenId = byToken.Id
	return one, nil
}
