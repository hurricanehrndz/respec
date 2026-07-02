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

	r, err := Render(cfg)
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
	r, err := Render(cfg)
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

	r, err := Render(cfg)
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
