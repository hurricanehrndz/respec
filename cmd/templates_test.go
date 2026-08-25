package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplatesListEmbedded(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home) // no config -> defaults, templates_dir empty
	out, err := runRespecErr(t, "templates", "list")
	if err != nil {
		t.Fatalf("templates list: %v\n%s", err, out)
	}
	for _, want := range []string{"skills/respec/SKILL.md", "skills/respec/agents/openai.yaml", "skills/rsx-research/SKILL.md", "skills/rsx-plan/SKILL.md", "skills/rsx-implement/SKILL.md", "[embedded]"} {
		if !strings.Contains(out, want) {
			t.Errorf("list output missing %q:\n%s", want, out)
		}
	}
}

func TestTemplatesEjectAndListShowsOverride(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	tdir := filepath.Join(home, "templates")
	runRespec(t, "config", "set", "templates_dir", tdir)

	// Eject a single skill.
	if _, err := runRespecErr(t, "templates", "eject", "rsx-research/SKILL.md", "--force=false"); err != nil {
		t.Fatalf("eject: %v", err)
	}
	ejected := filepath.Join(tdir, "skills", "rsx-research", "SKILL.md")
	if _, err := os.Stat(ejected); err != nil {
		t.Fatalf("ejected file missing: %v", err)
	}

	// list --json should now report the research skill sourced from the override dir.
	out := runRespec(t, "templates", "list", "--json")
	var infos []templateInfo
	if err := json.Unmarshal([]byte(out), &infos); err != nil {
		t.Fatalf("decode list json: %v\n%s", err, out)
	}
	found := false
	for _, i := range infos {
		if i.Name == "skills/rsx-research/SKILL.md" {
			found = true
			if i.Source != tdir {
				t.Errorf("rsx-research source = %q, want %q", i.Source, tdir)
			}
		}
	}
	if !found {
		t.Error("rsx-research not in list output")
	}

	// Re-eject without force must refuse; with force must succeed.
	if _, err := runRespecErr(t, "templates", "eject", "rsx-research/SKILL.md", "--force=false"); err == nil {
		t.Error("expected eject to refuse overwrite without --force")
	}
	if _, err := runRespecErr(t, "templates", "eject", "rsx-research/SKILL.md", "--force"); err != nil {
		t.Errorf("eject --force failed: %v", err)
	}
}

func TestTemplatesEjectRequiresTemplatesDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home) // templates_dir unset
	if _, err := runRespecErr(t, "templates", "eject", "--force=false"); err == nil {
		t.Fatal("expected error when templates_dir is unset")
	}
}
