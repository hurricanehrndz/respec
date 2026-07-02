package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/hurricanehrndz/respec/internal/store"
)

func TestStatusReportsPerArtifact(t *testing.T) {
	spec := "---\ntitle: s\nstatus: approved\n---\n# S\n\n## Requirements\nr\n"
	plan := "---\ntitle: p\nstatus: in-progress\n---\n# Plan\n"
	dir := fixtureChange(t, spec, plan)
	if err := os.WriteFile(store.ResearchPath(dir),
		[]byte("---\ntopic: t\nstatus: complete\n---\n# R\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res := statusJSON(t, dir)

	got := map[string]artifactStatus{}
	for _, a := range res.Artifacts {
		got[a.Artifact] = a
	}
	if a := got["research.md"]; !a.Present || a.Status != "complete" {
		t.Errorf("research.md = %+v, want present/complete", a)
	}
	if a := got["spec.md"]; !a.Present || a.Status != "approved" {
		t.Errorf("spec.md = %+v, want present/approved", a)
	}
	if a := got["plan.md"]; !a.Present || a.Status != "in-progress" {
		t.Errorf("plan.md = %+v, want present/in-progress", a)
	}
}

func TestStatusPlainKeepsStateWordFirst(t *testing.T) {
	dir := fixtureChange(t, "spec\n", "# Plan no frontmatter\n")
	out, err := runRespecErr(t, "status", dir, "--json=false")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	// First line must remain the bare state word for scripting.
	first := strings.SplitN(out, "\n", 2)[0]
	if first != "unstamped" {
		t.Errorf("first line = %q, want state word 'unstamped'", first)
	}
	if !strings.Contains(out, "spec.md") {
		t.Errorf("plain status should list artifacts:\n%s", out)
	}
}
