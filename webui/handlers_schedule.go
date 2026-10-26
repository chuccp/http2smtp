package webui

import (
	"net/http"
	"strconv"

	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/util"
)

// ScheduleListData is the scheduled-tasks list page.
type ScheduleListData struct {
	Base
	Rows      []*model.Schedule
	Pager     *Pager
	Name      string
	AdminOnly bool
}

func (p *Pages) scheduleList(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	page := listPage(req)
	name := req.Query("name")
	adminOnly := boolQuery(req, "adminOnly")

	result, err := p.scheduleSvc.GetPage(page, user.Id, user.IsAdmin, name, adminOnly)
	if err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	pb, ok2 := result.(*web.PageAble[*model.Schedule])
	if !ok2 || pb == nil {
		return p.renderErr(req, 0, "common.failed", "unexpected page result")
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
	return p.render(req, "schedule_list", data)
}

// ScheduleFormData is the scheduled-task create/edit form.
type ScheduleFormData struct {
	Base
	Id           uint
	Name         string
	Cron         string
	Url          string
	Method       string
	Body         string
	UseTemplate  bool
	Template     string
	OnlyByError  bool
	IsUse        bool
	Headers      []*model.Header
	TokenOptions []SelectOption
}

func (p *Pages) scheduleForm(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	data := &ScheduleFormData{
		Base:    p.base(req, user, "schedule.addSchedule", false),
		Method:  "GET",
		IsUse:   true,
		Headers: []*model.Header{{}, {}},
	}
	var selectedToken uint
	if s := req.Param("id"); s != "" {
		id, err := idParam(req, "id")
		if err != nil {
			return p.renderErr(req, http.StatusBadRequest, "common.failed", err.Error())
		}
		row, err := p.scheduleOwnedBy(id, user)
		if err != nil || row == nil {
			return p.renderErr(req, http.StatusNotFound, "common.failed", "schedule not found")
		}
		selectedToken = row.TokenId
		data.Id = row.Id
		data.Name = row.Name
		data.Cron = row.Cron
		data.Url = row.Url
		data.Method = row.Method
		data.Body = row.Body
		data.UseTemplate = row.UseTemplate
		data.Template = row.Template
		data.OnlyByError = row.IsOnlySendByError
		data.IsUse = row.IsUse
		if len(row.Headers) > 0 {
			data.Headers = row.Headers
		}
		data.Title = translatorFor(data.Lang).T("schedule.editSchedule")
	}
	if err := p.fillTokenOptions(data, user, selectedToken); err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	return p.render(req, "schedule_form", data)
}

// fillTokenOptions loads the caller's enabled tokens into the form's select; only
// in-use tokens may drive a schedule.
func (p *Pages) fillTokenOptions(data *ScheduleFormData, user *model.User, selected uint) error {
	tokens, err := p.tokenModel.Query().Where("user_id = ? AND state = 0", user.Id).Order("id desc").All()
	if err != nil {
		return err
	}
	for _, t := range tokens {
		data.TokenOptions = append(data.TokenOptions, SelectOption{
			Value:    strconv.Itoa(int(t.Id)),
			Label:    t.Name,
			Selected: t.Id == selected,
		})
	}
	return nil
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

func (p *Pages) scheduleSave(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	st := &model.Schedule{
		UserId:            user.Id,
		Name:              req.GetFormParam("name"),
		Cron:              req.GetFormParam("cron"),
		Url:               req.GetFormParam("url"),
		Method:            req.GetFormParam("method"),
		Body:              req.GetFormParam("body"),
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

	fail := func(msg string) (any, error) {
		data := &ScheduleFormData{
			Base:        p.base(req, user, "schedule.addSchedule", false),
			Id:          st.Id,
			Name:        st.Name,
			Cron:        st.Cron,
			Url:         st.Url,
			Method:      st.Method,
			Body:        st.Body,
			UseTemplate: st.UseTemplate,
			Template:    st.Template,
			OnlyByError: st.IsOnlySendByError,
			IsUse:       st.IsUse,
			Headers:     st.Headers,
		}
		data.Flash = Flash{Err: msg}
		if len(data.Headers) == 0 {
			data.Headers = []*model.Header{{}, {}}
		}
		_ = p.fillTokenOptions(data, user, st.TokenId)
		req.Response().WriteStatus(http.StatusBadRequest)
		return p.render(req, "schedule_form", data)
	}
	if st.Name == "" {
		return fail("common.failed")
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
	// The token must belong to the caller — ScheduleService only checks existence.
	owned, err := p.tokenModel.Query().Where("id = ? AND user_id = ?", st.TokenId, user.Id).One()
	if err != nil || owned == nil {
		return fail("token not found")
	}

	if st.Id != 0 {
		exist, err := p.scheduleOwnedBy(st.Id, user)
		if err != nil || exist == nil {
			return fail("schedule not found")
		}
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
		exist.Headers = st.Headers
		if err := p.scheduleSvc.Edit(exist); err != nil {
			return fail(err.Error())
		}
	} else {
		if err := p.scheduleSvc.Save(st); err != nil {
			return fail(err.Error())
		}
	}
	return p.redirect(req, backTo("/schedule", "common.success", "", ""))
}
