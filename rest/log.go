package rest

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
	"github.com/chuccp/http2smtp/smtp"
	"go.uber.org/zap"
)

type Log struct {
	context     *core.Context
	logModel    *model.LogModel
	userService *service.UserService
}

// The log page is the instance's whole send record for an admin and the caller's own
// entries for everyone else, matching the other lists: every query below adds the owner
// filter only for a non-admin, and only then is a miss reported as "not found" rather
// than "forbidden" so existence does not leak.

func (l *Log) getOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, l.context)
	if user == nil {
		return nil, err
	}
	query := l.logModel.Query().Where("id = ?", uint(atoi))
	if !user.IsAdmin {
		query = query.Where("user_id = ?", user.Id)
	}
	one, err := query.One()
	if err != nil {
		return nil, err
	}
	if one != nil {
		one.StatusStr = entity.StatusText(one.Status)
		if user.IsAdmin {
			l.userService.FillUserNames([]uint{one.UserId}, func(uid uint, name string) {
				one.UserName = name
			})
		}
	}
	return one, nil
}

func (l *Log) getPage(req *web.Request) (any, error) {
	page, err := req.Page()
	if err != nil {
		return nil, err
	}
	page = capPage(page)
	user, err := auth.User(req, l.context)
	if user == nil {
		return nil, err
	}
	searchKey := req.GetFormParam("searchKey")
	adminOnly := req.Query("adminOnly") == "true"
	query := l.logModel.Query()
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
		return nil, err
	}
	for _, one := range result.List {
		if one != nil {
			one.StatusStr = entity.StatusText(one.Status)
		}
	}
	if user.IsAdmin {
		userIds := make([]uint, 0)
		for _, one := range result.List {
			if one != nil && one.UserId > 0 {
				userIds = append(userIds, one.UserId)
			}
		}
		l.userService.FillUserNames(userIds, func(uid uint, name string) {
			for _, one := range result.List {
				if one != nil && one.UserId == uid {
					one.UserName = name
				}
			}
		})
	}
	return result, nil
}

// deleteOne removes one entry and the attachments it recorded. Deleting entries is an
// admin-only cleanup: the page hides the button from everyone else.
func (l *Log) deleteOne(req *web.Request) (any, error) {
	id, err := strconv.Atoi(req.Param("id"))
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, l.context)
	if user == nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New("admin access required")
	}
	query := l.logModel.Query().Where("id = ?", uint(id))
	if !user.IsAdmin {
		query = query.Where("user_id = ?", user.Id)
	}
	one, err := query.One()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New("log not found")
	}
	if err := l.logModel.DeleteByPK(uint(id)); err != nil {
		return nil, err
	}
	removeAttachments(one.Files)
	recordOp(l.context, user.Id, entity.OpDelete, entity.TargetLog, uint(id), one.Subject)
	return web.Ok("ok"), nil
}

// cleanUp deletes the caller's own entries — all of them when days is 0 — together with
// their attachments, and reports how many went. The scope stays the caller's own entries
// even for an admin: wiping the instance's whole send record should be a decision, not a
// button, and cleaning up after yourself is the thing that actually needs doing.
func (l *Log) cleanUp(req *web.Request) (any, error) {
	user, err := auth.User(req, l.context)
	if user == nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New("admin access required")
	}
	days := req.GetJsonIntValueOrDefault("days", 0)
	query := l.logModel.Query().Where("user_id = ?", user.Id)
	if days > 0 {
		query = query.Where("create_time < ?", time.Now().AddDate(0, 0, -days))
	}
	rows, err := query.All()
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(rows))
	files := make([]string, 0, len(rows))
	for _, one := range rows {
		if one == nil {
			continue
		}
		ids = append(ids, one.Id)
		files = append(files, one.Files)
	}
	if len(ids) == 0 {
		return web.Data(map[string]any{"deleted": 0}), nil
	}
	// Rows first: if that fails nothing has been touched yet, and the files of a row that
	// is still listed must stay on disk.
	if err := l.logModel.Delete().Where("id IN (?)", ids).Delete(); err != nil {
		return nil, err
	}
	for _, one := range files {
		removeAttachments(one)
	}
	recordOp(l.context, user.Id, entity.OpCleanup, entity.TargetLog, 0, "")
	return web.Data(map[string]any{"deleted": len(ids)}), nil
}

// removeAttachments deletes the files a log recorded. Best effort: a file that is already
// gone, or one another entry still points at, must not stop the cleanup.
func removeAttachments(filesJSON string) {
	if filesJSON == "" {
		return
	}
	var files []*smtp.File
	if err := json.Unmarshal([]byte(filesJSON), &files); err != nil {
		return
	}
	for _, file := range files {
		if file != nil && file.FilePath != "" {
			_ = os.Remove(file.FilePath)
		}
	}
}

// downLoad serves an attachment recorded on a log the caller may read: their own, or any
// entry when they are an admin. The path is resolved from the log record, never from
// client input.
func (l *Log) downLoad(req *web.Request) (any, error) {
	logId, err := strconv.Atoi(req.GetFormParam("logId"))
	if err != nil {
		return nil, errors.New("logId is required")
	}
	fileName := req.GetFormParam("file")
	if fileName == "" {
		return nil, errors.New("file is required")
	}
	user, err := auth.User(req, l.context)
	if user == nil {
		return nil, err
	}
	query := l.logModel.Query().Where("id = ?", uint(logId))
	if !user.IsAdmin {
		query = query.Where("user_id = ?", user.Id)
	}
	one, err := query.One()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New("log not found")
	}
	filePath, err := FilePathOf(one.Files, fileName)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(filePath); err != nil {
		return nil, errors.New("attachment file is no longer available")
	}
	log.Info("downLoad", zap.Uint("logId", uint(logId)), zap.String("file", fileName), zap.Uint("userId", user.Id))
	return web.CreateFileResponse(filePath), nil
}

// FilePathOf resolves the stored path of a named attachment from a log's files JSON.
func FilePathOf(filesJSON string, name string) (string, error) {
	if filesJSON == "" {
		return "", errors.New("log has no attachments")
	}
	var files []*smtp.File
	if err := json.Unmarshal([]byte(filesJSON), &files); err != nil {
		return "", errors.New("invalid attachment record")
	}
	for _, file := range files {
		if file != nil && file.Name == name && file.FilePath != "" {
			return file.FilePath, nil
		}
	}
	return "", errors.New("attachment not found")
}

func (l *Log) Init(context *core.Context) error {
	l.context = context
	l.logModel = wf.GetModel[*model.LogModel](context)
	l.userService = wf.GetService[*service.UserService](context)
	context.Get("/api/log/:id", l.getOne).WithMeta(auth2.WithLogin())
	context.Get("/api/log", l.getPage).WithMeta(auth2.WithLogin())
	context.Delete("/api/log/:id", l.deleteOne).WithMeta(auth2.WithLogin())
	context.Post("/api/log/cleanup", l.cleanUp).WithMeta(auth2.WithLogin())
	context.Get("/api/download", l.downLoad).WithMeta(auth2.WithLogin())
	return nil

}
