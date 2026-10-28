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

func (l *ScheduleService) GetPage(page *web.Page, userId uint, isAdmin bool, name string, adminOnly bool) (any, error) {
	var result *web.PageAble[*model.Schedule]
	var err error
	if isAdmin {
		query := l.scheduleModel.Query()
		if adminOnly {
			query = query.Where("user_id IN (SELECT id FROM t_user WHERE is_admin = ?)", true)
		}
		if name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		result, err = query.Order("id desc").PageForWeb(page)
	} else {
		query := l.scheduleModel.Query().Where("user_id = ?", userId)
		if name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		result, err = query.Order("id desc").PageForWeb(page)
	}
	if err != nil {
		return nil, err
	}
	// The row is what the edit dialog loads, so it must carry the header list the same way
	// the detail endpoint does — otherwise the dialog shows no headers and saving would
	// drop the stored ones.
	for _, item := range result.List {
		expandHeaders(item)
	}
	if isAdmin {
		userIds := make([]uint, 0)
		for _, item := range result.List {
			if item.UserId > 0 {
				userIds = append(userIds, item.UserId)
			}
		}
		userService := wf.GetService[*UserService](l.context)
		userService.FillUserNames(userIds, func(uid uint, name string) {
			for _, item := range result.List {
				if item.UserId == uid {
					item.UserName = name
				}
			}
		})
		admins := userService.AdminIds(userIds)
		for _, item := range result.List {
			item.CreatedByAdmin = admins[item.UserId]
		}
	}
	return result, nil
}
func (l *ScheduleService) Edit(sd *model.Schedule) error {
	if err := applyHeaders(sd); err != nil {
		return err
	}

	v, err := l.tokenModel.FindByPK(sd.TokenId)
	if err != nil {
		return err
	}
	if v == nil {
		return errors.New("token not found")
	}
	// Save, not UpdateByPK: the update path skips zero values, so turning a switch off
	// (is_use, use_template, is_only_send_by_error) or clearing the body would be
	// silently ignored while the request still reported success.
	return l.scheduleModel.Save(sd)
}
func (l *ScheduleService) Save(sd *model.Schedule) error {
	if err := applyHeaders(sd); err != nil {
		return err
	}
	v, err := l.tokenModel.FindByPK(sd.TokenId)
	if err != nil {
		return err
	}
	if v == nil {
		return errors.New("token not found")
	}
	return l.scheduleModel.Save(sd)

}

// expandHeaders decodes the stored header JSON into the transient Headers field, which is
// what clients read and write. A malformed column yields an empty list rather than an error.
func expandHeaders(one *model.Schedule) {
	if util.IsNotBlank(one.HeaderStr) {
		if err := json.Unmarshal([]byte(one.HeaderStr), &one.Headers); err != nil {
			one.Headers = []*model.Header{}
		}
		return
	}
	one.Headers = []*model.Header{}
}

// applyHeaders mirrors the header list into the column that is actually stored; the two
// are separate fields, and only HeaderStr survives a round trip through the database.
// A nil list means the caller did not send one and the stored value is left alone.
func applyHeaders(sd *model.Schedule) error {
	if sd.Headers == nil {
		return nil
	}
	if len(sd.Headers) == 0 {
		sd.HeaderStr = ""
		return nil
	}
	jsonData, err := json.Marshal(sd.Headers)
	if err != nil {
		return err
	}
	sd.HeaderStr = string(jsonData)
	return nil
}

func (l *ScheduleService) GetOne(id int, userId uint) (*model.Schedule, error) {
	one, err := l.scheduleModel.Query().Where("id = ? AND user_id = ?", uint(id), userId).One()
	if err != nil {
		return nil, err
	}
	return l.prepare(one)
}

// GetOneAny loads a schedule without the owner filter, for admin views: the list already
// shows admins every user's schedules, so the row-level read must not refuse them.
func (l *ScheduleService) GetOneAny(id int) (*model.Schedule, error) {
	one, err := l.scheduleModel.Query().Where("id = ?", uint(id)).One()
	if err != nil {
		return nil, err
	}
	return l.prepare(one)
}

// prepare expands the stored header JSON and checks that the referenced token still exists.
func (l *ScheduleService) prepare(one *model.Schedule) (*model.Schedule, error) {
	if one == nil {
		return nil, errors.New("schedule not found")
	}
	expandHeaders(one)

	byToken, err := l.tokenModel.FindByPK(one.TokenId)
	if err != nil {
		return nil, err
	}
	if byToken == nil {
		return nil, errors.New("token not found")
	}
	one.TokenId = byToken.Id
	return one, nil
}
