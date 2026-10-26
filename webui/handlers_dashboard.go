package webui

import (
	"github.com/chuccp/go-web-frame/web"
)

// DashboardData carries the four counters. Counting in SQL keeps the dashboard cheap:
// the Vue page used to pull four whole lists just to read their totals.
type DashboardData struct {
	Base
	SmtpCount  int
	MailCount  int
	TokenCount int
	SchedCount int
}

func (p *Pages) dashboard(req *web.Request) (any, error) {
	user, ok := p.requireUser(req)
	if !ok {
		return nil, nil
	}
	data := &DashboardData{Base: p.base(req, user, "dashboard.dashboard", true)}

	// Admins see the instance's totals, users only their own — the same scope the
	// list pages use.
	{
		q := p.smtpModel.Query()
		if !user.IsAdmin {
			q = q.Where("user_id = ?", user.Id)
		}
		n, err := q.Count()
		if err != nil {
			return p.renderErr(req, 0, "common.failed", err.Error())
		}
		data.SmtpCount = n
	}
	{
		q := p.mailModel.Query()
		if !user.IsAdmin {
			q = q.Where("user_id = ?", user.Id)
		}
		n, err := q.Count()
		if err != nil {
			return p.renderErr(req, 0, "common.failed", err.Error())
		}
		data.MailCount = n
	}
	{
		q := p.tokenModel.Query()
		if !user.IsAdmin {
			q = q.Where("user_id = ?", user.Id)
		}
		n, err := q.Count()
		if err != nil {
			return p.renderErr(req, 0, "common.failed", err.Error())
		}
		data.TokenCount = n
	}
	{
		q := p.scheduleModel.Query()
		if !user.IsAdmin {
			q = q.Where("user_id = ?", user.Id)
		}
		n, err := q.Count()
		if err != nil {
			return p.renderErr(req, 0, "common.failed", err.Error())
		}
		data.SchedCount = n
	}
	return p.render(req, "dashboard", data)
}
