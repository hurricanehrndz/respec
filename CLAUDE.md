# respec — project conventions

## Design principle: deterministic when possible

The CLI owns everything that code can answer; the agent (running the installed
workflow skills) is used only for judgment — research, planning,
implementation. If a value can be computed deterministically (dates, git metadata, hashes, YAML
manipulation, formatting), it belongs in the CLI, not in a skill instructing
the agent to shell out and hand-write it.

Concretely:

- `respec stamp` writes all deterministic frontmatter (provenance, spec hash)
  and reflows the artifacts. Skills must never instruct the agent to run
  `git`/`date` or hand-write those fields.
- `respec lint` validates what the schema can enforce (fields, status enums,
  required sections). Enforcement lives in `internal/artifact`, not in skill
  prose.
- New features should ask first: "can the CLI do this?" Only what genuinely
  requires judgment goes into the skill templates.
- For explicitly chosen external executors, `respec agents` probes CLIs and
  `respec agent-cmd` builds the child command line. Skills must never assemble
  harness flags themselves. Those details belong in `internal/agents`.

### Delegation: native by default

Skills use the current environment's native subagent mechanism unless the
operator explicitly selects an external executor in the plan or session.
Model, effort, and budget preferences alone do not select a CLI; a choice for
one role does not select executors for other roles. Native worker preferences
belong in plan prose, without an `**Agent:**` line.

`respec agents` cannot discover app-native workers, and `respec agent-cmd`
cannot launch them. Even a `codex` spec means the external CLI, not an
app-native worker. Keep CLI discovery, flags, and prompt files confined to
external delegation. Existing CLI assignments keep their meaning; ask the
operator to settle any conflict with a request for native workers.

### Delegation: three layers, split by lifetime

- **Standing defaults** live under `agents:` in `config.yaml`. Planning can
  read them through `respec config print` and offer them to the operator.
  Never bake them into installed skills or treat them as effort assignments.
- **Roster** — volatile, so it is never stored. `respec agents` probes external
  CLIs at the moment the operator asks; the current environment manages native
  worker availability.
- **Worker choices** belong to each effort. Planning settles them with the
  operator and records implementation, review, commit, and fallback preferences
  in `plan.md` under an optional **Worker preferences** section. Phase prose
  overrides plan-wide choices for that role. External implementers still use
  per-phase `**Agent:** <spec> — <why>` lines; native hints stay in prose.

Implementation follows the plan, never live agent config. Replanning preserves
recorded choices unless the operator changes them. Go templates handle
install-time settings and target syntax, not per-effort model choices.

Two invariants follow. Plans record a spec, never a validated catalogue entry:
`internal/agents` checks the harness and effort level, which are stable, and
leaves the model id opaque because a plan outlives any catalogue we could check
it against. And **stated nothing means invent nothing** — with no preferences,
skills write no agent lines and the harness does what it is configured to do.
That is what keeps plans portable between machines and keeps every plan written
before delegation existed valid.

### One skill per phase, not one per mode

There is no `plan-auto`/`implement-auto`. Whether a plan runs unattended is a
property of the plan (`execution_mode`), so a second skill cannot enforce it —
only lint can, and only lint holds for a plan hand-edited afterward. The
auto-mode rule lives in `internal/artifact`: an `Operator:` check outside the
final phase is rejected, because it would stall an unattended run.

`implement` reads `execution_mode` to decide when to stop for the operator. The
work is identical in both modes; only the gating differs. If a change makes the
two modes diverge structurally, that is the signal something belongs in lint
rather than in skill prose.

### Branch per effort: skill guidance, not enforcement

The implement skill puts an effort's commits on their own branch, defaulting
to `respec/<slug>`, and leaves an already-checked-out non-default branch alone.
That stays prose rather than a `respec branch` command or a lint rule for two
reasons: the default name is the change-dir basename, so there is nothing to
compute; and an operator or repository with its own naming convention has to
win over respec's default, which a lint rule cannot express. Pushing and PR
creation remain the operator's call — the skill creates and switches, never
publishes.

### Acknowledged compromise: repo-slug derivation

We aim for determinism but it isn't always practical. The `<owner-repo>`
directory name under the store is derived **by the agent** from the origin
remote, following the recipe in `internal/templates/assets/skills/rsx-research/SKILL.md`
— a deliberate trade (operator's call) to avoid growing the CLI surface with a
`respec new`-style command. Known cost: the hand-derivation can drift on
unusual remotes (e.g. GitLab subgroups), scattering one repo's efforts across
directories and weakening bare-slug `depends_on` resolution. If that drift
becomes a real problem, the fix is a canonical `RepoSlug()` in
`internal/gitmeta` surfaced to the agent — not more skill prose.

## Design provenance

The skill structure derives from the RPI (Research → Plan →
Implement) Goose recipes kept as unmodified reference in
`docs/external/rpi-recipes/` (see its README for origin and where respec
deliberately deviates). Consult them when iterating on
`internal/templates/assets/`.

## Layout invariants

- Store: `<store>/<owner-repo>/<slug>/` with `research.md`, `spec.md`,
  `plan.md` siblings. The effort date lives in frontmatter, not the path.
- Artifacts never land in the worked-on repo — only in the central store.
- Staleness is one-directional (spec→plan) by design.
- Section names validated by `respec lint` are load-bearing: the skill
  skeletons in `internal/templates/assets/` must keep them verbatim, and
  changes to `internal/artifact` validators must update the skills in the
  same commit.

## Development

- `just build` / `just test` / `just lint` (devenv shell provides go, hugo,
  just, golangci-lint).
- errcheck is enforced: check or explicitly discard every error return.
- Test git fixtures must disable commit signing (`-c commit.gpgsign=false`)
  for hermeticity.
