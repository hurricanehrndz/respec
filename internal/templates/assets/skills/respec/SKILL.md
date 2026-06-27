---
name: respec
description: The respec spec-driven workflow (research → plan → implement) over a single central store. Use when running /rsx:research, /rsx:plan, or /rsx:implement, or when creating/editing research.md, spec.md, or plan.md artifacts, or when calling the respec CLI (stamp, status, format, render, serve).
---

# respec workflow

respec drives a `research → plan → implement` workflow. **All artifacts live in one central
store**, never in the worked-on repo:

    {{.Store}}

The CLI does deterministic plumbing only. Reasoning (research, planning, implementation) is your
job; the CLI handles config, install, staleness stamping, prose reflow, and site rendering.

## Store layout

    {{.Store}}/<problem-space>/<YYYY-MM-DD-slug>/
        research.md   # provenance-stamped findings
        spec.md       # requirements / desired behavior
        plan.md       # phased implementation + checkboxes; frontmatter holds spec_sha256

A `<problem-space>/<YYYY-MM-DD-slug>/` directory is one "change". `spec.md` and `plan.md` are
siblings.

## Phases

### Research (`/rsx:research <topic>`)
Clarify the topic, explore the repo, then write `research.md` into a new change directory with
frontmatter: `date`, `repo`, `git_commit`, `topic`. Nothing is written to the worked-on repo.

### Plan (`/rsx:plan [change-dir]`)
Read `research.md`. Co-generate `spec.md` + `plan.md` (shared deliverables list, expressed as
prose). Ask clarifying questions; offer options. If a plan already exists and is **stale**, amend
it surgically rather than regenerating. After writing both files, stamp the plan:

    respec stamp <change-dir>

### Implement (`/rsx:implement <change-dir>`)
First gate on staleness:

    respec status <change-dir> --json

- `stale`   → the spec changed after stamping; STOP and re-plan.
- `unstamped` → STOP and run `/rsx:plan` to stamp.
- `fresh`   → proceed: execute phases in order, ticking `plan.md` checkboxes as verification passes.

## Staleness model

`respec stamp` records the sha256 of `spec.md` into `plan.md` frontmatter (`spec_sha256`).
`respec status` recomputes and compares. The relationship is one-directional: editing `spec.md`
marks the plan stale; editing `plan.md` does not. This is intentional — the plan must always track
the current spec.

## CLI reference

    respec config get|set|path        # configuration (~/.config/respec/config.yaml)
    respec install                    # (re)install these prompts + this skill at user scope
    respec stamp <change-dir>         # record spec.md's sha256 into plan.md frontmatter
    respec status <change-dir> [--json]  # fresh | stale | unstamped
    respec format <file>              # reflow prose only; non-prose left byte-identical
    respec render [--out <dir>]       # build the store as a browsable Hugo site
    respec serve                      # serve the store with live reload

## Hard rules

- Never write research/spec/plan artifacts into the worked-on repo — only into the central store.
- The plan is the source of truth during implementation.
- Keep `spec.md` and `plan.md` consistent; re-stamp after any spec change.
{{if .Context}}
## Shared context

{{.Context}}
{{end}}