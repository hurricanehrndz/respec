package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hurricanehrndz/respec/internal/config"
)

func TestRenderSubstitutesStoreAndInjects(t *testing.T) {
	cfg := config.Defaults()
	cfg.Store = "/tmp/my-store"
	cfg.Context = "SHARED-CONTEXT-MARKER"
	cfg.Rules.Research = "RESEARCH-RULE-MARKER"

	r, err := Render(cfg, TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	research, ok := r.Prompts["research.md"]
	if !ok {
		t.Fatal("research.md not rendered")
	}
	if !strings.Contains(research, "/tmp/my-store") {
		t.Error("store path not substituted into research prompt")
	}
	if !strings.Contains(research, "SHARED-CONTEXT-MARKER") {
		t.Error("context not injected into research prompt")
	}
	if !strings.Contains(research, "RESEARCH-RULE-MARKER") {
		t.Error("research rule not injected")
	}

	// No unexpanded template directives should remain anywhere.
	for name, content := range r.Prompts {
		if strings.Contains(content, "{{") {
			t.Errorf("prompt %s still has unexpanded {{ }}", name)
		}
	}
	skill, ok := r.Skills["respec/SKILL.md"]
	if !ok {
		t.Fatal("respec/SKILL.md not rendered")
	}
	if strings.Contains(skill, "{{") {
		t.Error("skill still has unexpanded {{ }}")
	}
	if !strings.Contains(skill, "/tmp/my-store") {
		t.Error("store path not substituted into skill")
	}
}

func TestRenderOmitsEmptyContextAndRules(t *testing.T) {
	cfg := config.Defaults()
	cfg.Store = "/tmp/s"
	r, err := Render(cfg, TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(r.Prompts["research.md"], "Shared context") {
		t.Error("empty context should not emit a Shared context section")
	}
}

func TestRenderLayersOverrideOverEmbedded(t *testing.T) {
	dir := t.TempDir()
	// Override just plan.md and the skill, plus add an override-only prompt.
	mustWrite(t, filepath.Join(dir, "prompts", "plan.md"), "store={{.Store}} PLAN-OVERRIDE\n")
	mustWrite(t, filepath.Join(dir, "prompts", "custom.md"), "CUSTOM-ONLY {{.Store}}\n")
	mustWrite(t, filepath.Join(dir, "skills", "respec", "SKILL.md"), "SKILL-OVERRIDE {{.Store}}\n")

	cfg := config.Defaults()
	cfg.Store = "/tmp/over"
	cfg.TemplatesDir = dir

	r, err := Render(cfg, TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// Override wins for plan.md.
	if !strings.Contains(r.Prompts["plan.md"], "PLAN-OVERRIDE") {
		t.Errorf("plan.md override not applied: %q", r.Prompts["plan.md"])
	}
	// Override-only prompt is included.
	if !strings.Contains(r.Prompts["custom.md"], "CUSTOM-ONLY") {
		t.Error("override-only custom.md not rendered")
	}
	// Embedded prompts still present (layering, not all-or-nothing).
	if _, ok := r.Prompts["research.md"]; !ok {
		t.Error("embedded research.md should still render under layering")
	}
	if _, ok := r.Prompts["implement.md"]; !ok {
		t.Error("embedded implement.md should still render under layering")
	}
	// Skill override wins.
	if !strings.Contains(r.Skills["respec/SKILL.md"], "SKILL-OVERRIDE") {
		t.Errorf("skill override not applied: %q", r.Skills["respec/SKILL.md"])
	}
}

func TestResolveReportsSource(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "prompts", "plan.md"), "x\n")
	cfg := config.Defaults()
	cfg.TemplatesDir = dir

	files, err := Resolve(cfg)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	sources := map[string]string{}
	for _, f := range files {
		sources[f.Name] = f.Source
	}
	if got := sources["prompts/plan.md"]; got != dir {
		t.Errorf("plan.md source = %q, want override dir %q", got, dir)
	}
	if got := sources["prompts/research.md"]; got != EmbeddedSource {
		t.Errorf("research.md source = %q, want %q", got, EmbeddedSource)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveErrorsOnMissingTemplatesDir(t *testing.T) {
	// A typo'd templates_dir must fail loud, not silently fall back to the
	// embedded defaults (the operator's overrides would be ignored).
	cfg := config.Defaults()
	cfg.TemplatesDir = filepath.Join(t.TempDir(), "no-such-dir")
	if _, err := Resolve(cfg); err == nil {
		t.Fatal("expected error for nonexistent templates_dir")
	}
	if _, err := Render(cfg, TargetPi, Features{}); err == nil {
		t.Fatal("Render should propagate the missing-dir error")
	}
}

func TestRenderPerTarget(t *testing.T) {
	cfg := config.Defaults()
	cfg.Store = "/tmp/s"

	pi, err := Render(cfg, TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render pi: %v", err)
	}
	claude, err := Render(cfg, TargetClaude, Features{})
	if err != nil {
		t.Fatalf("Render claude: %v", err)
	}

	// pi keeps its bash-flavored placeholders, skill path, and colon namespace.
	if !strings.Contains(pi.Prompts["research.md"], "Topic: $@") {
		t.Error("pi research.md should pass arguments via $@")
	}
	if !strings.Contains(pi.Prompts["plan.md"], "${1:-<pick latest>}") {
		t.Error("pi plan.md should default the change dir via ${1:-...}")
	}
	if !strings.Contains(pi.Prompts["research.md"], "~/.pi/agent/skills/respec/SKILL.md") {
		t.Error("pi prompts should reference the pi skill path")
	}
	if !strings.Contains(pi.Prompts["implement.md"], "/rsx:plan") {
		t.Error("pi implement.md should reference /rsx:plan")
	}
	if !strings.Contains(pi.Prompts["implement-auto.md"], "pi --print --no-session --thinking medium") {
		t.Error("pi implement-auto.md should spawn isolated medium-effort pi children")
	}
	if !strings.Contains(pi.Prompts["implement-auto.md"], "pi --print --no-session --thinking low") {
		t.Error("pi implement-auto.md should delegate commits to low-effort pi children")
	}
	if !strings.Contains(pi.Prompts["plan-auto.md"], "execution_mode: auto") ||
		!strings.Contains(pi.Prompts["plan-auto.md"], "runnable E2E/smoke command") {
		t.Error("pi plan-auto.md should persist auto mode and require end-to-end feedback")
	}
	if !strings.Contains(pi.Prompts["implement-auto.md"], "/rsx:plan-auto") {
		t.Error("pi implement-auto.md should reference /rsx:plan-auto")
	}
	// The required change-dir must use a placeholder pi actually substitutes.
	// pi supports $@/$1/${1:-default} but NOT bash's ${1:?msg}, which would
	// pass through literally and read to the agent as a missing argument.
	if !strings.Contains(pi.Prompts["implement.md"], "missing): $@") {
		t.Error("pi implement.md should pass the required change dir via $@")
	}
	if strings.Contains(pi.Prompts["implement.md"], "${1:?") {
		t.Error("pi implement.md must not use ${1:?...}; pi leaves it literal")
	}
	if !strings.Contains(pi.Prompts["implement-auto.md"], "missing): $@") {
		t.Error("pi implement-auto.md should pass the required change dir via $@")
	}

	// Claude Code has no bash-style placeholders and no colons in command names.
	for name, content := range claude.Prompts {
		for _, forbidden := range []string{"$@", "${1", "/rsx:", "~/.pi/"} {
			if strings.Contains(content, forbidden) {
				t.Errorf("claude prompt %s contains pi-ism %q", name, forbidden)
			}
		}
	}
	if !strings.Contains(claude.Prompts["research.md"], "Topic: $ARGUMENTS") {
		t.Error("claude research.md should pass arguments via $ARGUMENTS")
	}
	if !strings.Contains(claude.Prompts["research.md"], "~/.claude/skills/respec/SKILL.md") {
		t.Error("claude prompts should reference the claude skill path")
	}
	if !strings.Contains(claude.Prompts["implement.md"], "/rsx-plan") {
		t.Error("claude implement.md should reference /rsx-plan")
	}
	if !strings.Contains(claude.Prompts["implement-auto.md"], "env -u CLAUDECODE claude --print --no-session-persistence --effort medium") {
		t.Error("claude implement-auto.md should spawn isolated medium-effort Claude children")
	}
	if !strings.Contains(claude.Prompts["implement-auto.md"], "env -u CLAUDECODE claude --print --no-session-persistence --effort low") {
		t.Error("claude implement-auto.md should delegate commits to low-effort Claude children")
	}
	if !strings.Contains(claude.Prompts["implement-auto.md"], "/rsx-plan-auto") {
		t.Error("claude implement-auto.md should reference /rsx-plan-auto")
	}
	if !strings.Contains(claude.Skills["respec/SKILL.md"], "/rsx-research") {
		t.Error("claude skill should reference /rsx-research")
	}
}

func TestRenderRejectsUnknownTarget(t *testing.T) {
	if _, err := Render(config.Defaults(), Target{Name: "emacs"}, Features{}); err == nil {
		t.Fatal("expected error for unknown target")
	}
}

func TestRenderGatesProbeGuidance(t *testing.T) {
	// Prompts must never reference tooling the operator does not have: probe
	// guidance appears only when install detected the binary.
	cfg := config.Defaults()

	with, err := Render(cfg, TargetPi, Features{Probe: true})
	if err != nil {
		t.Fatalf("Render with probe: %v", err)
	}
	if !strings.Contains(with.Prompts["research.md"], "probe search") {
		t.Error("probe guidance missing from research.md when probe is available")
	}

	without, err := Render(cfg, TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render without probe: %v", err)
	}
	for name, content := range without.Prompts {
		if strings.Contains(content, "probe") {
			t.Errorf("prompt %s references probe although it is not installed", name)
		}
	}
}
