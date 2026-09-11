---
name: rsx-plan
description: "Plan or iterate on a respec change: execution mode, spec.md, phased plan.md, staffing, and verification."
disable-model-invocation: true
---

> **Read the respec skill at `{{.SkillPath}}` before doing anything else.** It contains the
> shared store layout, stamp and lint mechanics, delegation policy, and hard rules.

# Plan a change

The change directory is the argument you were invoked with, delivered as a trailing `User:`
message. If it is omitted, run `respec list`. Rows are most-recent-first within each repository, so
use the top effort for the worked-on repository. Ask instead of guessing when there is no clear top
row or the session is not inside the worked-on repository.

Read `research.md` completely. Its **Decisions** are settled constraints. Do not re-run or
re-litigate the research from `{{.Cmd "research"}}`.

## Settle execution_mode first

Before planning, ask the operator one question: will this plan run unattended, or will they be
looped in at each phase? Record the answer as `execution_mode`.

- `manual` means the operator is looped in. Confirm the phase outline and delegation matrix before
  writing detail. Phases may include `Operator:` checks when a human must inspect something.
- `auto` means the operator is not looped in until final acceptance. Gather and ask every material
  question now. Design phases that finish without stopping. `respec lint` rejects an `Operator:`
  check before the final phase.

The work is otherwise the same in both modes.

## Plan the work

Close every gap before choosing an approach. Ask only what inspection cannot answer. If the
operator corrects your understanding, verify the correction against the code before using it. The
finished plan has no unresolved questions.
Delegate broad sweeps such as caller audits and package-wide pattern surveys under the shared
policy. Keep design judgment and operator conversation in this session.

Find the highest meaningful end-to-end feedback loop available: an E2E or integration suite, CLI
smoke test, rendered output, or real request through the public boundary. Ask early whether it
needs credentials, services, hardware, or setup. An auto plan must name a runnable E2E,
integration, or smoke command.

In manual mode, propose each phase's name and outcome, then confirm the outline with the operator
before adding detail. In auto mode, decide the outline yourself.

## Staff the phases

Ask the operator which worker preferences they want for this effort: native defaults, particular
models or effort levels for implementation, review, and commits, or explicit external executors.
Do not re-ask choices already settled in this session or the research decisions. To consult
standing defaults, run `respec config print` and read `agents:`. Treat config as optional defaults,
not assignments; the operator's choices for this effort win in both execution modes.

Preserve an existing plan's worker choices unless the operator changes them. Do not replace them
with newer config values when amending the plan. If no preferences are stated, **do not invent
any**. Omit worker preferences and `**Agent:**` lines and use native defaults.

Record the chosen implementer, reviewer, committer, and fallback in `plan.md` under **Worker
preferences**, omitting roles with no preference. Put phase-specific overrides in that phase's
prose. For native workers, omit `**Agent:**`. Express model and effort hints using controls the
environment exposes; do not select a CLI merely to express those hints in an agent spec. An
external reviewer or committer is recorded as an explicit external agent spec in the role's prose.

Only for an explicit external executor choice, probe with `respec agents` before recording the
assignment. Add `--filter <substr>` on a large catalogue. This probes CLIs, not native workers.
Report an unreachable preference; do not silently choose another executor.

For a phase assigned to an external implementer, record the choice and why it fits:

    ## Phase 2: <name>
    **Agent:** pi:openai-codex/gpt-5.6-sol@medium — operator chose pi; mechanical work

The reason lets an implementer find an equivalent model on the chosen executor if needed. Model
and effort are optional. In manual mode, recommend assignments beside the phase outline and let
the operator settle them. In auto mode, resolve them within the operator's explicit executor
choices and ask when no reachable choice fits. Prefer a reviewer from a different model family
when the chosen mechanism supports it.

## Write the artifacts

Co-generate `spec.md` and `plan.md` from one deliverables list. `spec.md` defines requirements and
desired behavior. `plan.md` defines the phased implementation. Use the shared frontmatter schema.

For an existing plan, amend instead of regenerating. Edit surgically and preserve completed
checkboxes unless the change invalidates them; tell the operator before clearing one. Scope or
behavior feedback changes `spec.md` first. Approach-only feedback changes `plan.md`. A stale status
means the spec changed, so amend the plan to match. Preserve `execution_mode` unless the operator
explicitly changes it. After any edit, make sure a runnable end-to-end path remains.

Use this `spec.md` body:

    # <title>

    ## Summary
    <what and why>

    ## Requirements
    <desired behavior and acceptance criteria>

    ## Out of Scope
    <what this does not cover>

Use this `plan.md` body:

    # <title> Implementation Plan

    ## Overview
    <what we are implementing, the chosen approach, and why>

    ## Current State
    <what exists now, key file:line discoveries, and constraints>

    ## Desired End State
    <the target and how we will know it is complete>

    ## What We're NOT Doing
    <explicit exclusions that prevent scope creep>

    ## Worker preferences
    <!-- Optional: omit when none stated; omit roles without a preference. -->
    - Implementation: <native model/effort hints, or external with per-phase Agent lines>
    - Review: <native model/effort hints, or explicit external agent spec>
    - Commits: <native model/effort hints, or explicit external agent spec>
    - Fallback: <what to do if a preference cannot be honored>

    ## Phase 1: <name>
    **Agent:** <spec> — <why this one>   <!-- external implementers only; omit for native workers -->
    <files to change and the exact changes>

    ### Automated Verification
    - [ ] tests pass: `<command>`
    - [ ] lint passes: `<command>`

    ### Manual Verification
    - [ ] Orchestrator: <behavior or diff property the agent can inspect>
    - [ ] Operator: <irreducibly human check; auto plans use this only in the final phase>

    ## Testing Strategy
    <unit, integration, and end-to-end coverage; runnable commands; key edge cases>

**Automated Verification** contains deterministic commands. **Manual Verification** contains work
that needs judgment, not necessarily a person. The orchestrator performs every feasible manual
item. Reserve `Operator:` for checks an agent cannot perform because each one stops the workflow.
Every plan needs a runnable end-to-end, integration, or smoke loop at the highest practical
boundary, not only unit tests.

`respec lint` requires **Requirements** in the spec and **Overview**, at least one **Phase**,
**Automated Verification**, and **Manual Verification** in the plan. Keep those headings verbatim.

Implementation should be mechanical. If the implementing session must invent design, add the
missing decisions and detail now.

Stamp, lint, and fix every finding:

    respec stamp <change-dir>
    respec lint <change-dir>

End with both artifact paths, the phases, execution mode, and risks. Ask the operator to review
both files before starting a fresh `{{.Cmd "implement"}}` session.
{{if .Rules.Spec}}
## Spec rules

{{.Rules.Spec}}
{{end}}{{if .Rules.Plan}}
## Plan rules

{{.Rules.Plan}}
{{end}}
