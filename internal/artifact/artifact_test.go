package artifact

import (
	"strings"
	"testing"
)

const validResearch = `---
topic: auth token rotation
status: complete
date: 2026-07-01T10:00:00-07:00
repo: git@github.com:me/repo.git
repo_path: /home/me/repo
git_commit: abc123
tags: [research, auth]
---
# Research: Auth

## Research Question
What exists?

## Summary
Findings summary.

## Findings
- ` + "`main.go:1`" + ` does things.

## Open Questions
None.
`

func problemsFor(t *testing.T, kind Kind, doc string) []string {
	t.Helper()
	art, err := Parse(kind, []byte(doc))
	if err != nil {
		t.Fatalf("Parse(%s): %v", kind, err)
	}
	return art.Validate()
}

func TestResearchValid(t *testing.T) {
	if p := problemsFor(t, KindResearch, validResearch); len(p) != 0 {
		t.Fatalf("valid research reported problems: %v", p)
	}
}

func TestResearchMissingProvenanceAndSections(t *testing.T) {
	// Body has only the Research Question heading; frontmatter lacks the
	// CLI-stamped provenance. This is the "wrote prose, forgot to stamp" state.
	doc := "---\ntopic: t\nstatus: complete\n---\n# R\n\n## Research Question\nq\n"
	got := problemsFor(t, KindResearch, doc)
	want := []string{
		"missing frontmatter: date",
		"missing frontmatter: repo",
		"missing frontmatter: repo_path",
		"missing frontmatter: git_commit",
		"missing section: Summary",
		"missing section: Findings",
		"missing section: Open Questions",
	}
	for _, w := range want {
		if !contains(got, w) {
			t.Errorf("missing expected problem %q in %v", w, got)
		}
	}
}

func TestResearchBadStatus(t *testing.T) {
	doc := strings.Replace(validResearch, "status: complete", "status: wip", 1)
	got := problemsFor(t, KindResearch, doc)
	if !contains(got, `invalid status "wip" (allowed: draft|complete)`) {
		t.Errorf("expected bad-status problem, got %v", got)
	}
}

func TestPlanRequiresPhaseAndVerification(t *testing.T) {
	// Overview present but no Phase / verification sections, and no spec_sha256.
	doc := "---\ntitle: p\nstatus: draft\n---\n# Plan\n\n## Overview\nx\n"
	got := problemsFor(t, KindPlan, doc)
	for _, w := range []string{
		"missing frontmatter: spec_sha256",
		"missing section: Automated Verification",
		"missing section: Manual Verification",
		"missing section: at least one Phase",
	} {
		if !contains(got, w) {
			t.Errorf("expected %q in %v", w, got)
		}
	}
}

func TestPlanValid(t *testing.T) {
	doc := "---\ntitle: p\nstatus: in-progress\nexecution_mode: auto\nspec_sha256: deadbeef\n---\n" +
		"# Plan\n\n## Overview\nx\n\n## Phase 1: Do it\n**Agent:** claude:opus@high\n\nchanges\n\n" +
		"### Automated Verification\n- [ ] just test\n\n### Manual Verification\n- [ ] works\n"
	if p := problemsFor(t, KindPlan, doc); len(p) != 0 {
		t.Fatalf("valid plan reported problems: %v", p)
	}
}

// A plan written before per-phase delegation existed has no **Agent:** lines
// and must still lint clean: absence means no preference was stated, not an
// error. See TestAgentLineIsNeverRequired for the per-mode assertion.
func TestPlanWithoutAgentLinesStaysValid(t *testing.T) {
	doc := "---\ntitle: p\nstatus: in-progress\nexecution_mode: auto\nspec_sha256: deadbeef\n---\n" +
		"# Plan\n\n## Overview\nx\n\n## Phase 1: Do it\nchanges\n\n" +
		"### Automated Verification\n- [ ] just test\n\n### Manual Verification\n- [ ] works\n"
	if got := problemsFor(t, KindPlan, doc); len(got) != 0 {
		t.Errorf("legacy plan should lint clean, got %v", got)
	}
}

func TestPlanRejectsBadExecutionMode(t *testing.T) {
	doc := "---\ntitle: p\nstatus: draft\nexecution_mode: sometimes\nspec_sha256: deadbeef\n---\n" +
		"# Plan\n\n## Overview\nx\n\n## Phase 1: Do it\n\n" +
		"### Automated Verification\n- [ ] test\n\n### Manual Verification\n- [ ] inspect\n"
	got := problemsFor(t, KindPlan, doc)
	if !contains(got, `invalid execution_mode "sometimes" (allowed: manual|auto)`) {
		t.Errorf("expected bad execution-mode problem, got %v", got)
	}
}

const pullRequestTable = `| Repository | PR | Phase | Base | Depends on | Merge order | Reason | State |
| --- | --- | --- | --- | --- | --- | --- | --- |
`

func TestPlanPullRequests(t *testing.T) {
	row := "| owner/api | [#11](https://example.com/owner/api/pull/11) | Phase 1 | main | none | 1 | exports API for owner/app#12 | open |\n"
	for _, tc := range []struct {
		name    string
		section string
		problem string
	}{
		{"absent", "", ""},
		{"empty record", "## Pull requests\n\n" + pullRequestTable, ""},
		{"open", "## Pull requests\n\n" + pullRequestTable + row, ""},
		{"draft", "## Pull requests\n\n" + pullRequestTable + strings.Replace(row, "| open |", "| draft |", 1), ""},
		{"merged", "## Pull requests\n\n" + pullRequestTable + strings.Replace(row, "| open |", "| merged |", 1), ""},
		{"closed", "## Pull requests\n\n" + pullRequestTable + strings.Replace(row, "| open |", "| closed |", 1), ""},
		{"superseded", "## Pull requests\n\n" + pullRequestTable + strings.Replace(row, "| open |", "| superseded |", 1), ""},
		{"missing table", "## Pull requests\n\nNone yet.\n", "Pull requests: missing table"},
		{"missing column", "## Pull requests\n\n" + strings.Replace(pullRequestTable, "Repository", "Repo", 1) + row, "Pull requests: missing column: Repository"},
		{"duplicate column", "## Pull requests\n\n" + strings.Replace(strings.Replace(pullRequestTable, "| State |", "| State | PR |", 1), "| --- |\n", "| --- | --- |\n", 1), "Pull requests: duplicate column: pr"},
		{"negative order", "## Pull requests\n\n" + pullRequestTable + strings.Replace(row, "| 1 |", "| -1 |", 1), "Pull requests row 1: Merge order must be a positive integer"},
		{"empty field", "## Pull requests\n\n" + pullRequestTable + strings.Replace(row, "| main |", "| |", 1), "Pull requests row 1: missing Base"},
		{"bad state", "## Pull requests\n\n" + pullRequestTable + strings.Replace(row, "| open |", "| pending |", 1), `Pull requests row 1: invalid State "pending" (allowed: open|draft|merged|closed|superseded)`},
		{"zero order", "## Pull requests\n\n" + pullRequestTable + strings.Replace(row, "| 1 |", "| 0 |", 1), "Pull requests row 1: Merge order must be a positive integer"},
		{"prose order", "## Pull requests\n\n" + pullRequestTable + strings.Replace(row, "| 1 |", "| first |", 1), "Pull requests row 1: Merge order must be a positive integer"},
		{"code is not a record", "```markdown\n## Pull requests\n\n" + pullRequestTable + "```\n", ""},
		{"other section table", "## Pull requests\n\n## Notes\n\n" + pullRequestTable + row, "Pull requests: missing table"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := planWithBody("manual", twoPhases+"\n"+tc.section).Validate()
			if tc.problem == "" {
				if len(got) != 0 {
					t.Fatalf("valid plan reported problems: %v", got)
				}
			} else if !contains(got, tc.problem) {
				t.Errorf("expected %q in %v", tc.problem, got)
			}
		})
	}
}

func TestPlanPullRequestRequiredCells(t *testing.T) {
	values := []string{"owner/api", "#11", "Phase 1", "main", "none", "1", "exports API", "open"}
	for i, name := range pullRequestColumns {
		t.Run(name, func(t *testing.T) {
			cells := append([]string(nil), values...)
			cells[i] = ""
			body := twoPhases + "\n## Pull requests\n\n" + pullRequestTable + "| " + strings.Join(cells, " | ") + " |\n"
			got := planWithBody("manual", body).Validate()
			if !contains(got, "Pull requests row 1: missing "+name) {
				t.Errorf("empty required cell must be rejected: %v", got)
			}
		})
	}
}

func TestPlanPullRequestColumnsCanBeReordered(t *testing.T) {
	body := twoPhases + `
## pull requests

| state | reason | merge order | depends on | base | phase | pr | repository |
| --- | --- | --- | --- | --- | --- | --- | --- |
| open | exports API | 1 | none | main | Phase 1 | #11 | owner/api |
`
	if got := planWithBody("manual", body).Validate(); len(got) != 0 {
		t.Fatalf("column order and case must not matter: %v", got)
	}
}

// Historical rows can retain their order, and optional evidence needs no issue tracker.
func TestPlanPullRequestsOptionalColumnsAndHistory(t *testing.T) {
	table := strings.Replace(pullRequestTable, "| State |", "| State | Issue | Reviewed head |", 1)
	table = strings.Replace(table, "| --- |\n", "| --- | --- | --- |\n", 1)
	body := twoPhases + "\n## Pull requests\n\n" + table +
		"| owner/api | #11 | Phase 1 | main | none | 1 | replaced by owner/api#13 | superseded | | abc123 |\n" +
		"| owner/api | #13 | Phase 1 | main | none | 1 | exports API | open | | |\n"
	if got := planWithBody("manual", body).Validate(); len(got) != 0 {
		t.Fatalf("history and optional columns should lint clean: %v", got)
	}
}

func TestSpecValid(t *testing.T) {
	doc := "---\ntitle: s\nstatus: approved\ndate: 2026-07-01T00:00:00Z\n---\n# S\n\n## Requirements\nr\n"
	if p := problemsFor(t, KindSpec, doc); len(p) != 0 {
		t.Fatalf("valid spec reported problems: %v", p)
	}
}

func TestParseUnknownKind(t *testing.T) {
	if _, err := Parse(Kind("bogus"), []byte("")); err == nil {
		t.Fatal("expected error for unknown kind")
	}
}

func TestPlanIsDone(t *testing.T) {
	if (Plan{Status: "in-progress"}).IsDone() {
		t.Error("in-progress plan reported done")
	}
	if !(Plan{Status: StatusDone}).IsDone() {
		t.Error("done plan not reported done")
	}
}

func TestStatusOf(t *testing.T) {
	cases := []struct {
		a    Artifact
		want string
	}{
		{Research{Status: "draft"}, "draft"},
		{Spec{Status: "approved"}, "approved"},
		{Plan{Status: "done"}, "done"},
	}
	for _, c := range cases {
		if got := StatusOf(c.a); got != c.want {
			t.Errorf("StatusOf(%T) = %q, want %q", c.a, got, c.want)
		}
	}
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
