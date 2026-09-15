# respec

`respec` drives a `research → plan → implement` workflow for a single operator.
All artifacts live in **one central store** (a git repo you own), never in the
repo you are working on. The actual reasoning is done by the agent — running in
[pi](https://github.com/earendil-works/pi),
[Prime Agent](https://github.com/PrimeIntellect-ai/prime-agent),
[Claude Code](https://code.claude.com), or
[OpenAI Codex](https://developers.openai.com/codex/) — executing installed
skills (`/skill:rsx-*` in pi and Prime Agent, `/rsx-*` in Claude Code,
`$rsx-*` in Codex); the `respec` CLI does deterministic plumbing only:

- hold config (store path, optional injected context/rules),
- render and install the respec skills (a shared workflow plus three phase entry points) at user scope,
- stamp deterministic frontmatter — provenance (date, repo, git commit) and the
  spec→plan staleness hash — and auto-reflow the artifacts,
- validate artifacts (`lint`) and reflow prose without touching structure,
- render/serve the store as a browsable Hugo site.

## Setup

### Dependencies

Required:

- **[Go](https://go.dev)** — installs respec itself (`go install`).
- **git** — respec's model is git-based: `respec stamp` reads provenance
  (repo, commit) from the worked-on repo, and the central store is a git repo
  you own.
- **An agent** — [pi](https://github.com/earendil-works/pi),
  [Prime Agent](https://github.com/PrimeIntellect-ai/prime-agent),
  [Claude Code](https://code.claude.com), and/or
  [OpenAI Codex](https://developers.openai.com/codex/); the skills run there.

Optional, for the best experience:

- **[Hugo](https://gohugo.io)** — required only by `respec render` / `respec
  serve`, which build the store into a browsable site (with Mermaid diagrams).
- **[pre-commit](https://pre-commit.com)** — if your store repo uses the
  framework, this repo ships `respec-format` hooks to guard Markdown
  formatting (a standalone `respec install-hook` alternative needs nothing
  extra).

[mise](https://mise.jdx.dev) can install Go and Hugo for a store repo:

```sh
mise use go@latest hugo@latest
```

Use `respec install-hook` or the published pre-commit hooks to check store
formatting.

### Install

```sh
go install github.com/hurricanehrndz/respec@latest

respec config set store ~/respec-store   # central store (default: ~/respec-store)
respec install --target pi               # render + install the respec skills
respec install --target prime-agent      # same, for Prime Agent
respec install --target claude           # same, for Claude Code
respec install --target codex            # same, for OpenAI Codex (run each agent you use)
```

Every target installs the same four skills under its own skills directory:
`respec` (the shared workflow) plus `rsx-research`, `rsx-plan`, and
`rsx-implement` (the phase entry points). `--target pi` writes them to
`~/.pi/agent/skills/`, `--target prime-agent` to `~/.prime/agent/skills/`,
`--target claude` to `~/.claude/skills/`, and `--target codex` to
`$CODEX_HOME/skills/` (defaulting `CODEX_HOME` to `~/.codex`). All four skills
carry `disable-model-invocation: true`, so they load only when invoked
explicitly (the phase skills read `respec` by path). The store path is baked in
for every target. Re-running is idempotent; it overwrites the respec-owned
files with a fresh render, so run it again after changing the store path,
context, rules, or templates. It removes any retired respec skill file it no
longer renders. Agent preferences are read during planning, not baked into
skills, so changing them does not require reinstalling.

Custom template overrides using `{{.Agents}}` must switch to planning-time
selection. That field is no longer part of the install-time template context.

## Workflow

From any repo, in an agent session (the pi names are used below; Claude Code spells them
`/rsx-research` etc., Codex `$rsx-research` etc.):

1. `/skill:rsx-research <topic>` — an interview-driven exploration that writes `research.md` into
   `<store>/<owner-repo>/<slug>/` (efforts are grouped by the primary repo they affect), then runs
   `respec stamp` to fill provenance (date, repo, git commit).
2. `/skill:rsx-plan` — reads the research, co-generates `spec.md` + `plan.md` in the same change dir
   (phased plan with per-phase Automated/Manual verification and an end-to-end feedback loop), then
   runs `respec stamp` to record the spec's hash. If a plan already exists, it amends it surgically
   and preserves its `execution_mode`.
3. `/skill:rsx-implement` — first runs `respec status --json`; if the spec changed since the plan was
   stamped (`stale`), it stops and tells you to re-plan. Otherwise it puts the work on a branch for
   the effort (`respec/<slug>` unless the repo or you say otherwise) and runs each phase through the
   same loop: implement, adversarial review, orchestrator gate, commit. Pushing stays yours.

Artifacts remain in the central store; implementation changes land only in the worked-on repo.

### Attended and unattended runs

There are three commands, not five. Whether a plan runs unattended is a property of the plan, not a
separate workflow, so `/skill:rsx-plan` asks once and records it as `execution_mode`:

- **`manual`** — you are looped in. Planning confirms the outline and staffing with you;
  implementation pauses at each phase gate and before each commit, and `Operator:` checks may
  appear in any phase.
- **`auto`** — you are looped in only at the end. Planning gathers its questions up front and must
  name a runnable E2E/smoke command; implementation runs through to a final consolidated
  acceptance.

The two modes do identical work — the only difference is when you are consulted. What keeps an auto
plan honest is enforced by `respec lint` rather than by a separate skill: an `Operator:` check
anywhere but the final phase is rejected, because it would stall an unattended run.

## Commands

| Command | What it does |
| --- | --- |
| `respec config get\|set <key> [value]` | Read/write `~/.config/respec/config.yaml` |
| `respec config path` | Print the config file path |
| `respec config print` | Print effective configuration, including optional planning defaults |
| `respec install --target <pi\|prime-agent\|claude\|codex>` | Render + install the respec skills at user scope for that agent |
| `respec stamp <change-dir> [--repo <path>]` | Write provenance + `spec_sha256` into the artifacts, then reflow them |
| `respec status <change-dir> [--json]` | Report `fresh` / `stale` / `unstamped` plus per-artifact status |
| `respec list [--json]` | List every effort in the store, grouped by repo, with status/staleness |
| `respec lint <change-dir> [--json]` | Validate frontmatter fields, `status` values, and required sections |
| `respec format <path>... [--check]` | Reflow prose in files or every `*.md` under a dir; non-prose stays byte-identical |
| `respec templates list\|eject [name]` | Inspect templates / copy embedded defaults into `templates_dir` |
| `respec agents [--filter S] [--all] [--json]` | Probe external agent CLIs, effort levels, and reachable models, not native workers |
| `respec agent-cmd <spec> [--prompt-file F]` | Print the external CLI command for `<harness>:<model>@<effort>` |
| `respec install-hook [--force]` | Install a store pre-commit hook that checks Markdown formatting |
| `respec render [--out <dir>]` | Build the store as a Hugo site (default `<cache>/respec/site/public`) |
| `respec serve [--port N] [--bind ADDR]` | Serve the store with live reload |

`respec stamp` owns the deterministic frontmatter the agent should never
hand-write: provenance (`date`, `repo`, `repo_path`, `git_commit`, read from the
worked-on repo and written **write-once** into `research.md`/`spec.md`) plus
`spec_sha256` (refreshed into `plan.md`). It reflows each artifact it touches.
`respec lint` then checks required fields, `status` values, and sections.

Efforts live at `<store>/<owner-repo>/<slug>/` — grouped by the primary repo,
identified by `slug` (the date is in frontmatter, not the path). `respec list`
gives a store-wide overview. A plan may declare `depends_on` (a bare `slug` for
the same repo, or `repo/slug` cross-repo); `respec list` shows a dependent as
`(blocked)` until its dependencies reach `status: done`. This is for visibility —
respec does not orchestrate or auto-run efforts.

Breaking change: earlier versions laid the store out as
`<problem-space>/<YYYY-MM-DD-slug>/`. respec does not migrate old stores —
move each effort directory to `<owner-repo>/<slug>/` by hand (the date lives
in frontmatter, which `respec stamp` fills).

Staleness is asymmetric by design: only spec→plan is tracked, so editing
`plan.md` (e.g. ticking checkboxes during implementation) never marks anything
stale.

## Delegation

Skills use the current environment's native subagents by default. Native workers
receive tasks through that environment's own tools, without Respec adding CLI
launch commands or prompt files. If native workers are unavailable, the agent
works in the current session and reports the limitation.

### Choose workers during planning

For each effort, planning asks which worker preferences you want and records
them in `plan.md`. Installed skills describe the process, not your model choices.
For native workers, an optional section can look like this:

```markdown
## Worker preferences

- Implementation: prefer <model> at medium effort
- Review: prefer <model> at high effort
- Commits: native defaults
- Fallback: native defaults; report unavailable preferences
```

Put phase-specific overrides in that phase's prose. Implementation follows the
plan's choices for every role, not live config; later config changes cannot
restaff an approved effort. Amending a plan preserves its choices unless you
change them. Omit the section when you state no preferences.

A model, effort, or budget preference alone does not opt into external execution.
Native hints apply only where the environment exposes the corresponding controls.
With no preferences, the environment's defaults apply. Existing plans need no
new fields or sections.

### Optional standing defaults

Planning can read `agents:` from `respec config print` and offer those values
as defaults. You can accept or replace them for each effort. They are not
assignments until recorded in the plan, and implementation does not reload them.
The three role fields below suggest external CLIs; `notes` can suggest native
model or budget preferences without selecting a CLI:

```yaml
agents:
  implementer: pi:openai-codex/gpt-5.6-sol   # writes the code for a phase
  reviewer:    claude:opus                   # independently reviews the diff
  committer:   pi:openai-codex/gpt-5.4-mini  # commits an accepted phase
  notes:       "cost matters more than speed; ask before anything above high"
```

```sh
respec config set agents.implementer 'pi:openai-codex/gpt-5.6-sol'
```

### Choosing external executors

An explicit external executor choice in the plan or current session opts that
work into CLI delegation. Each role is independent: an external reviewer does
not imply an external implementer. If an existing assignment conflicts with a
request for native workers, the agent asks you to settle it.

A spec is `<harness>[:<model>][@<effort>]`, where harness is `pi`,
`prime-agent`, `claude`, or `codex`. **Model and effort are both optional** —
`claude:opus`, `pi@high`, or bare `pi` all work, and whatever you leave out
stays at the harness's own
default rather than something respec chose. Specs are validated when you set
them, so a bad effort level fails at the keyboard instead of mid-phase.

Every harness name in a spec identifies an external CLI. In particular,
`codex:...` selects `codex exec`, not an app-native Codex worker.

The child harness is independent of the one running the orchestrator, so a
Claude Code session can delegate a phase to gpt-5.6-sol through pi, and a pi
session can delegate to Opus through Claude Code. Beyond cost control that buys
review independence: a phase implemented by one model family and reviewed by
another is genuinely adversarial, rather than a model checking its own blind
spots.

For phases using an explicitly chosen external implementer, the plan records
the assignment and its reasoning:

    ## Phase 2: Parse the roster
    **Agent:** pi:openai-codex/gpt-5.6-sol@medium — mechanical, cheap model is enough

If that model is retired or unreachable when the plan runs, the implementer
looks for an equivalent on the chosen executor and reports the substitution.
Switching executors requires your approval. External reviewer and committer
specs go in the corresponding role's prose under **Worker preferences**, with
any phase-specific overrides in the phase.

The roster itself is **discovered, never stored**. Installed harnesses and
model catalogues drift, so a saved list would be wrong shortly after writing:

```sh
respec agents                        # external CLIs this machine can reach
respec agents --filter gpt-5.6       # narrow a large catalogue
respec agent-cmd 'claude:opus@high'  # the exact child command
```

These commands do not discover or launch native workers. For external workers,
`respec agent-cmd` builds the command with the correct per-CLI flags and stdin
redirection. The orchestrator executes it with a private prompt file, then
removes the file. Skills never assemble those flags themselves.

`respec lint` never requires an `**Agent:**` line. Absence leaves implementation
with the environment's native mechanism; existing CLI assignments retain their
meaning and validation.

### Adversarial review

Each phase gets an independent read of its diff before the orchestrator judges
it, from the reviewer recorded in the plan and the environment's native
subagent otherwise. The reviewer receives the phase text and spec alongside the
diff, and answers one narrow question: what does the diff do that the phase
does not ask for, and what does the phase ask for that it does not do?

Its findings are claims, not verdicts. Prompted to find problems a model will
find them, including invented ones, so the orchestrator adjudicates and says
which it kept. If a reviewer is unavailable, the orchestrator follows the
plan's fallback. Without a stricter fallback, it reviews the diff itself and
reports that no independent reviewer was used.

`respec format` reflows paragraph prose only — tables, fenced code, headings,
inline HTML, and bare URLs are left byte-identical, and inline code / links /
URLs are never split across lines.

To guard the store repo, either install the standalone hook
(`respec install-hook`) or, if the store uses the
[pre-commit](https://pre-commit.com) framework, reference this repo's hooks in
its `.pre-commit-config.yaml` (`respec-format` checks; `respec-format-fix`
rewrites in place):

```yaml
repos:
  - repo: https://github.com/hurricanehrndz/respec
    rev: <tag-or-sha>
    hooks:
      - id: respec-format
```

`respec render` / `respec serve` require `hugo` on PATH, preserve inline HTML
(`markup.goldmark.renderer.unsafe = true`), and render ` ```mermaid ` fences as
diagrams (mermaid.js is loaded from the jsDelivr CDN, so diagrams need network
to display).

## Configuration

`~/.config/respec/config.yaml`:

```yaml
store: ~/respec-store   # central store path
reflow_width: 80        # prose reflow width for `respec format`
templates_dir: ""       # optional override dir, layered per-file over the embedded defaults
                        # (override one skill without re-supplying the rest); empty = embedded
context: ""             # optional shared context injected into the shared respec skill
rules:                  # optional per-artifact rules injected into the matching skill
  research: ""
  spec: ""
  plan: ""
  implement: ""
agents:                 # optional defaults offered during planning, not assignments
  implementer: ""       # <harness>[:<model>][@<effort>], e.g. pi:openai-codex/gpt-5.6-sol
  reviewer: ""          # e.g. claude:opus
  committer: ""         # e.g. pi:openai-codex/gpt-5.4-mini
  notes: ""             # free prose: budget stance, models to avoid
```

Path values (`store` and `templates_dir`) expand a leading `~`, `$VAR`, and
`${VAR}` when used. Keys are addressed with dots on the CLI, e.g.
`respec config set rules.plan "..."` or
`respec config set agents.reviewer 'claude:opus'`. The three agent keys are
validated on write; `notes` is free prose. These are optional planning inputs;
set effort-specific worker choices in `plan.md`, not config.

## End-to-end walkthrough

A full check of the v1 definition of done, using a throwaway store and repo:

```sh
# 1. point respec at a throwaway store and install
respec config set store /tmp/respec-store
mkdir -p /tmp/respec-store /tmp/scratch-repo
respec install --target pi

# 2. from the unrelated repo, run the workflow in a pi session
cd /tmp/scratch-repo && git init -q .
pi   # then: /skill:rsx-research <topic> → /skill:rsx-plan → /skill:rsx-implement

# 3. artifacts landed only in the store; the scratch repo is untouched
find /tmp/respec-store -name '*.md'
git -C /tmp/scratch-repo status --short   # expect: empty

# 4. staleness gates implement
CHANGE=$(dirname "$(find /tmp/respec-store -name plan.md | head -1)")
respec status "$CHANGE"                   # fresh
echo change >> "$CHANGE/spec.md"
respec status "$CHANGE"                   # stale → /skill:rsx-implement refuses to proceed

# 5. validate, format + browse
respec lint "$CHANGE"                      # frontmatter/status/sections OK
respec format "$CHANGE"                    # reflow every *.md in the change dir
respec serve                              # http://127.0.0.1:1313
```

## Development

```sh
mise install    # installs the versions pinned in mise.toml
just build      # binary → build/respec
just test       # go vet + go test
just lint       # golangci-lint
```
