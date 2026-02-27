// Package assets embeds the staged .opencode configuration tree.
// Run `go generate ./assets/` (or `make generate`) to refresh the staged files
// from the source .opencode directory before building.
package assets

import (
	"embed"
	"io/fs"
)

//go:embed opencode
var embeddedFS embed.FS

// FS returns a sub-filesystem rooted at the embedded "opencode" directory,
// mirroring the layout of .opencode/ at the repo root.
func FS() (fs.FS, error) {
	return fs.Sub(embeddedFS, "opencode")
}
