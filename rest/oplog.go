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

// recordOp writes one operation-log entry. Best effort: the service swallows write
// failures, so an audit problem never fails the operation it describes.
func recordOp(ctx *core.Context, userId uint, action, target string, targetId uint, targetName string) {
	core.GetService[*service.OperationLogService](ctx).Record(userId, action, target, targetId, targetName)
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
	one, err := o.oplogModel.FindByPK(uint(id))
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New("operation log not found")
	}
	if err := o.oplogModel.DeleteByPK(uint(id)); err != nil {
		return nil, err
	}
	// Deleting audit rows is itself audited.
	recordOp(o.context, user.Id, entity.OpDelete, entity.TargetOplog, uint(id), one.TargetName)
	return web.Ok("ok"), nil
}

func (o *Oplog) cleanUp(req *web.Request) (any, error) {
	user, err := o.adminUser(req)
	if err != nil {
		return nil, err
	}
	days := req.GetJsonIntValueOrDefault("days", 0)
	query := o.oplogModel.Query()
	if days > 0 {
		query = query.Where("create_time < ?", time.Now().AddDate(0, 0, -days))
	}
	rows, err := query.All()
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(rows))
	for _, one := range rows {
		if one != nil {
			ids = append(ids, one.Id)
		}
	}
	if len(ids) == 0 {
		return web.Data(map[string]any{"deleted": 0}), nil
	}
	if err := o.oplogModel.Delete().Where("id IN (?)", ids).Delete(); err != nil {
		return nil, err
	}
	recordOp(o.context, user.Id, entity.OpCleanup, entity.TargetOplog, 0, "")
	return web.Data(map[string]any{"deleted": len(ids)}), nil
}

func (o *Oplog) Init(context *core.Context) error {
	o.context = context
	o.oplogModel = core.GetModel[*model.OperationLogModel](context)
	context.Delete("/api/oplog/:id", o.deleteOne).WithMeta(auth2.WithLogin())
	context.Post("/api/oplog/cleanup", o.cleanUp).WithMeta(auth2.WithLogin())
	return nil
}
