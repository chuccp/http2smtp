package rest

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
	context         *core.Context
	logModel        *model.LogModel
	attachmentModel *model.AttachmentModel
	userService     *service.UserService
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
	attachments, err := l.attachmentModel.FindByLogId(uint(id))
	if err != nil {
		return nil, err
	}
	if err := l.logModel.DeleteByPK(uint(id)); err != nil {
		return nil, err
	}
	l.deleteAttachmentRows([]uint{uint(id)})
	removeLogFiles(one.Files, attachments)
	recordOp(l.context, req, user.Id, entity.OpDelete, entity.TargetLog, uint(id), one.Subject)
	return web.Ok("ok"), nil
}

// cleanUp deletes the instance's log entries older than days (0 clears everything),
// together with their attachments, and reports how many went. It is admin-only: the log
// page is the instance's send record, so the admin maintains it as a whole — the same
// scope single-row deletes already have.
func (l *Log) cleanUp(req *web.Request) (any, error) {
	user, err := auth.User(req, l.context)
	if user == nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New("admin access required")
	}
	days := req.GetJsonIntValueOrDefault("days", 0)
	query := l.logModel.Query()
	if days > 0 {
		query = query.Where("create_time < ?", time.Now().AddDate(0, 0, -days))
	}
	rows, err := query.All()
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(rows))
	filesByID := make(map[uint]string, len(rows))
	for _, one := range rows {
		if one == nil {
			continue
		}
		ids = append(ids, one.Id)
		filesByID[one.Id] = one.Files
	}
	if len(ids) == 0 {
		return web.Data(map[string]any{"deleted": 0}), nil
	}
	// Rows first: if the delete fails the attachments of a row that is still listed stay
	// on disk. Attachments go per batch, right after the rows that recorded them.
	err = deleteInBatches(ids, func(batch []uint) error {
		if err := l.logModel.Delete().Where("id IN (?)", batch).Delete(); err != nil {
			return err
		}
		attachments, aerr := l.attachmentModel.FindByLogIds(batch)
		if aerr != nil {
			// The legacy files JSON still carries every path, so the cleanup below can
			// proceed even when the attachment rows cannot be read.
			log.Error("read attachment rows failed", zap.Error(aerr))
		}
		if derr := l.attachmentModel.Delete().Where("log_id IN ?", batch).Delete(); derr != nil {
			log.Error("delete attachment rows failed", zap.Error(derr))
		}
		byLog := make(map[uint][]*model.Attachment, len(batch))
		for _, attachment := range attachments {
			if attachment != nil {
				byLog[attachment.LogId] = append(byLog[attachment.LogId], attachment)
			}
		}
		for _, id := range batch {
			removeLogFiles(filesByID[id], byLog[id])
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	recordOp(l.context, req, user.Id, entity.OpCleanup, entity.TargetLog, 0, "")
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

// deleteAttachmentRows removes the attachment rows of the given logs. Best effort: the
// legacy files JSON still carries the paths, so a failed row delete never blocks the
// cache cleanup.
func (l *Log) deleteAttachmentRows(logIds []uint) {
	if l.attachmentModel == nil || len(logIds) == 0 {
		return
	}
	if err := deleteInBatches(logIds, func(batch []uint) error {
		return l.attachmentModel.Delete().Where("log_id IN ?", batch).Delete()
	}); err != nil {
		log.Error("delete attachment rows failed", zap.Error(err))
	}
}

// removeLogFiles deletes the cache copies recorded for a log in either place: the legacy
// files JSON and the attachment table.
func removeLogFiles(filesJSON string, attachments []*model.Attachment) {
	removeAttachments(filesJSON)
	for _, attachment := range attachments {
		if attachment != nil && attachment.FilePath != "" {
			_ = os.Remove(attachment.FilePath)
		}
	}
}

// downLoad serves an attachment the caller may read. A new-style link carries the
// attachment id, whose log is the permission boundary; the older logId+index and
// logId+file forms still resolve through the log's files JSON.
func (l *Log) downLoad(req *web.Request) (any, error) {
	attachmentIdRaw := req.GetFormParam("id")
	logIdRaw := req.GetFormParam("logId")
	indexRaw := req.GetFormParam("index")
	fileName := req.GetFormParam("file")
	if attachmentIdRaw == "" && logIdRaw == "" {
		return nil, errors.New("id or logId is required")
	}
	user, err := auth.User(req, l.context)
	if user == nil {
		return nil, err
	}
	var file *smtp.File
	var logId uint
	index := -1
	if attachmentIdRaw != "" {
		attachmentId, err := strconv.Atoi(attachmentIdRaw)
		if err != nil {
			return nil, errors.New("invalid attachment id")
		}
		attachment, err := l.attachmentModel.FindByPK(uint(attachmentId))
		if err != nil {
			return nil, err
		}
		if attachment == nil {
			return nil, errors.New("attachment not found")
		}
		// The attachment's log is the permission boundary: a row whose log is gone, or
		// belongs to someone else, stays unreachable.
		query := l.logModel.Query().Where("id = ?", attachment.LogId)
		if !user.IsAdmin {
			query = query.Where("user_id = ?", user.Id)
		}
		entry, err := query.One()
		if err != nil {
			return nil, err
		}
		if entry == nil {
			return nil, errors.New("attachment not found")
		}
		logId = attachment.LogId
		file = &smtp.File{Name: attachment.Name, FilePath: attachment.FilePath}
	} else {
		if indexRaw == "" && fileName == "" {
			return nil, errors.New("file or index is required")
		}
		parsed, err := strconv.Atoi(logIdRaw)
		if err != nil {
			return nil, errors.New("logId is required")
		}
		logId = uint(parsed)
		query := l.logModel.Query().Where("id = ?", logId)
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
		if indexRaw != "" {
			// Older links address the recorded array by position.
			index, err = strconv.Atoi(indexRaw)
			if err != nil {
				return nil, errors.New("invalid file index")
			}
			file, err = attachmentAt(one.Files, index)
		} else {
			// The oldest links carry only the name; a duplicate resolves to the first copy.
			file, err = attachmentByName(one.Files, fileName)
		}
		if err != nil {
			return nil, err
		}
	}
	if _, err := os.Stat(file.FilePath); err != nil {
		return nil, errors.New("attachment file is no longer available")
	}
	log.Info("downLoad", zap.Uint("logId", logId), zap.Int("index", index), zap.String("file", file.Name), zap.Uint("userId", user.Id))
	// Serve it under the name it was sent with, not the hashed cache filename.
	name := filepath.Base(file.Name)
	if name == "." || name == ".." || name == string(filepath.Separator) {
		return web.CreateFileResponse(file.FilePath), nil
	}
	return web.CreateFileResponseWithName(file.FilePath, name), nil
}

// attachmentAt resolves the attachment at index in a log's files array.
func attachmentAt(filesJSON string, index int) (*smtp.File, error) {
	files, err := decodeAttachments(filesJSON)
	if err != nil {
		return nil, err
	}
	if index < 0 || index >= len(files) || files[index] == nil || files[index].FilePath == "" {
		return nil, errors.New("attachment not found")
	}
	return files[index], nil
}

// attachmentByName is the name lookup used by older links.
func attachmentByName(filesJSON string, name string) (*smtp.File, error) {
	files, err := decodeAttachments(filesJSON)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file != nil && file.Name == name && file.FilePath != "" {
			return file, nil
		}
	}
	return nil, errors.New("attachment not found")
}

func decodeAttachments(filesJSON string) ([]*smtp.File, error) {
	if filesJSON == "" {
		return nil, errors.New("log has no attachments")
	}
	var files []*smtp.File
	if err := json.Unmarshal([]byte(filesJSON), &files); err != nil {
		return nil, errors.New("invalid attachment record")
	}
	return files, nil
}

func (l *Log) Init(context *core.Context) error {
	l.context = context
	l.logModel = wf.GetModel[*model.LogModel](context)
	l.attachmentModel = wf.GetModel[*model.AttachmentModel](context)
	l.userService = wf.GetService[*service.UserService](context)
	context.Get("/api/log/:id", l.getOne).WithMeta(auth2.WithLogin())
	context.Get("/api/log", l.getPage).WithMeta(auth2.WithLogin())
	context.Delete("/api/log/:id", l.deleteOne).WithMeta(auth2.WithLogin())
	context.Post("/api/log/cleanup", l.cleanUp).WithMeta(auth2.WithLogin())
	context.Get("/api/download", l.downLoad).WithMeta(auth2.WithLogin())
	return nil

}
