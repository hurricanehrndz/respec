// Package templates embeds the default /rsx:* prompt-templates and the backing
// respec skill, and renders them with the configured store path and injected
// context/rules (R-3, R-12, R-13).
//
// Resolution is layered: when templates_dir is set, each template file is taken
// from the override dir if present there, otherwise from the embedded defaults.
// This lets an operator override a single prompt without re-supplying the rest.
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

	"github.com/hurricanehrndz/respec/internal/config"
)

//go:embed all:assets
var embedded embed.FS

// EmbeddedSource is the reported source label for built-in templates.
const EmbeddedSource = "embedded"

// RenderData is the template context substituted into each asset.
type RenderData struct {
	Store   string
	Context string
	Rules   config.Rules
}

// Rendered holds the rendered assets keyed by their install-relative path.
type Rendered struct {
	Prompts map[string]string // base filename -> content (e.g. "research.md")
	Skills  map[string]string // path under skills/ -> content (e.g. "respec/SKILL.md")
}

// TemplateFile describes one resolved template and where it comes from.
type TemplateFile struct {
	Name    string // logical path, e.g. "prompts/research.md" or "skills/respec/SKILL.md"
	Key     string // install key: base filename (prompts) or path under skills/ (skills)
	IsSkill bool
	Source  string // EmbeddedSource, or the override dir when overridden
	fsys    fs.FS
	path    string
}

// embeddedFS returns the embedded assets subtree.
func embeddedFS() (fs.FS, error) {
	return fs.Sub(embedded, "assets")
}

// Resolve returns every template file, layering the override dir (when set) over
// the embedded defaults. Override entries win on name collisions; override-only
// templates are appended after the embedded ones.
func Resolve(cfg config.Config) ([]TemplateFile, error) {
	emb, err := embeddedFS()
	if err != nil {
		return nil, err
	}

	index := map[string]int{} // logical name -> position in files
	var files []TemplateFile
	add := func(f TemplateFile) {
		if i, ok := index[f.Name]; ok {
			files[i] = f // override wins, keeps original position
			return
		}
		index[f.Name] = len(files)
		files = append(files, f)
	}

	collect := func(fsys fs.FS, source string) error {
		prompts, err := fs.Glob(fsys, "prompts/*.md")
		if err != nil {
			return err
		}
		for _, p := range prompts {
			add(TemplateFile{Name: p, Key: path.Base(p), Source: source, fsys: fsys, path: p})
		}
		if _, err := fs.Stat(fsys, "skills"); err != nil {
			return nil // no skills tree in this layer
		}
		return fs.WalkDir(fsys, "skills", func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || path.Base(p) != "SKILL.md" {
				return nil
			}
			add(TemplateFile{Name: p, Key: strings.TrimPrefix(p, "skills/"), IsSkill: true, Source: source, fsys: fsys, path: p})
			return nil
		})
	}

	if err := collect(emb, EmbeddedSource); err != nil {
		return nil, err
	}
	if cfg.TemplatesDir != "" {
		dir := config.Expand(cfg.TemplatesDir)
		// A missing override dir is a misconfiguration, not "no overrides" —
		// fail loud instead of silently falling back to embedded defaults.
		if info, err := os.Stat(dir); err != nil {
			return nil, fmt.Errorf("templates_dir %q: %w", dir, err)
		} else if !info.IsDir() {
			return nil, fmt.Errorf("templates_dir %q is not a directory", dir)
		}
		if err := collect(os.DirFS(dir), dir); err != nil {
			return nil, err
		}
	}
	return files, nil
}

// Embedded returns the built-in template files only (ignoring any override dir).
// Used by `templates eject` to copy defaults out.
func Embedded() ([]TemplateFile, error) {
	return Resolve(config.Config{})
}

// ReadRaw returns the unrendered bytes of a template file.
func (f TemplateFile) ReadRaw() ([]byte, error) {
	return fs.ReadFile(f.fsys, f.path)
}

// Render reads every prompt and skill asset and renders it with config values.
func Render(cfg config.Config) (Rendered, error) {
	files, err := Resolve(cfg)
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
	for _, f := range files {
		raw, err := f.ReadRaw()
		if err != nil {
			return Rendered{}, err
		}
		rendered, err := renderBytes(f.Name, raw, data)
		if err != nil {
			return Rendered{}, err
		}
		if f.IsSkill {
			out.Skills[f.Key] = rendered
		} else {
			out.Prompts[f.Key] = rendered
		}
	}
	if len(out.Prompts) == 0 {
		return Rendered{}, fmt.Errorf("no prompt templates found in source")
	}
	return out, nil
}

func renderBytes(name string, raw []byte, data RenderData) (string, error) {
	tmpl, err := template.New(name).Option("missingkey=error").Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parsing template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("rendering template %s: %w", name, err)
	}
	return buf.String(), nil
}
