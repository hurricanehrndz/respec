---
name: respec
description: The respec spec-driven workflow (research → plan → implement) over a single central store. Use when running {{.Cmd "research"}}, {{.Cmd "plan"}}, {{.Cmd "plan-auto"}}, {{.Cmd "implement"}}, or {{.Cmd "implement-auto"}}, or when creating/editing research.md, spec.md, or plan.md artifacts, or when calling the respec CLI (stamp, status, lint, format, templates, render, serve).
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
| plan.md | `title`, `status` (draft\|approved\|in-progress\|done), optional `execution_mode` (manual\|auto), optional `depends_on` | `spec_sha256` (refreshed) |

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
operator, then co-generate `spec.md` + `plan.md` (shared deliverables). Find the highest practical
end-to-end feedback loop and ask early for any access/setup it needs. Each phase has **Automated
Verification** (deterministic commands) and **Manual Verification** (judgment). The orchestrator
owns every feasible manual check; label only irreducibly external checks `Operator:`. New normal
plans use `execution_mode: manual`.

### Auto Plan (`{{.Cmd "plan-auto"}} [change-dir]`)
Perform the same planning work without outline approval or incremental confirmation. Resolve
ordinary choices from evidence, conventions, and safe reversible defaults. Ask all material
questions and E2E prerequisites together at the start; do not plan around a blocker. Write
`execution_mode: auto`, include a runnable E2E/integration/smoke command, and make every phase
independently implementable, reviewable, verifiable, and committable. Record the autonomous
execution contract in the overview.

Revisiting any plan: amend surgically, never regenerate; preserve completed checkboxes unless the
change invalidates them. Preserve `execution_mode` unless the operator explicitly switches it;
invoking Auto Plan on a manual plan is an explicit switch to auto. Scope/behavior feedback updates
`spec.md` first; approach-only feedback updates `plan.md`. Re-check affected gates and E2E, then:

    respec stamp <change-dir>
    respec lint <change-dir>

### Implement (`{{.Cmd "implement"}} <change-dir>`)
Gate on `respec status <change-dir> --json`; stale/unstamped plans stop for the matching planner.
A fresh manual plan runs in phase order. Tick passing Automated checks, perform and tick feasible
Manual checks yourself, and pause only for remaining `Operator:` checks. Run the planned E2E path.
If reality contradicts the plan, stop and report expected vs. found instead of improvising.

### Auto Implement (`{{.Cmd "implement-auto"}} <change-dir>`)
Require a fresh plan with `execution_mode: auto`. Work sequentially by phase:

1. Spawn a fresh medium-effort implementation child in the worked-on repo:
{{if .IsClaude}}
       env -u CLAUDECODE claude --print --no-session-persistence --effort medium "<phase task>"
{{else}}
       pi --print --no-session --thinking medium "<phase task>"
{{end}}{{if .IsClaude}}
   Claude children deliberately unset the parent's `CLAUDECODE` nesting guard.{{end}}
2. The orchestrator reviews the complete diff, runs all automated and E2E checks, and performs all
   feasible manual checks. Failed review goes to a fresh medium-effort correction child; the
   orchestrator never accepts a child's self-report as proof.
3. After acceptance, spawn a fresh low-effort child to commit exactly the phase:
{{if .IsClaude}}
       env -u CLAUDECODE claude --print --no-session-persistence --effort low "<commit task>"
{{else}}
       pi --print --no-session --thinking low "<commit task>"
{{end}}
4. Verify the commit, update checkboxes, and continue. Never run phase workers concurrently in one
   working tree.

Do not interrupt the operator until the final consolidated acceptance unless blocked by a plan
mismatch, required access, unsafe dirty state, scope change, or high-impact action. Set `done` only
when all checks are complete; otherwise remain `in-progress` for final operator acceptance.

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
- Every plan seeks an end-to-end feedback loop; auto plans require a runnable E2E, integration, or
  smoke check and surface prerequisites during planning, not implementation.
{{if .Context}}
## Shared context

{{.Context}}
{{end}}
