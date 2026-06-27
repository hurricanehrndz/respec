---
description: Co-generate spec.md + plan.md from research, then stamp the plan
argument-hint: "[change-dir]"
---
You are in the **Plan** phase of the respec workflow (research → plan → implement).

Change directory (defaults to the most recent under the store if omitted): ${1:-<pick latest>}

Central store:

    {{.Store}}

Load the detailed workflow first: read `~/.pi/agent/skills/respec/SKILL.md` (or `/skill:respec`).

Your job this phase:

1. Read `research.md` in the target change directory.
2. **Co-generate** `spec.md` and `plan.md` together in that same change directory, sharing one
   deliverables list expressed as prose:
   - `spec.md` — the requirements / desired behavior.
   - `plan.md` — phased implementation with checkboxes and per-phase verification.
3. Ask clarifying questions and offer options before committing to an approach.
4. If a `plan.md` already exists and `respec status <change-dir>` reports **stale**, amend the plan
   **surgically** to match the changed spec rather than regenerating it from scratch.
5. After writing `spec.md` and `plan.md`, record the spec's hash into the plan's frontmatter by
   running:

       respec stamp <change-dir>

Write only into the central store change directory. Nothing goes into the worked-on repo.
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
