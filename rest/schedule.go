package rest

import (
	"io"
	"strconv"
	"time"

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
	user, err := auth.User(req, schedule.context)
	if user == nil {
		return nil, err
	}
	// The list shows admins every user's schedules, so the detail read must not refuse them
	if user.IsAdmin {
		return schedule.scheduleService.GetOneAny(atoi)
	}
	return schedule.scheduleService.GetOne(atoi, user.Id)
}
func (schedule *Schedule) deleteOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, schedule.context)
	if user == nil {
		return nil, err
	}
	exist, err := schedule.scheduleModel.Query().Where("id = ? AND user_id = ?", uint(atoi), user.Id).One()
	if err != nil {
		return nil, err
	}
	if exist == nil {
		return nil, errors.New("schedule not found")
	}
	err = schedule.scheduleModel.DeleteByPK(uint(atoi))
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
	user, err := auth.User(req, schedule.context)
	if user == nil {
		return nil, err
	}
	name := req.Query("name")
	adminOnly := req.Query("adminOnly") == "true"
	return schedule.scheduleService.GetPage(page, user.Id, user.IsAdmin, name, adminOnly)
}
func (schedule *Schedule) postOne(req *web.Request) (any, error) {
	var st model.Schedule
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, schedule.context)
	if user == nil {
		return nil, err
	}
	// Create is create: a client-supplied id must not turn this into an UPDATE of an
	// existing row (gorm Save updates when the primary key is set), which would let any
	// signed-in user overwrite another account's schedule and take ownership of it.
	st.Id = 0
	st.UserId = user.Id
	err = schedule.validate(&st)
	if err != nil {
		return nil, err
	}
	if err = schedule.ensureTokenOwnedBy(st.TokenId, st.UserId); err != nil {
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
	user, err := auth.User(req, schedule.context)
	if user == nil {
		return nil, err
	}
	exist, err := schedule.scheduleModel.FindByPK(st.Id)
	if err != nil {
		return nil, err
	}
	// Report "not found" rather than "forbidden" so we don't leak existence
	if exist == nil || (exist.UserId != user.Id && !user.IsAdmin) {
		return nil, errors.New("schedule not found")
	}
	// Keep the original owner: an admin editing someone else's row must not take it over
	exist.Name = st.Name
	exist.TokenId = st.TokenId
	exist.Cron = st.Cron
	exist.Url = st.Url
	exist.Method = st.Method
	exist.Body = st.Body
	exist.UseTemplate = st.UseTemplate
	exist.Template = st.Template
	exist.IsUse = st.IsUse
	exist.IsOnlySendByError = st.IsOnlySendByError
	// Headers omitted (nil) means "leave the stored list alone"; an empty array clears it
	if st.Headers != nil {
		exist.Headers = st.Headers
	}
	exist.UpdateTime = time.Now()
	err = schedule.validate(exist)
	if err != nil {
		return nil, err
	}
	if err = schedule.ensureTokenOwnedBy(exist.TokenId, exist.UserId); err != nil {
		return nil, err
	}
	err = schedule.scheduleService.Edit(exist)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}

// ensureTokenOwnedBy rejects a schedule that points at a token the owner does not hold.
func (schedule *Schedule) ensureTokenOwnedBy(tokenId uint, userId uint) error {
	byToken, err := schedule.tokenService.GetOne(tokenId, userId)
	if err != nil {
		return err
	}
	if byToken == nil {
		return errors.New("token not found")
	}
	return nil
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
	user, err := auth.User(req, schedule.context)
	if user == nil {
		return nil, err
	}
	if err = schedule.ensureTokenOwnedBy(st.TokenId, user.Id); err != nil {
		return nil, err
	}
	st.UserId = user.Id
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
	user, err := auth.User(req, schedule.context)
	if user == nil {
		return nil, err
	}
	var st *model.Schedule
	if user.IsAdmin {
		st, err = schedule.scheduleService.GetOneAny(atoi)
	} else {
		st, err = schedule.scheduleService.GetOne(atoi, user.Id)
	}
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
	context.Any("/scheduleTestApi", schedule.scheduleTestApi).WithMeta(auth2.WithLogin())
	schedule.scheduleService = wf.GetService[*service.ScheduleService](schedule.context)
	schedule.tokenService = wf.GetService[*service.TokenService](schedule.context)
	schedule.scheduleModel = wf.GetModel[*model.ScheduleModel](schedule.context)
	return nil

}
