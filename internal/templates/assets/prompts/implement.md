---
description: Implement an approved plan, gating on spec/plan staleness
argument-hint: "<change-dir>"
---
{{- /* Workflow mechanics here (staleness gate, status lifecycle) are deliberately duplicated in
skills/respec/SKILL.md so each file stands alone — edit both together. */}}
You are in the **Implement** phase of the respec workflow (research → plan → implement).

Change directory (required — stop and ask if missing): {{.Arg1Req "provide the change directory under the store"}}

Central store:

    {{.Store}}

Load the detailed workflow first: read `{{.SkillPath}}` (or `{{.SkillCmd}}`).

**Gate before doing any work.** Run:

    respec status <change-dir> --json

- If the state is `stale`, STOP. The spec changed after the plan was stamped. Tell the operator to
  re-plan with `{{.Cmd "plan"}} <change-dir>` and do not implement.
- If the state is `unstamped`, STOP and tell the operator to run `{{.Cmd "plan"}}` so the plan is stamped.
- If the state is `fresh`, proceed.

Re-run this gate whenever you resume after a pause (e.g. after a Manual Verification stop) — the
operator may have edited `spec.md` in the meantime.

When fresh:

1. Read `plan.md` and `spec.md` in the change directory completely. If the plan already has
   checkmarks, trust them: resume from the first unchecked item and do not silently redo
   completed work.
2. Set the plan's `status` to `in-progress` when you begin.
3. Execute the plan's phases in order, making the smallest changes that satisfy each phase.
   **Trust the plan** — do not re-search what it already documents; search only when the plan is
   ambiguous, reality mismatches it, or verification requires it.
4. For each phase, run its **Automated Verification** commands and tick those checkboxes in
   `plan.md` as they pass. If a command fails, debug within the phase's scope; if the fix would
   change the plan, stop and ask.
5. At each phase boundary, stop at the **Manual Verification** items: do not tick them yourself —
   report which automated checks passed and which manual checks await, then pause for operator
   confirmation, unless told to run consecutively.

If a planned file, symbol, command, or behavior is missing, stop and report before improvising:

    Issue in Phase <N>:
    Expected: <what the plan says>
    Found: <actual situation>
    Why this matters: <impact on correctness or scope>
    Suggested next step: <narrow plan update or operator guidance>

When the final phase's Manual checks are confirmed, set the plan's `status` to `done`, run
`respec format <change-dir>` (your checkbox and status edits must keep the store's formatting
hook green), and summarize: files changed, verification performed, manual checks still pending,
and any deviations from the plan.

The plan is the source of truth. Code changes land in the worked-on repo; plan/spec/research edits
land only in the central store change directory.
{{if .Context}}
## Shared context

{{.Context}}
{{end}}{{if .Rules.Implement}}
## Implement rules

{{.Rules.Implement}}
{{end}}
