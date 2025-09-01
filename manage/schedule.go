package manage

import (
	"errors"
	"github.com/chuccp/http2smtp/core"
	"github.com/chuccp/http2smtp/db"
	"github.com/chuccp/http2smtp/util"
	"github.com/chuccp/http2smtp/web"
	"go.uber.org/zap"
	"io"
	"strconv"
)

type Schedule struct {
	context *core.Context
}

func (schedule *Schedule) getOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	return schedule.context.GetScheduleService().GetOne(atoi)
}
func (schedule *Schedule) deleteOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	err = schedule.context.GetDb().GetScheduleModel().DeleteOne(uint(atoi))
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (schedule *Schedule) getPage(req *web.Request) (any, error) {
	page := req.GetPage()
	return schedule.context.GetScheduleService().GetPage(page)
}
func (schedule *Schedule) postOne(req *web.Request) (any, error) {
	var st db.Schedule
	err := req.ShouldBindBodyWithJSON(&st)
	if err != nil {
		return nil, err
	}
	err = schedule.validate(&st)
	if err != nil {
		return nil, err
	}
	err = schedule.context.GetScheduleService().Save(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (schedule *Schedule) putOne(req *web.Request) (any, error) {
	var st db.Schedule
	err := req.ShouldBindBodyWithJSON(&st)
	if err != nil {
		return nil, err
	}
	err = schedule.validate(&st)
	if err != nil {
		return nil, err
	}
	err = schedule.context.GetScheduleService().Edit(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (schedule *Schedule) sendMail(req *web.Request) (any, error) {
	var st db.Schedule
	err := req.ShouldBindBodyWithJSON(&st)
	err = schedule.validate(&st)
	if err != nil {
		return nil, err
	}
	st.IsOnlySendByError = false
	err = schedule.context.GetScheduleService().SendMail(&st)
	if err != nil {
		return nil, err
	}

	return "ok", nil
}
func (schedule *Schedule) validate(st *db.Schedule) error {
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
	request := req.GetRawRequest()
	params["header"] = request.Header
	body, _ := io.ReadAll(request.Body)
	params["body"] = string(body)
	params["method"] = request.Method
	params["url"] = request.URL.String()
	schedule.context.GetLog().Debug("scheduleTestApi", zap.String("body", string(body)))
	return params, nil
}

func (schedule *Schedule) Init(context *core.Context, server core.IHttpServer) {
	schedule.context = context
	server.GETAuth("/schedule/:id", schedule.getOne)
	server.DELETEAuth("/schedule/:id", schedule.deleteOne)
	server.GETAuth("/schedule", schedule.getPage)
	server.POSTAuth("/schedule", schedule.postOne)
	server.PUTAuth("/schedule", schedule.putOne)
	server.POSTAuth("/sendMailBySchedule", schedule.sendMail)
	server.Any("/scheduleTestApi", schedule.scheduleTestApi)

}
