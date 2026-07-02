package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedChange writes a change under <store>/<repo>/<slug>/ with the given plan
// frontmatter body, returning the store root (set via HOME/config beforehand).
func seedChange(t *testing.T, store, repo, slug, planFM string) {
	t.Helper()
	dir := filepath.Join(store, repo, slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("# s\n\n## Requirements\nr\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plan.md"), []byte(planFM), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListGroupsAndBlocks(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	store := filepath.Join(home, "store")
	runRespec(t, "config", "set", "store", store)

	// token-rotation is in-progress; typeahead depends on it -> blocked.
	seedChange(t, store, "hurricanehrndz-respec", "token-rotation",
		"---\ntitle: TR\nstatus: in-progress\n---\n# P\n")
	seedChange(t, store, "hurricanehrndz-respec", "typeahead",
		"---\ntitle: TA\nstatus: draft\ndepends_on: [token-rotation]\n---\n# P\n")
	// A cross-repo effort that depends on a done effort -> not blocked.
	seedChange(t, store, "acme-billing", "proration",
		"---\ntitle: PR\nstatus: done\n---\n# P\n")
	seedChange(t, store, "acme-billing", "invoicing",
		"---\ntitle: IN\nstatus: draft\ndepends_on: [proration]\n---\n# P\n")

	out := runRespec(t, "list", "--json")
	var rows []listRow
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		t.Fatalf("decode list json: %v\n%s", err, out)
	}
	by := map[string]listRow{}
	for _, r := range rows {
		by[r.Repo+"/"+r.Slug] = r
	}
	if len(rows) != 4 {
		t.Fatalf("want 4 rows, got %d: %+v", len(rows), rows)
	}
	if r := by["hurricanehrndz-respec/typeahead"]; !r.Blocked {
		t.Errorf("typeahead should be blocked (depends on in-progress token-rotation): %+v", r)
	}
	if r := by["acme-billing/invoicing"]; r.Blocked {
		t.Errorf("invoicing should not be blocked (depends on done proration): %+v", r)
	}
	if r := by["hurricanehrndz-respec/token-rotation"]; r.Blocked {
		t.Errorf("token-rotation has no deps, should not be blocked: %+v", r)
	}
}

func TestListPlainTableAndEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	store := filepath.Join(home, "store")
	runRespec(t, "config", "set", "store", store)

	// Empty store.
	out, err := runRespecErr(t, "list", "--json=false")
	if err != nil {
		t.Fatalf("list empty: %v", err)
	}
	if !strings.Contains(out, "no changes found") {
		t.Errorf("empty store should say so, got:\n%s", out)
	}

	// One effort -> table with header.
	seedChange(t, store, "acme-web", "login", "---\ntitle: L\nstatus: draft\n---\n# P\n")
	out, err = runRespecErr(t, "list", "--json=false")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out, "REPO") || !strings.Contains(out, "acme-web") || !strings.Contains(out, "login") {
		t.Errorf("table missing expected content:\n%s", out)
	}
}

func TestListSurfacesMalformedPlan(t *testing.T) {
	// A plan.md with broken frontmatter must show spec "error" and warn on
	// stderr — not masquerade as an unstamped effort (fail loud).
	home := t.TempDir()
	t.Setenv("HOME", home)
	store := filepath.Join(home, "store")
	runRespec(t, "config", "set", "store", store)
	seedChange(t, store, "acme-web", "broken", "---\ntitle: a: b\n---\n# P\n")

	out, err := runRespecErr(t, "list", "--json")
	if err != nil {
		t.Fatalf("list: %v\n%s", err, out)
	}
	// stderr is merged into out by runRespecErr; the JSON line still decodes.
	if !strings.Contains(out, "warning:") {
		t.Errorf("expected a warning for malformed plan.md:\n%s", out)
	}
	if !strings.Contains(out, `"spec":"error"`) {
		t.Errorf("expected spec state 'error':\n%s", out)
	}
}
