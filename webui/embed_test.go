package webui

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmbeddedUIAssetsAreComplete(t *testing.T) {
	required := []string{
		"templates/layout.html",
		"templates/login.html",
		"templates/compatibility.html",
		"static/app.css",
		"static/ui.css",
		"static/app.js",
		"static/ui.js",
		"static/favicon.svg",
	}
	for _, path := range required {
		data, err := uiAsset(path)
		if err != nil {
			t.Errorf("embedded asset %s: %v", path, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("embedded asset %s is empty", path)
		}
	}

	templateFiles := 0
	if err := fs.WalkDir(templatesFS, "templates", func(path string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() && strings.HasSuffix(path, ".html") {
			templateFiles++
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if templateFiles < 20 {
		t.Fatalf("embedded template count = %d, expected the complete UI", templateFiles)
	}
}

func uiAsset(path string) ([]byte, error) {
	if strings.HasPrefix(path, "templates/") {
		return templatesFS.ReadFile(path)
	}
	return staticFS.ReadFile(path)
}
