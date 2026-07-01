# respec

`respec` drives a `research → plan → implement` workflow for a single operator.
All artifacts live in **one central store** (a git repo you own), never in the
repo you are working on. The actual reasoning is done by
[pi](https://github.com/earendil-works/pi) executing installed `/rsx:*`
prompt-templates; the `respec` CLI does deterministic plumbing only:

- hold config (store path, optional injected context/rules),
- render and install the `/rsx:research|plan|implement` prompts + backing skill at user scope,
- stamp/compare spec→plan staleness (sha256 in frontmatter),
- reflow prose without touching structure,
- render/serve the store as a browsable Hugo site.

## Setup

Requires Go and Hugo (both provided by the repo's `devenv` shell) and pi.

```sh
go install github.com/hurricanehrndz/respec@latest

respec config set store ~/respec-store   # central store (default: ~/respec-store)
respec install                           # render + install /rsx:* prompts and the respec skill
```

`respec install` writes the prompts to
`~/.pi/agent/prompts/rsx:{research,plan,implement}.md` and the skill to
`~/.pi/agent/skills/respec/`, with the store path baked in. Re-running it is
idempotent; it overwrites the respec-owned files with a fresh render, so run it
again after changing config.

## Workflow

From any repo, in a pi session:

1. `/rsx:research <topic>` — explores and writes a provenance-stamped `research.md` into
   `<store>/<problem-space>/<YYYY-MM-DD-slug>/`.
2. `/rsx:plan` — reads the research, co-generates `spec.md` + `plan.md` in the same change dir,
   then runs `respec stamp` to record the spec's hash. If a plan already exists and the spec
   changed, it amends the plan surgically instead of regenerating.
3. `/rsx:implement` — first runs `respec status --json`; if the spec changed since the plan was
   stamped (`stale`), it stops and tells you to re-plan. Otherwise it executes the plan's phases,
   ticking checkboxes in `plan.md`.

Nothing is ever written to the repo you invoke from.

## Commands

| Command | What it does |
| --- | --- |
| `respec config get\|set <key> [value]` | Read/write `~/.config/respec/config.yaml` |
| `respec config path` | Print the config file path |
| `respec install` | Render + install the `/rsx:*` prompts and skill at user scope |
| `respec stamp <change-dir>` | Record `spec.md`'s sha256 in `plan.md` frontmatter |
| `respec status <change-dir> [--json]` | Report `fresh` / `stale` / `unstamped` |
| `respec format <file> [--check]` | Reflow prose to `reflow_width`; non-prose stays byte-identical |
| `respec render [--out <dir>]` | Build the store as a Hugo site (default `<cache>/respec/site/public`) |
| `respec serve [--port N] [--bind ADDR]` | Serve the store with live reload |

Staleness is asymmetric by design: only spec→plan is tracked, so editing
`plan.md` (e.g. ticking checkboxes during implementation) never marks anything
stale.

`respec format` reflows paragraph prose only — tables, fenced code, headings,
inline HTML, and bare URLs are left byte-identical, and inline code / links /
URLs are never split across lines.

`respec render` / `respec serve` require `hugo` on PATH and preserve inline HTML
(`markup.goldmark.renderer.unsafe = true`).

## Configuration

`~/.config/respec/config.yaml`:

```yaml
store: ~/respec-store   # central store path
reflow_width: 80        # prose reflow width for `respec format`
templates_dir: ""       # optional dir of prompt-template overrides; empty = embedded defaults
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
respec install

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

# 5. format + browse
respec format "$CHANGE/plan.md"
respec serve                              # http://127.0.0.1:1313
```

## Development

```sh
devenv shell
go build ./... && go vet ./... && go test ./...
```
