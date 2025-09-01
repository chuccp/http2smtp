package db

import (
	"github.com/chuccp/http2smtp/web"
	"gorm.io/gorm"
	"time"
)

const (
	SUCCESS byte = iota
	WARM
	ERROR
)

type Log struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name       string    `gorm:"column:name" json:"name"`
	Mail       string    `gorm:"column:mail" json:"mail"`
	Token      string    `gorm:"column:token" json:"token"`
	SMTP       string    `gorm:"column:smtp" json:"smtp"`
	Subject    string    `gorm:"column:subject" json:"subject"`
	Content    string    `gorm:"column:content" json:"content"`
	Files      string    `gorm:"column:files" json:"files"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time" json:"updateTime"`
	Status     byte      `gorm:"column:status" json:"status"`
	StatusStr  string    `gorm:"-" json:"statusStr"`
	Result     string    `gorm:"column:result" json:"result"`
}

func (log *Log) SetCreateTime(createTime time.Time) {
	log.CreateTime = createTime
}
func (log *Log) SetUpdateTime(updateTime time.Time) {
	log.UpdateTime = updateTime
}
func (log *Log) GetId() uint {
	return log.Id
}
func (log *Log) SetId(id uint) {
	log.Id = id
}

type LogModel struct {
	*Model[*Log]
	db        *gorm.DB
	tableName string
}

func NewLogModel(db *gorm.DB, tableName string) *LogModel {
	return &LogModel{db: db, tableName: tableName, Model: NewModel[*Log](db, tableName)}
}

func (a *LogModel) CreateTable() error {
	return a.Model.CreateTable(&Log{})
}
func (a *LogModel) DeleteTable() error {
	return a.Model.DeleteTable(&Log{})
}

func (a *LogModel) Page(page *web.Page) (*Page[*Log], error) {
	var logs []*Log
	num, err := a.Model.Page(page, &logs)
	if err != nil {
		return nil, err
	}
	for _, log := range logs {
		if log.Status == SUCCESS {
			log.StatusStr = "success"
		}
		if log.Status == WARM {
			log.StatusStr = "warm"
		}
		if log.Status == ERROR {
			log.StatusStr = "error"
		}
	}

	return ToPage[*Log](num, logs), nil
}
func (a *LogModel) Save(log *Log) error {
	return a.Model.Save(log)
}
func (a *LogModel) GetOne(id uint) (*Log, error) {
	var log Log
	err := a.Model.GetOne(id, &log)
	if err != nil {
		return nil, err
	}

	if log.Status == SUCCESS {
		log.StatusStr = "success"
	}
	if log.Status == WARM {
		log.StatusStr = "warm"
	}
	if log.Status == ERROR {
		log.StatusStr = "error"
	}

	return &log, nil
}
