package db

import (
	"time"

	"github.com/chuccp/http2smtp/web"
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

type STMPModel struct {
	*Model[*SMTP]
	db        *gorm.DB
	tableName string
}

func NewSMTPModel(db *gorm.DB, tableName string) *STMPModel {
	return &STMPModel{db: db, tableName: tableName, Model: NewModel[*SMTP](db, tableName)}
}

func (a *STMPModel) CreateTable() error {
	return a.Model.CreateTable(&SMTP{})
}
func (a *STMPModel) DeleteTable() error {
	return a.Model.DeleteTable(&SMTP{})
}

func (a *STMPModel) Save(stmp *SMTP) error {
	return a.Model.Save(stmp)
}
func (a *STMPModel) GetOne(id uint) (*SMTP, error) {
	var smtp SMTP
	err := a.Model.GetOne(id, &smtp)
	if err != nil {
		return nil, err
	}
	smtp.Name = smtp.Username
	return &smtp, nil
}

func (a *STMPModel) GetByIds(id []uint) ([]*SMTP, error) {
	var smtps []*SMTP
	err := a.Model.GetByIds(id, &smtps)
	if err != nil {
		return nil, err
	}
	for _, smtp := range smtps {
		smtp.Name = smtp.Username
	}
	return smtps, nil
}

func (a *STMPModel) GetMapByIds(id []uint) (map[uint]*SMTP, error) {
	SMTPs, err := a.GetByIds(id)
	if err != nil {
		return nil, err
	}
	var SMTPMap = make(map[uint]*SMTP)
	for _, st := range SMTPs {
		SMTPMap[st.Id] = st
	}
	return SMTPMap, nil
}

func (a *STMPModel) DeleteOne(id uint) error {
	return a.Model.DeleteOne(id, &SMTP{})
}

func (a *STMPModel) Edit(smtp *SMTP) error {
	return a.Model.Edit(smtp)
}

func (a *STMPModel) NewModel(db *gorm.DB) *STMPModel {
	return &STMPModel{db: db, tableName: a.tableName}
}
func (a *STMPModel) Page(page *web.Page) (*Page[*SMTP], error) {
	var smtps []*SMTP
	num, err := a.Model.Page(page, &smtps)
	if err != nil {
		return nil, err
	}
	for _, smtp := range smtps {
		smtp.Name = smtp.Username
	}
	return ToPage[*SMTP](num, smtps), nil
}
