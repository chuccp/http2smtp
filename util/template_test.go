package util

import "testing"

func TestParseTemplate(t *testing.T) {
	out, err := ParseTemplate("Status: {{.status}}", `{"status":"ok"}`)
	if err != nil || out != "Status: ok" {
		t.Errorf("ParseTemplate = %q, %v; want \"Status: ok\"", out, err)
	}

	out, err = ParseTemplate("{{.nested.key}}", `{"nested":{"key":"v"}}`)
	if err != nil || out != "v" {
		t.Errorf("nested ParseTemplate = %q, %v; want v", out, err)
	}

	if _, err := ParseTemplate("{{.x}}", "not json"); err == nil {
		t.Error("invalid JSON should be reported")
	}
	if _, err := ParseTemplate("{{.x", "{}"); err == nil {
		t.Error("invalid template should be reported")
	}
}
