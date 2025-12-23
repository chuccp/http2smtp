package db

import (
	"time"

	"github.com/chuccp/http2smtp/web"
	"gorm.io/gorm"
)

type IModel interface {
	SetCreateTime(createTime time.Time)
	SetUpdateTime(updateTIme time.Time)
	GetId() uint
	SetId(id uint)
}
type Model[T IModel] struct {
	db        *gorm.DB
	tableName string
}

func NewModel[T IModel](db *gorm.DB, tableName string) *Model[T] {

	return &Model[T]{db: db, tableName: tableName}
}

func (a *Model[T]) IsExist() bool {
	return a.db.Migrator().HasTable(a.tableName)
}
func (a *Model[T]) CreateTable(t T) error {
	if a.IsExist() {
		return nil
	}
	err := a.db.Table(a.tableName).AutoMigrate(t)
	return err
}
func (a *Model[T]) DeleteTable(t T) error {
	tx := a.db.Table(a.tableName).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(t)
	return tx.Error
}

func (a *Model[T]) Save(t T) error {
	t.SetCreateTime(time.Now())
	t.SetUpdateTime(time.Now())
	tx := a.db.Table(a.tableName).Create(t)
	return tx.Error
}
func (a *Model[T]) GetOne(id uint, t T) error {
	t.SetId(id)
	tx := a.db.Table(a.tableName).First(&t)
	if tx.Error == nil {
		return nil
	}
	return tx.Error
}

func (a *Model[T]) GetByIds(id []uint, ts *[]T) error {
	tx := a.db.Table(a.tableName).Where("`id` in (?) ", id).Find(ts)
	if tx.Error == nil {
		return nil
	}
	return tx.Error
}
func (a *Model[T]) DeleteOne(id uint, t T) error {
	tx := a.db.Table(a.tableName).Where("`id` = ? ", id).Delete(t)
	return tx.Error
}

func (a *Model[T]) Edit(t T) error {
	t.SetUpdateTime(time.Now())
	tx := a.db.Table(a.tableName).Updates(t)
	return tx.Error
}
func (a *Model[T]) EditForMap(id uint, data map[string]interface{}) error {
	tx := a.db.Table(a.tableName).Where("`id` = ? ", id).Updates(data)
	return tx.Error
}

func (a *Model[T]) NewModel(db *gorm.DB) *Model[T] {
	return &Model[T]{db: db, tableName: a.tableName}
}
func (a *Model[T]) Page(page *web.Page, ts *[]T) (int, error) {
	tx := a.db.Table(a.tableName).Order("`id` desc").Offset((page.PageNo - 1) * page.PageSize).Limit(page.PageSize).Find(ts)
	if tx.Error == nil {
		var num int64
		tx = a.db.Table(a.tableName).Count(&num)
		if tx.Error == nil {
			return int(num), nil
		}
	}
	return 0, tx.Error
}
