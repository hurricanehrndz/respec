---
description: Implement an approved plan, gating on spec/plan staleness
argument-hint: "<change-dir>"
---
You are in the **Implement** phase of the respec workflow (research → plan → implement).

Change directory: ${1:?provide the change directory under the store}

Central store:

    {{.Store}}

Load the detailed workflow first: read `~/.pi/agent/skills/respec/SKILL.md` (or `/skill:respec`).

**Gate before doing any work.** Run:

    respec status <change-dir> --json

- If the state is `stale`, STOP. The spec changed after the plan was stamped. Tell the operator to
  re-plan with `/rsx:plan <change-dir>` and do not implement.
- If the state is `unstamped`, STOP and tell the operator to run `/rsx:plan` so the plan is stamped.
- If the state is `fresh`, proceed.

When fresh:

1. Read `plan.md` and `spec.md` in the change directory completely.
2. Execute the plan's phases in order, making the smallest changes that satisfy each phase.
3. Tick checkboxes in `plan.md` as each phase's automated verification passes.
4. Pause for manual verification at each phase boundary unless told to run consecutively.

The plan is the source of truth. Code changes land in the worked-on repo; plan/spec/research
edits land only in the central store change directory.
{{if .Context}}
## Shared context

{{.Context}}
{{end}}{{if .Rules.Implement}}
## Implement rules

{{.Rules.Implement}}
{{end}}
