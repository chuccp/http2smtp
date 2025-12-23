package model

import (
	"time"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/model"
	"gorm.io/gorm"
)

type Mail struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name       string    `gorm:"column:name" json:"name"`
	Mail       string    `gorm:"column:mail" json:"mail"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time" json:"updateTime"`
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
	*model.EntryModel[*Mail]
	db      *gorm.DB
	context *core.Context
}

func (t *MailModel) Init(context *core.Context) {
	t.db = context.GetDB()
	t.context = context
	t.EntryModel = model.NewEntryModel[*Mail](t.db, t.GetTableName(), &Mail{})
}
func (t *MailModel) GetTableName() string {
	return "t_mail"
}
func (t *MailModel) Name() string {
	return t.GetTableName()
}

func (t *MailModel) FindMapByIds(id []uint) (map[uint]*Mail, error) {

	mails, err := t.FindAllByIds(id...)
	if err != nil {
		return nil, err
	}
	var mailMap = make(map[uint]*Mail)
	for _, mail := range mails {
		mailMap[mail.Id] = mail
	}
	return mailMap, nil
}
