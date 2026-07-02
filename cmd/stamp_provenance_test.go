package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hurricanehrndz/respec/internal/frontmatter"
	"github.com/hurricanehrndz/respec/internal/store"
)

// initGitRepo creates a temp git repo with one commit and returns its path.
func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "f.txt")
	git(t, dir, "-c", "user.email=t@t", "-c", "user.name=t", "-c", "commit.gpgsign=false", "commit", "-m", "one")
	return dir
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func fm(t *testing.T, path, key string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	v, _, err := frontmatter.GetString(data, key)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestStampWritesProvenanceWriteOnce(t *testing.T) {
	repo := initGitRepo(t)
	dir := fixtureChange(t, "the spec\n", "---\n---\n# Plan\n\nbody\n")
	research := "---\ntopic: t\nstatus: complete\n---\n# R\n\n## Research Question\nq\n"
	if err := os.WriteFile(store.ResearchPath(dir), []byte(research), 0o644); err != nil {
		t.Fatal(err)
	}

	runRespec(t, "stamp", dir, "--repo", repo)

	rp := store.ResearchPath(dir)
	if got := fm(t, rp, "date"); got == "" {
		t.Error("research.md missing stamped date")
	}
	if got := fm(t, rp, "repo_path"); got == "" {
		t.Error("research.md missing stamped repo_path")
	}
	commit1 := fm(t, rp, "git_commit")
	if len(commit1) != 40 {
		t.Fatalf("git_commit = %q, want 40-char sha", commit1)
	}
	// spec.md gets a date too.
	if got := fm(t, store.SpecPath(dir), "date"); got == "" {
		t.Error("spec.md missing stamped date")
	}

	// A new commit in the repo must NOT change research.md's provenance
	// (write-once preserves the research-time snapshot).
	if err := os.WriteFile(filepath.Join(repo, "g.txt"), []byte("y\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, repo, "add", "g.txt")
	git(t, repo, "-c", "user.email=t@t", "-c", "user.name=t", "-c", "commit.gpgsign=false", "commit", "-m", "two")

	runRespec(t, "stamp", dir, "--repo", repo)
	if commit2 := fm(t, rp, "git_commit"); commit2 != commit1 {
		t.Errorf("git_commit changed on re-stamp: %q -> %q (want write-once)", commit1, commit2)
	}
}

func TestStampNonGitRepoFailsLoud(t *testing.T) {
	// respec's provenance model is git-based; lint requires repo/git_commit,
	// so silently skipping them would dead-end the workflow. Stamp must error.
	nonRepo := t.TempDir() // not a git repo
	dir := fixtureChange(t, "spec\n", "---\n---\n# Plan\n")
	research := "---\ntopic: t\n---\n# R\n"
	if err := os.WriteFile(store.ResearchPath(dir), []byte(research), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runRespecErr(t, "stamp", dir, "--repo", nonRepo); err == nil {
		t.Fatal("expected stamp to fail for a non-git worked-on repo")
	}
	// Nothing was stamped (no partial mutation).
	if got := fm(t, store.ResearchPath(dir), "date"); got != "" {
		t.Errorf("date = %q, want unstamped after failed provenance read", got)
	}
}

func TestStampResearchOnlyChangeDir(t *testing.T) {
	// The research phase stamps a dir holding only research.md — no plan.md
	// or spec.md exists yet, and that must not be an error.
	repo := initGitRepo(t)
	dir := t.TempDir()
	research := "---\ntopic: t\nstatus: draft\n---\n# R\n"
	if err := os.WriteFile(store.ResearchPath(dir), []byte(research), 0o644); err != nil {
		t.Fatal(err)
	}
	runRespec(t, "stamp", dir, "--repo", repo)
	rp := store.ResearchPath(dir)
	if got := fm(t, rp, "date"); got == "" {
		t.Error("research-only stamp did not fill date")
	}
	if got := fm(t, rp, "git_commit"); len(got) != 40 {
		t.Errorf("git_commit = %q, want 40-char sha", got)
	}
}

func TestStampFillsEmptyScaffoldedKeys(t *testing.T) {
	// A scaffolded empty key (`date:`) must be filled, not treated as
	// already-stamped by the write-once check.
	repo := initGitRepo(t)
	dir := t.TempDir()
	research := "---\ntopic: t\ndate:\nrepo:\n---\n# R\n"
	if err := os.WriteFile(store.ResearchPath(dir), []byte(research), 0o644); err != nil {
		t.Fatal(err)
	}
	runRespec(t, "stamp", dir, "--repo", repo)
	if got := fm(t, store.ResearchPath(dir), "date"); got == "" || got == "<nil>" {
		t.Errorf("empty date: key not filled, got %q", got)
	}
	if got := fm(t, store.ResearchPath(dir), "repo"); got == "" || got == "<nil>" {
		t.Errorf("empty repo: key not filled, got %q", got)
	}
}
