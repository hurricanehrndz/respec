---
name: respec
description: "Shared respec workflow: store layout, frontmatter schema, delegation, staleness, and CLI reference."
disable-model-invocation: true
---

# respec workflow

respec uses a `research → plan → implement` workflow. The CLI handles deterministic plumbing:
configuration, installation, provenance and staleness stamping, prose reflow, validation, and site
rendering. The agent handles research, planning, and implementation.

## Store layout

All artifacts live in one central store, never in the worked-on repository:

    {{.Store}}/<owner-repo>/<slug>/
        research.md   # motivation, findings, options, and operator decisions
        spec.md       # requirements and desired behavior
        plan.md       # phased implementation and checkboxes

One `<owner-repo>/<slug>/` directory is one change effort. Efforts are grouped by the primary
repository they affect. The slug identifies an effort within that repository. The effort date
lives in frontmatter, not the path.

Derive `<owner-repo>` from the worked-on repository's `origin` remote. Take the last two path
segments (`owner/repo`), remove a trailing `.git`, then slugify them: lowercase and replace runs of
non-alphanumeric characters with hyphens. For example,
`git@github.com:hurricanehrndz/respec.git` becomes `hurricanehrndz-respec`. If there is no
`origin`, use the repository directory's basename.

Run `respec list` before creating an effort. If the same repository already has that slug, ask
whether this is a new version. If so, use a distinct slug such as `<slug>-v2`.

## Frontmatter schema

Write the human fields. `respec stamp` fills the deterministic fields.

| File | You write | `respec stamp` fills |
| --- | --- | --- |
| research.md | `topic`, `status` (`draft` or `complete`), optional `tags` | `date`, `repo`, `repo_path`, `git_commit` (write-once) |
| spec.md | `title`, `status` (`draft` or `approved`), optional `tags` | `date` (write-once) |
| plan.md | `title`, `status` (`draft`, `approved`, `in-progress`, or `done`), `execution_mode` (`manual` or `auto`), optional `depends_on` | `spec_sha256` (refreshed) |

A `depends_on` value is a bare slug for an effort in the same repository or `repo/slug` for a
cross-repository dependency. `respec lint <change-dir>` validates frontmatter, status values, and
required body sections.

## Delegation policy

Use a subagent when you need only a task's conclusion, not its full output. Delegate broad caller
audits, pattern surveys, bulk reads, and whole implementation phases. Keep a question that takes
two reads in the current session. Keep operator dialogue and synthesis in the orchestrating
session because those require the full context and live steering.

Use the current environment's native subagent mechanism by default for research, implementation,
review, and commits. Let it deliver tasks and manage worker sessions; do not add CLI discovery,
shell commands, or prompt files to native delegation. Use the environment's defaults unless the
operator states a preference. Apply native model and effort choices only through controls the
environment actually exposes; report unsupported preferences rather than switching executors. If
native subagents are unavailable, work in the current session and report the limitation, unless
the plan's fallback requires stopping.

Worker choices belong to the effort's `plan.md`, settled during planning, not to installed skills.
Standing `agents:` config values are optional planning inputs; implementation follows the plan,
not live config. Only an explicit external executor choice in the plan or this session opts that
work into CLI delegation. Model, effort, and cost preferences alone do not authorize an external CLI.
A choice for one role does not opt other roles into external execution. If an existing CLI assignment conflicts with a
request for native workers, ask the operator to settle the assignment before proceeding.

Run children one at a time when they share a working tree.

### External CLI delegation

For explicitly chosen external executors, discover available CLIs and build the launch command:

    respec agents                                  # external CLIs, effort levels, and models
    respec agents --filter <substr>                # narrow a large catalogue
    respec agent-cmd '<spec>' --prompt-file <file>   # print the external child command

These commands do not discover or launch app-native workers. An agent spec is
`<harness>[:<model>][@<effort>]`, for example `pi:openai-codex/gpt-5.6-sol@medium`, `claude:opus`,
or bare `pi`. Every harness name identifies an external CLI, including `codex`, not an app-native
worker. Model and effort are optional; omitting either uses that CLI's configured default.

Write the external child's task to a private temporary file, pass it with `--prompt-file`, execute
the printed command, and remove the file when the task ends. Never assemble child flags yourself,
interpolate a prompt into a shell command, or use `eval`.

## Staleness model

`respec stamp` records the sha256 of `spec.md` in `plan.md` as `spec_sha256`. `respec status`
recomputes and compares it. Editing `spec.md` makes the plan stale; editing `plan.md` does not.
This is one-directional because the plan must track the current spec.

## CLI reference

    respec config print|get|set|path     # effective config or individual values
    respec install --target <agent>      # install at user scope (pi | prime-agent | claude | codex)
    respec stamp <change-dir> [--repo <path>]
                                         # write provenance and spec_sha256, then reflow artifacts
                                         # use --repo for the worked-on repository when running elsewhere;
                                         # stamp refuses to record the store itself as provenance
    respec status <change-dir> [--json]  # artifact status plus fresh | stale | unstamped
    respec list [--json]                 # efforts grouped by repository, with staleness
    respec lint <change-dir> [--json]    # validate frontmatter, sections, status, and agent specs
    respec agents [--filter S] [--all]   # external CLIs only; does not probe native workers
    respec agent-cmd <spec> [--prompt-file F]
                                         # external child command for <harness>[:<model>][@<effort>]
    respec format <path>... [--check]    # reflow prose; leave non-prose byte-identical
    respec templates list|eject          # inspect or customize skill templates
    respec install-hook                  # store pre-commit hook for Markdown formatting
    respec render [--out <dir>]          # build the store as a browsable Hugo site
    respec serve                         # serve the store with live reload

## Hard rules

- Run each phase in a fresh session. The artifacts pass context between phases.
- Write artifacts for a human co-engineer. Use Mermaid diagrams or inline HTML when they explain
  structure better than prose. GitHub and `respec render`/`serve` render both.
- Never write `research.md`, `spec.md`, or `plan.md` into the worked-on repository. They belong only
  in the central store.
- Never compute dates, git provenance, or hashes by hand. Run `respec stamp`.
- Never build on a silent assumption. Verify it, ask the operator, or record an open question.
- Treat the plan as the source of truth during implementation.
- Keep `spec.md` and `plan.md` consistent. Re-stamp after every spec change.
- Every plan needs the highest practical end-to-end feedback loop. An auto plan requires a
  runnable E2E, integration, or smoke command and identifies its prerequisites during planning.
{{if .Context}}
## Shared context

{{.Context}}
{{end}}
