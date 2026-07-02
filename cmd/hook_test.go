package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupStoreRepo(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	store := filepath.Join(home, "respec-store")
	if err := os.MkdirAll(filepath.Join(store, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestInstallHook(t *testing.T) {
	store := setupStoreRepo(t)
	if _, err := runRespecErr(t, "install-hook", "--force=false"); err != nil {
		t.Fatalf("install-hook: %v", err)
	}
	hook := filepath.Join(store, ".git", "hooks", "pre-commit")
	info, err := os.Stat(hook)
	if err != nil {
		t.Fatalf("hook not written: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("hook not executable: mode %v", info.Mode())
	}
	data, err := os.ReadFile(hook)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), hookMarker) {
		t.Error("hook missing respec marker")
	}
	if !strings.HasPrefix(string(data), "#!/bin/sh") {
		t.Error("hook missing shebang")
	}
}

func TestInstallHookRefusesForeignWithoutForce(t *testing.T) {
	store := setupStoreRepo(t)
	hooksDir := filepath.Join(store, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	hook := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\n# hand written\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := runRespecErr(t, "install-hook", "--force=false"); err == nil {
		t.Fatal("expected refusal to overwrite foreign hook")
	}
	// --force overwrites.
	if _, err := runRespecErr(t, "install-hook", "--force"); err != nil {
		t.Fatalf("install-hook --force: %v", err)
	}
	data, _ := os.ReadFile(hook)
	if !strings.Contains(string(data), hookMarker) {
		t.Error("force did not install respec hook")
	}
}

func TestInstallHookErrorsWhenStoreNotGit(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// respec-store exists but has no .git.
	if err := os.MkdirAll(filepath.Join(home, "respec-store"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := runRespecErr(t, "install-hook", "--force=false"); err == nil {
		t.Fatal("expected error when store is not a git repo")
	}
}
