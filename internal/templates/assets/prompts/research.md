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

Research here is **interview-driven context-building**, not neutral archaeology. The goal is to
gather enough shared context that the plan phase can proceed efficiently. So:

1. **Interview the operator.** Clarify the topic; offer interpretations when it is ambiguous. Ask
   the questions that code inspection cannot answer, and do not guess on decisions that change
   direction.
2. **Explore the code** to ground every finding in reality; capture concrete `file:line` references.
3. **Weigh options.** When more than one path exists, lay out the candidates with pros and cons.
   Surface tradeoffs rather than hiding them.
4. **Record decisions.** When the operator leans toward an approach, write it down with the
   reasoning so the plan phase treats it as a settled constraint.

Choose a problem-space name and a short slug, then create the change directory
`{{.Store}}/<problem-space>/<YYYY-MM-DD-slug>/` and write `research.md` there.

Frontmatter you write: `topic`, `status` (`draft` until the interview settles, then `complete`),
and optional `tags`. **Do not** run `git` or `date` or hand-write provenance — `respec stamp` fills
`date`, `repo`, `repo_path`, and `git_commit` for you.

Body skeleton:

    # Research: <topic>

    ## Research Question
    <what we are trying to understand>

    ## Summary
    <the short version of what you learned>

    ## Findings
    <what exists today, with file:line references>

    ## Options & Tradeoffs
    <candidate paths, each with pros and cons — omit if there is genuinely one path>

    ## Decisions
    <what the operator leaned toward, and why — omit until something is decided>

    ## Open Questions
    <anything still unresolved after the interview>

When the document is written, stamp it (fills provenance, reflows prose):

    respec stamp <change-dir>

Do **not** write anything into the worked-on repo. Every file goes under the central store.
{{if .Context}}
## Shared context

{{.Context}}
{{end}}{{if .Rules.Research}}
## Research rules

{{.Rules.Research}}
{{end}}
