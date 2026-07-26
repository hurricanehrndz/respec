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
