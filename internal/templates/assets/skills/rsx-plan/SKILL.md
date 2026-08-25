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
{{if .HasProbe}}
`probe` is installed. Use it to verify that a symbol or signature the plan names still exists
without reading an unrelated whole file:

    probe extract <file>#<symbol>    # extract one function or class by name
    probe extract <file>:<line>      # extract the block containing a line
{{end}}
Delegate broad sweeps such as caller audits and package-wide pattern surveys under the shared
policy. Keep design judgment and operator conversation in this session.

Find the highest meaningful end-to-end feedback loop available: an E2E or integration suite, CLI
smoke test, rendered output, or real request through the public boundary. Ask early whether it
needs credentials, services, hardware, or setup. An auto plan must name a runnable E2E,
integration, or smoke command.

In manual mode, propose each phase's name and outcome, then confirm the outline with the operator
before adding detail. In auto mode, decide the outline yourself.

## Staff the phases

{{if .Agents.Stated -}}
The operator has stated these delegation preferences:
{{if .Agents.Implementer}}
- implementer: `{{.Agents.Implementer}}`
{{- end}}{{if .Agents.Reviewer}}
- reviewer: `{{.Agents.Reviewer}}`
{{- end}}{{if .Agents.Committer}}
- committer: `{{.Agents.Committer}}`
{{- end}}{{if .Agents.Notes}}
- notes: {{.Agents.Notes}}
{{- end}}

Treat these as the operator's standing stance, not as a per-phase decision. Apply them where they
fit and depart only with a stated reason — a phase whose difficulty clearly warrants something
else, or a preference naming a model this machine cannot reach.
{{- else -}}
The operator has stated no delegation preferences. Unless they say otherwise in this session,
**do not invent any** — write no `**Agent:**` lines and let each harness do what it is already
configured to do. That is a legitimate, complete plan and keeps plans portable between machines.
Record staffing only if the operator asks for it here.
{{- end}}

When staffing is in play, check what this machine can actually reach before naming anything —
`respec agents` (add `--filter <substr>` on a large catalogue). A preference naming a model that is
not installed here is worth telling the operator about rather than silently substituting.

Record each choice and why it fits the phase:

    ## Phase 2: <name>
    **Agent:** pi:openai-codex/gpt-5.6-sol@medium — mechanical work; a cheaper model is enough

The reason lets an implementer substitute an equivalent if the named model later becomes
unreachable. Model and effort are optional. In manual mode, recommend the delegation matrix beside
the phase outline and let the operator settle it. In auto mode, resolve it yourself and ask only
when no reachable choice fits. Prefer a reviewer from a different model family than the
implementer.

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

    ## Phase 1: <name>
    **Agent:** <spec> — <why this one>   <!-- omit when the operator stated no preference -->
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
