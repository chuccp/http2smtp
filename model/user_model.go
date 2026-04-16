package model

import (
	"time"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/model"
)

type UserModel struct {
	*model.EntryModel[*User]
	db      *db.DB
	context *core.Context
}

func (t *UserModel) Init(db *db.DB, context *core.Context) error {
	t.db = db
	t.context = context
	t.EntryModel = model.NewEntryModel[*User](t.db, t.GetTableName())
	return nil
}

func (t *UserModel) ReNew(db *db.DB, c *core.Context) core.IModel {
	return &UserModel{
		EntryModel: model.NewEntryModel[*User](db, t.GetTableName()),
		db:         db,
		context:    c,
	}
}

func (t *UserModel) GetTableName() string {
	return "t_user"
}

func (t *UserModel) FindOneByName(name string) (*User, error) {
	return t.FindOne("name = ?", name)
}

func (t *UserModel) FindOneByToken(token string) (*User, error) {
	return nil, nil
}

func (u *User) SetCreateTime(createTime time.Time) {
	u.CreateTime = createTime
}

func (u *User) SetUpdateTime(updateTime time.Time) {
	u.UpdateTime = updateTime
}

func (u *User) GetId() uint {
	return u.Id
}

func (u *User) SetId(id uint) {
	u.Id = id
}
