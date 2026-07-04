# respec

`respec` drives a `research → plan → implement` workflow for a single operator.
All artifacts live in **one central store** (a git repo you own), never in the
repo you are working on. The actual reasoning is done by the agent — running in
[pi](https://github.com/earendil-works/pi) or
[Claude Code](https://code.claude.com) — executing installed prompt-templates
(`/rsx:*` in pi, `/rsx-*` in Claude Code); the `respec` CLI does deterministic
plumbing only:

- hold config (store path, optional injected context/rules),
- render and install the research|plan|implement prompts + backing skill at user scope,
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
- **An agent** — [pi](https://github.com/earendil-works/pi) and/or
  [Claude Code](https://code.claude.com); the prompts run there.

Optional, for the best experience:

- **[Hugo](https://gohugo.io)** — required only by `respec render` / `respec
  serve`, which build the store into a browsable site (with Mermaid diagrams).
- **[probe](https://github.com/probelabs/probe)** — local semantic code
  search; when on PATH at install time, the research prompt steers the agent
  to explore with it (see below).
- **[pre-commit](https://pre-commit.com)** — if your store repo uses the
  framework, this repo ships `respec-format` hooks to guard Markdown
  formatting (a standalone `respec install-hook` alternative needs nothing
  extra).

On Nix, a [devenv](https://devenv.sh) shell can provide all of it. An example
`devenv.nix` for the store repo, which also wires the `respec format --check`
guard as a pre-commit hook (devenv generates `.pre-commit-config.yaml`; the
hooks block additionally needs
`devenv inputs add git-hooks github:cachix/git-hooks.nix --follows nixpkgs`):

```nix
{ pkgs, ... }:

{
  packages = with pkgs; [
    git
    hugo # respec render / serve
  ];

  # for `go install github.com/hurricanehrndz/respec@latest`
  languages.go.enable = true;

  # probe is not in nixpkgs: `npm install -g @probelabs/probe`
  # languages.javascript.enable = true;

  git-hooks.hooks.respec-format = {
    enable = true;
    name = "respec format --check";
    entry = "respec format --check";
    types = [ "markdown" ];
  };
}
```

### Install

```sh
go install github.com/hurricanehrndz/respec@latest

respec config set store ~/respec-store   # central store (default: ~/respec-store)
respec install --target pi               # render + install prompts and the respec skill
respec install --target claude           # same, for Claude Code (run both if you use both)
```

`--target pi` writes the prompts to
`~/.pi/agent/prompts/rsx:{research,plan,implement}.md` and the skill to
`~/.pi/agent/skills/respec/`. `--target claude` writes the commands to
`~/.claude/commands/rsx-{research,plan,implement}.md` (Claude Code command
names cannot contain a colon, so there the workflow is
`/rsx-research|plan|implement`) and the skill to `~/.claude/skills/respec/`.
The store path is baked in either way. Re-running is idempotent; it overwrites
the respec-owned files with a fresh render, so run it again after changing
config.

If [probe](https://github.com/probelabs/probe) (local semantic code search) is
on PATH at install time, the research prompt additionally steers the agent to
explore with `probe search`/`probe extract`; without it the prompts never
mention probe. Install probe, then re-run `respec install`, to enable it.

## Workflow

From any repo, in an agent session (Claude Code names are `/rsx-research` etc.):

1. `/rsx:research <topic>` — an interview-driven exploration that writes `research.md` into
   `<store>/<owner-repo>/<slug>/` (efforts are grouped by the primary repo they affect), then runs
   `respec stamp` to fill provenance (date, repo, git commit).
2. `/rsx:plan` — reads the research, co-generates `spec.md` + `plan.md` in the same change dir
   (phased plan with per-phase Automated/Manual verification), then runs `respec stamp` to record
   the spec's hash. If a plan already exists and the spec changed, it amends the plan surgically
   instead of regenerating.
3. `/rsx:implement` — first runs `respec status --json`; if the spec changed since the plan was
   stamped (`stale`), it stops and tells you to re-plan. Otherwise it executes the plan's phases,
   ticking Automated checkboxes as they pass and pausing at each phase's Manual checks.

Nothing is ever written to the repo you invoke from.

## Commands

| Command | What it does |
| --- | --- |
| `respec config get\|set <key> [value]` | Read/write `~/.config/respec/config.yaml` |
| `respec config path` | Print the config file path |
| `respec install --target <pi\|claude>` | Render + install the prompts and skill at user scope for that agent |
| `respec stamp <change-dir> [--repo <path>]` | Write provenance + `spec_sha256` into the artifacts, then reflow them |
| `respec status <change-dir> [--json]` | Report `fresh` / `stale` / `unstamped` plus per-artifact status |
| `respec list [--json]` | List every effort in the store, grouped by repo, with status/staleness |
| `respec lint <change-dir> [--json]` | Validate frontmatter fields, `status` values, and required sections |
| `respec format <path>... [--check]` | Reflow prose in files or every `*.md` under a dir; non-prose stays byte-identical |
| `respec templates list\|eject [name]` | Inspect templates / copy embedded defaults into `templates_dir` |
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
                        # (override one prompt without re-supplying the rest); empty = embedded
context: ""             # optional shared context injected into every prompt
rules:                  # optional per-artifact rules injected into the matching prompt
  research: ""
  spec: ""
  plan: ""
  implement: ""
```

Keys are addressed with dots on the CLI, e.g.
`respec config set rules.plan "..."`.

## End-to-end walkthrough

A full check of the v1 definition of done, using a throwaway store and repo:

```sh
# 1. point respec at a throwaway store and install
respec config set store /tmp/respec-store
mkdir -p /tmp/respec-store /tmp/scratch-repo
respec install --target pi

# 2. from the unrelated repo, run the workflow in a pi session
cd /tmp/scratch-repo && git init -q .
pi   # then: /rsx:research <topic> → /rsx:plan → /rsx:implement

# 3. artifacts landed only in the store; the scratch repo is untouched
find /tmp/respec-store -name '*.md'
git -C /tmp/scratch-repo status --short   # expect: empty

# 4. staleness gates implement
CHANGE=$(dirname "$(find /tmp/respec-store -name plan.md | head -1)")
respec status "$CHANGE"                   # fresh
echo change >> "$CHANGE/spec.md"
respec status "$CHANGE"                   # stale → /rsx:implement refuses to proceed

# 5. validate, format + browse
respec lint "$CHANGE"                      # frontmatter/status/sections OK
respec format "$CHANGE"                    # reflow every *.md in the change dir
respec serve                              # http://127.0.0.1:1313
```

## Development

```sh
devenv shell    # provides go, hugo, just, golangci-lint
just build      # binary → build/respec
just test       # go vet + go test
just lint       # golangci-lint
```
