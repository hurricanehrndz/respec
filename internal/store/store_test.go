package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Hello World":         "hello-world",
		"  Spaces  ":          "spaces",
		"Mixed_Case/Slashes!": "mixed-case-slashes",
		"already-slug":        "already-slug",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestChangeDir(t *testing.T) {
	s := New("/root/store")
	date := time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC)
	got := s.ChangeDir("Auth System", "Add OAuth Login", date)
	want := filepath.Join("/root/store", "auth-system", "2026-06-25-add-oauth-login")
	if got != want {
		t.Errorf("ChangeDir = %q, want %q", got, want)
	}
}

func TestSiblingPaths(t *testing.T) {
	dir := "/root/store/space/2026-06-25-slug"
	if got := SpecPath(dir); got != filepath.Join(dir, "spec.md") {
		t.Errorf("SpecPath = %q", got)
	}
	if got := PlanPath(dir); got != filepath.Join(dir, "plan.md") {
		t.Errorf("PlanPath = %q", got)
	}
	if got := ResearchPath(dir); got != filepath.Join(dir, "research.md") {
		t.Errorf("ResearchPath = %q", got)
	}
}

func TestValidateChangeDir(t *testing.T) {
	dir := t.TempDir()
	if err := ValidateChangeDir(dir); err == nil {
		t.Error("expected error for change dir without plan.md")
	}
	if err := os.WriteFile(PlanPath(dir), []byte("# plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateChangeDir(dir); err != nil {
		t.Errorf("ValidateChangeDir on valid dir: %v", err)
	}
	if err := ValidateChangeDir(filepath.Join(dir, "nope")); err == nil {
		t.Error("expected error for missing dir")
	}
}

func TestSHA256File(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(p, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := SHA256File(p)
	if err != nil {
		t.Fatalf("SHA256File: %v", err)
	}
	// sha256("abc")
	want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got != want {
		t.Errorf("SHA256File = %q, want %q", got, want)
	}
}
