package model

import (
	"time"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/model"
	"gorm.io/gorm"
)

type Token struct {
	Id               uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Token            string    `gorm:"unique;column:token" json:"token"`
	Name             string    `gorm:"-" json:"name"`
	Subject          string    `gorm:"column:subject" json:"subject"`
	ReceiveEmailIds  string    `gorm:"column:receive_emails" json:"receiveEmailIds"`
	ReceiveEmails    []*Mail   `gorm:"-" json:"receiveEmails"`
	ReceiveEmailsStr string    `gorm:"-" json:"receiveEmailsStr"`
	SMTPId           uint      `gorm:"column:SMTP_Id" json:"SMTPId"`
	SMTP             *SMTP     `gorm:"-" json:"SMTP"`
	SMTPStr          string    `gorm:"-" json:"SMTPStr"`
	IsUse            bool      `gorm:"column:is_use" json:"isUse"`
	CreateTime       time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime       time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (token *Token) SetCreateTime(createTime time.Time) {
	token.CreateTime = createTime
}
func (token *Token) SetUpdateTime(updateTIme time.Time) {
	token.UpdateTime = updateTIme
}
func (token *Token) GetId() uint {
	return token.Id
}
func (token *Token) SetId(id uint) {
	token.Id = id
}

type TokenModel struct {
	*model.EntryModel[*Token]
	db      *gorm.DB
	context *core.Context
}

func (t *TokenModel) Init(context *core.Context) {
	t.db = context.GetDB()
	t.context = context
	t.EntryModel = model.NewEntryModel[*Token](t.db, t.GetTableName(), &Token{})
}
func (t *TokenModel) GetTableName() string {
	return "t_token"
}
func (t *TokenModel) Name() string {
	return t.GetTableName()
}
