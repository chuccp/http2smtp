package webui

import (
	"bytes"
	"strings"
	"testing"
)

func templateTestBase() Base {
	return Base{Lang: "zh-cn", L: translatorFor("zh-cn")}
}

func renderContent(t *testing.T, page string, data any) string {
	t.Helper()
	tmpls, err := parseTemplates()
	if err != nil {
		t.Fatalf("parsing templates: %v", err)
	}
	tmpl := tmpls[page]
	if tmpl == nil {
		t.Fatalf("no template registered for %q", page)
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "content", data); err != nil {
		t.Fatalf("rendering %s: %v", page, err)
	}
	return buf.String()
}

// The create form must post an empty id; a nonzero id turns the same POST into an
// update of that row. A failed edit must keep the id so it does not become a create.
func TestFormHiddenIdKeepsCreateAndEditApart(t *testing.T) {
	cases := []struct {
		page   string
		create func() any
		edit   func() any
	}{
		{"smtp_form",
			func() any { return &SmtpFormData{Base: templateTestBase()} },
			func() any { return &SmtpFormData{Base: templateTestBase(), Id: 7} }},
		{"mail_form",
			func() any { return &MailFormData{Base: templateTestBase()} },
			func() any { return &MailFormData{Base: templateTestBase(), Id: 7} }},
		{"token_form",
			func() any {
				return &TokenFormData{Base: templateTestBase(), SelectedMails: map[uint]bool{}}
			},
			func() any {
				return &TokenFormData{Base: templateTestBase(), Id: 7, SelectedMails: map[uint]bool{}}
			}},
		{"schedule_form",
			func() any { return &ScheduleFormData{Base: templateTestBase()} },
			func() any { return &ScheduleFormData{Base: templateTestBase(), Id: 7} }},
		{"users_form",
			func() any { return &UserFormData{Base: templateTestBase()} },
			func() any { return &UserFormData{Base: templateTestBase(), Id: 7} }},
	}
	for _, c := range cases {
		create := renderContent(t, c.page, c.create())
		if !strings.Contains(create, `name="id" value=""`) {
			t.Errorf("%s create form does not submit an empty id", c.page)
		}
		edit := renderContent(t, c.page, c.edit())
		if !strings.Contains(edit, `name="id" value="7"`) {
			t.Errorf("%s edit form lost its id", c.page)
		}
	}
}

// The record selectors must carry data-picker so ui.js turns them into the tag +
// nested picker control (and keeps the native select for validation/FormData).
func TestRecordSelectorsArePickers(t *testing.T) {
	token := renderContent(t, "token_form", &TokenFormData{
		Base: templateTestBase(), SelectedMails: map[uint]bool{},
	})
	if strings.Count(token, "data-picker") < 2 {
		t.Errorf("token form should enhance both SMTP and recipient selects:\n%s", token)
	}
	if !strings.Contains(token, `name="smtpId"`) || !strings.Contains(token, `name="receiveEmailIds"`) {
		t.Error("token form lost one of its two record selects")
	}
	if !strings.Contains(token, "multiple") {
		t.Error("recipient select lost the multiple attribute")
	}

	schedule := renderContent(t, "schedule_form", &ScheduleFormData{Base: templateTestBase()})
	if !strings.Contains(schedule, `name="tokenId"`) || !strings.Contains(schedule, "data-picker") {
		t.Error("schedule form token select is not a picker")
	}

	send := renderContent(t, "smtp_send", &SmtpSendData{Base: templateTestBase()})
	if !strings.Contains(send, `name="toEmail"`) || !strings.Contains(send, "data-picker") {
		t.Error("smtp send recipient select is not a picker")
	}
	// The placeholder must stay the first option so an add dialog cannot silently send
	// to the first contact.
	if !strings.Contains(send, `<option value="">`) {
		t.Error("smtp send recipient select lost its placeholder option")
	}
}

// Clearing the days field maps to 0; both cases must keep their extra guard in place.
func TestLogCleanupGuardsEmptyAndZero(t *testing.T) {
	base := templateTestBase()
	base.User = &NavUser{Id: 1, Name: "admin", IsAdmin: true}
	out := renderContent(t, "logs_list", &LogListData{Base: base})
	for _, want := range []string{"raw === ''", "days === 0", "UI.confirm"} {
		if !strings.Contains(out, want) {
			t.Errorf("log cleanup script lost its %q guard", want)
		}
	}
}
