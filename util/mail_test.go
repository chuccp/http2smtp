package util

import "testing"

func TestParseMail(t *testing.T) {
	cases := []struct {
		in, name, mail string
	}{
		{`"Alice" <alice@example.com>`, "Alice", "alice@example.com"},
		{"Bob <bob@example.com>", "Bob", "bob@example.com"},
		{"carol@example.com", "", "carol@example.com"},
	}
	for _, c := range cases {
		name, addr, err := ParseMail(c.in)
		if err != nil {
			t.Errorf("ParseMail(%q) error: %v", c.in, err)
			continue
		}
		if name != c.name || addr != c.mail {
			t.Errorf("ParseMail(%q) = %q, %q; want %q, %q", c.in, name, addr, c.name, c.mail)
		}
	}

	if _, _, err := ParseMail("not-an-address"); err == nil {
		t.Error("ParseMail accepted an address without an @")
	}
}

func TestFormatMail(t *testing.T) {
	if got := FormatMail("", "a@example.com"); got != "a@example.com" {
		t.Errorf("FormatMail without a name = %q", got)
	}
	if got := FormatMail("Alice", "a@example.com"); got != `"Alice" <a@example.com>` {
		t.Errorf("FormatMail = %q", got)
	}
}
