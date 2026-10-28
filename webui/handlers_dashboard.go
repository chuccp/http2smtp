package webui

import (
	"fmt"
	"net"
	"strings"

	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
)

// DashboardData carries the four counters. Counting in SQL keeps the dashboard cheap
// rather than loading four whole lists just to read their totals.
type DashboardData struct {
	Base
	SmtpCount  int
	MailCount  int
	TokenCount int
	SchedCount int
	ApiURL     string
}

func (p *Pages) dashboard(view *web.Template) error {
	req := view.Request()
	user, ok := p.requireUser(view)
	if !ok {
		return nil
	}
	data := &DashboardData{Base: p.base(req, user, "dashboard.dashboard", true)}

	// The public API address shares the host the UI was reached on, on the configured
	// API port. Splitting the Host header keeps the management port out of the URL; an
	// IPv6 literal is wrapped back into brackets.
	apiPort := model.ApiPort
	if cfg, err := model.GetConfig(p.ctx.GetConfig()); err == nil && cfg.Api != nil && cfg.Api.Port > 0 {
		apiPort = cfg.Api.Port
	}
	hostname := req.Request().Host
	if h, _, err := net.SplitHostPort(hostname); err == nil {
		hostname = h
	}
	if strings.Contains(hostname, ":") {
		hostname = "[" + hostname + "]"
	}
	data.ApiURL = fmt.Sprintf("http://%s:%d/sendMail", hostname, apiPort)

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
