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
        research.md   # motivation, findings, options, and operator decisions
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
inheriting a context window polluted with tool output.

## Delegation

Subagents are available in **every** phase. This section is the whole policy — the phase sections
below do not restate it, and nothing in them forbids delegating.

The test is whether you need a task's *output* or only its *conclusion*. Sweeping thirty files to
find the three that matter produces pages nobody re-reads: delegate it and keep the finding. A
question answerable in two reads is not worth the round trip. Gauge scale before deciding.

What stays in the orchestrating session is the operator dialogue and the synthesis — a child cannot
be steered mid-flight, and the alignment is the part that must not be outsourced. Everything else
is fair game: bulk reads, caller audits, pattern surveys, whole implementation phases.

Discover what this machine can reach, then build the command:

    respec agents                                   # harnesses, effort levels, models
    respec agents --filter <substr>                 # narrow a large catalogue
    respec agent-cmd '<spec>' --prompt-file <file>  # exact child command

A spec is `<harness>[:<model>][@<effort>]` — e.g. `pi:openai-codex/gpt-5.6-sol@medium`,
`claude:opus`, or bare `pi`. Model and effort are both optional, and omitting one means "whatever
the harness is configured to do". Write the child's task to a temp file so the prompt is piped,
never interpolated. Run children one at a time; they share a working tree. An agent without a
subagent feature spawns a fresh instance of itself the same way.

Two layers decide staffing, and neither is a roster. The operator's standing preferences live in
`config.yaml` under `agents:` (implementer, reviewer, committer, notes); the plan records the
per-phase judgement made in light of them. **Where the operator has stated nothing, invent
nothing** — no agent line, and the harness does what it is already configured to do.

### Research (`{{.Cmd "research"}} <topic>`)
Drives toward alignment — not neutral documentation. Establish why the change is being requested
or initiated and why now; do not infer unclear motivation. Explore the code (`file:line` refs),
separate verified facts from doc/vendor claims, ask the operator only what evidence cannot answer,
weigh options with pros/cons and recommend one, and record the operator's calls under Decisions.
Write `research.md` with a Motivation section (required lint sections remain Research Question,
Summary, Findings, Open Questions; add Options & Tradeoffs / Decisions as needed), then
`respec stamp <change-dir>` and
`respec lint <change-dir>`. Done when a fresh plan session could work from the artifact and the
code it references without re-asking anything settled. Nothing goes to the worked-on repo.

### Plan (`{{.Cmd "plan"}} [change-dir]`)
Read `research.md` (its Decisions are settled constraints), then co-generate `spec.md` + `plan.md`
(shared deliverables). Find the highest practical end-to-end feedback loop and ask early for any
access/setup it needs. Each phase has **Automated Verification** (deterministic commands) and
**Manual Verification** (judgment). The orchestrator owns every feasible manual check; label only
irreducibly external checks `Operator:`.

Settle `execution_mode` first by asking the operator whether the plan runs unattended. `manual`
means confirming the outline and staffing with them and allowing `Operator:` checks anywhere;
`auto` means gathering every material question up front, naming a runnable E2E/smoke command, and
keeping `Operator:` checks out of every phase but the last — `respec lint` enforces that, which is
why there is no separate auto-planning command.

Staff the phases in the same confirmation, guided by the operator's stated preferences. Run
`respec agents` (add `--filter <substr>` on a large catalogue) to see which harnesses and models
this machine can actually reach — the roster is discovered, never configured, because installed
tooling drifts. Record `**Agent:** <spec> — <why>` per phase, keeping the reasoning so an
unreachable model can be re-derived later. Where the operator has stated no preference, write no
agent lines at all: the harness then does what it is configured to do, and the plan stays portable.
Prefer a different model family for the reviewer than the implementer, so review is adversarial
rather than self-checking.

Revisiting any plan: amend surgically, never regenerate; preserve completed checkboxes unless the
change invalidates them. Preserve `execution_mode` unless the operator explicitly switches it.
Scope/behavior feedback updates `spec.md` first; approach-only feedback updates `plan.md`. Re-check
affected gates and E2E, then:

    respec stamp <change-dir>
    respec lint <change-dir>

### Implement (`{{.Cmd "implement"}} <change-dir>`)
Gate on `respec status <change-dir> --json`; stale or unstamped plans stop for the planner. The
plan's `execution_mode` decides only **when you stop for the operator** — the work is identical
either way. `manual` pauses at each phase gate and before each commit; `auto` runs through to the
final consolidated acceptance. Either way, stop for a plan mismatch, required access, unsafe dirty
state, scope change, or high-impact action.

Per phase, sequentially:

1. **Implement.** Use the phase's `**Agent:**` spec via `respec agent-cmd '<spec>' --prompt-file
   <file>`; with no spec, use the harness's native subagent and its defaults. The child harness
   comes from the plan, so an orchestrator in one harness can drive children in another.
2. **Adversarial review.** Get an independent read of the diff, giving the reviewer the phase text
   and spec alongside it and asking one narrow question: what does the diff do that the phase does
   not ask for, and what does the phase ask for that it does not do? Findings are claims, not
   verdicts — the orchestrator adjudicates. A missing reviewer degrades quality but never halts
   the run.
3. **Gate.** The orchestrator inspects the full diff, runs Automated and E2E checks, performs and
   ticks feasible Manual checks, and never accepts a child's self-report as proof. In manual mode
   this is where you present evidence and wait.
4. **Commit** the accepted phase with the cheapest capable delegate, verify the commit, update
   checkboxes, and continue. Phase workers share one working tree, so running them concurrently
   would have each reviewing and committing the others' half-finished edits.

Set `done` only when all checks are complete; otherwise remain `in-progress` for final operator
acceptance.

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
    respec lint <change-dir> [--json]    # validate frontmatter, status, required sections, agent specs
    respec agents [--filter S] [--all]   # harnesses, effort levels, and models reachable here
    respec agent-cmd <spec> [--prompt-file F]  # exact child command for <harness>[:<model>][@<effort>]
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
