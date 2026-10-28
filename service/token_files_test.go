package service

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/smtp"
)

func TestAppendBase64Files(t *testing.T) {
	svc := &TokenService{cachePath: t.TempDir()}
	data := base64.StdEncoding.EncodeToString([]byte("hello"))
	files, err := svc.appendBase64Files(nil, []entity.File{{Name: "note.txt", Data: data}})
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if len(files) != 1 || files[0].Name != "note.txt" || files[0].FilePath == "" {
		t.Fatalf("files = %+v, want one named note.txt with a cache path", files)
	}
	t.Cleanup(func() {
		if files[0].File != nil {
			_ = files[0].File.Close()
		}
		_ = os.Remove(files[0].FilePath)
	})
	written, err := os.ReadFile(files[0].FilePath)
	if err != nil {
		t.Fatalf("reading the written file: %v", err)
	}
	if string(written) != "hello" {
		t.Errorf("written content = %q, want hello", written)
	}
}

// A decode failure returns the files already written, so the caller's cleanup can remove
// them, and never claims an entry it did not write.
func TestAppendBase64FilesInvalidData(t *testing.T) {
	svc := &TokenService{cachePath: t.TempDir()}
	good := base64.StdEncoding.EncodeToString([]byte("ok"))
	files, err := svc.appendBase64Files(nil, []entity.File{
		{Name: "first.txt", Data: good},
		{Name: "bad.txt", Data: "not base64!"},
	})
	if err == nil {
		t.Fatal("invalid base64 should be an error")
	}
	if len(files) != 1 || files[0].Name != "first.txt" {
		t.Fatalf("files = %+v, want only the successfully written one", files)
	}
	_ = files[0].File.Close()
	_ = os.Remove(files[0].FilePath)
}

// Attachments that never reach a log entry must be closed and removed, or the cache
// keeps copies nothing references.
func TestRemoveSendFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	removeSendFiles([]*smtp.File{
		{File: f, Name: "report.txt", FilePath: path},
		nil,
	})

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("attachment still on disk after cleanup: %v", err)
	}
	if _, err := f.Write([]byte("y")); err == nil {
		t.Error("attachment handle should be closed")
	}
}
