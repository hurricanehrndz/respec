package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runInstall(t *testing.T, target string) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", "") // an inherited CODEX_HOME would install outside the temp HOME
	rootCmd.SetArgs([]string{"install", "--target", target})
	rootCmd.SetOut(os.NewFile(0, os.DevNull))
	rootCmd.SetErr(os.NewFile(0, os.DevNull))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("install --target %s: %v", target, err)
	}
	return home
}

func TestInstallPiWritesUserScopeFiles(t *testing.T) {
	home := runInstall(t, "pi")

	prompts := []string{"rsx:research.md", "rsx:plan.md", "rsx:implement.md"}
	for _, name := range prompts {
		p := filepath.Join(home, ".pi", "agent", "prompts", name)
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("expected installed prompt %s: %v", p, err)
		}
		store := filepath.Join(home, "respec-store")
		if !strings.Contains(string(data), store) {
			t.Errorf("%s missing resolved store path %q", name, store)
		}
		if strings.Contains(string(data), "{{") {
			t.Errorf("%s has unexpanded template directives", name)
		}
	}

	skill := filepath.Join(home, ".pi", "agent", "skills", "respec", "SKILL.md")
	if _, err := os.ReadFile(skill); err != nil {
		t.Fatalf("expected installed skill %s: %v", skill, err)
	}
}

func TestInstallClaudeWritesUserScopeFiles(t *testing.T) {
	home := runInstall(t, "claude")

	prompts := []string{"rsx-research.md", "rsx-plan.md", "rsx-implement.md"}
	for _, name := range prompts {
		p := filepath.Join(home, ".claude", "commands", name)
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("expected installed command %s: %v", p, err)
		}
		if strings.Contains(string(data), "{{") {
			t.Errorf("%s has unexpanded template directives", name)
		}
		// No pi-isms may leak into the Claude Code render.
		for _, forbidden := range []string{"$@", "${1", "~/.pi/"} {
			if strings.Contains(string(data), forbidden) {
				t.Errorf("%s contains pi-ism %q", name, forbidden)
			}
		}
	}

	skill := filepath.Join(home, ".claude", "skills", "respec", "SKILL.md")
	if _, err := os.ReadFile(skill); err != nil {
		t.Fatalf("expected installed skill %s: %v", skill, err)
	}
}

func TestInstallCodexWritesUserScopeFiles(t *testing.T) {
	home := runInstall(t, "codex")
	codexHome := filepath.Join(home, ".codex")

	prompts := []string{"rsx-research.md", "rsx-plan.md", "rsx-implement.md"}
	for _, name := range prompts {
		p := filepath.Join(codexHome, "prompts", name)
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("expected installed prompt %s: %v", p, err)
		}
		if strings.Contains(string(data), "{{") {
			t.Errorf("%s has unexpanded template directives", name)
		}
		for _, forbidden := range []string{"$@", "${1:", "/rsx:", "~/.pi/", "~/.claude/"} {
			if strings.Contains(string(data), forbidden) {
				t.Errorf("%s contains foreign-agent syntax %q", name, forbidden)
			}
		}
	}

	// Codex reads the skill's display metadata from agents/openai.yaml, so the
	// whole bundle must land, not just SKILL.md.
	for _, rel := range []string{"SKILL.md", filepath.Join("agents", "openai.yaml")} {
		p := filepath.Join(codexHome, "skills", "respec", rel)
		if _, err := os.ReadFile(p); err != nil {
			t.Fatalf("expected installed skill file %s: %v", p, err)
		}
	}
}

func TestInstallCodexHonorsCodexHome(t *testing.T) {
	home := t.TempDir()
	codexHome := filepath.Join(home, "custom-codex")
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", codexHome)
	rootCmd.SetArgs([]string{"install", "--target", "codex"})
	rootCmd.SetOut(os.NewFile(0, os.DevNull))
	rootCmd.SetErr(os.NewFile(0, os.DevNull))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("install --target codex: %v", err)
	}
	if _, err := os.ReadFile(filepath.Join(codexHome, "prompts", "rsx-plan.md")); err != nil {
		t.Fatalf("expected Codex prompt under CODEX_HOME: %v", err)
	}
	if _, err := os.ReadFile(filepath.Join(codexHome, "skills", "respec", "SKILL.md")); err != nil {
		t.Fatalf("expected Codex skill under CODEX_HOME: %v", err)
	}
}

func TestInstallRejectsUnknownTarget(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	rootCmd.SetArgs([]string{"install", "--target", "emacs"})
	rootCmd.SetOut(os.NewFile(0, os.DevNull))
	rootCmd.SetErr(os.NewFile(0, os.DevNull))
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown --target")
	}
	if !strings.Contains(err.Error(), "emacs") {
		t.Errorf("error should name the bad target, got: %v", err)
	}
}

// A skill file that is no longer rendered must go: the agent reads whatever is
// in the bundle, so a retired reference file would keep being loaded next to
// its replacement. Other operators' skills in the same directory are not ours.
func TestInstallPrunesRetiredSkillFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", "")

	skillsDir := filepath.Join(home, ".pi", "agent", "skills")
	retired := filepath.Join(skillsDir, "respec", "references", "old.md")
	foreign := filepath.Join(skillsDir, "someone-elses", "SKILL.md")
	for _, p := range []string{retired, foreign} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	rootCmd.SetArgs([]string{"install", "--target", "pi"})
	rootCmd.SetOut(os.NewFile(0, os.DevNull))
	rootCmd.SetErr(os.NewFile(0, os.DevNull))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("install: %v", err)
	}

	if _, err := os.Stat(retired); !os.IsNotExist(err) {
		t.Errorf("retired skill file %s should have been removed", retired)
	}
	if _, err := os.Stat(filepath.Dir(retired)); !os.IsNotExist(err) {
		t.Errorf("emptied skill dir %s should have been removed", filepath.Dir(retired))
	}
	if _, err := os.Stat(foreign); err != nil {
		t.Errorf("another skill in the shared dir must be left alone: %v", err)
	}
	// The bundle respec does render must still be there afterwards.
	for _, rel := range []string{"SKILL.md", filepath.Join("agents", "openai.yaml")} {
		if _, err := os.Stat(filepath.Join(skillsDir, "respec", rel)); err != nil {
			t.Errorf("pruning removed a rendered skill file %s: %v", rel, err)
		}
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	home := runInstall(t, "pi")
	research := filepath.Join(home, ".pi", "agent", "prompts", "rsx:research.md")
	first, err := os.ReadFile(research)
	if err != nil {
		t.Fatal(err)
	}

	// Re-run install in the same HOME.
	rootCmd.SetArgs([]string{"install", "--target", "pi"})
	rootCmd.SetOut(os.NewFile(0, os.DevNull))
	rootCmd.SetErr(os.NewFile(0, os.DevNull))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("second install: %v", err)
	}
	second, err := os.ReadFile(research)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Error("install not idempotent: second run produced different bytes")
	}
}

// Retiring a prompt must remove its installed file, not just stop rewriting it.
// A leftover rsx: command keeps working and points the operator at a workflow
// that no longer exists.
func TestInstallPrunesRetiredPrompts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	promptsDir := filepath.Join(home, ".pi", "agent", "prompts")
	if err := os.MkdirAll(promptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	retired := filepath.Join(promptsDir, "rsx:plan-auto.md")
	mine := filepath.Join(promptsDir, "my-own.md")
	for _, p := range []string{retired, mine} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	rootCmd.SetArgs([]string{"install", "--target", "pi"})
	rootCmd.SetOut(os.NewFile(0, os.DevNull))
	rootCmd.SetErr(os.NewFile(0, os.DevNull))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("install: %v", err)
	}

	if _, err := os.Stat(retired); !os.IsNotExist(err) {
		t.Errorf("retired prompt %s should have been removed", retired)
	}
	// Only the rsx namespace is respec's to prune.
	if _, err := os.Stat(mine); err != nil {
		t.Errorf("non-respec prompt %s must be left alone: %v", mine, err)
	}
}
