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

Your job this phase:

1. Read `research.md` in the target change directory — its **Decisions** are settled constraints;
   do not re-litigate them or re-run the research.
2. Close the remaining gaps before committing to an approach: ask only what inspection cannot
   answer, and when the operator corrects your understanding, verify the correction against the
   code before building on it. No silent assumptions — resolve by inspection or by asking; the
   final plan carries no unresolved questions. Identify the highest meaningful end-to-end
   feedback loop available (existing E2E/integration suite, CLI smoke test, rendered output, or a
   real request through the public boundary). Ask early if credentials, services, hardware, or
   setup are required so verification is not discovered to be impossible during implementation.{{if .HasProbe}}
   `probe` is installed — prefer it over plain grep-and-read for inspection; it returns whole
   semantic blocks (functions, classes) ranked by relevance, which keeps context small:

       probe search "<terms>" <path>    # Elasticsearch-style query: AND / OR / NOT
       probe extract <file>#<symbol>    # pull one function/class by name
       probe extract <file>:<line>      # pull the block containing a line
{{end}}
3. Propose the phase outline first — each phase's name and what it accomplishes — and confirm it
   with the operator before writing the detailed plan.
4. **Co-generate** `spec.md` and `plan.md` together, sharing one deliverables list:
   - `spec.md` — the requirements / desired behavior.
   - `plan.md` — the phased implementation.
5. Revisiting an existing plan? Amend **surgically** — never regenerate from scratch — and
   preserve completed checkboxes unless the change invalidates them (call that out first):
   - `respec status <change-dir>` reports **stale** → the spec changed; amend the plan to match.
   - Operator feedback while the spec is unchanged → scope or behavior feedback belongs in
     `spec.md` first (staleness then forces the re-plan); approach-only feedback is a direct
     surgical plan edit.
   - Preserve its `execution_mode`. An auto plan remains auto unless the operator explicitly asks
     to switch it; keep its unattended phase/verification/commit contract and re-check that edited
     phases still have a runnable end-to-end path. A new normal plan uses `execution_mode: manual`.
6. After writing both files, stamp the plan (records the spec hash, reflows prose), then lint and
   fix any findings:

       respec stamp <change-dir>
       respec lint <change-dir>

   If this session is not running inside the worked-on repo (`repo_path` in `research.md` names
   it), pass `respec stamp --repo <worked-on-repo> <change-dir>` — stamp refuses to run against
   the store itself.

Frontmatter you write: `spec.md` → `title`, `status` (`draft`|`approved`), optional `tags`;
`plan.md` → `title`, `status` (`draft`|`approved`|`in-progress`|`done`), optional
`execution_mode` (`manual`|`auto`), and optional `depends_on` (efforts that must finish first — a
bare slug for the same repo, or `repo/slug` cross-repo).
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
    <the changes: which files, and what changes>

    ### Automated Verification
    - [ ] tests pass: `<command>`
    - [ ] lint passes: `<command>`

    ### Manual Verification
    - [ ] Orchestrator: <behavior or diff property the agent can inspect>
    - [ ] Operator: <irreducibly human/external check, only when necessary>

    ## Testing Strategy
    <unit / integration / end-to-end coverage, runnable commands, and key edge cases>

Split verification deliberately: **Automated** = deterministic commands; **Manual** = judgment.
The orchestrating agent must perform every feasible Manual item itself and tick it when verified;
reserve `Operator:` items for checks the agent truly cannot perform. Add phases as needed. A plan
must include a runnable end-to-end, integration, or smoke feedback loop at the highest practical
boundary, not just unit tests. In manual mode, pause only for remaining `Operator:` items. In auto
mode, move prerequisites early and avoid intermediate operator gates.

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

End by summarizing the spec and plan paths, the phases, and the risks the implementer should
watch — and remind the operator to review both artifacts before starting a fresh implement
session.
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
