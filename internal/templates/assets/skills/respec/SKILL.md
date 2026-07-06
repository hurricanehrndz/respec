---
name: respec
description: The respec spec-driven workflow (research → plan → implement) over a single central store. Use when running {{.Cmd "research"}}, {{.Cmd "plan"}}, or {{.Cmd "implement"}}, or when creating/editing research.md, spec.md, or plan.md artifacts, or when calling the respec CLI (stamp, status, lint, format, templates, render, serve).
---
{{- /* Workflow mechanics here are deliberately duplicated in the prompts/*.md templates so each
file stands alone — edit both together. */}}

# respec workflow

respec drives a `research → plan → implement` workflow. **All artifacts live in one central
store**, never in the worked-on repo:

    {{.Store}}

The CLI does deterministic plumbing only. Reasoning (research, planning, implementation) is your
job; the CLI handles config, install, provenance/staleness stamping, prose reflow, validation, and
site rendering. Never compute `date`, git metadata, or hashes by hand — `respec stamp` does it.

## Store layout

    {{.Store}}/<owner-repo>/<slug>/
        research.md   # findings, options, and operator decisions
        spec.md       # requirements / desired behavior
        plan.md       # phased implementation + checkboxes

A `<owner-repo>/<slug>/` directory is one "change" (effort). Efforts are grouped by the primary
repo they affect; `slug` is the effort's identity within that repo. `spec.md` and `plan.md` are
siblings. The effort date lives in frontmatter, not the path.

`<owner-repo>` is derived from the worked-on repo's `origin` remote: the last two path segments
(`owner/repo`, `.git` stripped) slugified — e.g. `git@github.com:hurricanehrndz/respec.git` →
`hurricanehrndz-respec` (basename if there is no origin). Run `respec list` before creating a new
effort; if the slug already exists for the repo, ask whether it is a new version (`<slug>-v2`).

## Frontmatter schema

You write the human fields; `respec stamp` fills the deterministic ones.

| File | You write | `respec stamp` fills |
| --- | --- | --- |
| research.md | `topic`, `status` (draft\|complete), `tags` | `date`, `repo`, `repo_path`, `git_commit` (write-once) |
| spec.md | `title`, `status` (draft\|approved), `tags` | `date` (write-once) |
| plan.md | `title`, `status` (draft\|approved\|in-progress\|done), optional `depends_on` | `spec_sha256` (refreshed) |

`respec lint <change-dir>` validates these fields, the `status` values, and the required body
sections.

## Phases

Run each phase in a **fresh session**. The artifacts are the context handoff: a phase gathers and
distills context in the open, in its own session, and the next phase reads the artifact instead of
inheriting a context window polluted with tool output. Default to **no subagents** for context
gathering — they cost the operator observability and steerability. Let codebase size direct you:
delegate bulk reads only when they would drown the session. An agent without a subagent feature
can spawn a fresh instance of itself via the shell.

### Research (`{{.Cmd "research"}} <topic>`)
Drives toward alignment — not neutral documentation. Explore the code (`file:line` refs), separate
verified facts from doc/vendor claims, ask the operator only what evidence cannot answer, weigh
options with pros/cons and recommend one, and record the operator's calls under Decisions. Write
`research.md` (required sections: Research Question, Summary, Findings, Open Questions; add
Options & Tradeoffs / Decisions as needed), then `respec stamp <change-dir>` and
`respec lint <change-dir>`. Done when a fresh plan session could work from the artifact and the
code it references without re-asking anything settled. Nothing goes to the worked-on repo.

### Plan (`{{.Cmd "plan"}} [change-dir]`)
Read `research.md` (its Decisions are settled constraints). Confirm the phase outline with the
operator, then co-generate `spec.md` + `plan.md` (shared deliverables). The plan is phased with,
per phase, an **Automated Verification** checklist (commands you can tick yourself) and a
**Manual Verification** checklist (operator judgment calls). The final plan carries no unresolved
questions and states the chosen approach and why.

Revisiting an existing plan: amend surgically, never regenerate; preserve completed checkboxes
unless the change invalidates them (say so first). Scope/behavior feedback goes into `spec.md`
first — staleness then forces the re-plan; approach-only feedback is a direct plan edit. Then:

    respec stamp <change-dir>
    respec lint <change-dir>

### Implement (`{{.Cmd "implement"}} <change-dir>`)
First gate on staleness:

    respec status <change-dir> --json

- `stale`   → the spec changed after stamping; STOP and re-plan.
- `unstamped` → STOP and run `{{.Cmd "plan"}}` to stamp.
- `fresh`   → proceed: set the plan's `status` to `in-progress`, execute phases in order, ticking
  Automated checks as they pass and pausing at each phase's Manual checks for operator
  confirmation. Trust the plan — search only when it is ambiguous or reality mismatches it; on a
  mismatch, stop and report (expected vs. found) instead of improvising. When the final Manual
  checks are confirmed, set `status` to `done`.

## Staleness model

`respec stamp` records the sha256 of `spec.md` into `plan.md` frontmatter (`spec_sha256`).
`respec status` recomputes and compares. The relationship is one-directional: editing `spec.md`
marks the plan stale; editing `plan.md` does not. This is intentional — the plan must always track
the current spec.

## CLI reference

    respec config get|set|path           # configuration (~/.config/respec/config.yaml)
    respec install --target <agent>      # (re)install these prompts + this skill at user scope (pi | claude)
    respec stamp <change-dir> [--repo <path>]  # write provenance + spec_sha256, then reflow the artifacts
                                         # (--repo: the worked-on repo, when the session runs elsewhere;
                                         #  stamp refuses to record the store itself as provenance)
    respec status <change-dir> [--json]  # per-artifact status + fresh | stale | unstamped
    respec list [--json]                 # every effort in the store, grouped by repo, with staleness
    respec lint <change-dir> [--json]    # validate frontmatter, status, and required sections
    respec format <path>... [--check]    # reflow prose only; non-prose left byte-identical
    respec templates list|eject          # inspect / customize the prompt + skill templates
    respec install-hook                  # store pre-commit hook that checks Markdown formatting
    respec render [--out <dir>]          # build the store as a browsable Hugo site
    respec serve                         # serve the store with live reload

## Hard rules

- Artifacts are for a human co-engineer as much as for you: prefer Mermaid diagrams (rendered by
  GitHub and `respec render`/`serve`) and inline HTML where they communicate better than prose.
- Never write research/spec/plan artifacts into the worked-on repo — only into the central store.
- Never hand-compute provenance or hashes; run `respec stamp`.
- No silent assumptions, in any phase: verify it, ask the operator, or record it as an open
  question — never build on a guess.
- The plan is the source of truth during implementation.
- Keep `spec.md` and `plan.md` consistent; re-stamp after any spec change.
{{if .Context}}
## Shared context

{{.Context}}
{{end}}
