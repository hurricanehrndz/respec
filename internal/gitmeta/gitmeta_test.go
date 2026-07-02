package gitmeta

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func gitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "f.txt")
	runGit(t, dir, "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-m", "init")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestReadReturnsCommitAndToplevel(t *testing.T) {
	dir := gitRepo(t)
	m, err := Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(m.Commit) != 40 {
		t.Errorf("Commit = %q, want 40-char sha", m.Commit)
	}
	// Toplevel should resolve to dir (allowing for symlinked temp roots).
	if got, err := filepath.EvalSymlinks(m.Toplevel); err != nil || got != mustEval(t, dir) {
		t.Errorf("Toplevel = %q, want %q", m.Toplevel, dir)
	}
	if m.Remote != "" {
		t.Errorf("Remote = %q, want empty (no origin configured)", m.Remote)
	}
}

func TestReadRemoteWhenConfigured(t *testing.T) {
	dir := gitRepo(t)
	runGit(t, dir, "remote", "add", "origin", "git@github.com:me/repo.git")
	m, err := Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if m.Remote != "git@github.com:me/repo.git" {
		t.Errorf("Remote = %q", m.Remote)
	}
}

func TestReadNonRepoErrors(t *testing.T) {
	if _, err := Read(t.TempDir()); err == nil {
		t.Fatal("expected error for non-git directory")
	}
}

func mustEval(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
