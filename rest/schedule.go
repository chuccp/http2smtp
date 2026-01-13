package rest

import (
	"io"
	"strconv"

	"emperror.dev/errors"
	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service2"
	"github.com/chuccp/http2smtp/util"
	"go.uber.org/zap"
)

type Schedule struct {
	context         *core.Context
	scheduleService *service2.ScheduleService
	tokenService    *service2.TokenService
	scheduleModel   *model.ScheduleModel
}

func (schedule *Schedule) getOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	return schedule.scheduleService.GetOne(atoi)
}
func (schedule *Schedule) deleteOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	err = schedule.scheduleModel.DeleteById(uint(atoi))
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (schedule *Schedule) getPage(req *web.Request) (any, error) {
	page, err := req.Page()
	if err != nil {
		return nil, err
	}
	return schedule.scheduleService.GetPage(page)
}
func (schedule *Schedule) postOne(req *web.Request) (any, error) {
	var st model.Schedule
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	err = schedule.validate(&st)
	if err != nil {
		return nil, err
	}
	err = schedule.scheduleService.Save(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (schedule *Schedule) putOne(req *web.Request) (any, error) {
	var st model.Schedule
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	err = schedule.validate(&st)
	if err != nil {
		return nil, err
	}
	err = schedule.scheduleService.Edit(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (schedule *Schedule) sendMail(req *web.Request) (any, error) {
	var st model.Schedule
	err := req.BindJSON(&st)
	err = schedule.validate(&st)
	if err != nil {
		return nil, err
	}
	st.IsOnlySendByError = false
	err = schedule.tokenService.SendApiCallMail(&st)
	if err != nil {
		return nil, err
	}

	return "ok", nil
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
	if len(st.Token) == 0 {
		return errors.New(" token  cannot be empty")
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

func (schedule *Schedule) Init(context *core.Context) error {
	schedule.context = context
	context.GetAuth("/schedule/:id", schedule.getOne)
	context.DeleteAuth("/schedule/:id", schedule.deleteOne)
	context.GetAuth("/schedule", schedule.getPage)
	context.PostAuth("/schedule", schedule.postOne)
	context.PutAuth("/schedule", schedule.putOne)
	context.PostAuth("/sendMailBySchedule", schedule.sendMail)
	context.Any("/scheduleTestApi", schedule.scheduleTestApi)
	schedule.scheduleService = wf.GetService[*service2.ScheduleService](schedule.context)
	schedule.tokenService = wf.GetService[*service2.TokenService](schedule.context)
	schedule.scheduleModel = wf.GetModel[*model.ScheduleModel](schedule.context)
	return nil

}
