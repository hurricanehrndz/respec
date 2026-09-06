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

// skillFiles is the bundle every target installs: the shared respec skill plus
// the three phase entry skills.
var skillFiles = []string{
	filepath.Join("respec", "SKILL.md"),
	filepath.Join("rsx-research", "SKILL.md"),
	filepath.Join("rsx-plan", "SKILL.md"),
	filepath.Join("rsx-implement", "SKILL.md"),
}

func TestInstallPiWritesUserScopeFiles(t *testing.T) {
	home := runInstall(t, "pi")
	skillsDir := filepath.Join(home, ".pi", "agent", "skills")

	for _, rel := range skillFiles {
		p := filepath.Join(skillsDir, rel)
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("expected installed skill %s: %v", p, err)
		}
		if strings.Contains(string(data), "{{") {
			t.Errorf("%s has unexpanded template directives", rel)
		}
	}
}

func TestInstallClaudeWritesUserScopeFiles(t *testing.T) {
	home := runInstall(t, "claude")
	skillsDir := filepath.Join(home, ".claude", "skills")

	for _, rel := range skillFiles {
		p := filepath.Join(skillsDir, rel)
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("expected installed skill %s: %v", p, err)
		}
		if strings.Contains(string(data), "{{") {
			t.Errorf("%s has unexpanded template directives", rel)
		}
		// No pi-isms may leak into the Claude Code render.
		for _, forbidden := range []string{"$@", "${1", "/rsx:", "~/.pi/", "/skill:"} {
			if strings.Contains(string(data), forbidden) {
				t.Errorf("%s contains foreign-agent syntax %q", rel, forbidden)
			}
		}
	}
}

func TestInstallPrimeAgentWritesUserScopeFiles(t *testing.T) {
	home := runInstall(t, "prime-agent")
	skillsDir := filepath.Join(home, ".prime", "agent", "skills")

	for _, rel := range skillFiles {
		p := filepath.Join(skillsDir, rel)
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("expected installed skill %s: %v", p, err)
		}
		if strings.Contains(string(data), "{{") {
			t.Errorf("%s has unexpanded template directives", rel)
		}
		for _, forbidden := range []string{"${1", "/rsx:", "~/.pi/", "~/.claude/", "${CODEX_HOME"} {
			if strings.Contains(string(data), forbidden) {
				t.Errorf("%s contains foreign-agent syntax %q", rel, forbidden)
			}
		}
	}
}

func TestInstallCodexWritesUserScopeFiles(t *testing.T) {
	home := runInstall(t, "codex")
	codexHome := filepath.Join(home, ".codex")
	skillsDir := filepath.Join(codexHome, "skills")

	for _, rel := range skillFiles {
		p := filepath.Join(skillsDir, rel)
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("expected installed skill %s: %v", p, err)
		}
		if strings.Contains(string(data), "{{") {
			t.Errorf("%s has unexpanded template directives", rel)
		}
		for _, forbidden := range []string{"$@", "${1:", "/rsx:", "~/.pi/", "~/.claude/", "/skill:"} {
			if strings.Contains(string(data), forbidden) {
				t.Errorf("%s contains foreign-agent syntax %q", rel, forbidden)
			}
		}
	}

	// Codex reads the skill's display metadata from agents/openai.yaml.
	meta := filepath.Join(skillsDir, "respec", "agents", "openai.yaml")
	if _, err := os.ReadFile(meta); err != nil {
		t.Fatalf("expected installed skill metadata %s: %v", meta, err)
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
	for _, rel := range skillFiles {
		if _, err := os.Stat(filepath.Join(skillsDir, rel)); err != nil {
			t.Errorf("pruning removed a rendered skill file %s: %v", rel, err)
		}
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	home := runInstall(t, "pi")
	research := filepath.Join(home, ".pi", "agent", "skills", "rsx-research", "SKILL.md")
	first, err := os.ReadFile(research)
	if err != nil {
		t.Fatal(err)
	}

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

// The pre-migration install wrote prompt templates; the skills-only install
// must remove them so they do not linger as slash commands pointing at a
// workflow that no longer exists. Other files in the same directory are not
// respec's.
func TestInstallPrunesRetiredPrompts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", "")

	promptsDir := filepath.Join(home, ".pi", "agent", "prompts")
	retired := []string{"rsx:research.md", "rsx:plan.md", "rsx:implement.md"}
	foreign := "someone-elses.md"
	for _, name := range append(retired, foreign) {
		if err := os.MkdirAll(promptsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(promptsDir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	rootCmd.SetArgs([]string{"install", "--target", "pi"})
	rootCmd.SetOut(os.NewFile(0, os.DevNull))
	rootCmd.SetErr(os.NewFile(0, os.DevNull))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("install: %v", err)
	}

	for _, name := range retired {
		if _, err := os.Stat(filepath.Join(promptsDir, name)); !os.IsNotExist(err) {
			t.Errorf("retired prompt %s should have been removed", name)
		}
	}
	if _, err := os.Stat(filepath.Join(promptsDir, foreign)); err != nil {
		t.Errorf("non-respec prompt must be left alone: %v", err)
	}
}

// A prompt directory emptied by the cleanup goes with it.
func TestInstallRemovesEmptiedPromptDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", "")

	promptsDir := filepath.Join(home, ".pi", "agent", "prompts")
	if err := os.MkdirAll(promptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(promptsDir, "rsx:research.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"install", "--target", "pi"})
	rootCmd.SetOut(os.NewFile(0, os.DevNull))
	rootCmd.SetErr(os.NewFile(0, os.DevNull))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("install: %v", err)
	}

	if _, err := os.Stat(promptsDir); !os.IsNotExist(err) {
		t.Errorf("emptied prompt dir %s should have been removed", promptsDir)
	}
}
