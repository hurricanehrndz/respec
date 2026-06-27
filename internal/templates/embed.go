// Package templates embeds the default /rsx:* prompt-templates and the backing
// respec skill, and renders them with the configured store path and injected
// context/rules (R-3, R-12, R-13).
package templates

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/chernand/respec/internal/config"
)

//go:embed all:assets
var embedded embed.FS

// RenderData is the template context substituted into each asset.
type RenderData struct {
	Store   string
	Context string
	Rules   config.Rules
}

// Rendered holds the rendered assets keyed by their install-relative path.
type Rendered struct {
	Prompts map[string]string // base filename -> content (e.g. "rsx:research.md")
	Skills  map[string]string // path under skills/ -> content (e.g. "respec/SKILL.md")
}

// sourceFS returns the asset filesystem: the override dir when templates_dir is
// set (R-12), otherwise the embedded defaults.
func sourceFS(cfg config.Config) (fs.FS, error) {
	if cfg.TemplatesDir != "" {
		return os.DirFS(config.Expand(cfg.TemplatesDir)), nil
	}
	return fs.Sub(embedded, "assets")
}

// Render reads every prompt and skill asset and renders it with config values.
func Render(cfg config.Config) (Rendered, error) {
	src, err := sourceFS(cfg)
	if err != nil {
		return Rendered{}, err
	}
	data := RenderData{
		Store:   cfg.StorePath(),
		Context: cfg.Context,
		Rules:   cfg.Rules,
	}
	out := Rendered{
		Prompts: map[string]string{},
		Skills:  map[string]string{},
	}

	prompts, err := fs.Glob(src, "prompts/*.md")
	if err != nil {
		return Rendered{}, err
	}
	for _, p := range prompts {
		rendered, err := renderFile(src, p, data)
		if err != nil {
			return Rendered{}, err
		}
		out.Prompts[path.Base(p)] = rendered
	}

	err = fs.WalkDir(src, "skills", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path.Base(p) != "SKILL.md" {
			return nil
		}
		rendered, rerr := renderFile(src, p, data)
		if rerr != nil {
			return rerr
		}
		out.Skills[strings.TrimPrefix(p, "skills/")] = rendered
		return nil
	})
	if err != nil {
		return Rendered{}, err
	}

	if len(out.Prompts) == 0 {
		return Rendered{}, fmt.Errorf("no prompt templates found in source")
	}
	return out, nil
}

func renderFile(src fs.FS, p string, data RenderData) (string, error) {
	raw, err := fs.ReadFile(src, p)
	if err != nil {
		return "", err
	}
	tmpl, err := template.New(p).Option("missingkey=error").Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parsing template %s: %w", p, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("rendering template %s: %w", p, err)
	}
	return buf.String(), nil
}
