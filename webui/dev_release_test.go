//go:build release

package webui

import "testing"

func TestReleaseBuildUsesEmbeddedAssets(t *testing.T) {
	if devMode {
		t.Fatal("release build must serve embedded templates and static files")
	}
}
