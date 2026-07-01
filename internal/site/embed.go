// Package site embeds a minimal Hugo project (config + layouts) used to render
// the central store as a browsable HTML site with inline HTML preserved (R-10).
package site

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:assets
var embedded embed.FS

// Materialize writes the embedded Hugo scaffold (hugo.toml + layouts) into dir,
// overwriting existing files (idempotent).
func Materialize(dir string) error {
	return fs.WalkDir(embedded, "assets", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel("assets", p)
		if err != nil {
			return err
		}
		dst := filepath.Join(dir, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		data, err := embedded.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o644)
	})
}
