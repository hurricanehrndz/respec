package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hurricanehrndz/respec/internal/config"
)

func TestRenderSubstitutesStoreAndInjects(t *testing.T) {
	cfg := config.Defaults()
	cfg.Store = "/tmp/my-store"
	cfg.Context = "SHARED-CONTEXT-MARKER"
	cfg.Rules.Research = "RESEARCH-RULE-MARKER"

	r, err := Render(cfg, TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// The store path and shared context live in the respec skill (single
	// source of truth); the phase skills point at it rather than duplicating.
	shared, ok := r.Skills["respec/SKILL.md"]
	if !ok {
		t.Fatal("respec/SKILL.md not rendered")
	}
	if !strings.Contains(shared, "/tmp/my-store") {
		t.Error("store path not substituted into respec skill")
	}
	if !strings.Contains(shared, "SHARED-CONTEXT-MARKER") {
		t.Error("context not injected into respec skill")
	}

	// Per-phase rules stay with their phase skill.
	research, ok := r.Skills["rsx-research/SKILL.md"]
	if !ok {
		t.Fatal("rsx-research/SKILL.md not rendered")
	}
	if !strings.Contains(research, "RESEARCH-RULE-MARKER") {
		t.Error("research rule not injected into research skill")
	}

	// No unexpanded template directives should remain anywhere.
	for name, content := range r.Skills {
		if strings.Contains(content, "{{") {
			t.Errorf("skill %s still has unexpanded {{ }}", name)
		}
	}
}

func TestRenderOmitsEmptyContextAndRules(t *testing.T) {
	cfg := config.Defaults()
	cfg.Store = "/tmp/s"
	r, err := Render(cfg, TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(r.Skills["respec/SKILL.md"], "Shared context") {
		t.Error("empty context should not emit a Shared context section")
	}
}

func TestRenderLayersOverrideOverEmbedded(t *testing.T) {
	dir := t.TempDir()
	// Override just the plan skill and the shared skill, plus an override-only skill.
	mustWrite(t, filepath.Join(dir, "skills", "rsx-plan", "SKILL.md"), "store={{.Store}} PLAN-OVERRIDE\n")
	mustWrite(t, filepath.Join(dir, "skills", "custom", "SKILL.md"), "CUSTOM-ONLY {{.Store}}\n")
	mustWrite(t, filepath.Join(dir, "skills", "respec", "SKILL.md"), "SKILL-OVERRIDE {{.Store}}\n")

	cfg := config.Defaults()
	cfg.Store = "/tmp/over"
	cfg.TemplatesDir = dir

	r, err := Render(cfg, TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(r.Skills["rsx-plan/SKILL.md"], "PLAN-OVERRIDE") {
		t.Errorf("rsx-plan override not applied: %q", r.Skills["rsx-plan/SKILL.md"])
	}
	if !strings.Contains(r.Skills["custom/SKILL.md"], "CUSTOM-ONLY") {
		t.Error("override-only custom skill not rendered")
	}
	if _, ok := r.Skills["rsx-research/SKILL.md"]; !ok {
		t.Error("embedded rsx-research should still render under layering")
	}
	if _, ok := r.Skills["rsx-implement/SKILL.md"]; !ok {
		t.Error("embedded rsx-implement should still render under layering")
	}
	if !strings.Contains(r.Skills["respec/SKILL.md"], "SKILL-OVERRIDE") {
		t.Errorf("respec skill override not applied: %q", r.Skills["respec/SKILL.md"])
	}
}

func TestResolveReportsSource(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "skills", "rsx-plan", "SKILL.md"), "x\n")
	cfg := config.Defaults()
	cfg.TemplatesDir = dir

	files, err := Resolve(cfg)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	sources := map[string]string{}
	for _, f := range files {
		sources[f.Name] = f.Source
	}
	if got := sources["skills/rsx-plan/SKILL.md"]; got != dir {
		t.Errorf("rsx-plan source = %q, want override dir %q", got, dir)
	}
	if got := sources["skills/rsx-research/SKILL.md"]; got != EmbeddedSource {
		t.Errorf("rsx-research source = %q, want %q", got, EmbeddedSource)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveErrorsOnMissingTemplatesDir(t *testing.T) {
	cfg := config.Defaults()
	cfg.TemplatesDir = filepath.Join(t.TempDir(), "no-such-dir")
	if _, err := Resolve(cfg); err == nil {
		t.Fatal("expected error for nonexistent templates_dir")
	}
	if _, err := Render(cfg, TargetPi, Features{}); err == nil {
		t.Fatal("Render should propagate the missing-dir error")
	}
}

func TestRenderPerTarget(t *testing.T) {
	cfg := config.Defaults()
	cfg.Store = "/tmp/s"

	pi, err := Render(cfg, TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render pi: %v", err)
	}
	claude, err := Render(cfg, TargetClaude, Features{})
	if err != nil {
		t.Fatalf("Render claude: %v", err)
	}
	codex, err := Render(cfg, TargetCodex, Features{})
	if err != nil {
		t.Fatalf("Render codex: %v", err)
	}
	primeAgent, err := Render(cfg, TargetPrimeAgent, Features{})
	if err != nil {
		t.Fatalf("Render prime-agent: %v", err)
	}

	// All four targets render the same four skills.
	for _, r := range []Rendered{pi, claude, codex, primeAgent} {
		for _, key := range []string{"respec/SKILL.md", "rsx-research/SKILL.md", "rsx-plan/SKILL.md", "rsx-implement/SKILL.md"} {
			if _, ok := r.Skills[key]; !ok {
				t.Errorf("missing skill %s", key)
			}
		}
	}

	// Skills take their argument as a trailing User: message, not $@ / $1.
	for name, content := range pi.Skills {
		if strings.Contains(content, "$@") || strings.Contains(content, "${1") {
			t.Errorf("pi skill %s still uses an argument placeholder", name)
		}
	}
	if !strings.Contains(pi.Skills["rsx-research/SKILL.md"], "trailing `User:` message") {
		t.Error("pi research skill should describe the trailing User: argument")
	}

	// pi references the pi skill path and /skill: invocation.
	if !strings.Contains(pi.Skills["rsx-research/SKILL.md"], "~/.pi/agent/skills/respec/SKILL.md") {
		t.Error("pi research skill should reference the pi respec skill path")
	}
	if !strings.Contains(pi.Skills["rsx-implement/SKILL.md"], "/skill:rsx-plan") {
		t.Error("pi implement skill should reference /skill:rsx-plan")
	}

	// Claude Code resolves skills as /name and keeps no pi-isms.
	for name, content := range claude.Skills {
		for _, forbidden := range []string{"$@", "${1", "/rsx:", "~/.pi/", "/skill:"} {
			if strings.Contains(content, forbidden) {
				t.Errorf("claude skill %s contains foreign-agent syntax %q", name, forbidden)
			}
		}
	}
	if !strings.Contains(claude.Skills["rsx-research/SKILL.md"], "~/.claude/skills/respec/SKILL.md") {
		t.Error("claude research skill should reference the claude respec skill path")
	}
	if !strings.Contains(claude.Skills["rsx-implement/SKILL.md"], "/rsx-plan") {
		t.Error("claude implement skill should reference /rsx-plan")
	}

	// Codex mentions skills as $name and keeps its own skill path.
	for name, content := range codex.Skills {
		for _, forbidden := range []string{"$@", "${1:", "/rsx:", "~/.pi/", "~/.claude/", "/skill:"} {
			if strings.Contains(content, forbidden) {
				t.Errorf("codex skill %s contains foreign-agent syntax %q", name, forbidden)
			}
		}
	}
	if !strings.Contains(codex.Skills["rsx-research/SKILL.md"], "${CODEX_HOME:-$HOME/.codex}/skills/respec/SKILL.md") {
		t.Error("codex research skill should reference the Codex respec skill path")
	}
	if !strings.Contains(codex.Skills["rsx-implement/SKILL.md"], "$rsx-plan") {
		t.Error("codex implement skill should reference $rsx-plan")
	}
	if meta, ok := codex.Skills["respec/agents/openai.yaml"]; !ok || !strings.Contains(meta, "$respec") {
		t.Error("the skill bundle should ship Codex's agents/openai.yaml metadata")
	}

	// Prime Agent uses /skill: invocation and its own skill path.
	for name, content := range primeAgent.Skills {
		for _, forbidden := range []string{"${1", "/rsx:", "~/.pi/", "~/.claude/", "${CODEX_HOME"} {
			if strings.Contains(content, forbidden) {
				t.Errorf("prime-agent skill %s contains foreign-agent syntax %q", name, forbidden)
			}
		}
	}
	if !strings.Contains(primeAgent.Skills["rsx-research/SKILL.md"], "~/.prime/agent/skills/respec/SKILL.md") {
		t.Error("prime-agent research skill should reference the prime-agent respec skill path")
	}
	if !strings.Contains(primeAgent.Skills["rsx-implement/SKILL.md"], "/skill:rsx-plan") {
		t.Error("prime-agent implement skill should reference /skill:rsx-plan")
	}
}

func TestRenderRejectsUnknownTarget(t *testing.T) {
	if _, err := Render(config.Defaults(), Target{Name: "emacs"}, Features{}); err == nil {
		t.Fatal("expected error for unknown target")
	}
}

// The child harness is chosen per phase in the plan, not frozen when the
// operator ran `respec install`. If a rendered skill hardcoded a child
// invocation again, a Claude-installed orchestrator could never delegate to pi
// (and vice versa), which is the whole point of the per-phase matrix.
func TestImplementDoesNotHardcodeChildHarness(t *testing.T) {
	cfg := config.Defaults()
	cfg.Store = "/tmp/s"

	for _, target := range []Target{TargetPi, TargetPrimeAgent, TargetClaude, TargetCodex} {
		r, err := Render(cfg, target, Features{})
		if err != nil {
			t.Fatalf("Render %s: %v", target.Name, err)
		}
		got := r.Skills["rsx-implement/SKILL.md"]
		for _, frozen := range []string{
			"pi --print --no-session --thinking",
			"prime-agent --print --no-session --thinking",
			"claude --print --no-session-persistence",
			"codex exec --model",
		} {
			if strings.Contains(got, frozen) {
				t.Errorf("%s implement skill hardcodes a child invocation %q; it must come from the plan's **Agent:** line via respec agent-cmd", target.Name, frozen)
			}
		}
		if !strings.Contains(got, "**Agent:**") {
			t.Errorf("%s implement skill should read the phase's **Agent:** line", target.Name)
		}
	}
}

// One implement skill serves both modes, reading execution_mode from the plan
// rather than being chosen at invocation.
func TestImplementHandlesBothExecutionModes(t *testing.T) {
	cfg := config.Defaults()
	cfg.Store = "/tmp/s"
	r, err := Render(cfg, TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	got := r.Skills["rsx-implement/SKILL.md"]
	for _, want := range []string{"execution_mode", "manual", "auto", "Adversarial review"} {
		if !strings.Contains(got, want) {
			t.Errorf("implement skill should cover %q", want)
		}
	}
}

// Commits must not pile onto the default branch.
func TestImplementWorksOnAnEffortBranch(t *testing.T) {
	cfg := config.Defaults()
	cfg.Store = "/tmp/s"
	r, err := Render(cfg, TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	for _, want := range []string{"default branch", "respec/<change-dir basename>", "stay"} {
		if !strings.Contains(r.Skills["rsx-implement/SKILL.md"], want) {
			t.Errorf("implement skill should cover %q", want)
		}
	}
}

// Absent preferences must not produce invented staffing.
func TestPlanTellsAgentToInventNothingWithoutPreferences(t *testing.T) {
	cfg := config.Defaults()
	cfg.Store = "/tmp/s"

	bare, err := Render(cfg, TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render without preferences: %v", err)
	}
	if !strings.Contains(bare.Skills["rsx-plan/SKILL.md"], "do not invent any") {
		t.Error("with no preferences, plan skill must tell the agent not to invent staffing")
	}

	cfg.Agents = config.Agents{
		Implementer: "pi:openai-codex/gpt-5.6-sol@medium",
		Reviewer:    "claude:opus",
	}
	stated, err := Render(cfg, TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render with preferences: %v", err)
	}
	for _, want := range []string{"pi:openai-codex/gpt-5.6-sol@medium", "claude:opus"} {
		if !strings.Contains(stated.Skills["rsx-plan/SKILL.md"], want) {
			t.Errorf("stated preference %q missing from plan skill", want)
		}
	}
	if strings.Contains(stated.Skills["rsx-plan/SKILL.md"], "do not invent any") {
		t.Error("with preferences stated, the no-preferences branch must not render")
	}
}

func TestRenderGatesProbeGuidance(t *testing.T) {
	cfg := config.Defaults()

	with, err := Render(cfg, TargetPi, Features{Probe: true})
	if err != nil {
		t.Fatalf("Render with probe: %v", err)
	}
	if !strings.Contains(with.Skills["rsx-research/SKILL.md"], "probe extract") {
		t.Error("probe guidance missing from research skill when probe is available")
	}

	without, err := Render(cfg, TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render without probe: %v", err)
	}
	for name, content := range without.Skills {
		if strings.Contains(content, "probe") {
			t.Errorf("skill %s references probe although it is not installed", name)
		}
	}
}

// All four skills are user-invoked (explicit entry points): the shared respec
// skill is read by path, and the phase skills load only when the operator
// invokes them. Nothing pays permanent context load.
func TestSkillsAreUserInvoked(t *testing.T) {
	r, err := Render(config.Defaults(), TargetPi, Features{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	for _, key := range []string{"respec/SKILL.md", "rsx-research/SKILL.md", "rsx-plan/SKILL.md", "rsx-implement/SKILL.md"} {
		if !strings.Contains(r.Skills[key], "disable-model-invocation: true") {
			t.Errorf("%s should set disable-model-invocation: true", key)
		}
	}
}
