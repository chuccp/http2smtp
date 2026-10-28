package webui

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/http2smtp/model"
)

// newOwnershipTestPages wires the models the token/schedule ownership checks use
// against a throwaway sqlite file, without a core.Context.
func newOwnershipTestPages(t *testing.T) *Pages {
	t.Helper()
	conn, err := db.ConnectionSQLite(filepath.Join(t.TempDir(), "ownership.db"))
	if err != nil {
		t.Fatalf("opening sqlite: %v", err)
	}
	sqlDB, err := conn.GetGorm().DB()
	if err != nil {
		t.Fatalf("getting sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	smtpModel := &model.SMTPModel{}
	if err := smtpModel.Init(conn, nil); err != nil {
		t.Fatalf("init smtp model: %v", err)
	}
	if err := smtpModel.CreateTable(); err != nil {
		t.Fatalf("create t_SMTP: %v", err)
	}
	mailModel := &model.MailModel{}
	if err := mailModel.Init(conn, nil); err != nil {
		t.Fatalf("init mail model: %v", err)
	}
	if err := mailModel.CreateTable(); err != nil {
		t.Fatalf("create t_mail: %v", err)
	}
	tokenModel := &model.TokenModel{}
	if err := tokenModel.Init(conn, nil); err != nil {
		t.Fatalf("init token model: %v", err)
	}
	if err := tokenModel.CreateTable(); err != nil {
		t.Fatalf("create t_token: %v", err)
	}
	userModel := &model.UserModel{}
	if err := userModel.Init(conn, nil); err != nil {
		t.Fatalf("init user model: %v", err)
	}
	if err := userModel.CreateTable(); err != nil {
		t.Fatalf("create t_user: %v", err)
	}
	return &Pages{smtpModel: smtpModel, mailModel: mailModel, tokenModel: tokenModel, userModel: userModel}
}

type ownershipFixture struct {
	adminSmtp, bobSmtp           uint
	adminMail, bobMail           uint
	adminToken, bobToken, bobOff uint
}

func seedOwnership(t *testing.T, p *Pages) ownershipFixture {
	t.Helper()
	saveSMTP := func(userId uint, name string) uint {
		row := &model.SMTP{UserId: userId, Name: name, Host: "localhost", Port: 25, Mail: name + "@example.com", Username: "u", Password: "p"}
		if err := p.smtpModel.Save(row); err != nil {
			t.Fatalf("save smtp %s: %v", name, err)
		}
		return row.Id
	}
	saveMail := func(userId uint, name string) uint {
		row := &model.Mail{UserId: userId, Name: name, Mail: name + "@example.com"}
		if err := p.mailModel.Save(row); err != nil {
			t.Fatalf("save mail %s: %v", name, err)
		}
		return row.Id
	}
	saveToken := func(userId uint, name string, smtpId, mailId uint, state uint8) uint {
		row := &model.Token{
			UserId:          userId,
			Name:            name,
			Token:           name + "-value",
			SMTPId:          smtpId,
			ReceiveEmailIds: strconv.Itoa(int(mailId)),
			State:           state,
		}
		if err := p.tokenModel.Save(row); err != nil {
			t.Fatalf("save token %s: %v", name, err)
		}
		return row.Id
	}

	f := ownershipFixture{}
	f.adminSmtp = saveSMTP(1, "admin-smtp")
	f.bobSmtp = saveSMTP(2, "bob-smtp")
	f.adminMail = saveMail(1, "admin-mail")
	f.bobMail = saveMail(2, "bob-mail")
	f.adminToken = saveToken(1, "admin-token", f.adminSmtp, f.adminMail, 0)
	f.bobToken = saveToken(2, "bob-token", f.bobSmtp, f.bobMail, 0)
	f.bobOff = saveToken(2, "bob-disabled-token", f.bobSmtp, f.bobMail, 1)
	return f
}

// Legacy rows (user_id = 0) stay editable by their caller; a normal row stays with its
// original owner even when an admin edits it.
func TestResourceOwnerId(t *testing.T) {
	if got := resourceOwnerId(2, 1); got != 2 {
		t.Errorf("normal row owner = %d, want 2", got)
	}
	if got := resourceOwnerId(0, 1); got != 1 {
		t.Errorf("legacy row owner = %d, want caller 1", got)
	}
}

// The token dialog must list the owner's SMTP servers and contacts, not the admin's.
func TestFillTokenFormOptionsUsesTokenOwner(t *testing.T) {
	p := newOwnershipTestPages(t)
	f := seedOwnership(t, p)

	bob := &TokenFormData{SelectedMails: map[uint]bool{f.bobMail: true}}
	if err := p.fillTokenFormOptions(bob, 2, f.bobSmtp); err != nil {
		t.Fatalf("fill bob options: %v", err)
	}
	if len(bob.SmtpOptions) != 1 || bob.SmtpOptions[0].Label != "bob-smtp (localhost)" || !bob.SmtpOptions[0].Selected {
		t.Errorf("bob smtp options = %+v", bob.SmtpOptions)
	}
	if len(bob.MailOptions) != 1 || bob.MailOptions[0].Label != "bob-mail <bob-mail@example.com>" || !bob.MailOptions[0].Selected {
		t.Errorf("bob mail options = %+v", bob.MailOptions)
	}

	admin := &TokenFormData{SelectedMails: map[uint]bool{}}
	if err := p.fillTokenFormOptions(admin, 1, f.adminSmtp); err != nil {
		t.Fatalf("fill admin options: %v", err)
	}
	if len(admin.SmtpOptions) != 1 || admin.SmtpOptions[0].Label != "admin-smtp (localhost)" {
		t.Errorf("admin smtp options = %+v", admin.SmtpOptions)
	}
	if len(admin.MailOptions) != 1 || admin.MailOptions[0].Label != "admin-mail <admin-mail@example.com>" {
		t.Errorf("admin mail options = %+v", admin.MailOptions)
	}
}

// The same resource ids must be rejected for a token owned by another user.
func TestCheckTokenResourcesAreOwnerScoped(t *testing.T) {
	p := newOwnershipTestPages(t)
	f := seedOwnership(t, p)

	bobMail := []string{strconv.Itoa(int(f.bobMail))}
	if err := p.checkTokenResources(2, f.bobSmtp, bobMail); err != nil {
		t.Fatalf("bob's own resources rejected: %v", err)
	}
	if err := p.checkTokenResources(2, f.adminSmtp, bobMail); err == nil || err.Error() != "smtp not found" {
		t.Errorf("admin smtp accepted for bob's token: %v", err)
	}
	if err := p.checkTokenResources(2, f.bobSmtp, []string{strconv.Itoa(int(f.adminMail))}); err == nil || err.Error() != "mail not found" {
		t.Errorf("admin mail accepted for bob's token: %v", err)
	}
	if err := p.checkTokenResources(1, f.adminSmtp, []string{strconv.Itoa(int(f.adminMail))}); err != nil {
		t.Errorf("admin's own resources rejected: %v", err)
	}
}

// A schedule may only reference an active token that belongs to the schedule owner.
func TestActiveTokenOwnedBy(t *testing.T) {
	p := newOwnershipTestPages(t)
	f := seedOwnership(t, p)

	if tok, err := p.activeTokenOwnedBy(2, f.bobToken); err != nil || tok == nil {
		t.Errorf("bob's active token not found: %v", err)
	}
	if tok, err := p.activeTokenOwnedBy(2, f.adminToken); err != nil || tok != nil {
		t.Errorf("admin's token accepted for bob's schedule: %v", err)
	}
	if tok, err := p.activeTokenOwnedBy(2, f.bobOff); err != nil || tok != nil {
		t.Errorf("disabled token accepted for a schedule: %v", err)
	}
	if tok, err := p.activeTokenOwnedBy(1, f.adminToken); err != nil || tok == nil {
		t.Errorf("admin's active token not found: %v", err)
	}
}

// A regular user's schedule form offers the schedule owner's active tokens; an admin
// may point the schedule at any user's active token.
func TestFillTokenOptionsUsesScheduleOwner(t *testing.T) {
	p := newOwnershipTestPages(t)
	f := seedOwnership(t, p)

	bob := &ScheduleFormData{}
	if err := p.fillTokenOptions(bob, &model.User{Id: 2}, 2, f.bobToken); err != nil {
		t.Fatalf("fill bob tokens: %v", err)
	}
	if len(bob.TokenOptions) != 1 || bob.TokenOptions[0].Label != "bob-token" || !bob.TokenOptions[0].Selected {
		t.Errorf("bob token options = %+v", bob.TokenOptions)
	}

	admin := &ScheduleFormData{}
	if err := p.fillTokenOptions(admin, &model.User{Id: 1, IsAdmin: true}, 1, f.adminToken); err != nil {
		t.Fatalf("fill admin tokens: %v", err)
	}
	if len(admin.TokenOptions) != 2 {
		t.Fatalf("admin should see every active token, got %+v", admin.TokenOptions)
	}
	labels := map[string]bool{}
	for _, opt := range admin.TokenOptions {
		labels[opt.Label] = true
		if opt.Value == strconv.Itoa(int(f.adminToken)) && !opt.Selected {
			t.Error("admin's token should stay selected")
		}
	}
	if !labels["admin-token"] || !labels["bob-token"] {
		t.Errorf("admin token options = %+v", admin.TokenOptions)
	}
}
