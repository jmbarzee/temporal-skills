# Skill: Temporal Go Author

**Goal:** Translate a validated TWF design into correct, idiomatic Temporal Go SDK code.

**Primary focus:** Translation fidelity. Design decisions are already made. The hard part is mapping DSL constructs to Go SDK patterns precisely — respecting determinism constraints, context propagation, error handling, and SDK idioms.

**Scope:**
- Produces: compilable Temporal Go SDK code
- Consumes: a validated `.twf` file produced by the `design` skill
- Does not: make workflow design decisions, modify the DSL, or produce code for other languages

**Authoritative references:**
- Temporal docs MCP server — consult for current Go SDK API, especially for patterns that change across SDK versions
- `tools/lsp/LANGUAGE_SPEC.md` — source of truth for DSL construct semantics when mapping to Go

**Entry point:** `SKILL.md` → `reference/` for specific construct mappings → consult Temporal docs MCP for SDK specifics
