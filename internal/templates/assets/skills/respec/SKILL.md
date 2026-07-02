---
name: respec
description: The respec spec-driven workflow (research → plan → implement) over a single central store. Use when running /rsx:research, /rsx:plan, or /rsx:implement, or when creating/editing research.md, spec.md, or plan.md artifacts, or when calling the respec CLI (stamp, status, lint, format, templates, render, serve).
---

# respec workflow

respec drives a `research → plan → implement` workflow. **All artifacts live in one central
store**, never in the worked-on repo:

    {{.Store}}

The CLI does deterministic plumbing only. Reasoning (research, planning, implementation) is your
job; the CLI handles config, install, provenance/staleness stamping, prose reflow, validation, and
site rendering. Never compute `date`, git metadata, or hashes by hand — `respec stamp` does it.

## Store layout

    {{.Store}}/<problem-space>/<YYYY-MM-DD-slug>/
        research.md   # interview-driven findings + decisions
        spec.md       # requirements / desired behavior
        plan.md       # phased implementation + checkboxes

A `<problem-space>/<YYYY-MM-DD-slug>/` directory is one "change". `spec.md` and `plan.md` are
siblings.

## Frontmatter schema

You write the human fields; `respec stamp` fills the deterministic ones.

| File | You write | `respec stamp` fills |
| --- | --- | --- |
| research.md | `topic`, `status` (draft\|complete), `tags` | `date`, `repo`, `repo_path`, `git_commit` (write-once) |
| spec.md | `title`, `status` (draft\|approved), `tags` | `date` (write-once) |
| plan.md | `title`, `status` (draft\|approved\|in-progress\|done) | `spec_sha256` (refreshed) |

`respec lint <change-dir>` validates these fields, the `status` values, and the required body
sections.

## Phases

### Research (`/rsx:research <topic>`)
Interview-driven context-building — not neutral documentation. Clarify the topic, explore the code
(`file:line` refs), weigh options with pros/cons, and record the decisions the operator leans
toward. Write `research.md` (required sections: Research Question, Summary, Findings, Open
Questions; add Options & Tradeoffs / Decisions as needed), then `respec stamp <change-dir>`.
Nothing goes to the worked-on repo.

### Plan (`/rsx:plan [change-dir]`)
Read `research.md` (its Decisions are settled constraints). Co-generate `spec.md` + `plan.md`
(shared deliverables). The plan is phased with, per phase, an **Automated Verification** checklist
(commands you can tick yourself) and a **Manual Verification** checklist (operator judgment calls).
If a plan exists and is **stale**, amend it surgically. Then:

    respec stamp <change-dir>

### Implement (`/rsx:implement <change-dir>`)
First gate on staleness:

    respec status <change-dir> --json

- `stale`   → the spec changed after stamping; STOP and re-plan.
- `unstamped` → STOP and run `/rsx:plan` to stamp.
- `fresh`   → proceed: execute phases in order, ticking Automated checks as they pass and pausing
  at each phase's Manual checks for operator confirmation.

## Staleness model

`respec stamp` records the sha256 of `spec.md` into `plan.md` frontmatter (`spec_sha256`).
`respec status` recomputes and compares. The relationship is one-directional: editing `spec.md`
marks the plan stale; editing `plan.md` does not. This is intentional — the plan must always track
the current spec.

## CLI reference

    respec config get|set|path           # configuration (~/.config/respec/config.yaml)
    respec install                       # (re)install these prompts + this skill at user scope
    respec stamp <change-dir>            # write provenance + spec_sha256, then reflow the artifacts
    respec status <change-dir> [--json]  # per-artifact status + fresh | stale | unstamped
    respec lint <change-dir> [--json]    # validate frontmatter, status, and required sections
    respec format <path> [--check]       # reflow prose only; non-prose left byte-identical
    respec templates list|eject          # inspect / customize the /rsx:* prompts + skill
    respec install-hook                  # store pre-commit hook that checks Markdown formatting
    respec render [--out <dir>]          # build the store as a browsable Hugo site
    respec serve                         # serve the store with live reload

## Hard rules

- Never write research/spec/plan artifacts into the worked-on repo — only into the central store.
- Never hand-compute provenance or hashes; run `respec stamp`.
- The plan is the source of truth during implementation.
- Keep `spec.md` and `plan.md` consistent; re-stamp after any spec change.
{{if .Context}}
## Shared context

{{.Context}}
{{end}}
