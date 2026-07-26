// Package agents discovers the delegation targets available on this machine
// and builds the exact command line for each one.
//
// The roster is deliberately *not* configuration. Which harnesses are
// installed, and which models they can reach, drifts as tools are added,
// upgraded, and retired — a stored list would be wrong shortly after it was
// written. Everything here is probed at the moment the operator asks, so the
// plan phase negotiates against reality rather than against a stale file.
package agents

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// Harness is a coding-agent CLI capable of running one delegated task.
type Harness string

const (
	HarnessPi     Harness = "pi"
	HarnessClaude Harness = "claude"
	HarnessCodex  Harness = "codex"
)

// Harnesses lists every harness respec knows how to drive, in preference
// order for display.
var Harnesses = []Harness{HarnessPi, HarnessClaude, HarnessCodex}

// efforts records the reasoning levels each harness accepts, weakest first.
// These come from the CLIs' own help output and are validated before a command
// is built so a plan cannot name a level its harness will reject at run time.
var efforts = map[Harness][]string{
	HarnessPi:     {"off", "minimal", "low", "medium", "high", "xhigh", "max"},
	HarnessClaude: {"low", "medium", "high", "xhigh", "max"},
	HarnessCodex:  {"minimal", "low", "medium", "high"},
}

// Efforts returns the reasoning levels a harness accepts, weakest first.
func Efforts(h Harness) []string { return efforts[h] }

// Known reports whether h is a harness respec can drive.
func Known(h Harness) bool { _, ok := efforts[h]; return ok }

// Availability is what probing found for one harness.
type Availability struct {
	Harness Harness  `json:"harness"`
	Path    string   `json:"path,omitempty"`    // resolved binary, empty when absent
	Efforts []string `json:"efforts,omitempty"` // accepted reasoning levels
	Models  []string `json:"models,omitempty"`  // enumerable models, may be empty
	// ModelsNote explains why Models is empty or partial, so the operator is
	// never left guessing whether "no models" means "none" or "cannot tell".
	ModelsNote string `json:"models_note,omitempty"`
}

// Installed reports whether the harness binary was found on PATH.
func (a Availability) Installed() bool { return a.Path != "" }

// Detect probes PATH for every known harness and enumerates what each can
// reach. Model enumeration runs a subprocess, so it is bounded by ctx; a
// harness that cannot be enumerated is still reported as installed.
func Detect(ctx context.Context) []Availability {
	out := make([]Availability, 0, len(Harnesses))
	for _, h := range Harnesses {
		a := Availability{Harness: h, Efforts: efforts[h]}
		path, err := exec.LookPath(string(h))
		if err != nil {
			out = append(out, a)
			continue
		}
		a.Path = path
		a.Models, a.ModelsNote = enumerate(ctx, h)
		out = append(out, a)
	}
	return out
}

// enumerate lists the models a harness can reach, or explains why it cannot.
func enumerate(ctx context.Context, h Harness) ([]string, string) {
	switch h {
	case HarnessPi:
		models, err := piModels(ctx)
		if err != nil {
			return nil, fmt.Sprintf("could not list pi models: %v", err)
		}
		return models, ""
	case HarnessClaude:
		// Claude Code has no list-models command; these are the aliases its
		// --model flag documents. Full model IDs are also accepted.
		return []string{"fable", "opus", "sonnet", "haiku"}, "aliases only; --model also accepts full model IDs"
	case HarnessCodex:
		return nil, "codex cannot enumerate models; pass one its account can reach"
	}
	return nil, ""
}

// piModels runs `pi --list-models` and returns provider/id pairs. pi's
// --model flag accepts the "provider/id" form directly, so provider is folded
// into the model string rather than carried separately.
func piModels(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pi", "--list-models")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var models []string
	seen := map[string]bool{}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		// Rows are "provider model context max-out thinking images"; the
		// header row names its first column "provider".
		if len(fields) < 2 || fields[0] == "provider" {
			continue
		}
		id := fields[0] + "/" + fields[1]
		if !seen[id] {
			seen[id] = true
			models = append(models, id)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	sort.Strings(models)
	return models, nil
}
