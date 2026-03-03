package model

import (
	"time"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/model"
)

type SMTP struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Host       string    `gorm:"column:host" json:"host"`
	Port       int       `gorm:"column:port" json:"port"`
	Mail       string    `gorm:"column:mail" json:"mail"`
	Username   string    `gorm:"column:username" json:"username"`
	Name       string    `gorm:"column:name" json:"name"`
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
	db      *db.DB
	context *core.Context
}

func (t *SMTPModel) Init(db *db.DB, context *core.Context) error {
	t.db = db
	t.context = context
	t.EntryModel = model.NewEntryModel[*SMTP](t.db, t.GetTableName())
	return nil
}

func (t *SMTPModel) ReNew(db *db.DB, c *core.Context) core.IModel {
	return &SMTPModel{
		EntryModel: model.NewEntryModel[*SMTP](db, t.GetTableName()),
		db:         db,
		context:    c,
	}
}

func (t *SMTPModel) GetTableName() string {
	return "t_SMTP"
}

func (t *SMTPModel) FindMapByIds(ids []uint) (map[uint]*SMTP, error) {
	SMTPs, err := t.FindAllByIds(ids...)
	if err != nil {
		return nil, err
	}
	var SMTPMap = make(map[uint]*SMTP)
	for _, st := range SMTPs {
		SMTPMap[st.Id] = st
	}
	return SMTPMap, nil
}
