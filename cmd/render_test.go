package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupStore configures a temp HOME with a store containing one artifact that
// has inline HTML and a GFM table, and returns the output dir to render into.
func setupStore(t *testing.T) (home, store, out string) {
	t.Helper()
	home = t.TempDir()
	t.Setenv("HOME", home)
	// UserCacheDir honors XDG_CACHE_HOME / falls back to HOME on macOS; pin it
	// under the temp HOME so the build dir is isolated.
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))

	store = filepath.Join(home, "respec-store")
	changeDir := filepath.Join(store, "demo-space", "2026-06-25-demo")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	artifact := "---\ntitle: Demo Spec\n---\n# Demo\n\nInline <span class=\"flag\">HTML</span> here.\n\n| a | b |\n|---|---|\n| 1 | 2 |\n"
	if err := os.WriteFile(filepath.Join(changeDir, "spec.md"), []byte(artifact), 0o644); err != nil {
		t.Fatal(err)
	}
	return home, store, filepath.Join(home, "out")
}

func TestRenderPreservesInlineHTMLAndTables(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not on PATH; skipping render integration test")
	}
	_, _, out := setupStore(t)

	runRespec(t, "render", "--out", out)

	// Find the rendered spec page.
	var html string
	err := filepath.Walk(out, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(p, ".html") {
			return err
		}
		b, _ := os.ReadFile(p)
		if strings.Contains(string(b), "Inline") {
			html = string(b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if html == "" {
		t.Fatal("no rendered page containing the artifact found")
	}
	if !strings.Contains(html, "<span class=\"flag\">HTML</span>") {
		t.Errorf("inline HTML not preserved in output:\n%s", html)
	}
	if !strings.Contains(html, "<table>") {
		t.Errorf("GFM table not rendered as <table>:\n%s", html)
	}
}

func TestRenderFailsLoudWithoutHugo(t *testing.T) {
	setupStore(t)
	// Force an empty PATH so hugo cannot be found.
	t.Setenv("PATH", t.TempDir())

	rootCmd.SetArgs([]string{"render", "--out", filepath.Join(t.TempDir(), "o")})
	var sink strings.Builder
	rootCmd.SetOut(&sink)
	rootCmd.SetErr(&sink)
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when hugo is not on PATH")
	}
	if !strings.Contains(err.Error(), "hugo") {
		t.Errorf("error should mention hugo, got: %v", err)
	}
}
