//go:build release

package webui

// devDefault is false in release builds (`go build -tags release`): the binary serves
// the embedded templates and assets and runs from anywhere, with no working-tree
// dependency.
const devDefault = false
