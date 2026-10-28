package service

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	stmp "github.com/chuccp/http2smtp/smtp"
)

// newTestLogService wires a LogService against a throwaway sqlite file; Log only touches
// the log model, so no core.Context is needed.
func newTestLogService(t *testing.T) *LogService {
	t.Helper()
	conn, err := db.ConnectionSQLite(filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatalf("opening sqlite: %v", err)
	}
	sqlDB, err := conn.GetGorm().DB()
	if err != nil {
		t.Fatalf("getting sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	logModel := &model.LogModel{}
	if err := logModel.Init(conn, nil); err != nil {
		t.Fatalf("init log model: %v", err)
	}
	if err := logModel.CreateTable(); err != nil {
		t.Fatalf("create t_log: %v", err)
	}
	attachmentModel := &model.AttachmentModel{}
	if err := attachmentModel.Init(conn, nil); err != nil {
		t.Fatalf("init attachment model: %v", err)
	}
	if err := attachmentModel.CreateTable(); err != nil {
		t.Fatalf("create t_attachment: %v", err)
	}
	return &LogService{logModel: logModel, attachmentModel: attachmentModel}
}

// Each file becomes a row linked to the log, keeping the name the sender used even when
// two files share it.
func TestLogServiceRecordsAttachments(t *testing.T) {
	s := newTestLogService(t)
	files := []*stmp.File{
		{Name: "report.pdf", FilePath: "/cache/one.pdf"},
		{Name: "report.pdf", FilePath: "/cache/two.pdf"},
		{Name: "empty.pdf"},
	}
	if err := s.Log(5, nil, nil, files, "", "", "subject", "body", nil); err != nil {
		t.Fatalf("log: %v", err)
	}
	logs, err := s.logModel.Query().All()
	if err != nil || len(logs) != 1 {
		t.Fatalf("logs = %d (%v), want 1", len(logs), err)
	}
	rows, err := s.attachmentModel.FindByLogId(logs[0].Id)
	if err != nil {
		t.Fatalf("reading attachments: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("attachments = %d, want 2 (the pathless entry is skipped)", len(rows))
	}
	if rows[0].Name != "report.pdf" || rows[0].FilePath != "/cache/one.pdf" {
		t.Errorf("first attachment = %+v", rows[0])
	}
	if rows[1].Name != "report.pdf" || rows[1].FilePath != "/cache/two.pdf" {
		t.Errorf("second attachment = %+v", rows[1])
	}
	if rows[0].LogId != logs[0].Id {
		t.Errorf("log id = %d, want %d", rows[0].LogId, logs[0].Id)
	}
	if rows[0].CreateTime.IsZero() {
		t.Error("attachment create time was not set")
	}
}

// The entry records the owner, the formatted SMTP/recipient strings and a status that
// follows the send outcome.
func TestLogServiceRecordsSendOutcome(t *testing.T) {
	s := newTestLogService(t)
	smtp := &model.SMTP{Username: "u", Mail: "from@example.com"}
	mails := []*model.Mail{{Name: "R", Mail: "to@example.com"}}

	if err := s.Log(7, smtp, mails, nil, "tok", "token-value", "subject", "body", nil); err != nil {
		t.Fatalf("success log: %v", err)
	}
	if err := s.Log(8, nil, nil, nil, "", "", "subj2", "body2", errors.New("boom")); err != nil {
		t.Fatalf("failure log: %v", err)
	}
	if err := s.Log(9, nil, nil, nil, "", "", "subj3", "body3",
		stmp.ToUserNotFoundError([]string{"gone@example.com"})); err != nil {
		t.Fatalf("recipient-not-found log: %v", err)
	}

	rows, err := s.logModel.Query().Order("id asc").All()
	if err != nil {
		t.Fatalf("reading rows: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}

	ok := rows[0]
	if ok.UserId != 7 || ok.Token != "token-value" || ok.Name != "tok" {
		t.Errorf("owner fields = %+v", ok)
	}
	if ok.SMTP != `"u" <from@example.com>` {
		t.Errorf("SMTP = %q", ok.SMTP)
	}
	if ok.Mail != `"R" <to@example.com>` {
		t.Errorf("Mail = %q", ok.Mail)
	}
	if ok.Subject != "subject" || ok.Content != "body" {
		t.Errorf("subject/content = %q/%q", ok.Subject, ok.Content)
	}
	if ok.Status != entity.SUCCESS || ok.Result != "success" {
		t.Errorf("status/result = %d/%q, want SUCCESS/success", ok.Status, ok.Result)
	}

	failed := rows[1]
	if failed.Status != entity.ERROR || failed.Result != "boom" {
		t.Errorf("failure status/result = %d/%q, want ERROR/boom", failed.Status, failed.Result)
	}
	if failed.SMTP != "" || failed.Mail != "" {
		t.Errorf("failure entry should carry empty SMTP/mail, got %q/%q", failed.SMTP, failed.Mail)
	}

	missing := rows[2]
	if missing.Status != entity.WARM {
		t.Errorf("recipient-not-found status = %d, want WARM", missing.Status)
	}
	if missing.Result == "" || missing.Result == "success" {
		t.Errorf("recipient-not-found result = %q, want the error text", missing.Result)
	}
}
