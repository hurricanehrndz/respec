// Package templates embeds the respec skills — the shared workflow plus three
// phase entry points — and renders them with the configured store path and
// injected context/rules (R-3, R-12, R-13).
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
	"strings"
	"text/template"

	"github.com/hurricanehrndz/respec/internal/config"
)

//go:embed all:assets
var embedded embed.FS

// EmbeddedSource is the reported source label for built-in templates.
const EmbeddedSource = "embedded"

// Target is an agent flavor the templates render for. The skill bodies are
// agent-agnostic; everything agent-specific — installed command names, where
// the skill lives — resolves through Target's methods.
type Target struct {
	Name string // "pi" | "prime-agent" | "claude" | "codex"
}

var (
	TargetPi         = Target{Name: "pi"}
	TargetPrimeAgent = Target{Name: "prime-agent"}
	TargetClaude     = Target{Name: "claude"}
	TargetCodex      = Target{Name: "codex"}

	// Targets maps a --target flag value to its Target.
	Targets = map[string]Target{"pi": TargetPi, "prime-agent": TargetPrimeAgent, "claude": TargetClaude, "codex": TargetCodex}
)

// IsClaude reports whether templates are being rendered for Claude Code.
func (t Target) IsClaude() bool { return t.Name == TargetClaude.Name }

// IsCodex reports whether templates are being rendered for OpenAI Codex.
func (t Target) IsCodex() bool { return t.Name == TargetCodex.Name }

// IsPrimeAgent reports whether templates are being rendered for Prime Agent.
func (t Target) IsPrimeAgent() bool { return t.Name == TargetPrimeAgent.Name }

// Cmd returns how to invoke a phase skill. pi and Prime Agent register skills
// as /skill:name; Claude Code resolves them as /name; Codex mentions them as
// $name.
func (t Target) Cmd(name string) string {
	if t.IsCodex() {
		return "$rsx-" + name
	}
	if t.IsClaude() {
		return "/rsx-" + name
	}
	return "/skill:rsx-" + name
}

// SkillPath is where the installed respec skill lives at user scope.
func (t Target) SkillPath() string {
	if t.IsCodex() {
		return "${CODEX_HOME:-$HOME/.codex}/skills/respec/SKILL.md"
	}
	if t.IsClaude() {
		return "~/.claude/skills/respec/SKILL.md"
	}
	if t.IsPrimeAgent() {
		return "~/.prime/agent/skills/respec/SKILL.md"
	}
	return "~/.pi/agent/skills/respec/SKILL.md"
}

// RenderData is the install-time context. Worker preferences are read during
// planning and recorded in the effort's plan, not baked into skills.
type RenderData struct {
	Store   string
	Context string
	Rules   config.Rules
	Target
}

// Rendered holds the rendered skills keyed by their install-relative path.
type Rendered struct {
	Skills map[string]string // path under skills/ -> content (e.g. "respec/SKILL.md")
}

// TemplateFile describes one resolved template and where it comes from.
type TemplateFile struct {
	Name    string // logical path, e.g. "skills/respec/SKILL.md"
	Key     string // install key: path under skills/ (e.g. "respec/SKILL.md")
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
		if _, err := fs.Stat(fsys, "skills"); err != nil {
			return nil // no skills tree in this layer
		}
		return fs.WalkDir(fsys, "skills", func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// Every file in the bundle ships, not just SKILL.md: Codex reads
			// agents/openai.yaml for the skill's display metadata, and an extra
			// file is inert for the targets that do not read it.
			if d.IsDir() {
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

// Render reads every skill asset and renders it with config values for the
// given target agent.
func Render(cfg config.Config, target Target) (Rendered, error) {
	if _, ok := Targets[target.Name]; !ok {
		return Rendered{}, fmt.Errorf("unknown render target %q", target.Name)
	}
	files, err := Resolve(cfg)
	if err != nil {
		return Rendered{}, err
	}
	data := RenderData{
		Store:   cfg.StorePath(),
		Context: cfg.Context,
		Rules:   cfg.Rules,
		Target:  target,
	}
	out := Rendered{Skills: map[string]string{}}
	for _, f := range files {
		raw, err := f.ReadRaw()
		if err != nil {
			return Rendered{}, err
		}
		rendered, err := renderBytes(f.Name, raw, data)
		if err != nil {
			return Rendered{}, err
		}
		out.Skills[f.Key] = rendered
	}
	if len(out.Skills) == 0 {
		return Rendered{}, fmt.Errorf("no skills found in source")
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
