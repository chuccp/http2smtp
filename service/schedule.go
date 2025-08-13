package service

import (
	"encoding/json"
	"errors"
	"github.com/chuccp/smtp2http/db"
	"github.com/chuccp/smtp2http/smtp"
	"github.com/chuccp/smtp2http/util"
	"github.com/chuccp/smtp2http/web"
)

type Schedule struct {
	db    *db.DB
	token *Token
}

func NewSchedule(db *db.DB, token *Token) *Schedule {
	return &Schedule{db: db, token: token}
}
func (schedule *Schedule) GetPage(page *web.Page) (any, error) {
	return schedule.db.GetScheduleModel().Page(page)
}
func (schedule *Schedule) Edit(sd *db.Schedule) error {
	if sd.Headers != nil && len(sd.Headers) > 0 {
		jsonData, err := json.Marshal(sd.Headers)
		if err != nil {
			return err
		} else {
			sd.HeaderStr = string(jsonData)
		}
	}

	v, err := schedule.db.GetTokenModel().GetOneByToken(sd.Token)
	if err != nil {
		return err
	}
	if v == nil {
		return errors.New("token not found")
	}
	return schedule.db.GetScheduleModel().Edit(sd)
}
func (schedule *Schedule) Save(sd *db.Schedule) error {
	v, err := schedule.db.GetTokenModel().GetOneByToken(sd.Token)
	if err != nil {
		return err
	}
	if v == nil {
		return errors.New("token not found")
	}
	return schedule.db.GetScheduleModel().Save(sd)

}

func (schedule *Schedule) GetOne(id int) (*db.Schedule, error) {

	one, err := schedule.db.GetScheduleModel().GetOne(uint(id))
	if err != nil {
		return nil, err
	}
	headerStr := one.HeaderStr
	if util.IsNotBlank(headerStr) {
		err := json.Unmarshal([]byte(headerStr), &one.Headers)
		if err != nil {
			one.Headers = []*db.Header{}
		}
	} else {
		one.Headers = []*db.Header{}
	}

	token := one.Token
	byToken, err := schedule.db.GetTokenModel().GetOneByToken(token)
	if err != nil {
		return nil, err
	}
	one.TokenId = byToken.Id
	return one, nil
}

func (schedule *Schedule) SendMail(sd *db.Schedule) error {
	byToken, err := schedule.token.GetOneByToken(sd.Token)
	if err != nil {
		return err
	}
	if byToken == nil {
		return errors.New("token not found")
	}
	_, err = smtp.SendAPIMail(sd, byToken.SMTP, byToken.ReceiveEmails)
	return err
}
