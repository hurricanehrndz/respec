---
description: Co-generate spec.md + plan.md from research, then stamp the plan
argument-hint: "[change-dir]"
---
You are in the **Plan** phase of the respec workflow (research → plan → implement).

Change directory (defaults to the most recent under the store if omitted): {{.Arg1Or "<pick latest>"}}

Central store:

    {{.Store}}

Load the detailed workflow first: read `{{.SkillPath}}` (or `{{.SkillCmd}}`).

Your job this phase:

1. Read `research.md` in the target change directory — its **Decisions** are settled constraints;
   do not re-litigate them or re-run the research.
2. Close the remaining gaps before committing to an approach: ask only what inspection cannot
   answer, and when the operator corrects your understanding, verify the correction against the
   code before building on it. No silent assumptions — resolve by inspection or by asking; the
   final plan carries no unresolved questions.
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
6. After writing both files, stamp the plan (records the spec hash, reflows prose), then lint and
   fix any findings:

       respec stamp <change-dir>
       respec lint <change-dir>

   If this session is not running inside the worked-on repo (`repo_path` in `research.md` names
   it), pass `respec stamp --repo <worked-on-repo> <change-dir>` — stamp refuses to run against
   the store itself.

Frontmatter you write: `spec.md` → `title`, `status` (`draft`|`approved`), optional `tags`;
`plan.md` → `title`, `status` (`draft`|`approved`|`in-progress`|`done`), and optional `depends_on`
(efforts that must finish first — a bare slug for the same repo, or `repo/slug` cross-repo).
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
    - [ ] <behavior to confirm by hand>

    ## Testing Strategy
    <unit / integration coverage and key edge cases>

Split verification deliberately: **Automated** = commands the implementer can run and tick off
itself; **Manual** = judgment calls that gate the phase and need the operator. Add more phases as
needed; each phase pauses at its Manual checks before the next begins.

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
