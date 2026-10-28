package rest

import (
	"errors"
	"strconv"
	"time"

	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
)

// Oplog serves the admin maintenance endpoints of the operation log: delete one entry,
// or clear entries older than N days (0 clears everything).
type Oplog struct {
	context    *core.Context
	oplogModel *model.OperationLogModel
}

// recordOp writes one operation-log entry, stamped with the caller's IP. Best effort:
// the service swallows write failures, so an audit problem never fails the operation it
// describes.
func recordOp(ctx *core.Context, req *web.Request, userId uint, action, target string, targetId uint, targetName string) {
	core.GetService[*service.OperationLogService](ctx).Record(userId, action, target, targetId, targetName, req.ClientIP())
}

func (o *Oplog) adminUser(req *web.Request) (*model.User, error) {
	user, err := auth.User(req, o.context)
	if user == nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New("admin access required")
	}
	return user, nil
}

func (o *Oplog) deleteOne(req *web.Request) (any, error) {
	user, err := o.adminUser(req)
	if err != nil {
		return nil, err
	}
	id, err := strconv.Atoi(req.Param("id"))
	if err != nil {
		return nil, err
	}
	one, err := deleteOplogRow(o.oplogModel, uint(id))
	if err != nil {
		return nil, err
	}
	// Deleting audit rows is itself audited.
	recordOp(o.context, req, user.Id, entity.OpDelete, entity.TargetOplog, uint(id), one.TargetName)
	return web.Ok("ok"), nil
}

func (o *Oplog) cleanUp(req *web.Request) (any, error) {
	user, err := o.adminUser(req)
	if err != nil {
		return nil, err
	}
	days := req.GetJsonIntValueOrDefault("days", 0)
	deleted, err := cleanupOplogRows(o.oplogModel, days)
	if err != nil {
		return nil, err
	}
	if deleted > 0 {
		recordOp(o.context, req, user.Id, entity.OpCleanup, entity.TargetOplog, 0, "")
	}
	return web.Data(map[string]any{"deleted": deleted}), nil
}

// deleteOplogRow removes one audit entry and returns it, so the caller can name the
// deleted row in its own audit entry.
func deleteOplogRow(m *model.OperationLogModel, id uint) (*model.OperationLog, error) {
	one, err := m.FindByPK(id)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New("operation log not found")
	}
	if err := m.DeleteByPK(id); err != nil {
		return nil, err
	}
	return one, nil
}

// cleanupOplogRows removes entries older than days — everything when days is 0 — and
// reports how many went.
func cleanupOplogRows(m *model.OperationLogModel, days int) (int, error) {
	query := m.Query()
	if days > 0 {
		query = query.Where("create_time < ?", time.Now().AddDate(0, 0, -days))
	}
	rows, err := query.All()
	if err != nil {
		return 0, err
	}
	ids := make([]uint, 0, len(rows))
	for _, one := range rows {
		if one != nil {
			ids = append(ids, one.Id)
		}
	}
	if len(ids) == 0 {
		return 0, nil
	}
	if err := m.Delete().Where("id IN (?)", ids).Delete(); err != nil {
		return 0, err
	}
	return len(ids), nil
}

func (o *Oplog) Init(context *core.Context) error {
	o.context = context
	o.oplogModel = core.GetModel[*model.OperationLogModel](context)
	context.Delete("/api/oplog/:id", o.deleteOne).WithMeta(auth2.WithLogin())
	context.Post("/api/oplog/cleanup", o.cleanUp).WithMeta(auth2.WithLogin())
	return nil
}
