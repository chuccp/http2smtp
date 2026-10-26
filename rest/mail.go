package rest

import (
	"errors"
	"net/mail"
	"strconv"
	"time"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
)

type Mail struct {
	context     *core.Context
	mailModel   *model.MailModel
	userService *service.UserService
}

func (m *Mail) getOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, m.context)
	if user == nil {
		return nil, err
	}
	one, err := m.mailModel.Query().Where("id = ? AND user_id = ?", uint(atoi), user.Id).One()
	if err != nil {
		return nil, err
	}
	return one, nil
}

func (m *Mail) deleteOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, m.context)
	if user == nil {
		return nil, err
	}
	exist, err := m.mailModel.Query().Where("id = ? AND user_id = ?", uint(atoi), user.Id).One()
	if err != nil {
		return nil, err
	}
	if exist == nil {
		return nil, errors.New("mail not found")
	}
	err = m.mailModel.DeleteByPK(uint(atoi))
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}
func (m *Mail) getPage(req *web.Request) (any, error) {
	page, err := req.Page()
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, m.context)
	if user == nil {
		return nil, err
	}
	name := req.Query("name")
	mail := req.Query("mail")
	adminOnly := req.Query("adminOnly") == "true"
	var result *web.PageAble[*model.Mail]
	if user.IsAdmin {
		query := m.mailModel.Query()
		if adminOnly {
			query = query.Where("user_id IN (SELECT id FROM t_user WHERE is_admin = ?)", true)
		}
		if name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		if mail != "" {
			query = query.Where("mail LIKE ?", "%"+mail+"%")
		}
		result, err = query.Order("id desc").PageForWeb(page)
	} else {
		query := m.mailModel.Query().Where("user_id = ?", user.Id)
		if name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		if mail != "" {
			query = query.Where("mail LIKE ?", "%"+mail+"%")
		}
		result, err = query.Order("id desc").PageForWeb(page)
	}
	if err != nil {
		return nil, err
	}
	if user.IsAdmin {
		userIds := make([]uint, 0)
		for _, item := range result.List {
			if item.UserId > 0 {
				userIds = append(userIds, item.UserId)
			}
		}
		m.userService.FillUserNames(userIds, func(uid uint, name string) {
			for _, item := range result.List {
				if item.UserId == uid {
					item.UserName = name
				}
			}
		})
	}
	return result, nil
}
func (m *Mail) postOne(req *web.Request) (any, error) {
	var st model.Mail
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	_, err = mail.ParseAddress(st.Mail)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, m.context)
	if user == nil {
		return nil, err
	}
	// Create is create: a client-supplied id must not turn this into an UPDATE of an
	// existing row (gorm Save updates when the primary key is set), which would let any
	// signed-in user overwrite another account's contact and take ownership of it.
	st.Id = 0
	st.UserId = user.Id
	err = m.mailModel.Save(&st)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}
func (m *Mail) putOne(req *web.Request) (any, error) {
	var st model.Mail
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}

	_, err = mail.ParseAddress(st.Mail)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, m.context)
	if user == nil {
		return nil, err
	}
	exist, err := m.mailModel.FindByPK(st.Id)
	if err != nil {
		return nil, err
	}
	// Report "not found" rather than "forbidden" so we don't leak existence
	if exist == nil || (exist.UserId != user.Id && !user.IsAdmin) {
		return nil, errors.New("mail not found")
	}
	// Keep the original owner: an admin editing someone else's row must not take it over
	exist.Name = st.Name
	exist.Mail = st.Mail
	exist.UpdateTime = time.Now()
	err = m.mailModel.Save(exist)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}
func (m *Mail) Init(context *core.Context) error {
	m.context = context
	context.Get("/mail/:id", m.getOne).WithMeta(auth2.WithLogin())
	context.Delete("/mail/:id", m.deleteOne).WithMeta(auth2.WithLogin())
	context.Get("/mail", m.getPage).WithMeta(auth2.WithLogin())
	context.Post("/mail", m.postOne).WithMeta(auth2.WithLogin())
	context.Put("/mail", m.putOne).WithMeta(auth2.WithLogin())
	m.mailModel = core.GetModel[*model.MailModel](context)
	m.userService = wf.GetService[*service.UserService](context)
	return nil
}
