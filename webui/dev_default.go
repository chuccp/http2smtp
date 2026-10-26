//go:build !release

package webui

// devDefault is true for ordinary local builds: `go build` / `go run` produce a binary
// that reads templates and static assets from the working tree, so a browser refresh
// shows edits. Release binaries are built with `-tags release` and serve the embedded
// copies instead.
const devDefault = true
