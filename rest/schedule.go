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
	"github.com/chuccp/http2smtp/entity"
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
	recordOp(schedule.context, req, user.Id, entity.OpDelete, entity.TargetSchedule, uint(atoi), exist.Name)
	return web.Ok("ok"), nil
}
func (schedule *Schedule) getPage(req *web.Request) (any, error) {
	page, err := req.Page()
	if err != nil {
		return nil, err
	}
	page = capPage(page)
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
	if err = schedule.ensureTokenUsable(st.TokenId, user); err != nil {
		return nil, err
	}
	err = schedule.scheduleService.Save(&st)
	if err != nil {
		return nil, err
	}
	recordOp(schedule.context, req, user.Id, entity.OpCreate, entity.TargetSchedule, st.Id, st.Name)
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
	exist.ContentType = st.ContentType
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
	if err = schedule.ensureTokenUsable(exist.TokenId, user); err != nil {
		return nil, err
	}
	err = schedule.scheduleService.Edit(exist)
	if err != nil {
		return nil, err
	}
	recordOp(schedule.context, req, user.Id, entity.OpUpdate, entity.TargetSchedule, exist.Id, exist.Name)
	return web.Ok("ok"), nil
}

// ensureTokenUsable enforces the token reference rule: a regular caller may only point
// a schedule at their own token, while an admin may use anyone's — the admin token list
// already exposes every token together with the SMTP configuration behind it.
func (schedule *Schedule) ensureTokenUsable(tokenId uint, user *model.User) error {
	if !user.IsAdmin {
		return schedule.ensureTokenOwnedBy(tokenId, user.Id)
	}
	byToken, err := schedule.tokenService.GetOneAny(tokenId)
	if err != nil {
		return err
	}
	if byToken == nil {
		return errors.New("token not found")
	}
	return nil
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
	if err = schedule.ensureTokenUsable(st.TokenId, user); err != nil {
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

// toggleOne flips a schedule's enabled flag. The row is read straight from the model
// (not GetOne) so a schedule whose token has since been deleted can still be switched
// off; ownership follows the trigger route, where admins reach every user's rows.
func (schedule *Schedule) toggleOne(req *web.Request) (any, error) {
	id, err := strconv.Atoi(req.Param("id"))
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, schedule.context)
	if user == nil {
		return nil, err
	}
	query := schedule.scheduleModel.Query().Where("id = ?", uint(id))
	if !user.IsAdmin {
		query = query.Where("user_id = ?", user.Id)
	}
	st, err := query.One()
	if err != nil {
		return nil, err
	}
	if st == nil {
		return nil, errors.New("schedule not found")
	}
	st.IsUse = !st.IsUse
	st.UpdateTime = time.Now()
	// Save, not UpdateByPK: turning the switch off must persist the zero value.
	if err := schedule.scheduleModel.Save(st); err != nil {
		return nil, err
	}
	action := entity.OpDisable
	if st.IsUse {
		action = entity.OpEnable
	}
	recordOp(schedule.context, req, user.Id, action, entity.TargetSchedule, st.Id, st.Name)
	return web.Ok("ok"), nil
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
	recordOp(schedule.context, req, user.Id, entity.OpTrigger, entity.TargetSchedule, st.Id, st.Name)
	return web.Ok("ok"), nil
}

func (schedule *Schedule) Init(context *core.Context) error {
	schedule.context = context
	context.Get("/api/schedule/:id", schedule.getOne).WithMeta(auth2.WithLogin())
	context.Delete("/api/schedule/:id", schedule.deleteOne).WithMeta(auth2.WithLogin())
	context.Get("/api/schedule", schedule.getPage).WithMeta(auth2.WithLogin())
	context.Post("/api/schedule", schedule.postOne).WithMeta(auth2.WithLogin())
	context.Put("/api/schedule", schedule.putOne).WithMeta(auth2.WithLogin())
	context.Post("/api/schedule/trigger/:id", schedule.triggerById).WithMeta(auth2.WithLogin())
	context.Post("/api/schedule/toggle/:id", schedule.toggleOne).WithMeta(auth2.WithLogin())
	context.Post("/api/sendMailBySchedule", schedule.sendMail).WithMeta(auth2.WithLogin())
	// Deliberately unauthenticated: a schedule calls this echo from the server itself,
	// without a browser session, so a login check would make the test link answer 403
	// every time. It only reflects the caller's own request and touches no stored data.
	context.Any("/api/scheduleTestApi", schedule.scheduleTestApi)
	schedule.scheduleService = wf.GetService[*service.ScheduleService](schedule.context)
	schedule.tokenService = wf.GetService[*service.TokenService](schedule.context)
	schedule.scheduleModel = wf.GetModel[*model.ScheduleModel](schedule.context)
	return nil

}
