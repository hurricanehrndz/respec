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

func TestRenderUsesOverrideDir(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "prompts", "custom.md"), "store={{.Store}} OVERRIDE-MARKER\n")
	mustWrite(t, filepath.Join(dir, "skills", "respec", "SKILL.md"), "skill {{.Store}}\n")

	cfg := config.Defaults()
	cfg.Store = "/tmp/over"
	cfg.TemplatesDir = dir

	r, err := Render(cfg)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	c, ok := r.Prompts["custom.md"]
	if !ok {
		t.Fatal("override prompt custom.md not rendered")
	}
	if !strings.Contains(c, "OVERRIDE-MARKER") || !strings.Contains(c, "/tmp/over") {
		t.Errorf("override dir not used / not rendered: %q", c)
	}
	if _, ok := r.Prompts["research.md"]; ok {
		t.Error("embedded defaults should not be used when templates_dir is set")
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
