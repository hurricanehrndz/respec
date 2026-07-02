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

1. Read `research.md` in the target change directory — its **Decisions** are settled constraints.
2. Ask any remaining clarifying questions and offer options before committing to an approach.
3. **Co-generate** `spec.md` and `plan.md` together, sharing one deliverables list:
   - `spec.md` — the requirements / desired behavior.
   - `plan.md` — the phased implementation.
4. If a `plan.md` already exists and `respec status <change-dir>` reports **stale**, amend the plan
   **surgically** to match the changed spec rather than regenerating it from scratch.
5. After writing both files, stamp the plan (records the spec hash, reflows prose):

       respec stamp <change-dir>

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
    <what we are implementing and why>

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

Write only into the central store change directory. Nothing goes into the worked-on repo.
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
