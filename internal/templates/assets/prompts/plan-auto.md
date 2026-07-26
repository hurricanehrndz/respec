---
description: Autonomously generate an implementation-ready spec and plan
argument-hint: "[change-dir]"
---
You are in the **Auto Plan** phase of the respec workflow (research → plan → implement).

Change directory (defaults to the most recent under the store if omitted): {{.Arg1Or "<pick latest>"}}

If omitted, run `respec list` and take the most-recent effort for the worked-on repo. Ask only if
there is no unambiguous match.

Central store:

    {{.Store}}

Load the detailed workflow first: read `{{.SkillPath}}` (or `{{.SkillCmd}}`).

## Autonomy contract

Proceed from inspection to finished artifacts without asking for outline approval or incremental
confirmation. Ask the operator only when a material decision cannot be resolved from research,
code, existing conventions, or a safe reversible default. Gather all such questions — including
credentials, services, hardware, destructive setup, or other end-to-end test prerequisites — and
ask them together as early as possible. Do not write a plan around an unresolved blocker.

## Process

1. Read `research.md` completely. Its **Decisions** are settled constraints; do not re-litigate
   them. Read existing `spec.md` and `plan.md` completely when present.{{if .HasProbe}}
   `probe` is installed — use it to pull a single definition without reading the whole file:

       probe extract <file>#<symbol>
       probe extract <file>:<line>
{{end}}
2. Inspect only enough code to close planning gaps and verify current file, symbol, command, and
   test references. Resolve ordinary choices yourself using repository conventions and the
   smallest approach that satisfies the research.
3. Find the highest meaningful end-to-end feedback loop the implementer can run: an existing E2E
   suite, integration test, CLI smoke test, rendered output check, or real request through the
   public boundary. An auto plan must include a runnable E2E/smoke command. If access or setup is
   required, ask for it now; if no honest end-to-end check is possible, stop and explain why.
4. Design phases that can run unattended. Each phase must be independently implementable,
   reviewable, verifiable, and committable. Record this execution contract in the plan overview:
   a medium-effort subagent implements one phase; the orchestrator reviews the diff and performs
   all verification; a low-effort subagent commits only after the orchestrator accepts the phase.
5. Write `spec.md` and `plan.md` together. Use `execution_mode: auto` in `plan.md`. Set statuses to
   `approved` only when requirements and implementation choices are settled.
6. Revisiting an existing plan is an **iteration**, not regeneration:
   - amend surgically and preserve valid completed checkboxes;
   - preserve `execution_mode: auto`; invoking this command on a manual plan explicitly converts
     it to auto and requires revising its gates for unattended execution;
   - scope or behavior changes update `spec.md` first; approach-only changes update `plan.md`;
   - re-check the E2E path and every affected phase after the edit.
7. Stamp, lint, and fix all findings:

       respec stamp <change-dir>
       respec lint <change-dir>

   If outside the worked-on repo named by `research.md`'s `repo_path`, use:

       respec stamp --repo <worked-on-repo> <change-dir>

Frontmatter you write: `spec.md` → `title`, `status` (`draft`|`approved`), optional `tags`;
`plan.md` → `title`, `status` (`draft`|`approved`|`in-progress`|`done`),
`execution_mode: auto`, and optional `depends_on`. `respec stamp` owns `date` and `spec_sha256`.

Use the standard required sections from the respec skill. In every phase:

    ### Automated Verification
    - [ ] <deterministic command and expected result>

    ### Manual Verification
    - [ ] Orchestrator: <specific inspection or behavior check the agent can perform>

**Manual** means judgment is required, not that a human operator must do it. Assign all feasible
manual checks to the orchestrator. An `Operator:` item is allowed only for an irreducibly external
condition and must not become a surprise intermediate gate — resolve its prerequisite during this
planning session. Put the end-to-end/smoke check in the earliest phase where it can catch an
integration failure and repeat it in the final phase when later work could regress it.

The plan must make implementation mechanical: exact files/symbols, intent, edge cases, commands,
expected outcomes, and phase boundaries. Write only into the central store.

End only after stamp and lint pass. Summarize artifact paths, phases, the E2E loop, assumptions you
resolved, and any remaining operator-only final acceptance. Do not ask for another planning pass
unless a blocker remains.
{{if .Context}}
## Shared context

{{.Context}}
{{end}}{{if .Rules.Spec}}
## Spec rules

{{.Rules.Spec}}
{{end}}{{if .Rules.Plan}}
## Plan rules

{{.Rules.Plan}}
{{end}}
