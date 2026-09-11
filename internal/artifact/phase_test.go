package artifact

import (
	"strings"
	"testing"
)

func planWithBody(mode, body string) Plan {
	return Plan{
		Title:         "T",
		Status:        "approved",
		SpecSHA:       "abc",
		ExecutionMode: mode,
		body:          []byte(body),
	}
}

const twoPhases = `## Overview
o

## Phase 1: Cheap mechanical edit
**Agent:** pi:openai-codex/gpt-5.4-mini@low

### Automated Verification
- [ ] x

## Phase 2: Subtle concurrency fix
**Agent:** claude:opus

### Manual Verification
- [ ] Orchestrator: y
`

func TestPhasesPairSpecsWithHeadings(t *testing.T) {
	got := planWithBody("auto", twoPhases).Phases()
	if len(got) != 2 {
		t.Fatalf("want 2 phases, got %d: %+v", len(got), got)
	}
	if got[0].Agent != "pi:openai-codex/gpt-5.4-mini@low" {
		t.Errorf("phase 1 agent: %q", got[0].Agent)
	}
	if got[1].Agent != "claude:opus" {
		t.Errorf("phase 2 agent: %q", got[1].Agent)
	}
}

// The **Agent:** line is optional in every mode. Requiring it would break
// native-worker plans and plans written before delegation existed.
func TestAgentLineIsNeverRequired(t *testing.T) {
	body := "## Overview\no\n\n## Phase 1: Unassigned\nchanges\n"
	for _, mode := range []string{"auto", "manual", ""} {
		for _, p := range planWithBody(mode, body).Validate() {
			if strings.Contains(p, agentLinePrefix) {
				t.Errorf("mode %q: agent line must not be required, got %q", mode, p)
			}
		}
	}
}

// Native model hints belong to the app, not the external CLI spec parser.
func TestNativeWorkerPreferencesStayProse(t *testing.T) {
	body := `## Overview
Use native workers.

## Worker preferences
- Implementation: prefer app-model at medium effort
- Review: prefer app-reviewer at high effort
- Commits: native defaults
- Fallback: native defaults; report unavailable preferences

## Phase 1: Implement
Implementation: prefer another-app-model for this phase.

### Automated Verification
- [ ] Run tests

### Manual Verification
- [ ] Orchestrator: inspect the diff
`
	for _, mode := range []string{"auto", "manual"} {
		plan := planWithBody(mode, body)
		if problems := plan.Validate(); len(problems) != 0 {
			t.Errorf("%s: native preferences must not need CLI specs: %v", mode, problems)
		}
		phases := plan.Phases()
		if len(phases) != 1 || phases[0].Agent != "" {
			t.Errorf("%s: native preferences became an external assignment: %+v", mode, phases)
		}
	}
}

// A malformed spec is always an error: it fails only once the phase is already
// running, which is the most expensive moment to find out.
func TestMalformedSpecIsAlwaysAnError(t *testing.T) {
	for _, mode := range []string{"auto", "manual"} {
		body := "## Overview\no\n\n## Phase 1: Bad\n**Agent:** claude:opus@off\n"
		got := planWithBody(mode, body).Validate()
		if !containsSubstr(got, "does not accept effort") {
			t.Errorf("%s: expected effort rejection, got %v", mode, got)
		}
	}
}

// An auto plan must reach the end without stopping for a human, so an
// Operator: check anywhere but the final phase is a blocker. This is the rule
// that replaces having a separate auto-planning prompt.
func TestAutoRejectsOperatorGateBeforeFinalPhase(t *testing.T) {
	body := `## Overview
o

## Phase 1: Needs a human midway

### Manual Verification
- [ ] Operator: confirm the staging DNS cutover

## Phase 2: Last

### Manual Verification
- [ ] Operator: final acceptance
`
	got := planWithBody("auto", body).Validate()
	if !containsSubstr(got, "blocks an unattended run") {
		t.Errorf("expected the mid-run operator gate to be flagged, got %v", got)
	}
	if containsSubstr(got, "final acceptance") {
		t.Errorf("an operator check in the final phase is allowed, got %v", got)
	}
}

// Manual plans are supposed to stop for the operator, so the same gate is fine.
func TestManualAllowsOperatorGatesAnywhere(t *testing.T) {
	body := "## Overview\no\n\n## Phase 1: A\n\n### Manual Verification\n- [ ] Operator: check\n\n## Phase 2: B\n"
	if got := planWithBody("manual", body).Validate(); containsSubstr(got, "blocks an unattended run") {
		t.Errorf("manual mode must allow operator gates, got %v", got)
	}
}

// Verification subsections belong to their phase; a sibling section does not.
func TestPhaseBoundariesFollowHeadingLevel(t *testing.T) {
	body := `## Overview
o

## Phase 1: Real
**Agent:** claude:opus

### Manual Verification
- [ ] Operator: mine

## Testing Strategy
**Agent:** claude:sonnet@low
- [ ] Operator: not a phase check
`
	got := planWithBody("auto", body).Phases()
	if len(got) != 1 {
		t.Fatalf("want 1 phase, got %d: %+v", len(got), got)
	}
	if got[0].Agent != "claude:opus" {
		t.Errorf("phase picked up a spec from a sibling section: %q", got[0].Agent)
	}
	if len(got[0].OperatorChecks) != 1 {
		t.Errorf("phase should own exactly its own operator check, got %v", got[0].OperatorChecks)
	}
}

func containsSubstr(problems []string, want string) bool {
	for _, p := range problems {
		if strings.Contains(p, want) {
			return true
		}
	}
	return false
}
