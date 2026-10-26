package webui

import (
	"net/http"
	"strconv"

	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/util"
)

// UserListData is the user management list page (admin only).
type UserListData struct {
	Base
	Rows  []*model.User
	Pager *Pager
}

func (p *Pages) userList(req *web.Request) (any, error) {
	user, ok := p.requireAdmin(req)
	if !ok {
		return nil, nil
	}
	page := listPage(req)
	result, err := p.userService.GetPage(page)
	if err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	pb, ok2 := result.(*web.PageAble[*model.User])
	if !ok2 || pb == nil {
		return p.renderErr(req, 0, "common.failed", "unexpected page result")
	}
	q := queryOf(req, "pageSize")
	data := &UserListData{
		Base:  p.base(req, user, "layout.sidebar.users", true),
		Rows:  pb.List,
		Pager: newPager(page.PageNo, page.PageSize, pb.Total, "/users", q),
	}
	return p.render(req, "users_list", data)
}

// UserFormData is the user create/edit form. Password is never echoed; blank on edit
// keeps the stored hash.
type UserFormData struct {
	Base
	Id      uint
	Name    string
	IsAdmin bool
	IsUse   bool
	Self    bool
}

func (p *Pages) userForm(req *web.Request) (any, error) {
	user, ok := p.requireAdmin(req)
	if !ok {
		return nil, nil
	}
	data := &UserFormData{Base: p.base(req, user, "layout.sidebar.users", false), IsUse: true}
	if s := req.Param("id"); s != "" {
		id, err := idParam(req, "id")
		if err != nil {
			return p.renderErr(req, http.StatusBadRequest, "common.failed", err.Error())
		}
		row, err := p.userService.GetOne(id)
		if err != nil {
			return p.renderErr(req, http.StatusNotFound, "common.failed", err.Error())
		}
		data.Id = row.Id
		data.Name = row.Name
		data.IsAdmin = row.IsAdmin
		data.IsUse = row.IsUse
		data.Self = row.Id == user.Id
	}
	return p.render(req, "users_form", data)
}

func (p *Pages) userSave(req *web.Request) (any, error) {
	caller, ok := p.requireAdmin(req)
	if !ok {
		return nil, nil
	}
	data := &UserFormData{
		Base:    p.base(req, caller, "layout.sidebar.users", false),
		Name:    req.GetFormParam("name"),
		IsAdmin: req.GetFormParam("isAdmin") == "on",
		IsUse:   req.GetFormParam("isUse") == "on",
	}
	fail := func(msg string) (any, error) {
		data.Flash = Flash{Err: msg}
		req.Response().WriteStatus(http.StatusBadRequest)
		return p.render(req, "users_form", data)
	}
	password := req.GetFormParam("password")

	if idStr := req.GetFormParam("id"); idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return fail("invalid id")
		}
		data.Id = uint(id)
		exist, err := p.userService.GetOne(uint(id))
		if err != nil {
			return fail(err.Error())
		}
		data.Self = exist.Id == caller.Id
		// Demoting or disabling the last active admin would lock the instance; the
		// Vue UI blocked editing the admin row entirely, this keeps the milder rule:
		// the admin account may change name/password but not its own flags.
		if exist.IsAdmin {
			data.IsAdmin = true
			data.IsUse = true
		}
		if data.Name == "" {
			return fail("common.failed")
		}
		if password != "" {
			if err := util.ValidatePassword(password); err != nil {
				return fail(err.Error())
			}
		}
		if err := p.userService.UpdateUser(uint(id), data.Name, password, data.IsAdmin, data.IsUse); err != nil {
			return fail(err.Error())
		}
	} else {
		if data.Name == "" || password == "" {
			return fail("common.failed")
		}
		if err := util.ValidatePassword(password); err != nil {
			return fail(err.Error())
		}
		if err := p.userService.CreateUser(data.Name, password, data.IsAdmin, data.IsUse); err != nil {
			return fail(err.Error())
		}
	}
	return p.redirect(req, backTo("/users", "common.success", "", ""))
}

// PasswordData is the change-own-password page.
type PasswordData struct {
	Base
}

func (p *Pages) passwordGet(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	return p.render(req, "password", &PasswordData{Base: p.base(req, user, "user.changePassword", false)})
}

func (p *Pages) passwordPost(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	data := &PasswordData{Base: p.base(req, user, "user.changePassword", false)}
	oldPw := req.GetFormParam("oldPassword")
	newPw := req.GetFormParam("newPassword")
	confirm := req.GetFormParam("confirm")
	fail := func(msgKeyOrText string) (any, error) {
		data.Flash = Flash{Err: msgKeyOrText}
		req.Response().WriteStatus(http.StatusBadRequest)
		return p.render(req, "password", data)
	}
	if oldPw == "" || newPw == "" {
		return fail("common.failed")
	}
	if confirm != newPw {
		return fail(translatorFor(data.Lang).T("setup.passwordsNotMatch"))
	}
	if err := util.ValidatePassword(newPw); err != nil {
		return fail(translatorFor(data.Lang).T("auth.passwordMinLength"))
	}
	if !util.CheckPasswordHash(oldPw, user.Password) {
		return fail("the current password is incorrect")
	}
	if err := p.userService.ChangeOwnPassword(user.Id, newPw); err != nil {
		return fail(err.Error())
	}
	// The salt rotated with the password, so this session is already dead — drop the
	// cookie rather than leave the browser holding a corpse, and sign in again.
	req.Cookie().Delete(entity.UserToken)
	return p.redirect(req, backTo("/login", "auth.passwordChanged", "", ""))
}
