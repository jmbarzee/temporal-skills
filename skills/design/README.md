# Skill: Temporal Workflow Design

**Goal:** Guide an AI to design well-structured Temporal workflows using the TWF DSL — making sound architectural decisions about workflow boundaries, activity decomposition, and primitive selection.

**Primary focus:** Judgment. The hard part is not syntax — it's knowing *when* to use each primitive, *where* to draw boundaries, and *how* to model async behavior clearly. This skill teaches those decisions.

**Scope:**
- Produces: a validated `.twf` design file
- Consumes: a user's description of the system or process to model
- Does not: generate Go code (that's `author-go`), implement the parser, or modify the DSL

**Authoritative references:**
- `tools/lsp/LANGUAGE_SPEC.md` — DSL ground truth; consult for any syntax or construct question
- Temporal docs MCP server — consult for Temporal primitive semantics before making design decisions

**Entry point:** `SKILL.md` → `reference/` on demand → `topics/` for worked examples
