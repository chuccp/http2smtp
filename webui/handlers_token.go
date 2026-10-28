package webui

import (
	"errors"
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

func (p *Pages) tokenList(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireUser(view)
	if !ok {
		return nil
	}
	page := listPage(req)
	name := req.Query("name")
	adminOnly := boolQuery(req, "adminOnly")

	result, err := p.tokenService.GetPage(page, user.Id, user.IsAdmin, name, adminOnly)
	if err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	pb, ok2 := result.(*web.PageAble[*model.Token])
	if !ok2 || pb == nil {
		return p.renderErr(view, 0, "common.failed", "unexpected page result")
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
	view.View("token_list").Layout(layoutFor("token_list")).Data(data)
	return nil
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

func (p *Pages) tokenForm(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireUser(view)
	if !ok {
		return nil
	}
	data := &TokenFormData{
		Base:          p.base(req, user, "token.generateToken", false),
		IsNew:         true,
		TokenValue:    util.GenerateRandomString(32, ""),
		SelectedMails: map[uint]bool{},
	}
	var selectedSmtp uint
	// A token belongs to its owner, not to whoever edits it: an admin editing another
	// user's token must see (and may only pick) that user's servers and contacts.
	ownerId := user.Id
	if s := req.Param("id"); s != "" {
		id, err := idParam(req, "id")
		if err != nil {
			return p.renderErr(view, http.StatusBadRequest, "common.failed", err.Error())
		}
		row, err := p.tokenOwnedBy(id, user)
		if err != nil || row == nil {
			return p.renderErr(view, http.StatusNotFound, "common.failed", "token not found")
		}
		data.Id = row.Id
		data.Name = row.Name
		data.TokenValue = row.Token
		data.Subject = row.Subject
		data.State = row.State
		data.IsNew = false
		selectedSmtp = row.SMTPId
		// user_id = 0 predates token ownership; keep it editable with the caller's
		// resources rather than stranding the row with no options at all.
		ownerId = resourceOwnerId(row.UserId, user.Id)
		for _, id := range util.StringToUintIds(row.ReceiveEmailIds) {
			data.SelectedMails[id] = true
		}
		data.Title = translatorFor(data.Lang).T("token.editToken")
	}

	if err := p.fillTokenFormOptions(data, ownerId, selectedSmtp); err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	view.View("token_form").Layout(layoutFor("token_form")).Data(data)
	return nil
}

// fillTokenFormOptions lists the token owner's servers and contacts — the API enforces
// the same ownership, the form just never offers what would be refused. Both the form
// and a failed save (which re-renders it) need this.
func (p *Pages) fillTokenFormOptions(data *TokenFormData, ownerId uint, selectedSmtp uint) error {
	smtps, err := p.smtpModel.Query().Where("user_id = ?", ownerId).Order("id desc").All()
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
	mails, err := p.mailModel.Query().Where("user_id = ?", ownerId).Order("id desc").All()
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

// resourceOwnerId keeps legacy rows (user_id = 0) editable by their caller while a
// normal row stays with its original owner, even when an admin edits it.
func resourceOwnerId(rowUserId, callerId uint) uint {
	if rowUserId == 0 {
		return callerId
	}
	return rowUserId
}

// checkTokenResources verifies that the selected SMTP server and recipient rows belong
// to the token owner, not to the caller. An admin editing someone else's token must
// not be able to point it at the admin's own credentials.
func (p *Pages) checkTokenResources(ownerId, smtpId uint, mailIds []string) error {
	owned, err := p.smtpModel.Query().Where("id = ? AND user_id = ?", smtpId, ownerId).One()
	if err != nil {
		return err
	}
	if owned == nil {
		return errors.New("smtp not found")
	}
	for _, n := range mailIds {
		mid, _ := strconv.ParseUint(n, 10, 32)
		m, err := p.mailModel.Query().Where("id = ? AND user_id = ?", uint(mid), ownerId).One()
		if err != nil {
			return err
		}
		if m == nil {
			return errors.New("mail not found")
		}
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

func (p *Pages) tokenSave(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireUser(view)
	if !ok {
		return nil
	}
	// The multi-select posts the same field repeatedly; collect them all.
	if err := req.Request().ParseForm(); err != nil {
		return err
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
	// The selected server and contacts must belong to the token's owner, not to the
	// caller: an admin editing someone else's token must not silently re-point it at
	// the admin's own resources. For a create the owner is the caller; for a legacy
	// user_id = 0 row it stays the caller too, so the row remains editable.
	ownerId := user.Id
	fail := func(msg string) error {
		data.Flash = Flash{Err: msg}
		if err := p.fillTokenFormOptions(data, ownerId, smtpId); err != nil {
			return p.renderErr(view, 0, "common.failed", err.Error())
		}
		return renderPageStatus(view, "token_form", data, http.StatusBadRequest)
	}

	// Resolve the row before validating anything, so a failed edit keeps its id and
	// re-renders as an edit, and so the owner is known when the selections are checked.
	var existing *model.Token
	if idStr := req.GetFormParam("id"); idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return fail("invalid id")
		}
		existing, err = p.tokenOwnedBy(uint(id), user)
		if err != nil || existing == nil {
			return fail("token not found")
		}
		data.Id = existing.Id
		data.IsNew = false
		data.Title = translatorFor(data.Lang).T("token.editToken")
		ownerId = resourceOwnerId(existing.UserId, user.Id)
	}

	if data.Name == "" || data.TokenValue == "" || smtpId == 0 {
		return fail("common.failed")
	}
	// The API enforces the same ownership; pointing at a server/contact the owner
	// does not hold would send through someone else's credentials.
	if err := p.checkTokenResources(ownerId, smtpId, idsDedup); err != nil {
		return fail(err.Error())
	}
	recvIds := strings.Join(idsDedup, ",")

	if existing != nil {
		existing.Name = data.Name
		existing.Subject = data.Subject
		existing.SMTPId = smtpId
		existing.ReceiveEmailIds = util.DeduplicateIds(recvIds)
		if data.TokenValue != existing.Token {
			existing.Token = data.TokenValue
		}
		existing.State = resolveTokenState(state, existing.State, user.IsAdmin)
		existing.UpdateTime = time.Now()
		if err := p.tokenModel.Save(existing); err != nil {
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
	return redirect(view, backTo("/tokens", "common.success", "", ""))
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

func (p *Pages) tokenSendForm(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireUser(view)
	if !ok {
		return nil
	}
	id, err := idParam(req, "id")
	if err != nil {
		return p.renderErr(view, http.StatusBadRequest, "common.failed", err.Error())
	}
	row, err := p.tokenOwnedBy(id, user)
	if err != nil || row == nil {
		return p.renderErr(view, http.StatusNotFound, "common.failed", "token not found")
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
	view.View("token_send").Layout(layoutFor("token_send")).Data(data)
	return nil
}
