package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/hurricanehrndz/respec/internal/store"
)

// runRespecErr executes the root command with args, returning output and error
// (unlike runRespec, which fails the test on any error).
func runRespecErr(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	rootCmd.SetArgs(args)
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	err := rootCmd.Execute()
	return out.String(), err
}

const goodPlan = "---\ntitle: p\nstatus: in-progress\nspec_sha256: abc\n---\n" +
	"# Plan\n\n## Overview\nx\n\n## Phase 1: Go\nchanges\n\n" +
	"### Automated Verification\n- [ ] just test\n\n### Manual Verification\n- [ ] works\n"

const goodSpec = "---\ntitle: s\nstatus: approved\ndate: 2026-07-01T00:00:00Z\n---\n# S\n\n## Requirements\nr\n"

func TestLintPasses(t *testing.T) {
	dir := fixtureChange(t, goodSpec, goodPlan)
	out, err := runRespecErr(t, "lint", dir)
	if err != nil {
		t.Fatalf("lint failed unexpectedly: %v\n%s", err, out)
	}
	if !strings.Contains(out, "plan.md: ok") || !strings.Contains(out, "spec.md: ok") {
		t.Errorf("expected ok lines, got:\n%s", out)
	}
}

func TestLintFailsAndExitsNonZero(t *testing.T) {
	// plan.md missing required sections and spec_sha256.
	dir := fixtureChange(t, goodSpec, "---\ntitle: p\nstatus: draft\n---\n# Plan\n\nno sections\n")
	out, err := runRespecErr(t, "lint", dir)
	if err == nil {
		t.Fatalf("expected lint to fail, output:\n%s", out)
	}
	if !strings.Contains(out, "missing section: Overview") {
		t.Errorf("expected section problem in output:\n%s", out)
	}
}

func TestLintSkipsAbsentArtifacts(t *testing.T) {
	// Only plan.md present (research.md/spec.md absent) — lint validates just plan.
	dir := t.TempDir()
	if err := os.WriteFile(store.PlanPath(dir), []byte(goodPlan), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runRespecErr(t, "lint", dir)
	if err != nil {
		t.Fatalf("lint failed: %v\n%s", err, out)
	}
	if strings.Contains(out, "research.md") || strings.Contains(out, "spec.md") {
		t.Errorf("absent artifacts should be skipped, got:\n%s", out)
	}
}
