package rest

import "testing"

func TestAttachmentByName(t *testing.T) {
	const files = `[{"name":"a.log","filePath":"/cache/a.log"},{"name":"b.pdf","filePath":"/cache/b.pdf"}]`

	t.Run("returns the stored path", func(t *testing.T) {
		file, err := attachmentByName(files, "b.pdf")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if file.FilePath != "/cache/b.pdf" {
			t.Errorf("got %q, want %q", file.FilePath, "/cache/b.pdf")
		}
	})

	t.Run("empty files means no attachments", func(t *testing.T) {
		if _, err := attachmentByName("", "a.log"); err == nil {
			t.Error("expected an error for an empty files field")
		}
	})

	t.Run("invalid json is rejected", func(t *testing.T) {
		if _, err := attachmentByName("not json", "a.log"); err == nil {
			t.Error("expected an error for invalid JSON")
		}
	})

	t.Run("name not in the record is rejected", func(t *testing.T) {
		if _, err := attachmentByName(files, "/etc/passwd"); err == nil {
			t.Error("expected an error for a name that is not recorded")
		}
	})

	t.Run("entries without a path are skipped", func(t *testing.T) {
		if _, err := attachmentByName(`[{"name":"a.log"}]`, "a.log"); err == nil {
			t.Error("expected an error when the stored path is blank")
		}
	})

	t.Run("paths are matched by name only", func(t *testing.T) {
		// A caller must not be able to reach a path that is not recorded under its own name
		if _, err := attachmentByName(files, "a.log/../b.pdf"); err == nil {
			t.Error("expected an error for a traversal-style name")
		}
	})
}

// Two attachments sent under one filename are stored under distinct cache paths; the
// index keeps each download tied to its own copy.
func TestAttachmentAt(t *testing.T) {
	const files = `[{"name":"report.pdf","filePath":"/cache/one.pdf"},{"name":"report.pdf","filePath":"/cache/two.pdf"}]`

	first, err := attachmentAt(files, 0)
	if err != nil || first.FilePath != "/cache/one.pdf" {
		t.Errorf("index 0 = %+v (%v), want /cache/one.pdf", first, err)
	}
	second, err := attachmentAt(files, 1)
	if err != nil || second.FilePath != "/cache/two.pdf" {
		t.Errorf("index 1 = %+v (%v), want /cache/two.pdf", second, err)
	}
	// The name lookup is what old links use; it resolves to the first copy.
	if one, err := attachmentByName(files, "report.pdf"); err != nil || one.FilePath != "/cache/one.pdf" {
		t.Errorf("name lookup = %+v (%v), want the first copy", one, err)
	}

	for _, bad := range []int{-1, 2} {
		if _, err := attachmentAt(files, bad); err == nil {
			t.Errorf("index %d should be out of range", bad)
		}
	}
	if _, err := attachmentAt(`[null,{"name":"x","filePath":""}]`, 0); err == nil {
		t.Error("a null entry should not resolve")
	}
	if _, err := attachmentAt(`[null,{"name":"x","filePath":""}]`, 1); err == nil {
		t.Error("an entry without a path should not resolve")
	}
}
