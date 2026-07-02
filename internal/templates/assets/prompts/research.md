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

Research here is **interview-driven context-building**, not neutral archaeology. The goal is to
gather enough shared context that the plan phase can proceed efficiently. So:

1. **Interview the operator.** Clarify the topic; offer interpretations when it is ambiguous. Ask
   the questions that code inspection cannot answer, and do not guess on decisions that change
   direction.
2. **Explore the code** to ground every finding in reality; capture concrete `file:line` references.
   Read any files the operator names **in full, first**, then work the topic as a few distinct
   research questions: *where* do the relevant files and components live, *how* does the specific
   code actually work, and *how* does this codebase already solve similar problems. **Default to
   exploring in the open, in this session** — the operator steers the exploration live (it is an
   interview), and everything worth keeping is distilled into `research.md`, the artifact that
   hands later sessions their context. Your context window is disposable; the artifact is not.
   Before reaching for subagents, gauge the scale (count and size the relevant files/dirs): only
   when the codebase is large enough that the bulk reading would drown this session, delegate the
   bulk reads — and keep the interview and synthesis here either way. An agent without a subagent
   feature can spawn a fresh instance of itself via the shell. If you run low on context, record
   what remains under **Open Questions** and continue in a fresh session that starts from the
   artifact.{{if .HasProbe}}
   `probe` is installed — prefer it over plain grep-and-read for exploration; it returns whole
   semantic blocks (functions, classes) ranked by relevance, which keeps context small:

       probe search "<terms>" <path>    # Elasticsearch-style query: AND / OR / NOT
       probe extract <file>#<symbol>    # pull one function/class by name
       probe extract <file>:<line>      # pull the block containing a line
{{end}}
3. **Weigh options.** When more than one path exists, lay out the candidates with pros and cons.
   Surface tradeoffs rather than hiding them.
4. **Record decisions.** When the operator leans toward an approach, write it down with the
   reasoning so the plan phase treats it as a settled constraint.

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
    <candidate paths, each with pros and cons — omit if there is genuinely one path>

    ## Decisions
    <what the operator leaned toward, and why — omit until something is decided>

    ## Open Questions
    <anything still unresolved after the interview>

`respec lint` validates the section headings **Research Question**, **Summary**, **Findings**, and
**Open Questions** verbatim — keep those names even when customizing this template.

Write for a **human co-engineer**, not just for agent execution. Where structure, flow, or
sequencing is easier to see than to read, use a Mermaid diagram (a fenced ` ```mermaid ` block) —
GitHub and `respec render`/`serve` both render them, and diagrams-as-text diff cleanly. Inline HTML
is also fine when it adds value; the store render preserves it.

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
