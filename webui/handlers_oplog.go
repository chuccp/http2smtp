package webui

import (
	"net/http"

	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
)

// recordOp stores one audit entry, stamped with the caller's IP. It is best effort: the
// service logs a write failure and the operation itself carries on.
func (p *Pages) recordOp(req *web.Request, userId uint, action, target string, targetId uint, targetName string) {
	if p.oplogService != nil {
		p.oplogService.Record(userId, action, target, targetId, targetName, req.ClientIP())
	}
}

// OplogRow is one audit row with its action and target translated for display.
type OplogRow struct {
	*model.OperationLog
	ActionLabel string
	ActionClass string
	TargetLabel string
}

// OplogListData is the operation-log page. Admins read every entry; everyone else sees
// their own.
type OplogListData struct {
	Base
	Rows          []OplogRow
	Pager         *Pager
	UserFilter    string
	ActionFilter  string
	TargetFilter  string
	NameFilter    string
	ActionOptions []SelectOption
	TargetOptions []SelectOption
}

var oplogActionKeys = map[string]string{
	entity.OpCreate:      "oplog.actionCreate",
	entity.OpUpdate:      "oplog.actionUpdate",
	entity.OpDelete:      "oplog.actionDelete",
	entity.OpEnable:      "oplog.actionEnable",
	entity.OpDisable:     "oplog.actionDisable",
	entity.OpTrigger:     "oplog.actionTrigger",
	entity.OpCleanup:     "oplog.actionCleanup",
	entity.OpRestart:     "oplog.actionRestart",
	entity.OpLogin:       "oplog.actionLogin",
	entity.OpLoginFailed: "oplog.actionLoginFailed",
	entity.OpLogout:      "oplog.actionLogout",
}

var oplogActionClasses = map[string]string{
	entity.OpCreate:      "tag-ok",
	entity.OpEnable:      "tag-ok",
	entity.OpLogin:       "tag-ok",
	entity.OpDelete:      "tag-off",
	entity.OpDisable:     "tag-off",
	entity.OpLoginFailed: "tag-off",
	entity.OpUpdate:      "tag-warn",
	entity.OpTrigger:     "tag-warn",
	entity.OpCleanup:     "tag-warn",
	entity.OpRestart:     "tag-warn",
	entity.OpLogout:      "tag-warn",
}

// The filter drop-downs list every action and target in a stable order.
var oplogActionOrder = []string{
	entity.OpLogin, entity.OpLoginFailed, entity.OpLogout,
	entity.OpCreate, entity.OpUpdate, entity.OpDelete, entity.OpEnable, entity.OpDisable,
	entity.OpTrigger, entity.OpCleanup, entity.OpRestart,
}

var oplogTargetOrder = []string{
	entity.TargetSMTP, entity.TargetMail, entity.TargetToken, entity.TargetSchedule,
	entity.TargetUser, entity.TargetLog, entity.TargetSettings, entity.TargetOplog,
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
	entity.TargetOplog:    "layout.sidebar.oplogs",
}

func (p *Pages) oplogList(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireUser(view)
	if !ok {
		return nil
	}
	page := listPage(req)
	userFilter := req.Query("user")
	actionFilter := req.Query("action")
	targetFilter := req.Query("target")
	nameFilter := req.Query("targetName")

	query := p.oplogModel.Query()
	if !user.IsAdmin {
		// Regular users read their own audit trail only.
		query = query.Where("user_id = ?", user.Id)
	} else if userFilter != "" {
		// The operator filter matches user names, so resolve them to ids first.
		operators, err := p.userModel.Query().Where("name LIKE ?", "%"+userFilter+"%").All()
		if err != nil {
			return p.renderErr(view, 0, "common.failed", err.Error())
		}
		ids := make([]uint, 0, len(operators))
		for _, u := range operators {
			ids = append(ids, u.Id)
		}
		if len(ids) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where("user_id IN ?", ids)
		}
	}
	if actionFilter != "" {
		query = query.Where("action = ?", actionFilter)
	}
	if targetFilter != "" {
		query = query.Where("target_type = ?", targetFilter)
	}
	if nameFilter != "" {
		query = query.Where("target_name LIKE ?", "%"+nameFilter+"%")
	}

	result, err := query.Order("id desc").PageForWeb(page)
	if err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	fillNames(p.userService, result.List,
		func(r *model.OperationLog) uint { return r.UserId },
		func(r *model.OperationLog, n string) { r.UserName = n })
	tr := translatorFor(langOf(req))
	rows := make([]OplogRow, 0, len(result.List))
	for _, r := range result.List {
		rows = append(rows, oplogRow(r, tr))
	}
	actionOptions := make([]SelectOption, 0, len(oplogActionOrder))
	for _, a := range oplogActionOrder {
		actionOptions = append(actionOptions, SelectOption{
			Value: a, Label: tr.T(oplogActionKeys[a]), Selected: a == actionFilter,
		})
	}
	targetOptions := make([]SelectOption, 0, len(oplogTargetOrder))
	for _, t := range oplogTargetOrder {
		targetOptions = append(targetOptions, SelectOption{
			Value: t, Label: tr.T(oplogTargetKeys[t]), Selected: t == targetFilter,
		})
	}
	q := queryOf(req, "user", "action", "target", "targetName", "pageSize")
	data := &OplogListData{
		Base:          p.base(req, user, "oplog.title", true),
		Rows:          rows,
		Pager:         newPager(page.PageNo, page.PageSize, result.Total, "/oplogs", q),
		UserFilter:    userFilter,
		ActionFilter:  actionFilter,
		TargetFilter:  targetFilter,
		NameFilter:    nameFilter,
		ActionOptions: actionOptions,
		TargetOptions: targetOptions,
	}
	view.View("oplog_list").Layout(layoutFor("oplog_list")).Data(data)
	return nil
}

// oplogRow translates one entry's action and target for display.
func oplogRow(r *model.OperationLog, tr *Translator) OplogRow {
	actionLabel := r.Action
	if key, ok := oplogActionKeys[r.Action]; ok {
		actionLabel = tr.T(key)
	}
	targetLabel := r.TargetType
	if key, ok := oplogTargetKeys[r.TargetType]; ok {
		targetLabel = tr.T(key)
	}
	return OplogRow{
		OperationLog: r,
		ActionLabel:  actionLabel,
		ActionClass:  oplogActionClasses[r.Action],
		TargetLabel:  targetLabel,
	}
}

// OplogDetailData is the operation-log detail dialog.
type OplogDetailData struct {
	Base
	Entry OplogRow
}

func (p *Pages) oplogDetail(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireUser(view)
	if !ok {
		return nil
	}
	id, err := idParam(req, "id")
	if err != nil {
		return p.renderErr(view, http.StatusBadRequest, "common.failed", err.Error())
	}
	one, err := p.oplogModel.FindByPK(id)
	if err != nil {
		return p.renderErr(view, 0, "common.failed", err.Error())
	}
	// A row that is not the caller's answers "not found", like the other lists.
	if one == nil || (!user.IsAdmin && one.UserId != user.Id) {
		return p.renderErr(view, http.StatusNotFound, "common.failed", "operation log not found")
	}
	fillNames(p.userService, []*model.OperationLog{one},
		func(r *model.OperationLog) uint { return r.UserId },
		func(r *model.OperationLog, n string) { r.UserName = n })
	data := &OplogDetailData{
		Base:  p.base(req, user, "oplog.detail", false),
		Entry: oplogRow(one, translatorFor(langOf(req))),
	}
	view.View("oplog_detail").Layout(layoutFor("oplog_detail")).Data(data)
	return nil
}
