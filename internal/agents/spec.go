package agents

import (
	"fmt"
	"strings"
)

// Spec names one delegation target: which harness runs the task, optionally
// which model it uses, and optionally how hard that model is asked to think.
//
// Its text form is `<harness>[:<model>][@<effort>]`, e.g.
//
//	pi                                  — harness defaults for both
//	pi@high                             — harness default model, pinned effort
//	pi:openai-codex/gpt-5.6-sol         — pinned model, harness default effort
//	pi:openai-codex/gpt-5.6-sol@medium  — both pinned
//	claude:opus
//
// Model and effort are optional because the common case is "just use the
// harness": an operator who has stated no preference should get whatever their
// harness is already configured to do, not a model respec chose for them.
//
// Model is left opaque on purpose — it may contain slashes (openrouter ids
// like `anthropic/claude-opus-5`) and respec never validates it against a
// catalogue, because the catalogue moves and a plan outlives it.
type Spec struct {
	Harness Harness
	Model   string // empty = harness default
	Effort  string // empty = harness default
}

// ParseSpec reads the `<harness>[:<model>][@<effort>]` form. The effort is
// taken from the last `@` and the harness from the first `:`, so slashes and
// colons remain legal inside a model id.
func ParseSpec(s string) (Spec, error) {
	var out Spec

	raw := strings.TrimSpace(s)
	if raw == "" {
		return out, fmt.Errorf("empty agent spec")
	}

	rest := raw
	if i := strings.LastIndex(rest, "@"); i >= 0 {
		out.Effort = strings.TrimSpace(rest[i+1:])
		rest = strings.TrimSpace(rest[:i])
		if out.Effort == "" {
			return out, fmt.Errorf("agent spec %q: trailing @ with no effort", raw)
		}
	}

	h, model, hasModel := strings.Cut(rest, ":")
	out.Harness = Harness(strings.TrimSpace(h))
	if !Known(out.Harness) {
		return out, fmt.Errorf("agent spec %q: unknown harness %q (known: %s)", raw, out.Harness, harnessList())
	}
	if hasModel {
		out.Model = strings.TrimSpace(model)
		if out.Model == "" {
			return out, fmt.Errorf("agent spec %q: trailing : with no model", raw)
		}
	}

	if out.Effort != "" && !validEffort(out.Harness, out.Effort) {
		return out, fmt.Errorf("agent spec %q: %s does not accept effort %q (one of: %s)",
			raw, out.Harness, out.Effort, strings.Join(efforts[out.Harness], ", "))
	}
	return out, nil
}

// String renders the canonical text form, omitting whatever was left default.
func (s Spec) String() string {
	out := string(s.Harness)
	if s.Model != "" {
		out += ":" + s.Model
	}
	if s.Effort != "" {
		out += "@" + s.Effort
	}
	return out
}

// IsHarnessDefault reports whether this spec pins nothing, meaning the child
// runs exactly as the harness is already configured.
func (s Spec) IsHarnessDefault() bool { return s.Model == "" && s.Effort == "" }

func validEffort(h Harness, effort string) bool {
	for _, e := range efforts[h] {
		if e == effort {
			return true
		}
	}
	return false
}

func harnessList() string {
	names := make([]string, 0, len(Harnesses))
	for _, h := range Harnesses {
		names = append(names, string(h))
	}
	return strings.Join(names, ", ")
}

// Command builds the argv for a one-shot, non-interactive run of this spec.
//
// The delegated prompt is always fed on **stdin**, never as an argument: phase
// prompts are long and contain quotes, backticks, and newlines, and every
// shell-quoting bug in a delegation loop traces back to trying to pass one as
// an argument. Callers pipe the prompt into the returned command.
func (s Spec) Command() ([]string, error) {
	if !Known(s.Harness) {
		return nil, fmt.Errorf("unknown harness %q", s.Harness)
	}
	if s.Effort != "" && !validEffort(s.Harness, s.Effort) {
		return nil, fmt.Errorf("%s does not accept effort %q (one of: %s)",
			s.Harness, s.Effort, strings.Join(efforts[s.Harness], ", "))
	}

	// An unset model or effort is omitted rather than defaulted here: the
	// harness already knows what the operator configured, and guessing on its
	// behalf would override a preference respec never saw.
	switch s.Harness {
	case HarnessPi:
		argv := []string{"pi", "--print", "--no-session"}
		if s.Model != "" {
			argv = append(argv, "--model", s.Model)
		}
		if s.Effort != "" {
			argv = append(argv, "--thinking", s.Effort)
		}
		return argv, nil
	case HarnessClaude:
		// CLAUDECODE is unset deliberately: Claude Code sets it in the parent
		// and otherwise refuses to start a nested CLI session.
		argv := []string{"env", "-u", "CLAUDECODE", "claude", "--print", "--no-session-persistence"}
		if s.Model != "" {
			argv = append(argv, "--model", s.Model)
		}
		if s.Effort != "" {
			argv = append(argv, "--effort", s.Effort)
		}
		return argv, nil
	case HarnessCodex:
		argv := []string{"codex", "exec"}
		if s.Model != "" {
			argv = append(argv, "--model", s.Model)
		}
		if s.Effort != "" {
			argv = append(argv, "-c", fmt.Sprintf("model_reasoning_effort=%q", s.Effort))
		}
		return argv, nil
	}
	return nil, fmt.Errorf("unknown harness %q", s.Harness)
}

// Shell renders the command as a copy-pasteable line reading its prompt from
// promptFile. Arguments are single-quoted so a model id or effort can never be
// re-interpreted by the shell.
func (s Spec) Shell(promptFile string) (string, error) {
	argv, err := s.Command()
	if err != nil {
		return "", err
	}
	quoted := make([]string, 0, len(argv))
	for _, a := range argv {
		quoted = append(quoted, shellQuote(a))
	}
	return strings.Join(quoted, " ") + " < " + shellQuote(promptFile), nil
}

// shellQuote wraps s in single quotes, escaping any embedded single quote.
func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " \t\n'\"\\$`&|;<>()*?[]{}!#~=") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
