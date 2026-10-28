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
	Rows             []*model.Log
	Pager            *Pager
	SearchKey        string
	AdminOnly        bool
	HasAttachment    bool
	AttachmentCounts map[uint]int
}

func (p *Pages) logList(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireUser(view)
	if !ok {
		return nil
	}
	page := listPage(req)
	searchKey := req.Query("searchKey")
	adminOnly := boolQuery(req, "adminOnly")
	hasAttachment := boolQuery(req, "hasAttachment")

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
	if hasAttachment {
		// New entries have attachment rows; older ones only carry the files JSON.
		query = query.Where("(id IN (SELECT log_id FROM t_attachment) OR files <> ?)", "")
	}
	result, err := query.Order("id desc").PageForWeb(page)
	if err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
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
	counts := p.attachmentCounts(result.List)
	q := queryOf(req, "searchKey", "adminOnly", "hasAttachment", "pageSize")
	data := &LogListData{
		Base:             p.base(req, user, "log.systemLogs", true),
		Rows:             result.List,
		Pager:            newPager(page.PageNo, page.PageSize, result.Total, "/logs", q),
		SearchKey:        searchKey,
		AdminOnly:        adminOnly,
		HasAttachment:    hasAttachment,
		AttachmentCounts: counts,
	}
	view.View("logs_list").Layout(layoutFor("logs_list")).Data(data)
	return nil
}

// attachmentCounts returns how many files each listed log carries. Logs from before the
// attachment table fall back to their files JSON.
func (p *Pages) attachmentCounts(rows []*model.Log) map[uint]int {
	counts := make(map[uint]int, len(rows))
	ids := make([]uint, 0, len(rows))
	for _, one := range rows {
		if one != nil {
			ids = append(ids, one.Id)
		}
	}
	if len(ids) > 0 {
		if attachments, err := p.attachmentModel.FindByLogIds(ids); err == nil {
			for _, attachment := range attachments {
				if attachment != nil {
					counts[attachment.LogId]++
				}
			}
		}
	}
	for _, one := range rows {
		if one == nil || counts[one.Id] > 0 || one.Files == "" {
			continue
		}
		var files []*smtp.File
		if err := json.Unmarshal([]byte(one.Files), &files); err == nil {
			counts[one.Id] = len(files)
		}
	}
	return counts
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

func (p *Pages) logDetail(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireUser(view)
	if !ok {
		return nil
	}
	id, err := idParam(req, "id")
	if err != nil {
		return p.renderErr(view, http.StatusBadRequest, "common.failed", err.Error())
	}
	query := p.logModel.Query().Where("id = ?", id)
	if !user.IsAdmin {
		query = query.Where("user_id = ?", user.Id)
	}
	one, err := query.One()
	if err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	if one == nil {
		return p.renderErr(view, http.StatusNotFound, "common.failed", "log not found")
	}
	one.StatusStr = entity.StatusText(one.Status)
	if user.IsAdmin {
		p.userService.FillUserNames([]uint{one.UserId}, func(uid uint, name string) {
			one.UserName = name
		})
	}

	data := &LogDetailData{Base: p.base(req, user, "log.logDetail", false), Entry: one}
	// The stored paths are resolved server-side by /api/download; the page only carries
	// the recorded ids and names. Logs written before the attachment table existed have
	// no rows and fall back to their files JSON.
	attachments, err := p.attachmentModel.FindByLogId(one.Id)
	if err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	if len(attachments) > 0 {
		for _, attachment := range attachments {
			if attachment == nil || attachment.Name == "" {
				continue
			}
			v := url.Values{}
			v.Set("id", strconv.Itoa(int(attachment.Id)))
			data.Attachments = append(data.Attachments, LogAttachment{
				Name: attachment.Name,
				Href: "/api/download?" + v.Encode(),
			})
		}
	} else {
		var files []*smtp.File
		if one.Files != "" {
			_ = json.Unmarshal([]byte(one.Files), &files)
		}
		for i, f := range files {
			if f == nil || f.Name == "" {
				continue
			}
			// Address the recorded position, not the name: two attachments may share one.
			v := url.Values{}
			v.Set("logId", strconv.Itoa(int(one.Id)))
			v.Set("index", strconv.Itoa(i))
			data.Attachments = append(data.Attachments, LogAttachment{
				Name: f.Name,
				Href: "/api/download?" + v.Encode(),
			})
		}
	}
	view.View("logs_detail").Layout(layoutFor("logs_detail")).Data(data)
	return nil
}
