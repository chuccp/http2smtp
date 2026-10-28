package util

import "testing"

func TestValidateURL(t *testing.T) {
	for _, ok := range []string{
		"http://example.com",
		"https://example.com/path?q=1",
		"HTTP://EXAMPLE.COM",
		"http://127.0.0.1:12567/api/scheduleTestApi",
	} {
		if err := ValidateURL(ok); err != nil {
			t.Errorf("ValidateURL(%q) = %v, want nil", ok, err)
		}
	}
	for _, bad := range []string{
		"",
		"   ",
		"example.com",
		"ftp://example.com",
		"javascript:alert(1)",
		"http://",
		"http:///path",
		"//example.com",
	} {
		if err := ValidateURL(bad); err == nil {
			t.Errorf("ValidateURL(%q) = nil, want an error", bad)
		}
	}
}
