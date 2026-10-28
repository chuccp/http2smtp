package webui

import (
	"github.com/chuccp/go-web-frame/web"
)

// DashboardData carries the four counters. Counting in SQL keeps the dashboard cheap
// rather than loading four whole lists just to read their totals.
type DashboardData struct {
	Base
	SmtpCount  int
	MailCount  int
	TokenCount int
	SchedCount int
}

func (p *Pages) dashboard(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireUser(view)
	if !ok {
		return nil
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
			return p.renderErr(view, 0, "common.failed", err.Error())
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
			return p.renderErr(view, 0, "common.failed", err.Error())
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
			return p.renderErr(view, 0, "common.failed", err.Error())
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
			return p.renderErr(view, 0, "common.failed", err.Error())
		}
		data.SchedCount = n
	}
	view.View("dashboard").Layout(layoutFor("dashboard")).Data(data)
	return nil
}
