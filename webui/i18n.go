package webui

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// The four locale JSON files are the UI's translation bundles and the source of truth
// for every string the templates render: they are embedded, so a release is a single
// binary, and webui/i18n_test.go keeps the four in key-for-key sync.
//
//go:embed i18n/zh-cn.json
var zhCnJSON []byte

//go:embed i18n/zh-tw.json
var zhTwJSON []byte

//go:embed i18n/en.json
var enJSON []byte

//go:embed i18n/ja.json
var jaJSON []byte

// DefaultLanguage is what a browser expressing no preference gets, and the fallback
// when nothing in Accept-Language matches a shipped language.
const DefaultLanguage = "zh-cn"

// Languages lists the supported language codes, in menu order.
var Languages = []string{"zh-cn", "zh-tw", "en", "ja"}

// Translator resolves dot-keys ("common.save") into the active language. It is
// immutable after load, so one instance per language is shared by all requests.
type Translator struct {
	m map[string]string
}

// placeholder matches the locale files' {count}-style parameter slots. Go-template
// text like {{.key}} inside a value never matches (the dot is not a word character),
// so template-help strings pass through untouched.
var placeholder = regexp.MustCompile(`\{[A-Za-z0-9_]+\}`)

// T returns the translation for key. When args are given, they replace the {name}
// placeholders in order of appearance — the locale values use brace placeholders
// rather than printf verbs. A key no language carries is returned as-is, visible in
// the UI rather than silently blank.
func (tr *Translator) T(key string, args ...any) string {
	v, ok := tr.m[key]
	if !ok {
		v = key
	}
	if len(args) == 0 {
		return v
	}
	i := 0
	return placeholder.ReplaceAllStringFunc(v, func(m string) string {
		if i >= len(args) {
			return m
		}
		s := fmt.Sprint(args[i])
		i++
		return s
	})
}

var translators = map[string]*Translator{}

func parseLocale(data []byte) (map[string]string, error) {
	m := make(map[string]string)
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// loadI18n parses the embedded locale files once at startup. Every language is layered
// over English, so a key added to en but not yet translated still shows English
// instead of the raw key.
func loadI18n() error {
	base, err := parseLocale(enJSON)
	if err != nil {
		return fmt.Errorf("parsing en locale: %w", err)
	}
	bundle := map[string][]byte{
		"en":    enJSON,
		"zh-cn": zhCnJSON,
		"zh-tw": zhTwJSON,
		"ja":    jaJSON,
	}
	for _, lang := range Languages {
		data, ok := bundle[lang]
		if !ok {
			return fmt.Errorf("no locale data for %q", lang)
		}
		own, err := parseLocale(data)
		if err != nil {
			return fmt.Errorf("parsing %s locale: %w", lang, err)
		}
		if lang != "en" {
			merged := make(map[string]string, len(base)+len(own))
			for k, v := range base {
				merged[k] = v
			}
			for k, v := range own {
				merged[k] = v
			}
			own = merged
		}
		translators[lang] = &Translator{m: own}
	}
	return nil
}

// translatorFor returns the loaded translator for a language code, falling back to
// the default language for anything unknown.
func translatorFor(lang string) *Translator {
	if tr, ok := translators[lang]; ok {
		return tr
	}
	return translators[DefaultLanguage]
}

// isSupportedLang reports whether the code names a language we ship.
func isSupportedLang(lang string) bool {
	_, ok := translators[lang]
	return ok
}

// pickLang chooses the request language: a valid cookie wins, otherwise the
// Accept-Language header is mapped onto what we ship, otherwise the default.
// Split out from the request handling so the mapping is testable on its own.
func pickLang(cookieLang, acceptLanguage string) string {
	if isSupportedLang(cookieLang) {
		return cookieLang
	}
	for _, tag := range parseAcceptLanguage(acceptLanguage) {
		switch {
		case tag == "zh-tw" || tag == "zh-hant" || strings.HasPrefix(tag, "zh-hk") || strings.HasPrefix(tag, "zh-mo"):
			return "zh-tw"
		case strings.HasPrefix(tag, "zh"):
			return "zh-cn"
		case strings.HasPrefix(tag, "ja"):
			return "ja"
		case strings.HasPrefix(tag, "en"):
			return "en"
		}
	}
	return DefaultLanguage
}

// parseAcceptLanguage splits an Accept-Language header into tags ordered by q-value.
func parseAcceptLanguage(header string) []string {
	type entry struct {
		tag string
		q   float64
	}
	var entries []entry
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		tag := part
		q := 1.0
		if i := strings.Index(part, ";"); i >= 0 {
			tag = strings.TrimSpace(part[:i])
			for _, param := range strings.Split(part[i+1:], ";") {
				param = strings.TrimSpace(param)
				if len(param) >= 2 && strings.EqualFold(param[:2], "q=") {
					fmt.Sscanf(param[2:], "%f", &q)
				}
			}
		}
		// q=0 marks a language the client explicitly does not accept.
		if tag != "" && tag != "*" && q > 0 {
			entries = append(entries, entry{strings.ToLower(tag), q})
		}
	}
	// Stable insertion sort by q descending — headers are tiny.
	for i := 1; i < len(entries); i++ {
		for j := i; j > 0 && entries[j].q > entries[j-1].q; j-- {
			entries[j], entries[j-1] = entries[j-1], entries[j]
		}
	}
	tags := make([]string, 0, len(entries))
	for _, e := range entries {
		tags = append(tags, e.tag)
	}
	return tags
}
