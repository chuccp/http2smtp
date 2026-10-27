package webui

import (
	"net/http"
	"net/mail"
	"strconv"
	"time"

	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
	stmp2 "github.com/chuccp/http2smtp/smtp"
)

// SmtpListData is the SMTP list page.
type SmtpListData struct {
	Base
	Rows      []*model.SMTP
	Pager     *Pager
	Name      string
	Host      string
	Username  string
	AdminOnly bool
}

func (p *Pages) smtpList(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	page := listPage(req)
	name := req.Query("name")
	host := req.Query("host")
	username := req.Query("username")
	adminOnly := boolQuery(req, "adminOnly")

	result, err := p.smtpService.PageForWeb(page, user.Id, user.IsAdmin, adminOnly, name, host, username)
	if err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	q := queryOf(req, "name", "host", "username", "adminOnly", "pageSize")
	data := &SmtpListData{
		Base:      p.base(req, user, "smtp.smtpManagement", true),
		Rows:      result.List,
		Pager:     newPager(page.PageNo, page.PageSize, result.Total, "/smtp", q),
		Name:      name,
		Host:      host,
		Username:  username,
		AdminOnly: adminOnly,
	}
	return p.render(req, "smtp_list", data)
}

// SmtpFormData is the create/edit form. Password is never echoed: blank means
// "keep the stored one" on edit.
type SmtpFormData struct {
	Base
	Id       uint
	Name     string
	Host     string
	Port     int
	Mail     string
	Username string
}

func (p *Pages) smtpForm(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	data := &SmtpFormData{Base: p.base(req, user, "smtp.addSMTP", false), Port: 25}
	if s := req.Param("id"); s != "" {
		id, err := idParam(req, "id")
		if err != nil {
			return p.renderErr(req, http.StatusBadRequest, "common.failed", err.Error())
		}
		row, err := p.smtpOwnedBy(id, user)
		if err != nil || row == nil {
			return p.renderErr(req, http.StatusNotFound, "common.failed", "smtp not found")
		}
		data.Id = row.Id
		data.Name = row.Name
		data.Host = row.Host
		data.Port = row.Port
		data.Mail = row.Mail
		data.Username = row.Username
		data.Title = translatorFor(data.Lang).T("smtp.editSMTP")
	}
	return p.render(req, "smtp_form", data)
}

// smtpOwnedBy loads a row the caller may edit: their own, or any when admin.
func (p *Pages) smtpOwnedBy(id uint, user *model.User) (*model.SMTP, error) {
	if user.IsAdmin {
		return p.smtpModel.FindByPK(id)
	}
	return p.smtpModel.Query().Where("id = ? AND user_id = ?", id, user.Id).One()
}

func (p *Pages) smtpSave(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	formErr := func(data *SmtpFormData, msg string) (any, error) {
		data.Flash = Flash{Err: msg}
		req.Response().WriteStatus(http.StatusBadRequest)
		return p.render(req, "smtp_form", data)
	}
	fill := func(data *SmtpFormData) {
		data.Name = req.GetFormParam("name")
		data.Host = req.GetFormParam("host")
		data.Port, _ = strconv.Atoi(req.GetFormParam("port"))
		data.Mail = req.GetFormParam("mail")
		data.Username = req.GetFormParam("username")
	}

	data := &SmtpFormData{Base: p.base(req, user, "smtp.addSMTP", false)}
	fill(data)
	// Resolve the row before validating, so a failed edit re-renders with its id
	// instead of turning into a create form.
	var existing *model.SMTP
	if idStr := req.GetFormParam("id"); idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return formErr(data, "invalid id")
		}
		existing, err = p.smtpOwnedBy(uint(id), user)
		if err != nil {
			return formErr(data, err.Error())
		}
		if existing == nil {
			return formErr(data, "smtp not found")
		}
		data.Id = existing.Id
		data.Title = translatorFor(data.Lang).T("smtp.editSMTP")
	}
	if data.Name == "" || data.Host == "" || data.Mail == "" {
		return formErr(data, "common.failed")
	}
	if _, err := mail.ParseAddress(data.Mail); err != nil {
		return formErr(data, err.Error())
	}
	password := req.GetFormParam("password")

	if existing != nil {
		existing.Name = data.Name
		existing.Host = data.Host
		existing.Port = data.Port
		existing.Mail = data.Mail
		existing.Username = data.Username
		// Blank means keep: the API's PUT overwrites with whatever it is given, but a
		// form that simply did not echo the password must not wipe it.
		if password != "" {
			existing.Password = password
		}
		existing.UpdateTime = time.Now()
		if err := p.smtpModel.Save(existing); err != nil {
			return formErr(data, err.Error())
		}
	} else {
		row := &model.SMTP{
			UserId:   user.Id,
			Name:     data.Name,
			Host:     data.Host,
			Port:     data.Port,
			Mail:     data.Mail,
			Username: data.Username,
			Password: password,
		}
		if row.Port == 0 {
			row.Port = 25
		}
		if password == "" {
			return formErr(data, "password is required")
		}
		if err := p.smtpModel.Save(row); err != nil {
			return formErr(data, err.Error())
		}
	}
	return p.redirect(req, backTo("/smtp", "common.success", "", ""))
}

// SmtpSendData backs the "send test mail" dialog opened from the SMTP list.
type SmtpSendData struct {
	Base
	Id         uint
	Name       string
	Mail       string
	Subject    string
	Recipients []*model.Mail
}

func (p *Pages) smtpSendForm(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	id, err := idParam(req, "id")
	if err != nil {
		return p.renderErr(req, http.StatusBadRequest, "common.failed", err.Error())
	}
	row, err := p.smtpOwnedBy(id, user)
	if err != nil || row == nil || row.UserId != user.Id {
		// The send API only lets a caller send through their own SMTP server, so an
		// admin opening another user's row would get a form that can never send.
		return p.renderErr(req, http.StatusNotFound, "common.failed", "smtp not found")
	}
	mails, err := p.mailModel.Query().Where("user_id = ?", user.Id).Order("id desc").All()
	if err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	data := &SmtpSendData{
		Base:       p.base(req, user, "smtp.sendTestMail", false),
		Id:         row.Id,
		Name:       row.Name,
		Mail:       row.Mail,
		Recipients: mails,
	}
	data.Subject = data.L.T("smtp.testEmailSubject")
	return p.render(req, "smtp_send", data)
}

// smtpTest runs the connection probe from the form footer without saving anything.
func (p *Pages) smtpTest(req *web.Request) (any, error) {
	if _, ok := p.requireUser(req); !ok {
		return nil, nil
	}
	st := &model.SMTP{
		Host:     req.GetFormParam("host"),
		Name:     req.GetFormParam("name"),
		Mail:     req.GetFormParam("mail"),
		Username: req.GetFormParam("username"),
		Password: req.GetFormParam("password"),
	}
	st.Port, _ = strconv.Atoi(req.GetFormParam("port"))
	if st.Port == 0 {
		st.Port = 25
	}
	if _, err := mail.ParseAddress(st.Mail); err != nil {
		return nil, err
	}
	if err := stmp2.SendTestMsg2(st); err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}
