package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const unformattedMD = "# T\n\nword word word word word word word word word word word word word word word word word word word word\n"

func TestFormatDirectoryReflows(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "nested")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(sub, "doc.md")
	if err := os.WriteFile(p, []byte(unformattedMD), 0o644); err != nil {
		t.Fatal(err)
	}

	// --check=false is passed explicitly to defeat cobra flag persistence
	// across tests in this package.
	if _, err := runRespecErr(t, "format", dir, "--check=false"); err != nil {
		t.Fatalf("format dir: %v", err)
	}
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	// The long paragraph should now wrap across multiple lines.
	para := strings.SplitN(string(got), "\n\n", 2)[1]
	if !strings.Contains(strings.TrimRight(para, "\n"), "\n") {
		t.Errorf("expected reflowed (multi-line) paragraph, got:\n%s", got)
	}
}

func TestFormatCheckFailsOnUnformattedDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "doc.md"), []byte(unformattedMD), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runRespecErr(t, "format", dir, "--check")
	if err == nil {
		t.Fatalf("expected check to fail, output:\n%s", out)
	}
	if !strings.Contains(err.Error(), "not formatted") {
		t.Errorf("error = %q, want 'not formatted'", err.Error())
	}
}

func TestFormatErrorsOnDirWithoutMarkdown(t *testing.T) {
	// A dir with zero *.md files must fail loud, not let --check pass vacuously.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runRespecErr(t, "format", dir, "--check"); err == nil {
		t.Fatal("expected error for dir without *.md files")
	}
	if _, err := runRespecErr(t, "format", dir, "--check=false"); err == nil {
		t.Fatal("expected error for dir without *.md files (write mode)")
	}
}
