package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runInstall(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	rootCmd.SetArgs([]string{"install"})
	rootCmd.SetOut(os.NewFile(0, os.DevNull))
	rootCmd.SetErr(os.NewFile(0, os.DevNull))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("install: %v", err)
	}
	return home
}

func TestInstallWritesUserScopeFiles(t *testing.T) {
	home := runInstall(t)

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

func TestInstallIsIdempotent(t *testing.T) {
	home := runInstall(t)
	research := filepath.Join(home, ".pi", "agent", "prompts", "rsx:research.md")
	first, err := os.ReadFile(research)
	if err != nil {
		t.Fatal(err)
	}

	// Re-run install in the same HOME.
	rootCmd.SetArgs([]string{"install"})
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
