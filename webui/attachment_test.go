package webui

import (
	"path/filepath"
	"testing"

	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/http2smtp/model"
)

// The mail-log list counts attachment rows first and falls back to the legacy files JSON
// for entries written before the table existed.
func TestAttachmentCountsUsesRowsAndJSONFallback(t *testing.T) {
	conn, err := db.ConnectionSQLite(filepath.Join(t.TempDir(), "attachments.db"))
	if err != nil {
		t.Fatalf("opening sqlite: %v", err)
	}
	sqlDB, err := conn.GetGorm().DB()
	if err != nil {
		t.Fatalf("getting sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	attachmentModel := &model.AttachmentModel{}
	if err := attachmentModel.Init(conn, nil); err != nil {
		t.Fatalf("init attachment model: %v", err)
	}
	if err := attachmentModel.CreateTable(); err != nil {
		t.Fatalf("create t_attachment: %v", err)
	}
	if err := attachmentModel.Saves([]*model.Attachment{
		{LogId: 1, Name: "a.txt", FilePath: "/cache/a"},
		{LogId: 1, Name: "b.txt", FilePath: "/cache/b"},
	}); err != nil {
		t.Fatalf("seed attachments: %v", err)
	}

	p := &Pages{attachmentModel: attachmentModel}
	rows := []*model.Log{
		{Id: 1, Files: `[{"name":"a.txt"},{"name":"b.txt"}]`},
		{Id: 2, Files: `[{"name":"legacy.txt"}]`},
		{Id: 3},
	}
	counts := p.attachmentCounts(rows)
	if counts[1] != 2 {
		t.Errorf("log 1 count = %d, want 2 from the attachment rows", counts[1])
	}
	if counts[2] != 1 {
		t.Errorf("log 2 count = %d, want 1 from the legacy files JSON", counts[2])
	}
	if counts[3] != 0 {
		t.Errorf("log 3 count = %d, want 0", counts[3])
	}
}
