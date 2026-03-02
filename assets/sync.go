//go:build ignore

// sync.go stages .picoclaw/ → assets/picoclaw/ excluding node_modules/, bin/, and .gitignore.
// Run via: go generate ./assets/
package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var skipDirs = map[string]bool{
	"node_modules": true,
	"bin":          true,
}

var skipFiles = map[string]bool{
	".gitignore": true,
}

func main() {
	src := filepath.Join("..", ".picoclaw")
	dst := filepath.Join("picoclaw")

	// Remove stale staged tree so deleted source files don't linger.
	if err := os.RemoveAll(dst); err != nil {
		fatalf("remove %s: %v", dst, err)
	}

	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		if rel == "." {
			return nil
		}

		// Check top-level component for skipped dirs/files.
		top := strings.SplitN(rel, string(filepath.Separator), 2)[0]
		if d.IsDir() && skipDirs[top] {
			return filepath.SkipDir
		}
		if !d.IsDir() && skipFiles[d.Name()] {
			return nil
		}

		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
	if err != nil {
		fatalf("walk: %v", err)
	}
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func fatalf(format string, args ...any) {
	os.Stderr.WriteString("sync: ")
	os.Stderr.WriteString(fmt.Sprintf(format, args...))
	os.Stderr.WriteString("\n")
	os.Exit(1)
}
