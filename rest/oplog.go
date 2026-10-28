package rest

import (
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/http2smtp/service"
)

// recordOp writes one operation-log entry. Best effort: the service swallows write
// failures, so an audit problem never fails the operation it describes.
func recordOp(ctx *core.Context, userId uint, action, target string, targetId uint, targetName string) {
	core.GetService[*service.OperationLogService](ctx).Record(userId, action, target, targetId, targetName)
}
