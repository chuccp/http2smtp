package webui

import "testing"

func TestPageBaseCarriesBuildVersion(t *testing.T) {
	page := NewPages("v1.2.3-test")
	base := page.newBase("en", "auth.login")
	if base.Version != "v1.2.3-test" {
		t.Errorf("base version = %q", base.Version)
	}
}

// safeNext only lets same-site paths through; anything a browser could resolve to
// another origin falls back to the dashboard.
func TestSafeNext(t *testing.T) {
	cases := map[string]string{
		"":                     "/dashboard",
		"/dashboard":           "/dashboard",
		"/smtp?name=x":         "/smtp?name=x",
		"dashboard":            "/dashboard",
		"//evil.example":       "/dashboard",
		"/\\evil.example":      "/dashboard",
		"https://evil.example": "/dashboard",
		"/\nevil":              "/dashboard",
	}
	for in, want := range cases {
		if got := safeNext(in); got != want {
			t.Errorf("safeNext(%q) = %q, want %q", in, got, want)
		}
	}
}
