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

// Every password input gets the eye toggle; a new password field without it should
// fail this test rather than ship inconsistent.
func TestEveryPasswordFieldHasToggle(t *testing.T) {
	cases := []struct {
		page string
		data any
		min  int
	}{
		{"login", &LoginData{Base: templateTestBase()}, 1},
		{"setup", &SetupData{Base: templateTestBase(), Step: 1}, 1},
		{"setup", &SetupData{Base: templateTestBase(), Step: 2}, 2},
		{"smtp_form", &SmtpFormData{Base: templateTestBase()}, 1},
		{"users_form", &UserFormData{Base: templateTestBase()}, 1},
		{"password", &PasswordData{Base: templateTestBase()}, 3},
		{"settings", &SettingsData{Base: templateTestBase(), DbType: "mysql"}, 1},
	}
	for _, c := range cases {
		out := renderContent(t, c.page, c.data)
		passwords := strings.Count(out, `type="password"`)
		toggles := strings.Count(out, `data-password-toggle`)
		if passwords < c.min {
			t.Errorf("%s rendered %d password fields, want at least %d", c.page, passwords, c.min)
		}
		if toggles != passwords {
			t.Errorf("%s has %d password fields but %d eye toggles", c.page, passwords, toggles)
		}
	}
}

// The form components must preserve the attributes the pages and ui.js rely on, and a
// missing optional dict key must never leak "<no value>" into the markup.
func TestFormComponentsKeepFieldAttributes(t *testing.T) {
	smtp := renderContent(t, "smtp_form", &SmtpFormData{Base: templateTestBase(), Port: 25})
	for _, want := range []string{
		`name="port" value="25"`,
		`placeholder="smtp.gmail.com"`,
		`id="smtp-test-btn"`,
		`data-password-toggle`,
	} {
		if !strings.Contains(smtp, want) {
			t.Errorf("smtp form component output lost %q:\n%s", want, smtp)
		}
	}

	login := renderContent(t, "login", &LoginData{Base: templateTestBase()})
	for _, want := range []string{"autofocus", "login-logo", "login-input", "login-submit"} {
		if !strings.Contains(login, want) {
			t.Errorf("login page lost %q", want)
		}
	}

	cases := []struct {
		page string
		data any
	}{
		{"smtp_form", &SmtpFormData{Base: templateTestBase()}},
		{"mail_form", &MailFormData{Base: templateTestBase()}},
		{"users_form", &UserFormData{Base: templateTestBase()}},
		{"password", &PasswordData{Base: templateTestBase()}},
		{"token_form", &TokenFormData{Base: templateTestBase(), SelectedMails: map[uint]bool{}}},
		{"schedule_form", &ScheduleFormData{Base: templateTestBase(), Method: "POST"}},
		{"smtp_send", &SmtpSendData{Base: templateTestBase()}},
		{"token_send", &TokenSendData{Base: templateTestBase()}},
		{"login", &LoginData{Base: templateTestBase()}},
	}
	for _, c := range cases {
		out := renderContent(t, c.page, c.data)
		if strings.Contains(out, "<no value>") {
			t.Errorf("%s form leaked a missing dict key into the markup", c.page)
		}
	}
}

// Checkboxes/radios and the setup defaults travel through the shared components too;
// their names, ids and default values are what the page scripts and handlers read.
func TestFormComponentsPreserveControls(t *testing.T) {
	users := renderContent(t, "users_form", &UserFormData{Base: templateTestBase()})
	for _, want := range []string{`name="isUse"`, `name="isAdmin"`} {
		if !strings.Contains(users, want) {
			t.Errorf("users form lost %s", want)
		}
	}

	token := renderContent(t, "token_form", &TokenFormData{Base: templateTestBase(), SelectedMails: map[uint]bool{}})
	for _, want := range []string{
		`name="state" value="on" checked`,
		`name="state" value="off"`,
		`class="copy-btn"`,
		`data-copy-input`,
		`name="token"`,
	} {
		if !strings.Contains(token, want) {
			t.Errorf("token form lost %q", want)
		}
	}

	schedule := renderContent(t, "schedule_form", &ScheduleFormData{Base: templateTestBase(), Method: "POST"})
	for _, want := range []string{
		`name="method" value="GET"`,
		`name="method" value="POST" checked`,
		`name="useTemplate" id="use-template"`,
		`name="onlyByError"`,
		`name="isUse"`,
	} {
		if !strings.Contains(schedule, want) {
			t.Errorf("schedule form lost %q", want)
		}
	}
	if strings.Contains(schedule, `<span class="hint">`) {
		t.Error("schedule form still renders an input hint as a second line")
	}
	if !strings.Contains(schedule, "0 0 9 * * *") {
		t.Error("schedule cron example is missing from the input placeholder")
	}

	setup := renderContent(t, "setup", &SetupData{Base: templateTestBase(), Step: 1, DbType: "mysql"})
	for _, want := range []string{`name="mysqlPort" value="3306"`, `name="mysqlCharset" value="utf8mb4"`, `data-password-toggle`} {
		if !strings.Contains(setup, want) {
			t.Errorf("setup form lost %q", want)
		}
	}
}

// A successful test send alerts first and only closes the dialog once the user has
// acknowledged it; it must not close silently or add a toast behind the alert.
func TestSendTestMailAlertsThenCloses(t *testing.T) {
	cases := []struct {
		page string
		data any
	}{
		{"smtp_send", &SmtpSendData{Base: templateTestBase()}},
		{"token_send", &TokenSendData{Base: templateTestBase()}},
	}
	for _, c := range cases {
		out := renderContent(t, c.page, c.data)
		alertAt := strings.Index(out, "UI.alert(form.getAttribute('data-success')")
		doneAt := strings.Index(out, "UI.dialogDone(btn, '')")
		if alertAt < 0 || doneAt < 0 || alertAt > doneAt {
			t.Errorf("%s must alert the send result before closing the dialog", c.page)
		}
		if strings.Contains(out, "UI.toast(") {
			t.Errorf("%s should not toast a test send; the alert is the confirmation", c.page)
		}
	}
}

// Add/edit saves and the settings save announce success with an alert; toasts are for
// background actions (delete/trigger), not for a completed save.
func TestSaveSuccessDoesNotToast(t *testing.T) {
	src, err := staticFS.ReadFile("static/ui.js")
	if err != nil {
		t.Fatal(err)
	}
	js := string(src)
	if strings.Contains(js, "toast(successText") {
		t.Error("save success still uses a toast")
	}
	for _, want := range []string{"alert(successText).then(done)", "alert(successText).then(refresh)"} {
		if !strings.Contains(js, want) {
			t.Errorf("save success is missing %q", want)
		}
	}
}

// The token field's copy button must copy the input's current value (the token can be
// edited) and restore its icon after the checkmark flash.
func TestCopyFieldCopiesCurrentInputValue(t *testing.T) {
	src, err := staticFS.ReadFile("static/app.js")
	if err != nil {
		t.Fatal(err)
	}
	js := string(src)
	for _, want := range []string{
		"[data-copy], [data-copy-input]",
		"hasAttribute('data-copy-input')",
		"querySelector('input')",
		"el._copyOriginal = el.innerHTML",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("copy handler is missing %q", want)
		}
	}
}

// The forms that keep two fields per row need a wider dialog so each label+control
// column fits; the width travels on the add/edit triggers.
func TestFormDialogWidthsMatchFieldCount(t *testing.T) {
	base := templateTestBase()
	base.User = &NavUser{Id: 1, Name: "admin", IsAdmin: true}
	for _, c := range []struct {
		page  string
		data  any
		width string
	}{
		{"smtp_list", &SmtpListData{Base: base}, "860px"},
		{"token_list", &TokenListData{Base: base}, "860px"},
		{"schedule_list", &ScheduleListData{Base: base}, "860px"},
		{"mail_list", &MailListData{Base: base}, "600px"},
		{"users_list", &UserListData{Base: base}, "600px"},
	} {
		out := renderContent(t, c.page, c.data)
		if !strings.Contains(out, `data-dialog-width="`+c.width+`"`) {
			t.Errorf("%s add/edit dialog does not open at %s", c.page, c.width)
		}
	}
}

// Forms with only a few controls use the login-like one-column .form-simple layout;
// only the larger forms keep the two-column .form-grid.
func TestSimpleFormsUseOneColumn(t *testing.T) {
	cases := []struct {
		page string
		data any
	}{
		{"mail_form", &MailFormData{Base: templateTestBase()}},
		{"users_form", &UserFormData{Base: templateTestBase()}},
		{"password", &PasswordData{Base: templateTestBase()}},
		{"smtp_send", &SmtpSendData{Base: templateTestBase()}},
		{"token_send", &TokenSendData{Base: templateTestBase()}},
		{"setup", &SetupData{Base: templateTestBase(), Step: 2}},
	}
	for _, c := range cases {
		out := renderContent(t, c.page, c.data)
		if !strings.Contains(out, "form-simple") {
			t.Errorf("%s is not using the simple one-column form layout", c.page)
		}
		if strings.Contains(out, `class="form-grid"`) {
			t.Errorf("%s still renders a two-column form grid", c.page)
		}
	}
}

// The schedule method radios reveal the request body and make it required only for
// POST; the save handler repeats that requirement server-side.
func TestSchedulePostRevealsRequiredBody(t *testing.T) {
	get := renderContent(t, "schedule_form", &ScheduleFormData{Base: templateTestBase(), Method: "GET"})
	if !strings.Contains(get, `id="body-field" hidden`) {
		t.Error("GET schedule should hide the request body field")
	}
	if strings.Contains(get, `name="body" required`) {
		t.Error("GET schedule should not require a request body")
	}
	post := renderContent(t, "schedule_form", &ScheduleFormData{Base: templateTestBase(), Method: "POST"})
	if strings.Contains(post, `id="body-field" hidden`) {
		t.Error("POST schedule should show the request body field")
	}
	if !strings.Contains(post, `name="body" required`) {
		t.Error("POST schedule should require a request body")
	}
	for _, want := range []string{"syncBody", "bodyField.hidden = !post", "bodyInput.required = post"} {
		if !strings.Contains(post, want) {
			t.Errorf("schedule method script is missing %q", want)
		}
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
