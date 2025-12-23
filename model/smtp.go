package model

import (
	"time"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/model"
	"gorm.io/gorm"
)

type SMTP struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Host       string    `gorm:"column:host" json:"host"`
	Port       int       `gorm:"column:port" json:"port"`
	Mail       string    `gorm:"column:mail" json:"mail"`
	Username   string    `gorm:"column:username" json:"username"`
	Name       string    `gorm:"-"  json:"name"`
	Password   string    `gorm:"column:password"  json:"password"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (smtp *SMTP) SetCreateTime(createTime time.Time) {
	smtp.CreateTime = createTime
}
func (smtp *SMTP) SetUpdateTime(updateTime time.Time) {
	smtp.UpdateTime = updateTime
}
func (smtp *SMTP) GetId() uint {
	return smtp.Id
}
func (smtp *SMTP) SetId(id uint) {
	smtp.Id = id
}

type SMTPModel struct {
	*model.EntryModel[*SMTP]
	db      *gorm.DB
	context *core.Context
}

func (t *SMTPModel) Init(context *core.Context) {
	t.db = context.GetDB()
	t.context = context
	t.EntryModel = model.NewEntryModel[*SMTP](t.db, t.GetTableName(), &SMTP{})
}
func (t *SMTPModel) GetTableName() string {
	return "t_SMTP"
}
func (t *SMTPModel) Name() string {
	return t.GetTableName()
}
