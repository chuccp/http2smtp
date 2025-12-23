package model

import (
	"fmt"
	"time"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/model"
	"github.com/chuccp/go-web-frame/util"
	"gorm.io/gorm"
)

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Schedule struct {
	Id                uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name              string    `gorm:"column:name" json:"name"`
	Token             string    `gorm:"column:token" json:"token"`
	TokenId           uint      `gorm:"-" json:"tokenId"`
	Cron              string    `gorm:"column:cron" json:"cron"`
	Url               string    `gorm:"column:url" json:"url"`
	Method            string    `gorm:"column:method" json:"method"`
	HeaderStr         string    `gorm:"column:header_str" json:"headerStr"`
	Headers           []*Header `gorm:"-" json:"headers"`
	Body              string    `gorm:"column:body" json:"body"`
	UseTemplate       bool      `gorm:"column:use_template" json:"useTemplate"`
	Template          string    `gorm:"column:template" json:"template"`
	IsUse             bool      `gorm:"column:is_use" json:"isUse"`
	IsOnlySendByError bool      `gorm:"column:is_only_send_by_error" json:"isSendOnlyByError"`
	CreateTime        time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime        time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (schedule *Schedule) Key() string {

	value := fmt.Sprintf("%s#%s#%s#%s#%s#%s#%s#%s#%t#%t#%t",
		schedule.Name,
		schedule.Token,
		schedule.Cron,
		schedule.Url,
		schedule.Method,
		schedule.HeaderStr,
		schedule.Body,
		schedule.Template,
		schedule.UseTemplate,
		schedule.IsUse,
		schedule.IsOnlySendByError)
	return util.MD5([]byte(value))
}

func (schedule *Schedule) SetCreateTime(createTime time.Time) {
	schedule.CreateTime = createTime
}
func (schedule *Schedule) SetUpdateTime(updateTIme time.Time) {
	schedule.UpdateTime = updateTIme
}
func (schedule *Schedule) GetId() uint {
	return schedule.Id
}
func (schedule *Schedule) SetId(id uint) {
	schedule.Id = id
}

type ScheduleModel struct {
	*model.EntryModel[*Schedule]
	db      *gorm.DB
	context *core.Context
}

func (t *ScheduleModel) Init(context *core.Context) {
	t.db = context.GetDB()
	t.context = context
	t.EntryModel = model.NewEntryModel[*Schedule](t.db, t.GetTableName(), &Schedule{})
}
func (t *ScheduleModel) GetTableName() string {
	return "t_mail"
}
func (t *ScheduleModel) Name() string {
	return t.GetTableName()
}
