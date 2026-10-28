package webui

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/util"
)

// defaultContentType is what a POST schedule sends unless the form picks another one.
const defaultContentType = "application/json"

// formContentTypes is the set the schedule form offers. A Content-Type header row stays
// as the escape hatch for any other type.
var formContentTypes = map[string]bool{
	"application/json":                  true,
	"application/x-www-form-urlencoded": true,
	"text/plain":                        true,
	"application/xml":                   true,
}

// normalizeContentType keeps a forged or stale form value from reaching the request.
func normalizeContentType(value string) string {
	if formContentTypes[value] {
		return value
	}
	return defaultContentType
}

// scheduleTestUrl is this instance's request-echo endpoint. A scheduled call is made
// inside the process, so the instance's own loopback address and management port stay
// reachable even when the UI is opened through a domain or a Docker port mapping.
func (p *Pages) scheduleTestUrl() string {
	port := model.ManagePort
	if cfg, err := model.GetConfig(p.ctx.GetConfig()); err == nil && cfg.Manage != nil && cfg.Manage.Port > 0 {
		port = cfg.Manage.Port
	}
	return fmt.Sprintf("http://127.0.0.1:%d/api/scheduleTestApi", port)
}

// ScheduleListData is the scheduled-tasks list page.
type ScheduleListData struct {
	Base
	Rows      []*model.Schedule
	Pager     *Pager
	Name      string
	AdminOnly bool
}

func (p *Pages) scheduleList(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireUser(view)
	if !ok {
		return nil
	}
	page := listPage(req)
	name := req.Query("name")
	adminOnly := boolQuery(req, "adminOnly")

	result, err := p.scheduleSvc.GetPage(page, user.Id, user.IsAdmin, name, adminOnly)
	if err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	pb, ok2 := result.(*web.PageAble[*model.Schedule])
	if !ok2 || pb == nil {
		return p.renderErr(view, 0, "common.failed", "unexpected page result")
	}
	if user.IsAdmin {
		fillNames(p.userService, pb.List,
			func(s *model.Schedule) uint { return s.UserId },
			func(s *model.Schedule, n string) { s.UserName = n })
	}
	q := queryOf(req, "name", "adminOnly", "pageSize")
	data := &ScheduleListData{
		Base:      p.base(req, user, "schedule.scheduleManagement", true),
		Rows:      pb.List,
		Pager:     newPager(page.PageNo, page.PageSize, pb.Total, "/schedule", q),
		Name:      name,
		AdminOnly: adminOnly,
	}
	view.View("schedule_list").Layout(layoutFor("schedule_list")).Data(data)
	return nil
}

// ScheduleFormData is the scheduled-task create/edit form.
type ScheduleFormData struct {
	Base
	Id           uint
	Name         string
	Cron         string
	Url          string
	TestUrl      string
	Method       string
	Body         string
	ContentType  string
	UseTemplate  bool
	Template     string
	OnlyByError  bool
	IsUse        bool
	Headers      []*model.Header
	TokenOptions []SelectOption
}

func (p *Pages) scheduleForm(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireUser(view)
	if !ok {
		return nil
	}
	testUrl := p.scheduleTestUrl()
	data := &ScheduleFormData{
		Base:        p.base(req, user, "schedule.addSchedule", false),
		TestUrl:     testUrl,
		Method:      "GET",
		ContentType: defaultContentType,
		IsUse:       true,
		Headers:     []*model.Header{{}, {}},
	}
	var selectedToken uint
	// A schedule belongs to its owner; an admin editing someone else's schedule must
	// pick from that user's tokens, not their own.
	ownerId := user.Id
	if s := req.Param("id"); s != "" {
		id, err := idParam(req, "id")
		if err != nil {
			return p.renderErr(view, http.StatusBadRequest, "common.failed", err.Error())
		}
		row, err := p.scheduleOwnedBy(id, user)
		if err != nil || row == nil {
			return p.renderErr(view, http.StatusNotFound, "common.failed", "schedule not found")
		}
		selectedToken = row.TokenId
		ownerId = resourceOwnerId(row.UserId, user.Id)
		data.Id = row.Id
		data.Name = row.Name
		data.Cron = row.Cron
		data.Url = row.Url
		data.Method = row.Method
		data.Body = row.Body
		if row.ContentType != "" {
			data.ContentType = row.ContentType
		}
		data.UseTemplate = row.UseTemplate
		data.Template = row.Template
		data.OnlyByError = row.IsOnlySendByError
		data.IsUse = row.IsUse
		if len(row.Headers) > 0 {
			data.Headers = row.Headers
		}
		data.Title = translatorFor(data.Lang).T("schedule.editSchedule")
	}
	if err := p.fillTokenOptions(data, user, ownerId, selectedToken); err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	view.View("schedule_form").Layout(layoutFor("schedule_form")).Data(data)
	return nil
}

// fillTokenOptions loads the enabled tokens a schedule may point at; only in-use
// tokens drive a schedule. Everyone but an admin sees just the schedule owner's
// tokens; an admin may pick any user's, so those options carry the owner name.
func (p *Pages) fillTokenOptions(data *ScheduleFormData, user *model.User, ownerId uint, selected uint) error {
	query := p.tokenModel.Query().Where("state = 0")
	if !user.IsAdmin {
		query = query.Where("user_id = ?", ownerId)
	}
	tokens, err := query.Order("id desc").All()
	if err != nil {
		return err
	}
	if user.IsAdmin {
		userIds := make([]uint, 0, len(tokens))
		for _, t := range tokens {
			if t.UserId > 0 {
				userIds = append(userIds, t.UserId)
			}
		}
		if nameMap, err := p.userModel.FindNameMapByIds(userIds); err == nil {
			for _, t := range tokens {
				t.UserName = nameMap[t.UserId]
			}
		}
	}
	for _, t := range tokens {
		label := t.Name
		if user.IsAdmin && t.UserName != "" {
			label += " (" + t.UserName + ")"
		}
		data.TokenOptions = append(data.TokenOptions, SelectOption{
			Value:    strconv.Itoa(int(t.Id)),
			Label:    label,
			Selected: t.Id == selected,
		})
	}
	return nil
}

// activeTokenOwnedBy returns an in-use token owned by ownerId, the only kind a
// schedule may point at.
func (p *Pages) activeTokenOwnedBy(ownerId, tokenId uint) (*model.Token, error) {
	return p.tokenModel.Query().Where("id = ? AND user_id = ? AND state = 0", tokenId, ownerId).One()
}

func (p *Pages) scheduleOwnedBy(id uint, user *model.User) (*model.Schedule, error) {
	if user.IsAdmin {
		return p.scheduleSvc.GetOneAny(int(id))
	}
	return p.scheduleSvc.GetOne(int(id), user.Id)
}

// maxHeaderRows caps the dynamic header rows a form may submit.
const maxHeaderRows = 64

// collectHeaders gathers the repeated header_name_i / header_value_i pairs, dropping
// rows where both sides are blank.
func collectHeaders(req *web.Request) []*model.Header {
	headers := make([]*model.Header, 0)
	for i := 0; i <= maxHeaderRows; i++ {
		name := req.GetFormParam("header_name_" + strconv.Itoa(i))
		value := req.GetFormParam("header_value_" + strconv.Itoa(i))
		if name == "" && value == "" {
			continue
		}
		headers = append(headers, &model.Header{Name: name, Value: value})
	}
	return headers
}

func (p *Pages) scheduleSave(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireUser(view)
	if !ok {
		return nil
	}
	st := &model.Schedule{
		UserId:            user.Id,
		Name:              req.GetFormParam("name"),
		Cron:              req.GetFormParam("cron"),
		Url:               req.GetFormParam("url"),
		Method:            req.GetFormParam("method"),
		Body:              req.GetFormParam("body"),
		ContentType:       normalizeContentType(req.GetFormParam("contentType")),
		UseTemplate:       req.GetFormParam("useTemplate") == "on",
		Template:          req.GetFormParam("template"),
		IsUse:             req.GetFormParam("isUse") == "on",
		IsOnlySendByError: req.GetFormParam("onlyByError") == "on",
		Headers:           collectHeaders(req),
	}
	if st.Method == "" {
		st.Method = "GET"
	}
	tokenId64, _ := strconv.ParseUint(req.GetFormParam("tokenId"), 10, 32)
	st.TokenId = uint(tokenId64)
	// Parse the id up front so a validation failure re-renders the edit form as an edit
	if idStr := req.GetFormParam("id"); idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err == nil {
			st.Id = uint(id)
		}
	}

	// The schedule's owner, not the caller, owns the token. Resolve the row before
	// validating the token so an admin editing someone else's schedule can neither be
	// offered the admin's tokens nor re-point the schedule at one.
	ownerId := user.Id
	var existing *model.Schedule
	fail := func(msg string) error {
		data := &ScheduleFormData{
			Base:        p.base(req, user, "schedule.addSchedule", false),
			Id:          st.Id,
			Name:        st.Name,
			Cron:        st.Cron,
			Url:         st.Url,
			TestUrl:     p.scheduleTestUrl(),
			Method:      st.Method,
			Body:        st.Body,
			ContentType: st.ContentType,
			UseTemplate: st.UseTemplate,
			Template:    st.Template,
			OnlyByError: st.IsOnlySendByError,
			IsUse:       st.IsUse,
			Headers:     st.Headers,
		}
		if data.Id != 0 {
			data.Title = translatorFor(data.Lang).T("schedule.editSchedule")
		}
		data.Flash = Flash{Err: msg}
		if len(data.Headers) == 0 {
			data.Headers = []*model.Header{{}, {}}
		}
		_ = p.fillTokenOptions(data, user, ownerId, st.TokenId)
		return renderPageStatus(view, "schedule_form", data, http.StatusBadRequest)
	}
	if st.Id != 0 {
		loaded, err := p.scheduleOwnedBy(st.Id, user)
		if err != nil || loaded == nil {
			return fail("schedule not found")
		}
		existing = loaded
		ownerId = resourceOwnerId(loaded.UserId, user.Id)
	}
	if st.Name == "" {
		return fail("common.failed")
	}
	if st.Method == "POST" && st.Body == "" {
		return fail("request body is required")
	}
	if err := util.ValidateURL(st.Url); err != nil {
		return fail(err.Error())
	}
	if err := util.ParserCron(st.Cron); err != nil {
		return fail(err.Error())
	}
	if st.TokenId == 0 {
		return fail("schedule.associatedToken")
	}
	// ScheduleService only checks existence, so the form enforces that the token is in
	// use — and, for a regular user, that it belongs to the schedule's owner. An admin
	// may point the schedule at any user's token.
	if user.IsAdmin {
		if t, err := p.tokenModel.Query().Where("id = ? AND state = 0", st.TokenId).One(); err != nil || t == nil {
			return fail("token not found")
		}
	} else if owned, err := p.activeTokenOwnedBy(ownerId, st.TokenId); err != nil || owned == nil {
		return fail("token not found")
	}

	if existing != nil {
		existing.Name = st.Name
		existing.TokenId = st.TokenId
		existing.Cron = st.Cron
		existing.Url = st.Url
		existing.Method = st.Method
		existing.Body = st.Body
		existing.ContentType = st.ContentType
		existing.UseTemplate = st.UseTemplate
		existing.Template = st.Template
		existing.IsUse = st.IsUse
		existing.IsOnlySendByError = st.IsOnlySendByError
		existing.Headers = st.Headers
		if err := p.scheduleSvc.Edit(existing); err != nil {
			return fail(err.Error())
		}
		p.recordOp(req, user.Id, entity.OpUpdate, entity.TargetSchedule, existing.Id, existing.Name)
	} else {
		if err := p.scheduleSvc.Save(st); err != nil {
			return fail(err.Error())
		}
		p.recordOp(req, user.Id, entity.OpCreate, entity.TargetSchedule, st.Id, st.Name)
	}
	return redirect(view, backTo("/schedule", "common.success", "", ""))
}
