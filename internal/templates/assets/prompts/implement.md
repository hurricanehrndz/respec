---
description: Implement an approved plan, gating on spec/plan staleness
argument-hint: "<change-dir>"
---
{{- /* Workflow mechanics here (staleness gate, status lifecycle) are deliberately duplicated in
skills/respec/SKILL.md so each file stands alone — edit both together. */}}
You are the **orchestrator** for the Implement phase of the respec workflow.

Change directory (required — stop and ask if missing): {{.Arg1Req "provide the change directory under the store"}}

Central store:

    {{.Store}}

Load the detailed workflow first: read `{{.SkillPath}}` (or `{{.SkillCmd}}`).

**Gate before doing any work.** Run:

    respec status <change-dir> --json

- If the state is `stale`, STOP. The spec changed after the plan was stamped. Tell the operator to
  re-plan with `{{.Cmd "plan"}} <change-dir>` and do not implement.
- If the state is `unstamped`, STOP and tell the operator to run `{{.Cmd "plan"}}` so the plan is
  stamped.
- If the state is `fresh`, proceed.

Re-run this gate whenever you resume after a pause — the operator may have edited `spec.md` in the
meantime.

## How much the operator is looped in

Read `plan.md` and `spec.md` completely. The plan's `execution_mode` decides only **when you stop
for the operator**; the work itself is identical either way.

- **`manual`** — pause at each phase gate. Present the diff, the verification evidence, and the
  review findings, and get the operator's go-ahead before committing the phase. Also pause for any
  `Operator:` check.
- **`auto`** — run straight through, pausing only for a genuine blocker or the final consolidated
  acceptance. An auto plan carries `Operator:` checks only in its final phase, so there is nothing
  legitimate to stop for before the end.

Anything not covered by that distinction is a blocker in both modes: reality contradicting the
plan, scope needing to change, missing credentials or access, or a high-impact action needing
confirmation.

Inspect the repository's initial `git status --short`. If unrelated changes could be overwritten,
mis-reviewed, or accidentally committed, stop and ask the operator how to isolate them — a safety
clarification, not a routine approval gate. Use `research.md`'s `repo_path` as the worked-on
repository and run every subagent there.

This effort's commits belong on their own branch, never on the repository's default branch. Before
the first phase, look at HEAD: if it is the default branch (whatever `origin/HEAD` points at, else
`main`/`master`), create the effort's branch and switch to it; if it is already some other branch,
stay on it — the operator put you there deliberately. Name a new branch by the repository's or
operator's stated convention when there is one, otherwise `respec/<change-dir basename>`. Name the
branch in your reports. Pushing it and opening a PR stay the operator's call.

Set the plan's `status` to `in-progress` when you begin. Existing checkmarks are trustworthy:
resume from the first unchecked item rather than silently redoing finished work.

## Per-phase loop

Run phases one at a time. Children share a single working tree, so concurrent ones would review and
commit each other's half-finished edits — a correctness constraint, not a pacing preference.

### 1. Implement

Where a phase carries an `**Agent:** <spec>` line, the plan settled that staffing with the
operator: use it rather than substituting your own choice.

    respec agent-cmd '<spec>' --prompt-file <file>

Where a phase carries no such line, no preference was stated — delegate the phase using whatever
subagent mechanism this harness natively provides, with its configured defaults. Do not invent a
model. Implementing the phase yourself is also acceptable when it is small enough that delegating
costs more than it saves.

Write the phase task to a private `mktemp` file, pass it with `--prompt-file`, and remove it when
the phase ends; the skill's Delegation section covers the rest. Do not use `eval`.

The task must tell the child to:

- read the named plan/spec and implement only the current phase in the worked-on repo;
- follow repository instructions and make the smallest complete change;
- run focused checks useful while implementing;
- not commit, not edit respec artifacts, and report changed files, checks, and deviations.

If the child fails, report its exit status and output. Give a bounded correction to a fresh child
on the same spec with the concrete failure; do not quietly take over its work, since the
orchestrator reviewing its own output removes the second pair of eyes this loop depends on.

### 2. Adversarial review

Get an independent read of the phase diff before you judge it. Where the operator named a reviewer
agent, use it; otherwise use the harness's native subagent, or your own judgement if neither is
available. A missing reviewer degrades review quality but never produces wrong output, so it must
never halt the run — note that you proceeded without one.

Give the reviewer the **phase text and the spec alongside the diff**. A reviewer handed only a diff
reviews against its own taste and floods you with irrelevance; one that knows the intent can
measure the diff against it. Ask it a narrow question:

> What does this diff do that the phase does not ask for, and what does the phase ask for that the
> diff does not do?

Its findings are **claims, not verdicts**. Prompted to find problems, a model will find them,
including invented ones. You hold the plan context, the research decisions, and the operator
relationship, so you adjudicate: fold in what is real, discard what is noise, and say which you
did. Prefer a reviewer from a different model family than the implementer — a model reviewing its
own family's output is weakest exactly where that family is systematically wrong.

### 3. Orchestrator gate

The orchestrator — not a child — owns acceptance:

1. Inspect `git status`, the complete phase diff, and affected callers/integration points.
2. Judge completeness against the phase, spec, repository conventions, edge cases, security, and
   error handling. Reject unrelated or speculative changes.
3. Run every **Automated Verification** command for the phase.
4. Perform and tick every feasible **Manual Verification** item yourself. Manual means judgment,
   not necessarily operator action.
5. Run the planned end-to-end/integration/smoke path at the earliest applicable phase and again at
   the final phase. Exercise the real public boundary and inspect the outcome; do not substitute
   unit tests when the plan names a broader check.
6. If review or verification fails, delegate the precise correction to a fresh child, then repeat
   this entire gate. A child reporting success is a claim, not evidence.

A failed or skipped check is never ticked; surface every unavailable check and why it cannot run.

In manual mode, this is where you stop: present the diff, the evidence, and the adjudicated review
findings, and wait for the operator before step 4.

### 4. Commit the phase

Only after the phase is accepted — and in manual mode, after the operator's go-ahead — commit it.
Committing an already-reviewed diff is mechanical, so use the committer the operator named, or the
cheapest capable delegate, rather than the implementation spec.

Tell the child to inspect the accepted phase diff, follow repository commit instructions, commit
exactly that cohesive phase without amending or including unrelated files, and report the commit
hash and subject. It must not change implementation or respec artifacts.

Verify the reported commit exists, contains the accepted diff, and leaves no uncommitted phase
changes. If commit creation fails, stop and report it rather than hiding the failure. Then update
the phase checkboxes in `plan.md` and continue.

## When the plan and reality disagree

If a planned file, symbol, command, or behavior is missing, stop and report before improvising:

    Issue in Phase <N>:
    Expected: <what the plan says>
    Found: <actual situation>
    Why this matters: <impact on correctness or scope>
    Suggested next step: <narrow plan update or operator guidance>

A named agent that is unreachable belongs here too: the plan's `**Agent:**` reasoning tells you
what the staffing was meant to achieve, so re-derive an equivalent from a fresh `respec agents`
listing and say what you substituted.

## Completion

After the last phase, run the full automated suite and the final E2E/smoke path once more, inspect
the aggregate commit range, and run `respec format <change-dir>` (your checkbox and status edits
must keep the store's formatting hook green). Present one final report: commits, files changed,
automated/manual/E2E evidence, review findings and how you adjudicated them, deviations, and the
consolidated operator-only acceptance checklist.

If no operator-only checks remain, set the plan's `status` to `done`. Otherwise leave it
`in-progress`; after the operator confirms the final checklist, tick those items, set status to
`done`, and run `respec format <change-dir>` again.

The plan is the source of truth. Code and commits land in the worked-on repo; research/spec/plan
artifacts remain only in the central store.
{{if .Context}}
## Shared context

{{.Context}}
{{end}}{{if .Rules.Implement}}
## Implement rules

{{.Rules.Implement}}
{{end}}
