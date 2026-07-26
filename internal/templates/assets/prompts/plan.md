---
description: Co-generate spec.md + plan.md from research, then stamp the plan
argument-hint: "[change-dir]"
---
{{- /* Workflow mechanics here (frontmatter split, stamp/lint sequence, lint-validated headings) are
deliberately duplicated in skills/respec/SKILL.md so each file stands alone — edit both together. */}}
You are in the **Plan** phase of the respec workflow (research → plan → implement).

Change directory (defaults to the most recent under the store if omitted): {{.Arg1Or "<pick latest>"}}

If omitted, pick the latest deterministically: run `respec list` — rows sort most-recent-first
within each repo — and take the top effort for the worked-on repo. If that pick is ambiguous
(no clear top row, or the session is not inside a worked-on repo), ask the operator instead of
guessing.

Central store:

    {{.Store}}

Load the detailed workflow first: read `{{.SkillPath}}` (or `{{.SkillCmd}}`).

## Settle the execution mode first

Ask the operator one question before planning: **will this plan run unattended, or will you be
looped in at each phase?** The answer sets `execution_mode` and changes how you work from here.

- **`manual`** — the operator is looped in. Confirm the phase outline and the delegation matrix
  with them before writing detail, and let phases carry `Operator:` checks wherever a human
  genuinely has to look.
- **`auto`** — the operator is not looped in until the end. Gather every material question now,
  ask them together, and design phases that reach the end without stopping. `respec lint` rejects
  an auto plan with an `Operator:` check anywhere but the final phase, so an unresolved
  prerequisite must be settled during this session or the plan is not auto.

Everything below applies to both modes; where they differ it is called out.

## Plan the work

Start from `research.md` in the target change directory. Its **Decisions** are settled constraints:
do not re-litigate them or re-run the research.

Close the remaining gaps before committing to an approach. Ask only what inspection cannot answer,
and when the operator corrects your understanding, verify the correction against the code before
building on it — a plan built on an unverified correction fails during implementation, where it is
most expensive. The finished plan carries no unresolved questions.{{if .HasProbe}}
`probe` is installed — use it to pull a single definition without reading the whole file,
which is usually how you verify that a symbol or signature the plan will name still exists:

    probe extract <file>#<symbol>    # pull one function/class by name
    probe extract <file>:<line>      # pull the block containing a line
{{end}}
Planning inspection can consume the context the plan itself needs, so delegate the broad sweeps —
auditing every caller of a function you intend to change, surveying a pattern across a package —
and keep their conclusions. The skill's Delegation section covers when and how. The design
judgment and the operator conversation stay here.

Find the highest meaningful end-to-end feedback loop available — an existing E2E or integration
suite, a CLI smoke test, rendered output, or a real request through the public boundary — and ask
early if credentials, services, hardware, or setup are needed. Discovering during implementation
that verification is impossible is the failure this question prevents. An auto plan must name a
runnable E2E/smoke command; without one there is nothing to catch a silent break in an unattended
run.

In manual mode, propose the phase outline — each phase's name and what it accomplishes — and
confirm it with the operator before writing detail. In auto mode, decide it yourself.

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

Treat these as the operator's standing stance, not as a per-phase decision. Your job is judgement
in light of them: apply them where they fit, and depart only with a stated reason — a phase whose
difficulty clearly warrants something else, or a preference naming a model this machine cannot
reach.
{{- else -}}
The operator has stated no delegation preferences. Unless they say otherwise in this session,
**do not invent any** — write no `**Agent:**` lines and let each harness do what it is already
configured to do. That is a legitimate, complete plan, and it is what keeps plans portable between
machines. Record staffing only if the operator asks for it here.
{{- end}}

When staffing is in play, check what this machine can actually reach before naming anything —
`respec agents` (add `--filter <substr>` on a large catalogue). A preference naming a model that is
not installed here is worth telling the operator about rather than silently substituting.

Record the decision on the phase, with the reasoning that produced it:

    ## Phase 2: <name>
    **Agent:** pi:openai-codex/gpt-5.6-sol@medium — mechanical, cheap model is enough

The reasoning is not decoration. If that model is retired or unreachable when the plan finally
runs, the implementer can re-derive an equivalent from the intent instead of guessing. Model and
effort are both optional (`claude:opus`, `pi@high`, or bare `pi`), and omitting one means "whatever
the harness is configured to do".

In manual mode, propose the matrix alongside the phase outline so the operator settles scope and
staffing in one pass — recommend, do not decide, since which models they trust and what they will
spend is not inferable from the code. In auto mode, resolve it yourself and ask only when nothing
reachable fits a phase.

Prefer a different model family for the reviewer than for the implementer: a model reviewing its
own family's output is weakest exactly where that family is systematically wrong.

## Write the artifacts

**Co-generate** `spec.md` and `plan.md` together, sharing one deliverables list: `spec.md` holds the
requirements and desired behavior, `plan.md` the phased implementation.

Revisiting an existing plan is an amendment, not a regeneration. Edit surgically and preserve
completed checkboxes unless the change invalidates them — call that out first. Where the feedback
lands depends on what it is: `respec status <change-dir>` reporting **stale** means the spec
changed, so amend the plan to match; operator feedback about scope or behavior belongs in `spec.md`
first, and the resulting staleness then forces the re-plan; approach-only feedback is a direct
surgical plan edit. Preserve the plan's `execution_mode` unless the operator explicitly switches
it, and after any edit re-check that a runnable end-to-end path survives.

After writing both files, stamp the plan (records the spec hash, reflows prose), then lint and fix
any findings:

    respec stamp <change-dir>
    respec lint <change-dir>

If this session is not running inside the worked-on repo (`repo_path` in `research.md` names
it), pass `respec stamp --repo <worked-on-repo> <change-dir>` — stamp refuses to run against
the store itself.

Frontmatter you write: `spec.md` → `title`, `status` (`draft`|`approved`), optional `tags`;
`plan.md` → `title`, `status` (`draft`|`approved`|`in-progress`|`done`), `execution_mode`
(`manual`|`auto`), and optional `depends_on` (efforts that must finish first — a bare slug for the
same repo, or `repo/slug` cross-repo).
`respec stamp` fills `spec.md`'s `date` and `plan.md`'s `spec_sha256` — do not write those by hand.

`spec.md` skeleton:

    # <title>

    ## Summary
    <what and why>

    ## Requirements
    <the desired behavior / acceptance criteria>

    ## Out of Scope
    <what this explicitly does not cover>

`plan.md` skeleton:

    # <title> Implementation Plan

    ## Overview
    <what we are implementing, the chosen approach, and why — written so the implementer can
    resolve small mismatches in its spirit>

    ## Current State
    <what exists now, key file:line discoveries, constraints>

    ## Desired End State
    <the target, and how we will know we are there>

    ## What We're NOT Doing
    <explicit out-of-scope, to prevent scope creep>

    ## Phase 1: <name>
    **Agent:** <spec> — <why this one>   <!-- omit entirely when no preference was stated -->
    <the changes: which files, and what changes>

    ### Automated Verification
    - [ ] tests pass: `<command>`
    - [ ] lint passes: `<command>`

    ### Manual Verification
    - [ ] Orchestrator: <behavior or diff property the agent can inspect>
    - [ ] Operator: <irreducibly human check — auto plans may use this only in the final phase>

    ## Testing Strategy
    <unit / integration / end-to-end coverage, runnable commands, and key edge cases>

The verification split is load-bearing: **Automated** means a deterministic command, **Manual**
means judgment is required — not that a human must do it. The orchestrating agent performs every
feasible Manual item itself and ticks it when verified, so `Operator:` is reserved for checks an
agent truly cannot perform; every one of them is a place the workflow stops and waits for a human.
Add phases as needed. A plan needs a runnable end-to-end, integration, or smoke loop at the highest
practical boundary, not just unit tests.

`respec lint` validates the headings **Requirements** (spec), **Overview**, at least one
**Phase**, **Automated Verification**, and **Manual Verification** (plan) verbatim — keep those
names even when customizing this template.

Both artifacts are read by a **human co-engineer** who participates in the effort, not just by the
implementing agent. Mermaid diagrams (fenced ` ```mermaid ` blocks) are strongly encouraged where
they beat prose — architecture sketches in the spec, phase/dependency flow in the plan. GitHub and
`respec render`/`serve` both render them, and diagrams-as-text diff cleanly. Inline HTML is also
fine when it adds value; the store render preserves it.

Quality bar: implementation should feel mechanical. If the implementing session would need to
invent design, the plan is not specific enough.

Write only into the central store change directory. Nothing goes into the worked-on repo.

End by summarizing the spec and plan paths, the phases, the execution mode, and the risks the
implementer should watch — and remind the operator to review both artifacts before starting a
fresh implement session.
{{if .Context}}
## Shared context

{{.Context}}
{{end}}{{if .Rules.Spec}}
## Spec rules

{{.Rules.Spec}}
{{end}}{{if .Rules.Plan}}
## Plan rules

{{.Rules.Plan}}
{{end}}
