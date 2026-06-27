---
description: Research a topic into the respec central store (writes research.md)
argument-hint: "<topic>"
---
You are in the **Research** phase of the respec workflow (research → plan → implement).

Topic: $@

Central store (all artifacts live here, never in the worked-on repo):

    {{.Store}}

Load the detailed workflow before doing anything else: read the `respec` skill at
`~/.pi/agent/skills/respec/SKILL.md` (or invoke `/skill:respec`). Follow it.

Your job this phase:

1. Clarify the topic with the operator if it is ambiguous; offer interpretations.
2. Explore the current repo / relevant code as needed to ground the research in reality.
3. Choose a problem-space name and a short slug, then create the change directory:
   `{{.Store}}/<problem-space>/<YYYY-MM-DD-slug>/`.
4. Write `research.md` into that change directory with provenance frontmatter:
   - `date` (today, with timezone offset)
   - `repo` (the worked-on repo name)
   - `git_commit` (current HEAD of the worked-on repo)
   - `topic`
5. Capture findings as prose with concrete file:line references where relevant.

Do **not** write anything into the worked-on repo. Every file you create goes under the
central store path above.
{{if .Context}}
## Shared context

{{.Context}}
{{end}}{{if .Rules.Research}}
## Research rules

{{.Rules.Research}}
{{end}}
