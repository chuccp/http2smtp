package model

import "time"

type User struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name       string    `gorm:"column:name;uniqueIndex" json:"name"`
	Password   string    `gorm:"column:password" json:"-"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time" json:"updateTime"`
	IsUse      bool      `gorm:"column:is_use" json:"isUse"`
	IsAdmin    bool      `gorm:"column:is_admin" json:"isAdmin"`
}
