---
name: rsx-implement
description: "Implement an approved respec plan: staleness gate, sequential phases, adversarial review, verification, and commits."
disable-model-invocation: true
---

> **Read the respec skill at `{{.SkillPath}}` before doing anything else.** It contains the
> shared store layout, stamp and lint mechanics, delegation policy, and hard rules.

# Implement a plan

The change directory is the argument you were invoked with, delivered as a trailing `User:`
message. It is required. Stop and ask for it if it is missing.

## Gate on plan status

Before any work, run:

    respec status <change-dir> --json

- If the state is `stale`, stop. The spec changed after the plan was stamped. Tell the operator to
  re-plan with `{{.Cmd "plan"}} <change-dir>`.
- If the state is `unstamped`, stop. Tell the operator to run `{{.Cmd "plan"}}` so the plan is
  stamped.
- Proceed only when the state is `fresh`.

Repeat this gate whenever work resumes after a pause because the operator may have edited
`spec.md`.

Read `plan.md` and `spec.md` completely. The plan's `execution_mode` controls when the operator is
looped in, not how the work is performed.

- In `manual` mode, pause at every phase gate. Present the diff, verification evidence, and review
  findings. Get the operator's approval before committing. Pause for every `Operator:` check.
- In `auto` mode, continue through all phases. Pause only for a blocker or final consolidated
  acceptance. An auto plan may have `Operator:` checks only in its final phase.

Both modes stop when reality contradicts the plan, scope must change, access is missing, the dirty
working tree is unsafe, or an action needs confirmation because it has high impact.

## Prepare the repository

Use `research.md`'s `repo_path` as the worked-on repository and run every child there. Inspect the
initial `git status --short`. If unrelated changes could be overwritten, mis-reviewed, or
committed, ask the operator how to isolate them.

Keep this effort's commits off the default branch. Resolve the default from `origin/HEAD`, falling
back to `main` or `master`. If HEAD is the default branch, create and switch to a branch. Follow the
repository's or operator's naming convention, or use `respec/<change-dir basename>`. If HEAD is
already another branch, stay there because the operator chose it. Report the branch name. Do not
push or open a pull request without the operator's request.

Set the plan status to `in-progress` when work starts. Trust existing checkmarks and resume at the
first unchecked item.

## Per-phase loop

Use the worker preferences recorded in `plan.md` for implementation, review, commits, and fallback.
Phase-specific choices override the plan-wide choices for that role. Do not reload standing agent
preferences from config or replace the plan's choices with installed skill defaults. If a plan
has no worker preferences, use native defaults while honoring any existing `**Agent:**` lines.

Run phases one at a time. Children share a working tree; concurrent work would mix incomplete
diffs, reviews, and commits.

### 1. Implement

Without an `**Agent:**` line, use native subagents under the shared delegation policy. Apply any
native-worker preferences from the plan and phase; otherwise use the environment's defaults.
Implement in the current session only when delegation costs more than the work or native subagents
are unavailable, subject to the plan's fallback. Report when no independent worker was used.

A phase's `**Agent:** <spec>` line names an external CLI. Honor it using the shared external CLI
delegation instructions, not by silently translating it to a native worker.

Tell the implementer to:

- read the named plan and spec, then implement only the current phase in the worked-on repository;
- follow repository instructions and make the smallest complete change;
- run focused checks while working;
- not commit or edit respec artifacts;
- report changed files, checks, and deviations.

If the child fails, report the failure and available diagnostics. Give a fresh child using the
same execution mechanism and assignment a bounded correction with the concrete failure. Do not
silently take over, since that removes the second pair of eyes used by the review loop.

### 2. Adversarial review

Get an independent reading of the phase diff before judging it. Apply the plan's reviewer choice
under the shared delegation policy. Otherwise use a native child. If no reviewer is available,
follow the plan's fallback; absent a stricter fallback, review it yourself and report that no
independent reviewer was used.

Give the reviewer the phase text, spec, and diff. Ask one narrow question:

> What does this diff do that the phase does not ask for, and what does the phase ask for that the
> diff does not do?

Treat findings as claims, not verdicts. Verify each claim against the plan, research decisions, and
code. Apply real findings, discard noise, and report both decisions. Prefer a reviewer from a
different model family when the chosen mechanism supports it.

### 3. Orchestrator gate

The orchestrator owns acceptance:

1. Inspect `git status`, the complete phase diff, and affected callers and integration points.
2. Compare the work with the phase, spec, repository conventions, edge cases, security needs, and
   error handling. Reject unrelated or speculative changes.
3. Run every **Automated Verification** command for the phase.
4. Perform and tick every feasible **Manual Verification** item. Manual means judgment, not
   necessarily operator action.
5. Run the planned E2E, integration, or smoke path at the earliest applicable phase and again in
   the final phase. Exercise its real public boundary; do not replace it with unit tests.
6. If review or verification fails, send the exact correction to a fresh child and repeat this
   gate. A child's success report is not evidence.

Never tick a failed or skipped check. Report each check that could not run and why. In manual mode,
present the diff, evidence, and adjudicated review findings here, then wait for approval before
committing.

### 4. Commit the phase

Commit only an accepted phase. In manual mode, wait for operator approval first. Apply the plan's
committer choice under the shared delegation policy. Otherwise use a native committer or commit in
the current session; do not choose an external CLI just to get a cheaper model.

Tell the committer to inspect the accepted diff, follow repository commit instructions, and commit
exactly that cohesive phase. It must not amend, include unrelated files, change implementation, or
edit respec artifacts. It reports the hash and subject.

Verify that the commit exists, contains the accepted diff, and leaves no uncommitted phase changes.
If commit creation fails, stop and report it. Update the phase checkboxes in `plan.md`, then
continue.

## When the plan and reality disagree

Do not improvise around a missing file, symbol, command, behavior, or agent. Stop and report:

    Issue in Phase <N>:
    Expected: <what the plan says>
    Found: <actual situation>
    Why this matters: <impact on correctness or scope>
    Suggested next step: <narrow plan update or operator guidance>

If a worker preference cannot be honored, follow the plan's fallback and report the limitation.
For an unreachable external assignment, run `respec agents` again. Use the `**Agent:**` reasoning
to find an equivalent model on the chosen executor and report the substitution. Ask before
switching executors. For native workers, use the environment's own availability and failure
signals, not the CLI roster.

## Complete the effort

After the last phase:

1. Run the full automated suite and final E2E, integration, or smoke path again.
2. Inspect the aggregate commit range.
3. Run `respec format <change-dir>` so checkbox and status edits pass the store formatting hook.
4. Report commits, files changed, automated and manual evidence, end-to-end evidence, review
   findings and their disposition, deviations, and the consolidated operator-only checklist.

If no operator-only checks remain, set the plan status to `done`. Otherwise leave it `in-progress`.
After the operator confirms the final checklist, tick those items, set status to `done`, and run
`respec format <change-dir>` again.

Code and commits belong in the worked-on repository. Research, spec, and plan artifacts remain in
the central store.
{{if .Rules.Implement}}
## Implement rules

{{.Rules.Implement}}
{{end}}
