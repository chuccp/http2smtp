package webui

import (
	"net/http"
	"net/mail"
	"strconv"
	"time"

	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
)

// MailListData is the recipients list page.
type MailListData struct {
	Base
	Rows      []*model.Mail
	Pager     *Pager
	Name      string
	MailAddr  string
	AdminOnly bool
}

func (p *Pages) mailList(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	page := listPage(req)
	name := req.Query("name")
	mailAddr := req.Query("mail")
	adminOnly := boolQuery(req, "adminOnly")

	query := p.mailModel.Query()
	if user.IsAdmin {
		if adminOnly {
			query = query.Where("user_id IN (SELECT id FROM t_user WHERE is_admin = ?)", true)
		}
	} else {
		query = query.Where("user_id = ?", user.Id)
	}
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if mailAddr != "" {
		query = query.Where("mail LIKE ?", "%"+mailAddr+"%")
	}
	result, err := query.Order("id desc").PageForWeb(page)
	if err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	if user.IsAdmin {
		fillNames(p.userService, result.List,
			func(row *model.Mail) uint { return row.UserId },
			func(row *model.Mail, n string) { row.UserName = n })
	}
	q := queryOf(req, "name", "mail", "adminOnly", "pageSize")
	data := &MailListData{
		Base:      p.base(req, user, "mail.mailManagement", true),
		Rows:      result.List,
		Pager:     newPager(page.PageNo, page.PageSize, result.Total, "/mail", q),
		Name:      name,
		MailAddr:  mailAddr,
		AdminOnly: adminOnly,
	}
	return p.render(req, "mail_list", data)
}

// MailFormData is the recipient create/edit form.
type MailFormData struct {
	Base
	Id   uint
	Name string
	Mail string
}

func (p *Pages) mailForm(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	data := &MailFormData{Base: p.base(req, user, "mail.addRecipient", false)}
	if s := req.Param("id"); s != "" {
		id, err := idParam(req, "id")
		if err != nil {
			return p.renderErr(req, http.StatusBadRequest, "common.failed", err.Error())
		}
		row, err := p.mailOwnedBy(id, user)
		if err != nil || row == nil {
			return p.renderErr(req, http.StatusNotFound, "common.failed", "mail not found")
		}
		data.Id = row.Id
		data.Name = row.Name
		data.Mail = row.Mail
		data.Title = translatorFor(data.Lang).T("mail.editRecipient")
	}
	return p.render(req, "mail_form", data)
}

func (p *Pages) mailOwnedBy(id uint, user *model.User) (*model.Mail, error) {
	if user.IsAdmin {
		return p.mailModel.FindByPK(id)
	}
	return p.mailModel.Query().Where("id = ? AND user_id = ?", id, user.Id).One()
}

func (p *Pages) mailSave(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	data := &MailFormData{Base: p.base(req, user, "mail.addRecipient", false)}
	data.Name = req.GetFormParam("name")
	data.Mail = req.GetFormParam("mail")
	fail := func(msg string) (any, error) {
		data.Flash = Flash{Err: msg}
		req.Response().WriteStatus(http.StatusBadRequest)
		return p.render(req, "mail_form", data)
	}
	if data.Name == "" {
		return fail("common.failed")
	}
	if _, err := mail.ParseAddress(data.Mail); err != nil {
		return fail(err.Error())
	}
	if idStr := req.GetFormParam("id"); idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return fail("invalid id")
		}
		exist, err := p.mailOwnedBy(uint(id), user)
		if err != nil {
			return fail(err.Error())
		}
		if exist == nil {
			return fail("mail not found")
		}
		data.Id = exist.Id
		exist.Name = data.Name
		exist.Mail = data.Mail
		exist.UpdateTime = time.Now()
		if err := p.mailModel.Save(exist); err != nil {
			return fail(err.Error())
		}
	} else {
		row := &model.Mail{UserId: user.Id, Name: data.Name, Mail: data.Mail}
		if err := p.mailModel.Save(row); err != nil {
			return fail(err.Error())
		}
	}
	return p.redirect(req, backTo("/mail", "common.success", "", ""))
}
