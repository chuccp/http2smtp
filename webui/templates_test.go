package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
)

func templateTestBase() Base {
	return Base{
		Lang: "zh-cn",
		L:    translatorFor("zh-cn"),
		User: &NavUser{Id: 1, Name: "tester", IsAdmin: true},
	}
}

func renderContent(t *testing.T, page string, data any) string {
	t.Helper()
	server := templateTestServer(t, func(view *web.Template) error {
		view.View(page).Layout(layoutFor(page)).Data(data)
		return nil
	})
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("rendering %s: status %d: %s", page, recorder.Code, recorder.Body.String())
	}
	return recorder.Body.String()
}

func templateTestServer(t *testing.T, handler web.TemplateHandler) http.Handler {
	t.Helper()
	templatesOnce.Do(func() {
		templateOptionForTests, templatesErr = web.LoadTemplateOption(
			templatesFS,
			"templates",
			web.WithTemplateFuncs(templateFuncs),
			web.WithLayouts("layout", "bare"),
			web.WithPartials("partials"),
		)
	})
	if templatesErr != nil {
		t.Fatalf("loading templates: %v", templatesErr)
	}

	server := web.DefaultServer()
	server.GetTemplate("/", handler, templateOptionForTests)
	return server.GetHandler()
}

var (
	templatesOnce          sync.Once
	templateOptionForTests web.TemplateOption
	templatesErr           error
)

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
	// The token form's two pickers page and search; only the multi-select recipient
	// picker pins the pending selection above the list.
	for _, want := range []string{`data-search`, `data-page-size="10"`, `data-page-info=`, `data-selected-header`} {
		if !strings.Contains(token, want) {
			t.Errorf("token form picker is missing %q", want)
		}
	}
	if strings.Count(token, "data-page-size=") != 2 {
		t.Error("both token form pickers should be paged")
	}

	schedule := renderContent(t, "schedule_form", &ScheduleFormData{Base: templateTestBase()})
	if !strings.Contains(schedule, `name="tokenId"`) || !strings.Contains(schedule, "data-picker") {
		t.Error("schedule form token select is not a picker")
	}
	// The single-select token picker pages and searches too, but carries no header.
	for _, want := range []string{`data-search`, `data-page-size="10"`, `data-page-info=`} {
		if !strings.Contains(schedule, want) {
			t.Errorf("schedule token picker is missing %q", want)
		}
	}
	if strings.Contains(schedule, "data-selected-header") {
		t.Error("single-select token picker should not pin a selection header")
	}

	// The test send picker is its own paged multi-select over /api/mail, not a <select>.
	send := renderContent(t, "smtp_send", &SmtpSendData{Base: templateTestBase()})
	for _, want := range []string{
		`name="toEmail"`,
		`id="recipient-picker"`,
		`data-page-info=`,
		`data-search-placeholder=`,
		`data-admin-label=`,
		`data-mine-label=`,
	} {
		if !strings.Contains(send, want) {
			t.Errorf("smtp send picker is missing %q", want)
		}
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
		`list="smtp-host-options"`,
		`<datalist id="smtp-host-options">`,
		`value="smtp.qq.com"`,
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

// The schedule method radios reveal the request body and its content type, and make the
// body required only for POST; the save handler repeats that requirement server-side.
func TestSchedulePostRevealsRequiredBody(t *testing.T) {
	get := renderContent(t, "schedule_form", &ScheduleFormData{Base: templateTestBase(), Method: "GET"})
	if !strings.Contains(get, `id="body-field" hidden`) {
		t.Error("GET schedule should hide the request body field")
	}
	if strings.Contains(get, `name="body" required`) {
		t.Error("GET schedule should not require a request body")
	}
	if !strings.Contains(get, `id="content-type-field" hidden`) {
		t.Error("GET schedule should hide the content type field")
	}
	post := renderContent(t, "schedule_form", &ScheduleFormData{Base: templateTestBase(), Method: "POST", ContentType: "application/json"})
	if strings.Contains(post, `id="body-field" hidden`) {
		t.Error("POST schedule should show the request body field")
	}
	if !strings.Contains(post, `name="body" required`) {
		t.Error("POST schedule should require a request body")
	}
	if strings.Contains(post, `id="content-type-field" hidden`) {
		t.Error("POST schedule should show the content type field")
	}
	if !strings.Contains(post, `<option value="application/json" selected>`) {
		t.Error("content type field lost its selected JSON option")
	}
	for _, want := range []string{"syncBody", "bodyField.hidden = !post", "bodyInput.required = post", "contentTypeField.hidden = !post"} {
		if !strings.Contains(post, want) {
			t.Errorf("schedule method script is missing %q", want)
		}
	}
	// The URL field carries the quick-fill button for the instance's request-test echo.
	if !strings.Contains(post, `id="use-test-url"`) || !strings.Contains(post, `data-test-url=`) {
		t.Error("schedule URL field lost its test-link button")
	}
}

// Triggering a schedule reports through a dialog; the row delete keeps the lighter toast.
func TestScheduleTriggerReportsViaAlert(t *testing.T) {
	out := renderContent(t, "schedule_list", &ScheduleListData{
		Base: templateTestBase(),
		Rows: []*model.Schedule{{Id: 7, Name: "demo", Cron: "0 0 9 * * *", Url: "http://127.0.0.1:12566/api/scheduleTestApi", Method: "GET", IsUse: true}},
	})
	if !strings.Contains(out, `data-alert="任务触发成功"`) {
		t.Error("trigger-now should report through an alert")
	}
	if strings.Contains(out, `data-href="/api/schedule/trigger/7" data-flash`) {
		t.Error("trigger-now still uses a toast")
	}
}

// The SMTP list stars rows owned by an administrator, and only for an admin viewer.
func TestSmtpListStarsAdminCreatedRows(t *testing.T) {
	rows := []*model.SMTP{{Id: 1, Name: "corp", Host: "smtp.example.com", CreatedByAdmin: true}}
	admin := renderContent(t, "smtp_list", &SmtpListData{Base: templateTestBase(), Rows: rows})
	if !strings.Contains(admin, "admin-star") {
		t.Error("admin viewer should see the admin-created star")
	}

	base := templateTestBase()
	base.User = &NavUser{Id: 2, Name: "bob"}
	owner := renderContent(t, "smtp_list", &SmtpListData{Base: base, Rows: rows})
	if strings.Contains(owner, "admin-star") {
		t.Error("non-admin viewer should not see the admin-created star")
	}
}

// The other admin-visible lists star admin-created rows the same way the SMTP list
// does; a regular user never sees the marker.
func TestAdminListsStarAdminCreatedRows(t *testing.T) {
	adminBase := templateTestBase()
	cases := []struct {
		page string
		data any
	}{
		{"mail_list", &MailListData{Base: adminBase, Rows: []*model.Mail{{Id: 1, Name: "a", Mail: "a@example.com", CreatedByAdmin: true}}}},
		{"token_list", &TokenListData{Base: adminBase, Rows: []*model.Token{{Id: 1, Name: "tok", CreatedByAdmin: true}}}},
		{"schedule_list", &ScheduleListData{Base: adminBase, Rows: []*model.Schedule{{Id: 1, Name: "job", CreatedByAdmin: true}}}},
	}
	for _, c := range cases {
		if out := renderContent(t, c.page, c.data); !strings.Contains(out, "admin-star") {
			t.Errorf("%s should star an admin-created row for an admin viewer", c.page)
		}
	}

	base := templateTestBase()
	base.User = &NavUser{Id: 2, Name: "bob"}
	nonAdmin := []struct {
		page string
		data any
	}{
		{"mail_list", &MailListData{Base: base, Rows: []*model.Mail{{Id: 1, Name: "a", Mail: "a@example.com", CreatedByAdmin: true}}}},
		{"token_list", &TokenListData{Base: base, Rows: []*model.Token{{Id: 1, Name: "tok", CreatedByAdmin: true}}}},
		{"schedule_list", &ScheduleListData{Base: base, Rows: []*model.Schedule{{Id: 1, Name: "job", CreatedByAdmin: true}}}},
	}
	for _, c := range nonAdmin {
		if out := renderContent(t, c.page, c.data); strings.Contains(out, "admin-star") {
			t.Errorf("%s should not star rows for a regular viewer", c.page)
		}
	}
}

// Admins can test-send through every row; a regular user only through servers they own.
func TestSmtpSendButtonFollowsAdmin(t *testing.T) {
	other := []*model.SMTP{{Id: 5, UserId: 9, Name: "other"}}
	admin := renderContent(t, "smtp_list", &SmtpListData{Base: templateTestBase(), Rows: other})
	if !strings.Contains(admin, `/smtp/send/5`) {
		t.Error("admin should get a send-test button on another user's server")
	}

	base := templateTestBase()
	base.User = &NavUser{Id: 2, Name: "bob"}
	owner := renderContent(t, "smtp_list", &SmtpListData{Base: base, Rows: other})
	if strings.Contains(owner, `/smtp/send/5`) {
		t.Error("regular user should not get a send-test button on another user's server")
	}
	own := renderContent(t, "smtp_list", &SmtpListData{Base: base, Rows: []*model.SMTP{{Id: 6, UserId: 2, Name: "own"}}})
	if !strings.Contains(own, `/smtp/send/6`) {
		t.Error("regular user should get a send-test button on their own server")
	}
}

// Both test-send dialogs can attach files; the shared partial and ui.js helpers turn
// the selection into base64 entries in the JSON body.
func TestSendDialogsOfferAttachments(t *testing.T) {
	cases := []struct {
		page string
		data any
	}{
		{"smtp_send", &SmtpSendData{Base: templateTestBase()}},
		{"token_send", &TokenSendData{Base: templateTestBase()}},
	}
	for _, c := range cases {
		out := renderContent(t, c.page, c.data)
		for _, want := range []string{`type="file"`, `name="files"`, `class="file-chips"`, `UI.fileChips`, `UI.readFiles`, `files: files`} {
			if !strings.Contains(out, want) {
				t.Errorf("%s is missing %q", c.page, want)
			}
		}
	}
}

// The operation-log page renders the translated action and target plus the operator.
func TestOplogListRendersLabels(t *testing.T) {
	out := renderContent(t, "oplog_list", &OplogListData{
		Base: templateTestBase(),
		Rows: []OplogRow{{
			OperationLog: &model.OperationLog{
				Id: 1, UserId: 1, UserName: "111111",
				Action: entity.OpDelete, TargetType: entity.TargetSchedule, TargetName: "nightly",
			},
			ActionLabel: "删除",
			ActionClass: "tag-off",
			TargetLabel: "定时任务",
		}},
	})
	for _, want := range []string{
		"删除", "定时任务", "nightly", "111111", "tag-off",
		`data-page-dialog="/oplogs/1"`,
		`data-href="/api/oplog/1"`,
		`id="oplog-cleanup"`,
		`name="user"`,
		`name="action"`,
		`name="target"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("operation log page is missing %q", want)
		}
	}
}

// A regular user reads their own operation log without the maintenance actions.
func TestOplogListMaintenanceIsAdminOnly(t *testing.T) {
	base := templateTestBase()
	base.User = &NavUser{Id: 2, Name: "bob"}
	out := renderContent(t, "oplog_list", &OplogListData{
		Base: base,
		Rows: []OplogRow{{
			OperationLog: &model.OperationLog{Id: 1, UserId: 2, UserName: "bob"},
			ActionLabel:  "登录", TargetLabel: "用户管理",
		}},
	})
	for _, banned := range []string{`id="oplog-cleanup"`, "/api/oplog/1", `name="user"`} {
		if strings.Contains(out, banned) {
			t.Errorf("regular user should not see %q", banned)
		}
	}
	if !strings.Contains(out, "/oplogs/1") {
		t.Error("regular user should still open their own entry")
	}
}

// Only admins can delete or clean up mail logs; regular users still open the details.
func TestLogListMaintenanceIsAdminOnly(t *testing.T) {
	rows := []*model.Log{{Id: 7, UserId: 1, Subject: "s", StatusStr: "success"}}
	admin := renderContent(t, "logs_list", &LogListData{Base: templateTestBase(), Rows: rows})
	if !strings.Contains(admin, `id="cleanup-btn"`) || !strings.Contains(admin, `data-href="/api/log/7"`) {
		t.Error("admin should get the mail-log delete and cleanup actions")
	}

	base := templateTestBase()
	base.User = &NavUser{Id: 2, Name: "bob"}
	user := renderContent(t, "logs_list", &LogListData{Base: base, Rows: rows})
	if strings.Contains(user, `id="cleanup-btn"`) || strings.Contains(user, "/api/log/7") {
		t.Error("regular user should not get the mail-log delete or cleanup actions")
	}
	if !strings.Contains(user, "/logs/7") {
		t.Error("regular user should still open the log detail")
	}
}

// The operation-log detail dialog shows the entry's translated fields.
func TestOplogDetailRendersEntry(t *testing.T) {
	out := renderContent(t, "oplog_detail", &OplogDetailData{
		Base: templateTestBase(),
		Entry: OplogRow{
			OperationLog: &model.OperationLog{
				Id: 3, UserId: 1, UserName: "111111",
				Action: entity.OpDisable, TargetType: entity.TargetSchedule, TargetId: 9, TargetName: "nightly",
			},
			ActionLabel: "停用",
			ActionClass: "tag-off",
			TargetLabel: "定时任务",
		},
	})
	for _, want := range []string{"停用", "定时任务", "nightly", "111111", "对象 ID", "9"} {
		if !strings.Contains(out, want) {
			t.Errorf("operation log detail is missing %q", want)
		}
	}
}

// The dashboard shows the public API address with a copy button.
func TestDashboardShowsAPIAddress(t *testing.T) {
	out := renderContent(t, "dashboard", &DashboardData{
		Base:   templateTestBase(),
		ApiURL: "http://example.com:12567/sendMail",
	})
	for _, want := range []string{
		"http://example.com:12567/sendMail",
		`data-copy="http://example.com:12567/sendMail"`,
		"API 地址",
		"JSON（POST）",
		"curl -X POST http://example.com:12567/sendMail",
		`"token":"YOUR_TOKEN"`,
		"表单（POST）",
		`-H 'Content-Type: application/x-www-form-urlencoded'`,
		`-d 'token=YOUR_TOKEN&subject=test&content=hello&recipients=ops@example.com'`,
		"GET 参数",
		`curl 'http://example.com:12567/sendMail?token=YOUR_TOKEN&subject=test&content=hello&recipients=ops@example.com'`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("dashboard is missing %q", want)
		}
	}
}

// After a port change the old origin is gone, so the interstitial counts down and then
// jumps to the same host on the new management port; a same-port restart just reloads.
func TestRestartingPageCountsDownToNewPort(t *testing.T) {
	moved := renderContent(t, "restarting", &RestartingData{
		Base:        templateTestBase(),
		Message:     "restarted",
		ManagePort:  12568,
		ApiPort:     12567,
		CurrentPort: 12566,
		PortChanged: true,
	})
	for _, want := range []string{
		`data-port="12568"`,
		`data-template="`,
		`{seconds}`,
		`location.hostname`,
		`location.href = target`,
	} {
		if !strings.Contains(moved, want) {
			t.Errorf("restarting page is missing %q", want)
		}
	}

	same := renderContent(t, "restarting", &RestartingData{
		Base: templateTestBase(), Message: "restarted", ManagePort: 12566, ApiPort: 12567, PortChanged: false,
	})
	if !strings.Contains(same, "location.reload") {
		t.Error("same-port restart should reload the current origin")
	}
	if strings.Contains(same, "countdown") {
		t.Error("same-port restart should not render a countdown")
	}
}

// Every management list can re-fetch its current content without leaving the page, and
// the admin-only filter checkbox applies on change instead of waiting for Search.
func TestListPagesOfferRefresh(t *testing.T) {
	cases := []struct {
		page   string
		data   any
		filter bool
	}{
		{"smtp_list", &SmtpListData{Base: templateTestBase()}, true},
		{"mail_list", &MailListData{Base: templateTestBase()}, true},
		{"token_list", &TokenListData{Base: templateTestBase()}, true},
		{"schedule_list", &ScheduleListData{Base: templateTestBase()}, true},
		{"logs_list", &LogListData{Base: templateTestBase()}, true},
		{"users_list", &UserListData{Base: templateTestBase()}, false},
	}
	for _, c := range cases {
		out := renderContent(t, c.page, c.data)
		if !strings.Contains(out, "data-refresh") {
			t.Errorf("%s has no refresh button", c.page)
		}
		if c.filter && !strings.Contains(out, "data-auto-submit") {
			t.Errorf("%s has no auto-submitting admin filter", c.page)
		}
	}
}

// The row toggle offers the opposite action of the current state, each with its own
// toast, and hits the dedicated toggle endpoint.
func TestScheduleListOffersToggle(t *testing.T) {
	out := renderContent(t, "schedule_list", &ScheduleListData{
		Base: templateTestBase(),
		Rows: []*model.Schedule{
			{Id: 7, Name: "on", Url: "http://example.com/a", IsUse: true},
			{Id: 8, Name: "off", Url: "http://example.com/b", IsUse: false},
		},
	})
	for _, want := range []string{
		`data-href="/api/schedule/toggle/7" data-flash="任务已停用"`,
		`>停用</button>`,
		`data-href="/api/schedule/toggle/8" data-flash="任务已启用"`,
		`>启用</button>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("schedule list toggle is missing %q", want)
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

// The login captcha carries its challenge into the form and keeps the submit button
// disabled until the script has seen every click.
// The login form carries the captcha hook and the hidden answer fields; the challenge
// itself opens in a modal and is fetched by the script, so the page stays compact.
func TestLoginRendersCaptchaHook(t *testing.T) {
	out := renderContent(t, "login", &LoginData{Base: templateTestBase()})
	for _, want := range []string{
		`name="captchaId"`,
		`name="captchaXs"`,
		`name="captchaYs"`,
		`data-captcha-title=`,
		`data-captcha-prompt-label=`,
		`data-captcha-progress=`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("login captcha hook is missing %q", want)
		}
	}
	if strings.Contains(out, "data:image/png") {
		t.Error("the login page should fetch the challenge instead of embedding it")
	}
}

// The general settings card carries the request-body cap next to the ports.
func TestSettingsRendersMaxBodySize(t *testing.T) {
	base := templateTestBase()
	base.User = &NavUser{Id: 1, Name: "admin", IsAdmin: true}
	out := renderContent(t, "settings", &SettingsData{
		Base:        base,
		WebPort:     12566,
		ApiPort:     12567,
		MaxBodySize: "10M",
		DbType:      "sqlite",
		SqliteFile:  "d-mail.db",
	})
	for _, want := range []string{
		`name="maxBodySize"`,
		`value="10M"`,
		`required`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("settings maxBodySize field is missing %q", want)
		}
	}
}

func TestSettingsDockerPortNoteUsesAlignedStyle(t *testing.T) {
	base := templateTestBase()
	base.User = &NavUser{Id: 1, Name: "admin", IsAdmin: true}
	out := renderContent(t, "settings", &SettingsData{
		Base:     base,
		IsDocker: true,
		WebPort:  12566,
		ApiPort:  12567,
		DbType:   "sqlite",
	})
	if !strings.Contains(out, `class="settings-docker-note"`) {
		t.Fatal("Docker port note is missing its aligned style")
	}
	if strings.Contains(out, `class="flash flash-err">Docker`) {
		t.Fatal("Docker port note should not use the full-width card flash")
	}
}
