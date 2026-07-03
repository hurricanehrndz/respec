# RPI recipes (external reference)

Third-party reference material, kept unmodified as design provenance — **not
part of respec**.

These are [Goose](https://github.com/block/goose) recipes implementing the
Research → Plan → Implement (RPI) context-engineering pattern (originating
from HumanLayer; recipe author `angiejones` per the YAML metadata), plus the
accompanying tutorial page (`rpi-readme.md`, captured from the Goose docs
site).

respec's embedded prompt templates (`internal/templates/assets/`) derived
their document structure from these recipes: the phased plan skeleton with
dual **Automated/Manual Verification**, "What We're NOT Doing", and the
research doc shape (Summary / Findings / Open Questions). Where respec
deviates deliberately — alignment-driven, opinionated research (evidence-based
recommendations, Options & Tradeoffs, a Decisions log) instead of RPI's
"no opinions" rule; the CLI stamping provenance instead of the agent;
scale-directed exploration instead of RPI's mandatory parallel sub-recipe
fan-out (default is in the open within the dedicated research session,
because the artifact is the context handoff and mid-session subagents cost
the operator observability and steerability during the live exploration;
fan-out only when codebase size makes bulk reading drown the session) —
the reasoning lives in the repo history and CLAUDE.md.

Consult these when iterating on the prompt templates; do not edit them.
