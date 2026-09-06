---
name: rsx-research
description: "Research and align a respec change: motivation, code evidence, options, decisions, and research.md."
disable-model-invocation: true
---

> **Read the respec skill at `{{.SkillPath}}` before doing anything else.** It contains the
> shared store layout, stamp and lint mechanics, delegation policy, and hard rules.

# Research a change

The topic is the argument you were invoked with, delivered as a trailing `User:` message. Ask for
it if it is missing.

Research drives toward alignment, not neutral archaeology. Explore the topic, form an
evidence-based view, and converge with the operator on the decisions the plan will use.

Read every file the operator names in full before branching out. Work the topic as a few concrete
questions: where the relevant components live, how the code works, and how this repository already
solves similar problems. Ground every finding in code you read, with `file:line` references.
External tools, APIs, and prior art may also be evidence. Finding that almost nothing exists yet is
still a useful result.
Delegate broad reading as described in the respec skill. Keep the operator conversation and final
synthesis here. Distill every durable finding into `research.md`; the context window is disposable,
but the artifact is not. If context runs low, record remaining work under **Open Questions** and
continue in a fresh session from the artifact.

## Reach alignment

Record the triggering problem or opportunity, why the change matters, and why it matters now. Ask
instead of inferring when the motivation is unclear.

Keep evidence honest. Separate facts verified by reading or running code from claims in docs,
comments, or vendor material. Prefer first-hand observation for external tools and services. When
a material point remains unclear, verify it, ask, or record it under **Open Questions**. Never
present an assumption as a finding.

Take a position. When several paths exist, describe their pros and cons and recommend one with your
reasoning. If the evidence does not favor one, say so and record the tradeoff. Bring the operator
questions that inspection cannot settle, such as intent, priority, or risk tolerance. Questions do
not replace investigation. Record each operator decision and its reasoning under **Decisions** so
the plan treats it as settled.

## Write research.md

Create or resume the effort directory using the respec skill's store rules, then write
`research.md` there. Use its frontmatter schema and this body:

    # Research: <topic>

    ## Research Question
    <what we are trying to understand>

    ## Motivation
    <why the change is requested or initiated, and why now>

    ## Summary
    <the short version of what you learned>

    ## Findings
    <what exists today, with file:line references>

    ## Options & Tradeoffs
    <candidate paths, pros and cons, and your recommendation; omit only when there is one path>

    ## Decisions
    <the operator's calls and reasoning; omit until a decision exists>

    ## Open Questions
    <anything unresolved after alignment>

`respec lint` requires the headings **Research Question**, **Summary**, **Findings**, and **Open
Questions** verbatim.

The artifact is complete when a fresh planning session can read only `research.md` and its code
references, then produce an informed plan without redoing the research or re-asking anything
settled. Mark its status `complete` only at that point.

Stamp, lint, and fix every finding:

    respec stamp <change-dir>
    respec lint <change-dir>

End with the artifact path and a short account of findings, decisions, and open questions. Ask the
operator to review `research.md` before starting a fresh `{{.Cmd "plan"}}` session.
{{if .Rules.Research}}
## Research rules

{{.Rules.Research}}
{{end}}
