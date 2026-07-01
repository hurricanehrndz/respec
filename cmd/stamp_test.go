package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hurricanehrndz/respec/internal/store"
)

// runRespec executes the root command with args, capturing stdout.
func runRespec(t *testing.T, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	rootCmd.SetArgs(args)
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("respec %v: %v\noutput: %s", args, err, out.String())
	}
	return out.String()
}

func fixtureChange(t *testing.T, spec, plan string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(store.SpecPath(dir), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.PlanPath(dir), []byte(plan), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func statusJSON(t *testing.T, dir string) statusResult {
	t.Helper()
	out := runRespec(t, "status", dir, "--json")
	var res statusResult
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatalf("decoding status json %q: %v", out, err)
	}
	return res
}

func TestStampStatusLifecycle(t *testing.T) {
	planBody := "---\ntitle: p\n---\n# Plan\n\nbody\n"
	dir := fixtureChange(t, "spec one\n", planBody)

	// Unstamped before stamping.
	if res := statusJSON(t, dir); res.State != "unstamped" {
		t.Fatalf("pre-stamp state = %q, want unstamped", res.State)
	}

	runRespec(t, "stamp", dir)

	if res := statusJSON(t, dir); res.State != "fresh" {
		t.Fatalf("post-stamp state = %q, want fresh", res.State)
	}

	// Editing spec.md -> stale.
	if err := os.WriteFile(store.SpecPath(dir), []byte("spec CHANGED\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if res := statusJSON(t, dir); res.State != "stale" {
		t.Fatalf("after spec edit state = %q, want stale", res.State)
	}
}

func TestStatusAsymmetryPlanEditStaysFresh(t *testing.T) {
	dir := fixtureChange(t, "the spec\n", "---\n---\n# Plan\n\noriginal\n")
	runRespec(t, "stamp", dir)
	if res := statusJSON(t, dir); res.State != "fresh" {
		t.Fatalf("post-stamp = %q, want fresh", res.State)
	}
	// Edit only the plan body (preserve its frontmatter/stamp).
	plan, _ := os.ReadFile(store.PlanPath(dir))
	edited := bytes.Replace(plan, []byte("original"), []byte("revised plan prose"), 1)
	if err := os.WriteFile(store.PlanPath(dir), edited, 0o644); err != nil {
		t.Fatal(err)
	}
	if res := statusJSON(t, dir); res.State != "fresh" {
		t.Errorf("after plan-only edit state = %q, want fresh (R-6 asymmetry)", res.State)
	}
}

func TestStatusUnstampedMissingKey(t *testing.T) {
	dir := fixtureChange(t, "spec\n", "# Plan with no frontmatter\n")
	res := statusJSON(t, dir)
	if res.State != "unstamped" {
		t.Errorf("state = %q, want unstamped", res.State)
	}
	if filepath.Base(store.PlanPath(dir)) != "plan.md" {
		t.Fatal("sanity: plan path")
	}
}
