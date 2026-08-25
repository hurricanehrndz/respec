package agents

import (
	"strings"
	"testing"
)

func TestParseSpecRoundTrips(t *testing.T) {
	// Model ids carry slashes (openrouter) and dots; parsing must not mangle
	// them, because a mangled id fails only once a phase is already running.
	for _, raw := range []string{
		"pi:openai-codex/gpt-5.6-sol@medium",
		"pi:openrouter/anthropic/claude-opus-5@max",
		"claude:opus@high",
		"codex:gpt-5.5@high",
		"prime-agent:openai-codex/gpt-5.6-sol@medium",
		// Model and effort are independently optional: an operator who stated
		// no preference must get harness defaults, not a respec-chosen model.
		"pi",
		"pi@high",
		"pi:openai-codex/gpt-5.6-sol",
		"claude:opus",
		"prime-agent",
	} {
		s, err := ParseSpec(raw)
		if err != nil {
			t.Fatalf("ParseSpec(%q): %v", raw, err)
		}
		if s.String() != raw {
			t.Errorf("round trip: got %q, want %q", s.String(), raw)
		}
	}
}

func TestParseSpecRejectsBadInput(t *testing.T) {
	cases := map[string]string{
		"":                  "empty",
		"emacs:opus@high":   "unknown harness",
		"pi:@medium":        "trailing : with no model",
		"pi:opus@":          "trailing @ with no effort",
		"claude:opus@off":   "does not accept effort",
		"codex:gpt-5.5@max": "does not accept effort",
	}
	for raw, want := range cases {
		_, err := ParseSpec(raw)
		if err == nil {
			t.Errorf("ParseSpec(%q): expected error", raw)
			continue
		}
		if !strings.Contains(err.Error(), want) {
			t.Errorf("ParseSpec(%q): error %q does not mention %q", raw, err, want)
		}
	}
}

// Effort levels differ per harness; "off" is pi-only and "max" is not a codex
// level. Accepting one the CLI rejects would fail mid-phase, so it is caught
// at parse time.
func TestEffortIsValidatedPerHarness(t *testing.T) {
	if _, err := ParseSpec("pi:openai-codex/gpt-5.4@off"); err != nil {
		t.Errorf("pi should accept 'off': %v", err)
	}
	if _, err := ParseSpec("claude:opus@minimal"); err == nil {
		t.Error("claude should reject 'minimal'")
	}
}

func TestCommandPerHarness(t *testing.T) {
	cases := []struct {
		spec string
		want []string
	}{
		{"pi:openai-codex/gpt-5.6-sol@medium", []string{
			"pi", "--print", "--no-session", "--model", "openai-codex/gpt-5.6-sol", "--thinking", "medium"}},
		{"claude:opus@high", []string{
			"env", "-u", "CLAUDECODE", "claude", "--print", "--no-session-persistence",
			"--model", "opus", "--effort", "high"}},
		{"codex:gpt-5.5@high", []string{
			"codex", "exec", "--model", "gpt-5.5", "-c", `model_reasoning_effort="high"`}},
		{"prime-agent:openai-codex/gpt-5.6-sol@medium", []string{
			"prime-agent", "--print", "--no-session", "--model", "openai-codex/gpt-5.6-sol", "--thinking", "medium"}},
		// Unpinned fields are omitted, not defaulted: the harness already knows
		// what the operator configured, and guessing would override it.
		{"pi", []string{"pi", "--print", "--no-session"}},
		{"pi@low", []string{"pi", "--print", "--no-session", "--thinking", "low"}},
		{"claude:opus", []string{
			"env", "-u", "CLAUDECODE", "claude", "--print", "--no-session-persistence", "--model", "opus"}},
		{"codex", []string{"codex", "exec"}},
		{"prime-agent", []string{"prime-agent", "--print", "--no-session"}},
	}
	for _, c := range cases {
		s, err := ParseSpec(c.spec)
		if err != nil {
			t.Fatalf("ParseSpec(%q): %v", c.spec, err)
		}
		got, err := s.Command()
		if err != nil {
			t.Fatalf("Command(%q): %v", c.spec, err)
		}
		if strings.Join(got, " ") != strings.Join(c.want, " ") {
			t.Errorf("Command(%q):\n got %q\nwant %q", c.spec, got, c.want)
		}
	}
}

// A nested claude session refuses to start unless CLAUDECODE is cleared, so
// the wrapper must survive any refactor of the argv builder.
func TestClaudeCommandClearsClaudecode(t *testing.T) {
	s, err := ParseSpec("claude:sonnet@low")
	if err != nil {
		t.Fatal(err)
	}
	argv, err := s.Command()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(argv, " ") != "env -u CLAUDECODE claude --print --no-session-persistence --model sonnet --effort low" {
		t.Errorf("claude argv lost its CLAUDECODE guard: %q", argv)
	}
}

// The prompt is piped, never interpolated: a phase task containing quotes must
// not be able to break out of the command line.
func TestShellQuotesPromptPath(t *testing.T) {
	s, err := ParseSpec("pi:openai-codex/gpt-5.4@low")
	if err != nil {
		t.Fatal(err)
	}
	line, err := s.Shell("/tmp/it's here/phase 1.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(line, `< '/tmp/it'\''s here/phase 1.md'`) {
		t.Errorf("prompt path not safely quoted: %s", line)
	}
	if !strings.Contains(line, "pi --print --no-session") {
		t.Errorf("unexpected command: %s", line)
	}
}

func TestEffortsExposedPerHarness(t *testing.T) {
	if len(Efforts(HarnessPi)) == 0 || len(Efforts(HarnessPrimeAgent)) == 0 || len(Efforts(HarnessClaude)) == 0 || len(Efforts(HarnessCodex)) == 0 {
		t.Fatal("every known harness must publish its effort levels")
	}
	if Known("emacs") {
		t.Error("emacs is not a known harness")
	}
}
