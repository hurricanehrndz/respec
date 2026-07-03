---
description: Research a topic into the respec central store (writes research.md)
argument-hint: "<topic>"
---
You are in the **Research** phase of the respec workflow (research → plan → implement).

Topic: {{.AllArgs}}

Central store (all artifacts live here, never in the worked-on repo):

    {{.Store}}

Load the detailed workflow before doing anything else: read the `respec` skill at
`{{.SkillPath}}` (or invoke `{{.SkillCmd}}`). Follow it.

Research here **drives toward alignment**, not neutral archaeology. You explore, form an
evidence-based view, and converge with the operator on the decisions the plan must stand on. So:

1. **Explore the code first** to ground every finding in reality; capture concrete `file:line`
   references. Read any files the operator names **in full, first**, then work the topic as a few
   distinct research questions: *where* do the relevant files and components live, *how* does the
   specific code actually work, and *how* does this codebase already solve similar problems. The
   ground truth may also be external — candidate tools, APIs, prior art — and "almost nothing
   exists in the codebase yet" is itself a finding. **Default to exploring in the open, in this
   session** — the operator steers the exploration live, and everything worth keeping is distilled
   into `research.md`, the artifact that hands later sessions their context. Your context window
   is disposable; the artifact is not. Before reaching for subagents, gauge the scale (count and
   size the relevant files/dirs): only when the codebase is large enough that the bulk reading
   would drown this session, delegate the bulk reads — and keep the dialogue and synthesis here
   either way. An agent without a subagent feature can spawn a fresh instance of itself via the
   shell. If you run low on context, record what remains under **Open Questions** and continue in
   a fresh session that starts from the artifact.{{if .HasProbe}}
   `probe` is installed — prefer it over plain grep-and-read for exploration; it returns whole
   semantic blocks (functions, classes) ranked by relevance, which keeps context small:

       probe search "<terms>" <path>    # Elasticsearch-style query: AND / OR / NOT
       probe extract <file>#<symbol>    # pull one function/class by name
       probe extract <file>:<line>      # pull the block containing a line
{{end}}
2. **Keep evidence honest.** Separate what you verified by reading code or probing first-hand
   from what a doc, comment, or vendor claims. For external tools and services, prefer running
   the thing over restating its README.
3. **Ask what evidence cannot answer.** Bring the operator the questions inspection cannot
   settle — intent, priorities, tradeoff calls — and offer interpretations when the topic is
   ambiguous. Questions are not a substitute for investigation.
4. **No silent assumptions.** When something material is not immediately obvious: verify it; if
   you cannot verify it, ask; if it stays unresolved, record it under **Open Questions** as an
   explicit unknown. Never present an assumption as a finding.
5. **Take a position.** When more than one path exists, lay out the candidates with pros and
   cons and recommend one, with reasoning. If the evidence genuinely favors neither, say it is a
   coin flip — a recorded coin flip is still alignment.
6. **Record decisions.** When the operator makes the call, write it into **Decisions** with the
   reasoning so the plan phase treats it as settled.

Create the change directory under the store, grouped by the **primary repo** you are working in
and named by a short slug:

    {{.Store}}/<owner-repo>/<slug>/

Derive `<owner-repo>` from the worked-on repo's `origin` remote so every effort for the same repo
lands together: take the last two path segments (`owner/repo`), drop a trailing `.git`, then
slugify (lowercase, runs of non-alphanumerics → hyphens). Examples: `git@github.com:hurricanehrndz/respec.git`
→ `hurricanehrndz-respec`; `https://github.com/Acme/Web-App.git` → `acme-web-app`. With no `origin`,
use the repository directory's basename. This is the one git value you read yourself (just to name
the directory) — everything else is stamped.

Before creating it, run `respec list` and check for an existing effort with the same slug **for this
repo**. If one exists, ask the operator whether this is a new version; if so, use a distinct slug
like `<slug>-v2` (two efforts for one repo cannot share a slug).

Then write `research.md` in that directory. Frontmatter you write: `topic`, `status` (`draft` until
the interview settles, then `complete`), and optional `tags`. **Do not** hand-write the provenance
fields — `respec stamp` fills `date`, `repo`, `repo_path`, and `git_commit` for you.

Body skeleton:

    # Research: <topic>

    ## Research Question
    <what we are trying to understand>

    ## Summary
    <the short version of what you learned>

    ## Findings
    <what exists today, with file:line references>

    ## Options & Tradeoffs
    <candidate paths with pros/cons and your recommendation — omit if there is genuinely one path>

    ## Decisions
    <the operator's call and the reasoning — omit until something is decided>

    ## Open Questions
    <anything still unresolved after the interview>

`respec lint` validates the section headings **Research Question**, **Summary**, **Findings**, and
**Open Questions** verbatim — keep those names even when customizing this template.

Write for a **human co-engineer**, not just for agent execution. Where structure, flow, or
sequencing is easier to see than to read, use a Mermaid diagram (a fenced ` ```mermaid ` block) —
GitHub and `respec render`/`serve` both render them, and diagrams-as-text diff cleanly. Inline HTML
is also fine when it adds value; the store render preserves it.

Quality bar: the artifact is complete when a fresh planning session, reading only `research.md`
and the code it references, could produce an informed plan without redoing the research or
re-asking the operator anything already settled.

When the document is written, stamp it (fills provenance, reflows prose), then lint and fix any
findings:

    respec stamp <change-dir>
    respec lint <change-dir>

Do **not** write anything into the worked-on repo. Every file goes under the central store.

End by summarizing the artifact path and the key findings, decisions, and open questions — and
remind the operator to review `research.md` before starting a fresh plan session.
{{if .Context}}
## Shared context

{{.Context}}
{{end}}{{if .Rules.Research}}
## Research rules

{{.Rules.Research}}
{{end}}
