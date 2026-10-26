package webui

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/smtp"
)

// LogListData is the mail-log list page.
type LogListData struct {
	Base
	Rows      []*model.Log
	Pager     *Pager
	SearchKey string
	AdminOnly bool
}

func (p *Pages) logList(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	page := listPage(req)
	searchKey := req.Query("searchKey")
	adminOnly := boolQuery(req, "adminOnly")

	query := p.logModel.Query()
	if user.IsAdmin {
		if adminOnly {
			query = query.Where("user_id IN (SELECT id FROM t_user WHERE is_admin = ?)", true)
		}
	} else {
		query = query.Where("user_id = ?", user.Id)
	}
	if searchKey != "" {
		query = query.Where("(name LIKE ? OR mail LIKE ? OR subject LIKE ?)", "%"+searchKey+"%", "%"+searchKey+"%", "%"+searchKey+"%")
	}
	result, err := query.Order("id desc").PageForWeb(page)
	if err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	for _, one := range result.List {
		if one != nil {
			one.StatusStr = entity.StatusText(one.Status)
		}
	}
	if user.IsAdmin {
		fillNames(p.userService, result.List,
			func(l *model.Log) uint { return l.UserId },
			func(l *model.Log, n string) { l.UserName = n })
	}
	q := queryOf(req, "searchKey", "adminOnly", "pageSize")
	data := &LogListData{
		Base:      p.base(req, user, "log.systemLogs", true),
		Rows:      result.List,
		Pager:     newPager(page.PageNo, page.PageSize, result.Total, "/logs", q),
		SearchKey: searchKey,
		AdminOnly: adminOnly,
	}
	return p.render(req, "logs_list", data)
}

// LogAttachment is one downloadable file recorded on a log entry.
type LogAttachment struct {
	Name string
	Href string
}

// LogDetailData is the log detail page.
type LogDetailData struct {
	Base
	Entry       *model.Log
	Attachments []LogAttachment
}

func (p *Pages) logDetail(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	id, err := idParam(req, "id")
	if err != nil {
		return p.renderErr(req, http.StatusBadRequest, "common.failed", err.Error())
	}
	query := p.logModel.Query().Where("id = ?", id)
	if !user.IsAdmin {
		query = query.Where("user_id = ?", user.Id)
	}
	one, err := query.One()
	if err != nil {
		return p.renderErr(req, 0, "common.failed", err.Error())
	}
	if one == nil {
		return p.renderErr(req, http.StatusNotFound, "common.failed", "log not found")
	}
	one.StatusStr = entity.StatusText(one.Status)

	data := &LogDetailData{Base: p.base(req, user, "log.logDetail", false), Entry: one}
	// The stored paths are resolved server-side by /api/download; the page only carries
	// the recorded names.
	var files []*smtp.File
	if one.Files != "" {
		_ = json.Unmarshal([]byte(one.Files), &files)
	}
	for _, f := range files {
		if f == nil || f.Name == "" {
			continue
		}
		v := url.Values{}
		v.Set("logId", strconv.Itoa(int(one.Id)))
		v.Set("file", f.Name)
		data.Attachments = append(data.Attachments, LogAttachment{
			Name: f.Name,
			Href: "/api/download?" + v.Encode(),
		})
	}
	return p.render(req, "logs_detail", data)
}
