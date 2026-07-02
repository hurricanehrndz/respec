# respec — project conventions

## Design principle: deterministic when possible

The CLI owns everything that code can answer; the agent (running the `/rsx:*`
prompts) is used only for judgment — research, planning, implementation. If a
value can be computed deterministically (dates, git metadata, hashes, YAML
manipulation, formatting), it belongs in the CLI, not in a prompt instructing
the agent to shell out and hand-write it.

Concretely:

- `respec stamp` writes all deterministic frontmatter (provenance, spec hash)
  and reflows the artifacts. Prompts must never instruct the agent to run
  `git`/`date` or hand-write those fields.
- `respec lint` validates what the schema can enforce (fields, status enums,
  required sections). Enforcement lives in `internal/artifact`, not in prompt
  prose.
- New features should ask first: "can the CLI do this?" Only what genuinely
  requires judgment goes into the prompt templates.

### Acknowledged compromise: repo-slug derivation

We aim for determinism but it isn't always practical. The `<owner-repo>`
directory name under the store is derived **by the agent** from the origin
remote, following the recipe in `internal/templates/assets/prompts/research.md`
— a deliberate trade (operator's call) to avoid growing the CLI surface with a
`respec new`-style command. Known cost: the hand-derivation can drift on
unusual remotes (e.g. GitLab subgroups), scattering one repo's efforts across
directories and weakening bare-slug `depends_on` resolution. If that drift
becomes a real problem, the fix is a canonical `RepoSlug()` in
`internal/gitmeta` surfaced to the agent — not more prompt prose.

## Design provenance

The prompt-template structure derives from the RPI (Research → Plan →
Implement) Goose recipes kept as unmodified reference in
`docs/external/rpi-recipes/` (see its README for origin and where respec
deliberately deviates). Consult them when iterating on
`internal/templates/assets/`.

## Layout invariants

- Store: `<store>/<owner-repo>/<slug>/` with `research.md`, `spec.md`,
  `plan.md` siblings. The effort date lives in frontmatter, not the path.
- Artifacts never land in the worked-on repo — only in the central store.
- Staleness is one-directional (spec→plan) by design.
- Section names validated by `respec lint` are load-bearing: the prompt
  skeletons in `internal/templates/assets/` must keep them verbatim, and
  changes to `internal/artifact` validators must update the prompts in the
  same commit.

## Development

- `just build` / `just test` / `just lint` (devenv shell provides go, hugo,
  just, golangci-lint).
- errcheck is enforced: check or explicitly discard every error return.
- Test git fixtures must disable commit signing (`-c commit.gpgsign=false`)
  for hermeticity.
