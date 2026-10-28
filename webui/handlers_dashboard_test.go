package webui

import "testing"

// The API address must normalise every Host-header shape, including bracketed and bare
// IPv6 literals, without doubling the brackets.
func TestAPIURL(t *testing.T) {
	cases := map[string]string{
		"example.com:12566": "http://example.com:12567/sendMail",
		"example.com":       "http://example.com:12567/sendMail",
		"localhost:12566":   "http://localhost:12567/sendMail",
		"10.0.0.5:8080":     "http://10.0.0.5:12567/sendMail",
		"[::1]:12566":       "http://[::1]:12567/sendMail",
		"[::1]":             "http://[::1]:12567/sendMail",
	}
	for host, want := range cases {
		if got := apiURL(host, 12567); got != want {
			t.Errorf("apiURL(%q) = %q, want %q", host, got, want)
		}
	}
}
