package model

import (
	"time"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/model"
)

// Attachment is one file recorded on a mail log. Name is the filename the sender used
// (two files may share one), while FilePath is the unique cache copy that is actually
// downloaded; the path never leaves the server.
type Attachment struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	LogId      uint      `gorm:"column:log_id;index" json:"logId"`
	Name       string    `gorm:"column:name" json:"name"`
	FilePath   string    `gorm:"column:file_path" json:"-"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (a *Attachment) SetCreateTime(createTime time.Time) {
	a.CreateTime = createTime
}
func (a *Attachment) SetUpdateTime(updateTime time.Time) {
	a.UpdateTime = updateTime
}
func (a *Attachment) GetId() uint {
	return a.Id
}
func (a *Attachment) SetId(id uint) {
	a.Id = id
}

type AttachmentModel struct {
	*model.EntryModel[*Attachment, uint]
	db      *db.DB
	context *core.Context
}

func (t *AttachmentModel) Init(db *db.DB, context *core.Context) error {
	t.db = db
	t.context = context
	t.EntryModel = model.NewEntryModel[*Attachment, uint](t.db, t.GetTableName())
	return nil
}

func (t *AttachmentModel) ReNew(db *db.DB, c *core.Context) core.IModel {
	return &AttachmentModel{
		EntryModel: model.NewEntryModel[*Attachment, uint](db, t.GetTableName()),
		db:         db,
		context:    c,
	}
}

func (t *AttachmentModel) GetTableName() string {
	return "t_attachment"
}

// FindByLogId returns one log's attachments in the order they were sent, so the detail
// page and any index-based lookup agree on the order.
func (t *AttachmentModel) FindByLogId(logId uint) ([]*Attachment, error) {
	return t.Query().Where("log_id = ?", logId).Order("id asc").All()
}

// FindByLogIds returns the attachments of several logs, oldest first.
func (t *AttachmentModel) FindByLogIds(logIds []uint) ([]*Attachment, error) {
	if len(logIds) == 0 {
		return []*Attachment{}, nil
	}
	return t.Query().Where("log_id IN ?", logIds).Order("id asc").All()
}
