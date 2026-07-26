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

// Target is an agent flavor the templates render for. The prompt bodies are
// agent-agnostic; everything agent-specific — argument placeholders, installed
// command names, where the skill lives — resolves through Target's methods.
type Target struct {
	Name string // "pi" | "claude"
}

var (
	TargetPi     = Target{Name: "pi"}
	TargetClaude = Target{Name: "claude"}

	// Targets maps a --target flag value to its Target.
	Targets = map[string]Target{"pi": TargetPi, "claude": TargetClaude}
)

// IsClaude reports whether templates are being rendered for Claude Code.
func (t Target) IsClaude() bool { return t.Name == TargetClaude.Name }

// AllArgs is the placeholder expanding to the invocation's whole argument string.
func (t Target) AllArgs() string {
	if t.IsClaude() {
		return "$ARGUMENTS"
	}
	return "$@"
}

// Arg1Or is the first-argument placeholder with a fallback when absent. Claude
// Code has no default syntax (bash-style ${1:-...} is unsupported), so there
// the surrounding prose must carry the fallback and the raw placeholder is
// emitted; pi embeds the default.
func (t Target) Arg1Or(def string) string {
	if t.IsClaude() {
		return "$ARGUMENTS"
	}
	return "${1:-" + def + "}"
}

// Arg1Req is the first-argument placeholder for a required argument. Neither
// target can enforce required-ness in the placeholder itself: pi's template
// engine substitutes $@, $1, ${1:-default}, and ${@:N} but NOT bash's ${1:?msg}
// error form — that pattern is left literal, so the agent sees the raw
// placeholder and treats the arg as missing. Claude Code has no such syntax at
// all. Both therefore emit a plain all-args placeholder and rely on the
// surrounding prose ("required — stop and ask if missing") to enforce it. msg
// is retained to document the argument at the call site.
func (t Target) Arg1Req(msg string) string {
	if t.IsClaude() {
		return "$ARGUMENTS"
	}
	return "$@"
}

// Cmd returns the installed command name for a prompt. pi namespaces with a
// colon (/rsx:plan); Claude Code user-scope command names cannot contain one,
// so the namespace flattens to a hyphen (/rsx-plan).
func (t Target) Cmd(name string) string {
	if t.IsClaude() {
		return "/rsx-" + name
	}
	return "/rsx:" + name
}

// SkillPath is where the installed respec skill lives at user scope.
func (t Target) SkillPath() string {
	if t.IsClaude() {
		return "~/.claude/skills/respec/SKILL.md"
	}
	return "~/.pi/agent/skills/respec/SKILL.md"
}

// SkillCmd is how the operator invokes the respec skill interactively.
func (t Target) SkillCmd() string {
	if t.IsClaude() {
		return "/respec"
	}
	return "/skill:respec"
}

// Features records optional operator tooling detected at install time; the
// templates gate matching guidance on these so prompts never reference tools
// that are not there.
type Features struct {
	Probe bool // the probe code-search binary is on PATH
}

// RenderData is the template context substituted into each asset.
type RenderData struct {
	Store    string
	Context  string
	Rules    config.Rules
	Agents   config.Agents
	HasProbe bool
	Target
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

// Render reads every prompt and skill asset and renders it with config values
// for the given target agent and detected features.
func Render(cfg config.Config, target Target, feats Features) (Rendered, error) {
	if _, ok := Targets[target.Name]; !ok {
		return Rendered{}, fmt.Errorf("unknown render target %q", target.Name)
	}
	files, err := Resolve(cfg)
	if err != nil {
		return Rendered{}, err
	}
	data := RenderData{
		Store:    cfg.StorePath(),
		Context:  cfg.Context,
		Rules:    cfg.Rules,
		Agents:   cfg.Agents,
		HasProbe: feats.Probe,
		Target:   target,
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
