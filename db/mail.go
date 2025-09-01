package db

import (
	"bytes"
	"github.com/chuccp/http2smtp/util"
	"github.com/chuccp/http2smtp/web"
	"gorm.io/gorm"
	"time"
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
	*Model[*Mail]
	db        *gorm.DB
	tableName string
}

func NewMailModel(db *gorm.DB, tableName string) *MailModel {
	return &MailModel{db: db, tableName: tableName, Model: NewModel[*Mail](db, tableName)}
}

func (a *MailModel) CreateTable() error {
	return a.Model.CreateTable(&Mail{})
}
func (a *MailModel) DeleteTable() error {
	return a.Model.DeleteTable(&Mail{})
}

func (a *MailModel) Save(mail *Mail) error {
	return a.Model.Save(mail)
}

func (a *MailModel) GetByIds(id []uint) ([]*Mail, error) {
	var mails []*Mail
	return mails, a.Model.GetByIds(id, &mails)
}
func (a *MailModel) GetMapByIds(id []uint) (map[uint]*Mail, error) {

	mails, err := a.GetByIds(id)
	if err != nil {
		return nil, err
	}
	var mailMap = make(map[uint]*Mail)
	for _, mail := range mails {
		mailMap[mail.Id] = mail
	}
	return mailMap, nil
}
func (a *MailModel) GetOne(id uint) (*Mail, error) {
	var mail Mail
	err := a.Model.GetOne(id, &mail)
	if err != nil {
		return nil, err
	}
	return &mail, nil
}

func (a *MailModel) DeleteOne(id uint) error {
	return a.Model.DeleteOne(id, &Mail{})
}

func (a *MailModel) Edit(mail *Mail) error {
	return a.Model.Edit(mail)
}

func (a *MailModel) NewModel(db *gorm.DB) *MailModel {
	return &MailModel{db: db, tableName: a.tableName}
}
func (a *MailModel) Page(page *web.Page) (*Page[*Mail], error) {
	var mails []*Mail
	num, err := a.Model.Page(page, &mails)
	if err != nil {
		return nil, err
	}
	return ToPage[*Mail](num, mails), nil
}

func GetMails(ids []uint, mailMap map[uint]*Mail) []*Mail {
	mails := make([]*Mail, 0)
	for _, id := range ids {
		v, ok := mailMap[id]
		if ok {
			mails = append(mails, v)
		}
	}
	return mails
}
func GetMailsStr(mails []*Mail) string {
	buffer := new(bytes.Buffer)
	for _, mail := range mails {
		buffer.WriteString("," + util.FormatMail(mail.Name, mail.Mail))
	}
	if buffer.Len() == 0 {
		return ""
	}
	return buffer.String()[1:]
}
