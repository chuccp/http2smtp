package model

import (
	"time"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/model"
)

// OperationLog is one recorded management action (create/update/delete/enable/disable,
// a login, a restart, ...). UserName is not a column — the admin list fills it.
type OperationLog struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserId     uint      `gorm:"column:user_id" json:"userId"`
	UserName   string    `gorm:"-" json:"userName"`
	Action     string    `gorm:"column:action" json:"action"`
	TargetType string    `gorm:"column:target_type" json:"targetType"`
	TargetId   uint      `gorm:"column:target_id" json:"targetId"`
	TargetName string    `gorm:"column:target_name" json:"targetName"`
	Ip         string    `gorm:"column:ip" json:"ip"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
}

func (l *OperationLog) SetCreateTime(createTime time.Time) {
	l.CreateTime = createTime
}
func (l *OperationLog) SetUpdateTime(updateTime time.Time) {
	// Audit rows are append-only; the framework still calls this on save.
}
func (l *OperationLog) GetId() uint {
	return l.Id
}
func (l *OperationLog) SetId(id uint) {
	l.Id = id
}

type OperationLogModel struct {
	*model.EntryModel[*OperationLog, uint]
	db      *db.DB
	context *core.Context
}

func (t *OperationLogModel) Init(db *db.DB, context *core.Context) error {
	t.db = db
	t.context = context
	t.EntryModel = model.NewEntryModel[*OperationLog, uint](t.db, t.GetTableName())
	return nil
}

func (t *OperationLogModel) ReNew(db *db.DB, c *core.Context) core.IModel {
	return &OperationLogModel{
		EntryModel: model.NewEntryModel[*OperationLog, uint](db, t.GetTableName()),
		db:         db,
		context:    c,
	}
}

func (t *OperationLogModel) GetTableName() string {
	return "t_operation_log"
}
