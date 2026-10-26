package webui

import (
	"strings"
	"testing"
)

func init() {
	if err := loadI18n(); err != nil {
		panic(err)
	}
}

// The four locale files have to stay in key-for-key sync — the generator enforces it
// at build time, and this guards the committed JSON against hand edits.
func TestI18n_AllLanguagesCarryTheSameKeys(t *testing.T) {
	var refLang string
	var ref map[string]string
	for _, lang := range Languages {
		tr := translators[lang]
		if tr == nil {
			t.Fatalf("no translator loaded for %q", lang)
		}
		if ref == nil {
			refLang, ref = lang, tr.m
			continue
		}
		for key := range ref {
			if _, ok := tr.m[key]; !ok {
				t.Errorf("%s is missing key %q (present in %s)", lang, key, refLang)
			}
		}
		for key := range tr.m {
			if _, ok := ref[key]; !ok {
				t.Errorf("%s has key %q that %s does not carry", lang, key, refLang)
			}
		}
	}
}

func TestTranslatorT(t *testing.T) {
	en := translatorFor("en")
	if got := en.T("common.save"); got != "Save" {
		t.Errorf("en common.save = %q, want %q", got, "Save")
	}

	zh := translatorFor("zh-cn")
	if got := zh.T("common.save"); got != "保存" {
		t.Errorf("zh-cn common.save = %q, want %q", got, "保存")
	}

	// An unknown key surfaces as the key itself rather than vanishing.
	if got := en.T("no.such.key"); got != "no.such.key" {
		t.Errorf("unknown key = %q, want the key back", got)
	}

	// An unknown language falls back to the default rather than nil-dereferencing.
	if got := translatorFor("xx").T("common.save"); got != "保存" {
		t.Errorf("fallback language common.save = %q, want %q", got, "保存")
	}
}

func TestPickLang(t *testing.T) {
	cases := []struct {
		cookie, header, want string
	}{
		{"zh-tw", "", "zh-tw"},
		{"ja", "zh-cn", "ja"}, // cookie beats header
		{"bogus", "ja-JP,ja;q=0.9", "ja"},
		{"", "zh-CN,zh;q=0.9", "zh-cn"},
		{"", "zh-TW,zh;q=0.9", "zh-tw"},
		{"", "zh-HK,zh;q=0.9", "zh-tw"},
		{"", "en-US,en;q=0.9", "en"},
		{"", "fr;q=0.9,de", "zh-cn"}, // no usable tag → default
		{"", "", "zh-cn"},
		// q-ordering: the higher q wins even when listed last
		{"", "en;q=0.5,ja;q=0.8", "ja"},
	}
	for _, c := range cases {
		if got := pickLang(c.cookie, c.header); got != c.want {
			t.Errorf("pickLang(%q, %q) = %q, want %q", c.cookie, c.header, got, c.want)
		}
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	got := parseAcceptLanguage("en-US,en;q=0.9,zh-CN;q=0.8")
	if len(got) != 3 || got[0] != "en-us" || got[1] != "en" || got[2] != "zh-cn" {
		t.Errorf("parseAcceptLanguage order = %v", got)
	}
	if tags := parseAcceptLanguage(""); len(tags) != 0 {
		t.Errorf("empty header produced %v", tags)
	}
	if tags := parseAcceptLanguage("*"); len(tags) != 0 {
		t.Errorf("wildcard produced %v", tags)
	}
}

// {name} placeholders consume the args positionally; values without placeholders are
// returned unchanged even when args are passed.
func TestTranslatorPlaceholderSubstitution(t *testing.T) {
	if err := loadI18n(); err != nil {
		t.Fatal(err)
	}
	en := translatorFor("en")
	got := en.T("log.cleanupDone", 12)
	if want := "Cleaned up 12 entries"; got != want {
		t.Errorf("cleanupDone = %q, want %q", got, want)
	}
	// Two placeholders, two args, in order of appearance
	got = en.T("settings.restartPortChanged", 12566, 12567)
	if !strings.Contains(got, "12566") || !strings.Contains(got, "12567") {
		t.Errorf("restartPortChanged lost its args: %q", got)
	}
	// A value whose braces are Go-template syntax must pass through untouched
	got = en.T("schedule.templatePlaceholder")
	if !strings.Contains(got, "{{.key}}") {
		t.Errorf("templatePlaceholder was mangled: %q", got)
	}
}
