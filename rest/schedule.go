package rest

import (
	"io"
	"strconv"

	"emperror.dev/errors"
	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
	"github.com/chuccp/http2smtp/util"
	"go.uber.org/zap"
)

type Schedule struct {
	context         *core.Context
	scheduleService *service.ScheduleService
	tokenService    *service.TokenService
	scheduleModel   *model.ScheduleModel
}

func (schedule *Schedule) getOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	user, _ := auth.User(req, schedule.context)
	return schedule.scheduleService.GetOne(atoi, user.Id)
}
func (schedule *Schedule) deleteOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	user, _ := auth.User(req, schedule.context)
	exist, err := schedule.scheduleModel.Query().Where("id = ? AND user_id = ?", uint(atoi), user.Id).One()
	if err != nil {
		return nil, err
	}
	if exist == nil {
		return nil, errors.New("schedule not found")
	}
	err = schedule.scheduleModel.DeleteById(uint(atoi))
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}
func (schedule *Schedule) getPage(req *web.Request) (any, error) {
	page, err := req.Page()
	if err != nil {
		return nil, err
	}
	user, _ := auth.User(req, schedule.context)
	return schedule.scheduleService.GetPage(page, user.Id)
}
func (schedule *Schedule) postOne(req *web.Request) (any, error) {
	var st model.Schedule
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	user, _ := auth.User(req, schedule.context)
	st.UserId = user.Id
	err = schedule.validate(&st)
	if err != nil {
		return nil, err
	}
	err = schedule.scheduleService.Save(&st)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}
func (schedule *Schedule) putOne(req *web.Request) (any, error) {
	var st model.Schedule
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	user, _ := auth.User(req, schedule.context)
	st.UserId = user.Id
	err = schedule.validate(&st)
	if err != nil {
		return nil, err
	}
	err = schedule.scheduleService.Edit(&st)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}
func (schedule *Schedule) sendMail(req *web.Request) (any, error) {
	var st model.Schedule
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	if st.TokenId == 0 {
		return nil, errors.New("tokenId is required")
	}
	err = schedule.validate(&st)
	if err != nil {
		return nil, err
	}
	st.IsOnlySendByError = false
	err = schedule.tokenService.SendApiCallMail(&st)
	if err != nil {
		return nil, err
	}

	return web.Ok("ok"), nil
}
func (schedule *Schedule) validate(st *model.Schedule) error {
	err := util.ValidateURL(st.Url)
	if err != nil {
		return err
	}
	err = util.ParserCron(st.Cron)
	if err != nil {
		return err
	}
	if st.TokenId == 0 {
		return errors.New("tokenId cannot be empty")
	}
	return nil
}
func (schedule *Schedule) scheduleTestApi(req *web.Request) (any, error) {
	params := make(map[string]any)
	request := req.Request()
	params["header"] = request.Header
	body, _ := io.ReadAll(request.Body)
	params["body"] = string(body)
	params["method"] = request.Method
	params["url"] = req.URL().String()
	log.Debug("scheduleTestApi", zap.String("body", string(body)))
	return params, nil
}

func (schedule *Schedule) triggerById(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	user, _ := auth.User(req, schedule.context)
	st, err := schedule.scheduleService.GetOne(atoi, user.Id)
	if err != nil {
		return nil, err
	}
	if st == nil {
		return nil, errors.New("schedule not found")
	}
	st.IsOnlySendByError = false
	err = schedule.validate(st)
	if err != nil {
		return nil, err
	}
	err = schedule.tokenService.SendApiCallMail(st)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}

func (schedule *Schedule) Init(context *core.Context) error {
	schedule.context = context
	context.Get("/schedule/:id", schedule.getOne).WithMeta(auth2.WithLogin())
	context.Delete("/schedule/:id", schedule.deleteOne).WithMeta(auth2.WithLogin())
	context.Get("/schedule", schedule.getPage).WithMeta(auth2.WithLogin())
	context.Post("/schedule", schedule.postOne).WithMeta(auth2.WithLogin())
	context.Put("/schedule", schedule.putOne).WithMeta(auth2.WithLogin())
	context.Post("/schedule/trigger/:id", schedule.triggerById).WithMeta(auth2.WithLogin())
	context.Post("/sendMailBySchedule", schedule.sendMail).WithMeta(auth2.WithLogin())
	context.Any("/scheduleTestApi", schedule.scheduleTestApi)
	schedule.scheduleService = wf.GetService[*service.ScheduleService](schedule.context)
	schedule.tokenService = wf.GetService[*service.TokenService](schedule.context)
	schedule.scheduleModel = wf.GetModel[*model.ScheduleModel](schedule.context)
	return nil

}
