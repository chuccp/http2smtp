package service

import (
	"path/filepath"
	"testing"

	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/http2smtp/model"
)

// newTestScheduleService wires a ScheduleService against a throwaway sqlite file;
// Edit/Save only touch the token and schedule models, so no core.Context is needed.
func newTestScheduleService(t *testing.T) (*ScheduleService, *model.TokenModel, *model.ScheduleModel) {
	t.Helper()
	conn, err := db.ConnectionSQLite(filepath.Join(t.TempDir(), "schedule.db"))
	if err != nil {
		t.Fatalf("opening sqlite: %v", err)
	}
	sqlDB, err := conn.GetGorm().DB()
	if err != nil {
		t.Fatalf("getting sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	tokenModel := &model.TokenModel{}
	if err := tokenModel.Init(conn, nil); err != nil {
		t.Fatalf("init token model: %v", err)
	}
	if err := tokenModel.CreateTable(); err != nil {
		t.Fatalf("create t_token: %v", err)
	}
	scheduleModel := &model.ScheduleModel{}
	if err := scheduleModel.Init(conn, nil); err != nil {
		t.Fatalf("init schedule model: %v", err)
	}
	if err := scheduleModel.CreateTable(); err != nil {
		t.Fatalf("create t_schedule: %v", err)
	}
	s := &ScheduleService{tokenModel: tokenModel, scheduleModel: scheduleModel}
	return s, tokenModel, scheduleModel
}

func mustCreateTokenRow(t *testing.T, tokenModel *model.TokenModel) *model.Token {
	t.Helper()
	tk := &model.Token{UserId: 1, Name: "t", Token: "schedule-tok", SMTPId: 1}
	if err := tokenModel.Save(tk); err != nil {
		t.Fatalf("save token: %v", err)
	}
	return tk
}

// The header list is transient; only HeaderStr survives the database. A nil list means
// "the caller did not send one" and keeps the stored column, an empty list clears it,
// and a non-empty list is marshalled into the column.
func TestScheduleEditHeaderHandling(t *testing.T) {
	s, tokenModel, scheduleModel := newTestScheduleService(t)
	tk := mustCreateTokenRow(t, tokenModel)

	row := &model.Schedule{
		UserId: 1, Name: "sched", TokenId: tk.Id, Cron: "0 0 * * * ?",
		Url: "http://example.com/hook", Method: "GET",
		HeaderStr: `[{"name":"X","value":"1"}]`, IsUse: true,
	}
	if err := s.Save(row); err != nil {
		t.Fatalf("save: %v", err)
	}

	// nil Headers: untouched.
	edit := *row
	edit.Headers = nil
	edit.Body = "changed"
	if err := s.Edit(&edit); err != nil {
		t.Fatalf("edit with nil headers: %v", err)
	}
	got, err := scheduleModel.FindByPK(row.Id)
	if err != nil || got == nil {
		t.Fatalf("re-read: %v", err)
	}
	if got.HeaderStr != `[{"name":"X","value":"1"}]` {
		t.Errorf("nil headers changed the stored column to %q", got.HeaderStr)
	}

	// Empty list: cleared.
	edit.Headers = []*model.Header{}
	if err := s.Edit(&edit); err != nil {
		t.Fatalf("edit with empty headers: %v", err)
	}
	if got, _ = scheduleModel.FindByPK(row.Id); got.HeaderStr != "" {
		t.Errorf("empty headers left the column at %q, want cleared", got.HeaderStr)
	}

	// Non-empty list: marshalled into the column.
	edit.Headers = []*model.Header{{Name: "A", Value: "2"}}
	if err := s.Edit(&edit); err != nil {
		t.Fatalf("edit with headers: %v", err)
	}
	if got, _ = scheduleModel.FindByPK(row.Id); got.HeaderStr != `[{"name":"A","value":"2"}]` {
		t.Errorf("stored headers = %q, want the marshalled list", got.HeaderStr)
	}

	// The detail read expands the column back into the list.
	one, err := s.GetOne(int(row.Id), 1)
	if err != nil {
		t.Fatalf("GetOne: %v", err)
	}
	if len(one.Headers) != 1 || one.Headers[0].Name != "A" || one.Headers[0].Value != "2" {
		t.Errorf("expanded headers = %+v, want the stored pair", one.Headers)
	}
}

// A schedule pointing at a token that no longer exists is refused, not silently saved:
// the runner would fire it into "token not found" forever.
func TestScheduleEditRejectsMissingToken(t *testing.T) {
	s, tokenModel, scheduleModel := newTestScheduleService(t)
	tk := mustCreateTokenRow(t, tokenModel)

	row := &model.Schedule{UserId: 1, Name: "sched", TokenId: tk.Id, IsUse: true}
	if err := s.Save(row); err != nil {
		t.Fatalf("save: %v", err)
	}

	edit := *row
	edit.TokenId = 9999
	if err := s.Edit(&edit); err == nil {
		t.Fatal("Edit accepted a schedule whose token does not exist")
	}
	edit.TokenId = 9999
	if err := s.Save(&model.Schedule{UserId: 1, Name: "new", TokenId: 9999}); err == nil {
		t.Fatal("Save accepted a schedule whose token does not exist")
	}
	if got, _ := scheduleModel.FindByPK(row.Id); got.TokenId != tk.Id {
		t.Errorf("stored token id = %d, want the original %d", got.TokenId, tk.Id)
	}
}

// Turning a switch off writes false. The update-by-pk path drops zero values, so a
// regression to it would silently ignore every disable.
func TestScheduleEditPersistsFalseSwitches(t *testing.T) {
	s, tokenModel, scheduleModel := newTestScheduleService(t)
	tk := mustCreateTokenRow(t, tokenModel)

	row := &model.Schedule{
		UserId: 1, Name: "sched", TokenId: tk.Id, IsUse: true,
		UseTemplate: true, IsOnlySendByError: true, Template: "tpl",
	}
	if err := s.Save(row); err != nil {
		t.Fatalf("save: %v", err)
	}

	edit := *row
	edit.IsUse = false
	edit.UseTemplate = false
	edit.IsOnlySendByError = false
	edit.Template = ""
	edit.Body = "" // clearing the body must persist too
	if err := s.Edit(&edit); err != nil {
		t.Fatalf("edit: %v", err)
	}
	got, err := scheduleModel.FindByPK(row.Id)
	if err != nil || got == nil {
		t.Fatalf("re-read: %v", err)
	}
	if got.IsUse {
		t.Error("is_use stayed true after being turned off")
	}
	if got.UseTemplate {
		t.Error("use_template stayed true after being turned off")
	}
	if got.IsOnlySendByError {
		t.Error("is_only_send_by_error stayed true after being turned off")
	}
	if got.Template != "" {
		t.Errorf("template = %q, want cleared", got.Template)
	}
}

// A malformed header column must not break the list or detail reads; it degrades to an
// empty header list.
func TestExpandHeadersMalformedJSON(t *testing.T) {
	for _, column := range []string{"not json", `[{"name":]`, ""} {
		sd := &model.Schedule{HeaderStr: column}
		expandHeaders(sd)
		if len(sd.Headers) != 0 {
			t.Errorf("HeaderStr %q expanded to %+v, want empty", column, sd.Headers)
		}
	}
	// Blank column clears a stale list as well.
	sd := &model.Schedule{HeaderStr: "  ", Headers: []*model.Header{{Name: "stale"}}}
	expandHeaders(sd)
	if len(sd.Headers) != 0 {
		t.Errorf("blank column kept %+v, want empty", sd.Headers)
	}
}
