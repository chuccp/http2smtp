package webui

import "embed"

// The whole management UI ships inside the binary: templates, static assets and the
// locale JSON (embedded in i18n.go). A release needs no web/ directory.
//
//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS
