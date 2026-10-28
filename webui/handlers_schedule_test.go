package webui

import "testing"

// A forged or stale select value must fall back to the default instead of becoming an
// arbitrary Content-Type header on the scheduled request.
func TestNormalizeContentType(t *testing.T) {
	cases := map[string]string{
		"":                                  defaultContentType,
		"application/json":                  "application/json",
		"application/x-www-form-urlencoded": "application/x-www-form-urlencoded",
		"text/plain":                        "text/plain",
		"application/xml":                   "application/xml",
		"text/html":                         defaultContentType,
		"application/json\r\nX-Evil: 1":     defaultContentType,
	}
	for in, want := range cases {
		if got := normalizeContentType(in); got != want {
			t.Errorf("normalizeContentType(%q) = %q, want %q", in, got, want)
		}
	}
}
