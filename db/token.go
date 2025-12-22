package db

import (
	"time"

	"github.com/chuccp/http2smtp/web"
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
	*Model[*Token]
	db        *gorm.DB
	tableName string
}

func NewTokenModel(db *gorm.DB, tableName string) *TokenModel {
	return &TokenModel{db: db, tableName: tableName, Model: NewModel[*Token](db, tableName)}
}

func (a *TokenModel) CreateTable() error {
	return a.Model.CreateTable(&Token{})
}
func (a *TokenModel) DeleteTable() error {
	return a.Model.DeleteTable(&Token{})
}

func (a *TokenModel) Save(token *Token) error {
	return a.Model.Save(token)
}
func (a *TokenModel) GetOne(id uint) (*Token, error) {
	var token Token
	err := a.Model.GetOne(id, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}
func (a *TokenModel) GetOneByToken(tokenStr string) (*Token, error) {
	var token Token
	token.Token = tokenStr
	tx := a.Model.db.Table(a.tableName).Where(token).First(&token)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &token, nil
}

func (a *TokenModel) DeleteOne(id uint) error {
	return a.Model.DeleteOne(id, &Token{})
}

func (a *TokenModel) Edit(token *Token) error {
	return a.Model.EditForMap(token.Id, map[string]interface{}{
		"is_use":         token.IsUse,
		"update_time":    time.Now(),
		"receive_emails": token.ReceiveEmailIds,
		"SMTP_Id":        token.SMTPId,
	})
}

func (a *TokenModel) NewModel(db *gorm.DB) *TokenModel {
	return &TokenModel{db: db, tableName: a.tableName}
}
func (a *TokenModel) Page(page *web.Page) (*Page[*Token], error) {
	var tokens []*Token
	num, err := a.Model.Page(page, &tokens)
	if err != nil {
		return nil, err
	}
	return ToPage[*Token](num, tokens), nil
}
