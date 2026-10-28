package service

import (
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/http2smtp/model"
	"go.uber.org/zap"
)

// OperationLogService records management actions for the admin audit page.
type OperationLogService struct {
	context    *core.Context
	oplogModel *model.OperationLogModel
}

func (s *OperationLogService) Init(ctx *core.Context) error {
	s.context = ctx
	s.oplogModel = core.GetModel[*model.OperationLogModel](ctx)
	return nil
}

// Record stores one action. Best effort: an audit write must never fail the operation
// it describes, so errors are logged and swallowed.
func (s *OperationLogService) Record(userId uint, action, target string, targetId uint, targetName, ip string) {
	row := &model.OperationLog{
		UserId:     userId,
		Action:     action,
		TargetType: target,
		TargetId:   targetId,
		TargetName: targetName,
		Ip:         ip,
	}
	if err := s.oplogModel.Save(row); err != nil {
		log.Error("operation log write failed",
			zap.String("action", action), zap.String("target", target), zap.Error(err))
	}
}
