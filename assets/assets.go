// Package assets embeds the staged .picoclaw configuration tree.
// Run `go generate ./assets/` (or `make generate`) to refresh the staged files
// from the source .picoclaw directory before building.
package assets

import (
	"embed"
	"io/fs"
)

//go:embed picoclaw
var embeddedFS embed.FS

// FS returns a sub-filesystem rooted at the embedded "picoclaw" directory,
// mirroring the layout of .picoclaw/ at the repo root.
func FS() (fs.FS, error) {
	return fs.Sub(embeddedFS, "picoclaw")
}
