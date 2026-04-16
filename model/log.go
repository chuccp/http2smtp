package model

import (
	"strings"
	"time"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/model"
	"github.com/chuccp/go-web-frame/web"
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
	*model.EntryModel[*Log]
	db      *db.DB
	context *core.Context
}

func (t *LogModel) Init(db *db.DB, context *core.Context) error {
	t.db = db
	t.context = context
	t.EntryModel = model.NewEntryModel[*Log](t.db, t.GetTableName())
	return nil
}

func (t *LogModel) ReNew(db *db.DB, c *core.Context) core.IModel {
	return &LogModel{
		EntryModel: model.NewEntryModel[*Log](db, t.GetTableName()),
		db:         db,
		context:    c,
	}
}

func (t *LogModel) PageBySearch(page *web.Page, key string) (*web.PageAble[*Log], error) {
	key = strings.TrimSpace(key)
	if len(key) == 0 {
		return t.PageForWeb(page)
	}
	return t.Query().Where("`content` like ? or `subject` like ?", "%"+key+"%", "%"+key+"%").Order("`id` desc").PageForWeb(page)

}

func (t *LogModel) GetTableName() string {
	return "t_log"
}
