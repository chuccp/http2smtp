package model

import (
	"time"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/model"
)

type Mail struct {
	Id       uint   `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserId   uint   `gorm:"column:user_id" json:"userId"`
	UserName string `gorm:"-" json:"userName"`
	// CreatedByAdmin marks rows whose owner is an administrator (admin list view only).
	CreatedByAdmin bool      `gorm:"-" json:"createdByAdmin"`
	Name           string    `gorm:"column:name" json:"name"`
	Mail           string    `gorm:"column:mail" json:"mail"`
	CreateTime     time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime     time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (mail *Mail) SetCreateTime(createTime time.Time) {
	mail.CreateTime = createTime
}
func (mail *Mail) SetUpdateTime(updateTime time.Time) {
	mail.UpdateTime = updateTime
}
func (mail *Mail) GetId() uint {
	return mail.Id
}
func (mail *Mail) SetId(id uint) {
	mail.Id = id
}

type MailModel struct {
	*model.EntryModel[*Mail, uint]
	db      *db.DB
	context *core.Context
}

func (t *MailModel) Init(db *db.DB, context *core.Context) error {
	t.db = db
	t.context = context
	t.EntryModel = model.NewEntryModel[*Mail, uint](t.db, t.GetTableName())
	return nil
}
func (t *MailModel) GetTableName() string {
	return "t_mail"
}
func (t *MailModel) ReNew(db *db.DB, c *core.Context) core.IModel {
	return &MailModel{
		EntryModel: model.NewEntryModel[*Mail, uint](db, t.GetTableName()),
		db:         db,
		context:    c,
	}
}
func (t *MailModel) FindMapByIds(id []uint) (map[uint]*Mail, error) {

	mails, err := t.FindAllByPK(id...)
	if err != nil {
		return nil, err
	}
	var mailMap = make(map[uint]*Mail)
	for _, mail := range mails {
		mailMap[mail.Id] = mail
	}
	return mailMap, nil
}
