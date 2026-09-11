package artifact

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/hurricanehrndz/respec/internal/agents"
)

// agentLinePrefix marks a phase's delegation assignment in plan.md:
//
//	## Phase 2: Parse the roster
//	**Agent:** pi:openai-codex/gpt-5.6-sol@medium
//
// The assignment lives in the body rather than frontmatter because it belongs
// to one phase, and phases are body sections; keying frontmatter by phase
// number would break the moment a phase is renamed or reordered.
//
// It is always optional. Without it, implementation uses native workers and
// any preferences recorded in plan prose. Older plans keep working unchanged.
const agentLinePrefix = "**Agent:**"

// operatorMarker labels a verification item only a human can perform.
const operatorMarker = "Operator:"

// Phase is one `## Phase N: ...` section of a plan.
type Phase struct {
	Heading        string   // heading text, e.g. "Phase 2: Parse the roster"
	Agent          string   // raw **Agent:** spec, empty when unassigned
	OperatorChecks []string // Manual Verification items marked `Operator:`
}

// Phases returns the plan's phase sections in document order.
func (pl Plan) Phases() []Phase {
	var out []Phase
	sc := bufio.NewScanner(bytes.NewReader(pl.body))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	cur := -1
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		if h, ok := headingText(line); ok {
			if strings.HasPrefix(strings.ToLower(h), "phase") {
				out = append(out, Phase{Heading: h})
				cur = len(out) - 1
			} else if !isSubsectionOfPhase(h) {
				// A sibling heading (## Testing Strategy) closes the phase; a
				// nested one (### Manual Verification) belongs to it.
				cur = -1
			}
			continue
		}
		if cur < 0 {
			continue
		}

		if strings.HasPrefix(line, agentLinePrefix) && out[cur].Agent == "" {
			out[cur].Agent = strings.TrimSpace(strings.TrimPrefix(line, agentLinePrefix))
			continue
		}
		if i := strings.Index(line, operatorMarker); i >= 0 && strings.HasPrefix(line, "-") {
			out[cur].OperatorChecks = append(out[cur].OperatorChecks,
				strings.TrimSpace(line[i+len(operatorMarker):]))
		}
	}
	return out
}

// isSubsectionOfPhase reports whether a heading is one of the per-phase
// verification subsections rather than a new top-level section.
func isSubsectionOfPhase(h string) bool {
	l := strings.ToLower(h)
	return strings.HasPrefix(l, "automated verification") || strings.HasPrefix(l, "manual verification")
}

// headingText reports whether line is an ATX heading and returns its text.
func headingText(line string) (string, bool) {
	if !strings.HasPrefix(line, "#") {
		return "", false
	}
	rest := strings.TrimLeft(line, "#")
	if rest == line || (rest != "" && !strings.HasPrefix(rest, " ")) {
		return "", false // "#hashtag" is not a heading
	}
	return strings.TrimSpace(rest), true
}

// appendPhaseProblems validates each phase's delegation assignment and, for
// auto plans, its operator gates.
//
// The `**Agent:**` line is never required. Native preferences stay in prose;
// only explicit external assignments need CLI spec validation. A malformed
// spec is always an error because it would fail during phase execution.
//
// Under `execution_mode: auto` an `Operator:` check outside the final phase is
// rejected. That is the whole point of an auto plan — it must reach the end
// without stopping for a human — and enforcing it here means it holds for any
// plan, including one hand-edited after it was written.
func appendPhaseProblems(p []string, pl Plan) []string {
	phases := pl.Phases()
	for i, ph := range phases {
		if ph.Agent != "" {
			if _, err := agents.ParseSpec(ph.Agent); err != nil {
				p = append(p, fmt.Sprintf("%s: %v", ph.Heading, err))
			}
		}
		if pl.ExecutionMode != "auto" || i == len(phases)-1 {
			continue
		}
		for _, chk := range ph.OperatorChecks {
			p = append(p, fmt.Sprintf(
				"%s: %s %q blocks an unattended run (execution_mode: auto allows operator checks only in the final phase)",
				ph.Heading, operatorMarker, chk))
		}
	}
	return p
}
