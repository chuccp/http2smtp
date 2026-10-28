package rest

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/http2smtp/model"
)

// newOplogTestModel wires an operation-log model against a throwaway sqlite file.
func newOplogTestModel(t *testing.T) (*model.OperationLogModel, *db.DB) {
	t.Helper()
	conn, err := db.ConnectionSQLite(filepath.Join(t.TempDir(), "oplog.db"))
	if err != nil {
		t.Fatalf("opening sqlite: %v", err)
	}
	sqlDB, err := conn.GetGorm().DB()
	if err != nil {
		t.Fatalf("getting sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	m := &model.OperationLogModel{}
	if err := m.Init(conn, nil); err != nil {
		t.Fatalf("init model: %v", err)
	}
	if err := m.CreateTable(); err != nil {
		t.Fatalf("create t_operation_log: %v", err)
	}
	return m, conn
}

// insertOplog writes a row with an explicit create_time, which Save would overwrite.
func insertOplog(t *testing.T, conn *db.DB, created time.Time) {
	t.Helper()
	err := conn.GetGorm().Exec(
		"INSERT INTO t_operation_log (user_id, action, target_type, target_id, target_name, ip, create_time) VALUES (?,?,?,?,?,?,?)",
		1, "create", "smtp", 1, "demo", "127.0.0.1", created,
	).Error
	if err != nil {
		t.Fatalf("insert operation log: %v", err)
	}
}

func TestOplogCleanupByAge(t *testing.T) {
	m, conn := newOplogTestModel(t)
	insertOplog(t, conn, time.Now().AddDate(0, 0, -40))
	insertOplog(t, conn, time.Now().AddDate(0, 0, -10))
	insertOplog(t, conn, time.Now())

	n, err := cleanupOplogRows(m, 30)
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if n != 1 {
		t.Errorf("cleanup deleted %d rows, want 1", n)
	}
	rows, err := m.Query().All()
	if err != nil || len(rows) != 2 {
		t.Errorf("remaining rows = %d (%v), want 2", len(rows), err)
	}

	n, err = cleanupOplogRows(m, 0)
	if err != nil {
		t.Fatalf("clear: %v", err)
	}
	if n != 2 {
		t.Errorf("clear deleted %d rows, want 2", n)
	}
	rows, err = m.Query().All()
	if err != nil || len(rows) != 0 {
		t.Errorf("remaining after clear = %d (%v), want 0", len(rows), err)
	}
}

func TestOplogDeleteOne(t *testing.T) {
	m, conn := newOplogTestModel(t)
	insertOplog(t, conn, time.Now())

	rows, err := m.Query().All()
	if err != nil || len(rows) != 1 {
		t.Fatalf("seed rows = %v (%v), want 1", rows, err)
	}
	one, err := deleteOplogRow(m, rows[0].Id)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if one == nil || one.TargetName != "demo" {
		t.Errorf("deleted row = %+v, want target name demo", one)
	}
	if _, err := deleteOplogRow(m, rows[0].Id); err == nil {
		t.Error("deleting a missing row should report an error")
	}
}
