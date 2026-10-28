package webui

import (
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
)

// recordOp stores one audit entry. It is best effort: the service logs a write failure
// and the operation itself carries on.
func (p *Pages) recordOp(userId uint, action, target string, targetId uint, targetName string) {
	if p.oplogService != nil {
		p.oplogService.Record(userId, action, target, targetId, targetName)
	}
}

// OplogRow is one audit row with its action and target translated for display.
type OplogRow struct {
	*model.OperationLog
	ActionLabel string
	ActionClass string
	TargetLabel string
}

// OplogListData is the operation-log page (admin only).
type OplogListData struct {
	Base
	Rows  []OplogRow
	Pager *Pager
}

var oplogActionKeys = map[string]string{
	entity.OpCreate:  "oplog.actionCreate",
	entity.OpUpdate:  "oplog.actionUpdate",
	entity.OpDelete:  "oplog.actionDelete",
	entity.OpEnable:  "oplog.actionEnable",
	entity.OpDisable: "oplog.actionDisable",
	entity.OpTrigger: "oplog.actionTrigger",
	entity.OpCleanup: "oplog.actionCleanup",
	entity.OpRestart: "oplog.actionRestart",
	entity.OpLogin:   "oplog.actionLogin",
}

var oplogActionClasses = map[string]string{
	entity.OpCreate:  "tag-ok",
	entity.OpEnable:  "tag-ok",
	entity.OpLogin:   "tag-ok",
	entity.OpDelete:  "tag-off",
	entity.OpDisable: "tag-off",
	entity.OpUpdate:  "tag-warn",
	entity.OpTrigger: "tag-warn",
	entity.OpCleanup: "tag-warn",
	entity.OpRestart: "tag-warn",
}

// Target types reuse the sidebar labels, so a row reads "定时任务 / 邮件通知".
var oplogTargetKeys = map[string]string{
	entity.TargetSMTP:     "layout.sidebar.smtp",
	entity.TargetMail:     "layout.sidebar.mail",
	entity.TargetToken:    "layout.sidebar.tokens",
	entity.TargetSchedule: "layout.sidebar.schedule",
	entity.TargetUser:     "layout.sidebar.users",
	entity.TargetLog:      "layout.sidebar.logs",
	entity.TargetSettings: "layout.sidebar.settings",
}

func (p *Pages) oplogList(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireAdmin(view)
	if !ok {
		return nil
	}
	page := listPage(req)
	result, err := p.oplogModel.Query().Order("id desc").PageForWeb(page)
	if err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	fillNames(p.userService, result.List,
		func(r *model.OperationLog) uint { return r.UserId },
		func(r *model.OperationLog, n string) { r.UserName = n })
	tr := translatorFor(langOf(req))
	rows := make([]OplogRow, 0, len(result.List))
	for _, r := range result.List {
		actionLabel := r.Action
		if key, ok := oplogActionKeys[r.Action]; ok {
			actionLabel = tr.T(key)
		}
		targetLabel := r.TargetType
		if key, ok := oplogTargetKeys[r.TargetType]; ok {
			targetLabel = tr.T(key)
		}
		rows = append(rows, OplogRow{
			OperationLog: r,
			ActionLabel:  actionLabel,
			ActionClass:  oplogActionClasses[r.Action],
			TargetLabel:  targetLabel,
		})
	}
	data := &OplogListData{
		Base:  p.base(req, user, "oplog.title", true),
		Rows:  rows,
		Pager: newPager(page.PageNo, page.PageSize, result.Total, "/oplogs", nil),
	}
	view.View("oplog_list").Layout(layoutFor("oplog_list")).Data(data)
	return nil
}
