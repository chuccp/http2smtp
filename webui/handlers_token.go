package webui

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/util"
)

// TokenListData is the token list page.
type TokenListData struct {
	Base
	Rows      []*model.Token
	Pager     *Pager
	Name      string
	AdminOnly bool
}

func (p *Pages) tokenList(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	page := listPage(req)
	name := req.Query("name")
	adminOnly := boolQuery(req, "adminOnly")

	result, err := p.tokenService.GetPage(page, user.Id, user.IsAdmin, name, adminOnly)
	if err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	pb, ok2 := result.(*web.PageAble[*model.Token])
	if !ok2 || pb == nil {
		return p.renderErr(req, 0, "common.failed", "unexpected page result")
	}
	if user.IsAdmin {
		fillNames(p.userService, pb.List,
			func(t *model.Token) uint { return t.UserId },
			func(t *model.Token, n string) { t.UserName = n })
	}
	q := queryOf(req, "name", "adminOnly", "pageSize")
	data := &TokenListData{
		Base:      p.base(req, user, "token.tokenManagement", true),
		Rows:      pb.List,
		Pager:     newPager(page.PageNo, page.PageSize, pb.Total, "/tokens", q),
		Name:      name,
		AdminOnly: adminOnly,
	}
	return p.render(req, "token_list", data)
}

// SelectOption is one entry of a form <select>.
type SelectOption struct {
	Value    string
	Label    string
	Selected bool
}

// TokenFormData is the token create/edit form.
type TokenFormData struct {
	Base
	Id            uint
	Name          string
	TokenValue    string
	IsNew         bool
	Subject       string
	State         uint8
	SmtpOptions   []SelectOption
	MailOptions   []SelectOption
	SelectedMails map[uint]bool
}

func (p *Pages) tokenForm(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	data := &TokenFormData{
		Base:          p.base(req, user, "token.generateToken", false),
		IsNew:         true,
		TokenValue:    util.GenerateRandomString(32, ""),
		SelectedMails: map[uint]bool{},
	}
	var selectedSmtp uint
	if s := req.Param("id"); s != "" {
		id, err := idParam(req, "id")
		if err != nil {
			return p.renderErr(req, http.StatusBadRequest, "common.failed", err.Error())
		}
		row, err := p.tokenOwnedBy(id, user)
		if err != nil || row == nil {
			return p.renderErr(req, http.StatusNotFound, "common.failed", "token not found")
		}
		data.Id = row.Id
		data.Name = row.Name
		data.TokenValue = row.Token
		data.Subject = row.Subject
		data.State = row.State
		data.IsNew = false
		selectedSmtp = row.SMTPId
		for _, id := range util.StringToUintIds(row.ReceiveEmailIds) {
			data.SelectedMails[id] = true
		}
		data.Title = translatorFor(data.Lang).T("token.editToken")
	}

	if err := p.fillTokenFormOptions(data, user, selectedSmtp); err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	return p.render(req, "token_form", data)
}

// fillTokenFormOptions lists the caller's own servers and contacts — the API enforces the
// same ownership, the form just never offers what would be refused. Both the form and
// a failed save (which re-renders it) need this.
func (p *Pages) fillTokenFormOptions(data *TokenFormData, user *model.User, selectedSmtp uint) error {
	smtps, err := p.smtpModel.Query().Where("user_id = ?", user.Id).Order("id desc").All()
	if err != nil {
		return err
	}
	for _, s := range smtps {
		data.SmtpOptions = append(data.SmtpOptions, SelectOption{
			Value:    strconv.Itoa(int(s.Id)),
			Label:    s.Name + " (" + s.Host + ")",
			Selected: s.Id == selectedSmtp,
		})
	}
	mails, err := p.mailModel.Query().Where("user_id = ?", user.Id).Order("id desc").All()
	if err != nil {
		return err
	}
	for _, m := range mails {
		data.MailOptions = append(data.MailOptions, SelectOption{
			Value:    strconv.Itoa(int(m.Id)),
			Label:    m.Name + " <" + m.Mail + ">",
			Selected: data.SelectedMails[m.Id],
		})
	}
	return nil
}

func (p *Pages) tokenOwnedBy(id uint, user *model.User) (*model.Token, error) {
	if user.IsAdmin {
		return p.tokenService.GetOneAny(id)
	}
	return p.tokenService.GetOne(id, user.Id)
}

// resolveTokenState maps a form switch onto the state the caller may set: the switch is
// only in-use/disabled-by-me, and a plain owner may not undo an admin disable.
func resolveTokenState(requested uint8, current uint8, isAdmin bool) uint8 {
	if requested == entity.TokenStateInUse {
		if !isAdmin && current == entity.TokenStateAdminDisabled {
			return entity.TokenStateAdminDisabled
		}
		return entity.TokenStateInUse
	}
	if isAdmin {
		return entity.TokenStateAdminDisabled
	}
	return entity.TokenStateUserDisabled
}

func (p *Pages) tokenSave(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	// The multi-select posts the same field repeatedly; collect them all.
	if err := req.Request().ParseForm(); err != nil {
		return nil, err
	}
	ids := req.Request().Form["receiveEmailIds"]
	idsDedup := util.RemoveDuplicates(ids)

	smtpId64, _ := strconv.ParseUint(req.GetFormParam("smtpId"), 10, 32)
	smtpId := uint(smtpId64)
	state := entity.TokenStateInUse
	if req.GetFormParam("state") == "off" {
		state = entity.TokenStateUserDisabled
	}

	data := &TokenFormData{
		Base:          p.base(req, user, "token.generateToken", false),
		Name:          req.GetFormParam("name"),
		TokenValue:    req.GetFormParam("token"),
		Subject:       req.GetFormParam("subject"),
		State:         state,
		IsNew:         req.GetFormParam("id") == "",
		SelectedMails: map[uint]bool{},
	}
	for _, s := range idsDedup {
		if n, err := strconv.ParseUint(s, 10, 32); err == nil {
			data.SelectedMails[uint(n)] = true
		}
	}
	fail := func(msg string) (any, error) {
		data.Flash = Flash{Err: msg}
		if err := p.fillTokenFormOptions(data, user, smtpId); err != nil {
			return p.renderErr(req, 0, "common.failed", err.Error())
		}
		req.Response().WriteStatus(http.StatusBadRequest)
		return p.render(req, "token_form", data)
	}

	if data.Name == "" || data.TokenValue == "" || smtpId == 0 {
		return fail("common.failed")
	}
	// Pointing at a server the caller does not hold would send through someone else's
	// credentials — the same rule the API applies.
	owned, err := p.smtpModel.Query().Where("id = ? AND user_id = ?", smtpId, user.Id).One()
	if err != nil || owned == nil {
		return fail("smtp not found")
	}
	for _, n := range idsDedup {
		mid, _ := strconv.ParseUint(n, 10, 32)
		m, err := p.mailModel.Query().Where("id = ? AND user_id = ?", uint(mid), user.Id).One()
		if err != nil || m == nil {
			return fail("mail not found")
		}
	}
	recvIds := strings.Join(idsDedup, ",")

	if idStr := req.GetFormParam("id"); idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return fail("invalid id")
		}
		exist, err := p.tokenOwnedBy(uint(id), user)
		if err != nil || exist == nil {
			return fail("token not found")
		}
		data.Id = exist.Id
		exist.Name = data.Name
		exist.Subject = data.Subject
		exist.SMTPId = smtpId
		exist.ReceiveEmailIds = util.DeduplicateIds(recvIds)
		if data.TokenValue != exist.Token {
			exist.Token = data.TokenValue
		}
		exist.State = resolveTokenState(state, exist.State, user.IsAdmin)
		exist.UpdateTime = time.Now()
		if err := p.tokenModel.Save(exist); err != nil {
			return fail(err.Error())
		}
	} else {
		row := &model.Token{
			UserId:          user.Id,
			Name:            data.Name,
			Token:           data.TokenValue,
			Subject:         data.Subject,
			SMTPId:          smtpId,
			ReceiveEmailIds: util.DeduplicateIds(recvIds),
			State:           resolveTokenState(state, entity.TokenStateInUse, user.IsAdmin),
		}
		if err := p.tokenModel.Save(row); err != nil {
			return fail(err.Error())
		}
	}
	return p.redirect(req, backTo("/tokens", "common.success", "", ""))
}

// TokenSendData is the "send test mail with this token" page.
type TokenSendData struct {
	Base
	TokenId    uint
	TokenName  string
	SmtpStr    string
	Recipients []string
	Subject    string
}

func (p *Pages) tokenSendForm(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	id, err := idParam(req, "id")
	if err != nil {
		return p.renderErr(req, http.StatusBadRequest, "common.failed", err.Error())
	}
	row, err := p.tokenOwnedBy(id, user)
	if err != nil || row == nil {
		return p.renderErr(req, http.StatusNotFound, "common.failed", "token not found")
	}
	data := &TokenSendData{
		Base:      p.base(req, user, "token.sendTestMail", false),
		TokenId:   row.Id,
		TokenName: row.Name,
		Subject:   row.Subject,
	}
	if row.SMTP != nil {
		data.SmtpStr = row.SMTP.Name
	}
	for _, m := range row.ReceiveEmails {
		if m != nil {
			data.Recipients = append(data.Recipients, m.Mail)
		}
	}
	return p.render(req, "token_send", data)
}
