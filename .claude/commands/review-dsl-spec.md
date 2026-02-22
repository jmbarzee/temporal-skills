# DSL Spec Review

Review the TWF language design against Temporal's actual primitives and patterns. Evaluate whether the DSL adequately represents the application domain — not whether the parser implements it correctly.

This review asks "does the DSL express Temporal well?" — not "does the parser handle it correctly?" Parser quality belongs in `/project:review-parser-internals`.

## Context

Before starting, read:
- `tools/lsp/LANGUAGE_SPEC.md` — the current formal grammar and semantics (the artifact under review)
- `AST_REVISIONS.md` — language changes in flight; don't re-report what's already planned

Use the **Temporal docs MCP server** (`mcp__temporal-docs__search_temporal_knowledge_sources`) as the authoritative reference for what Temporal offers. When evaluating whether the DSL covers a concept, search the docs — don't rely on memory.

## Review Rubric

### 1. Coverage
- Does the DSL have a representation for every major Temporal primitive: workflows, activities, signals, queries, updates, timers, promises, child workflows, Nexus operations, heartbeats, task queues, workers?
- Are there Temporal patterns the DSL cannot express at all, or can only express awkwardly?
- Are there constructs in the DSL that don't correspond to any real Temporal primitive?

### 2. Representation Quality
- For each construct, is the notation minimal and readable? Would a developer immediately understand what it describes?
- Are complex Temporal concepts (e.g., async fan-out, signal-driven state machines) expressible without verbose workarounds?
- Does the DSL's indentation-based structure help or hinder the representation of nested or parallel constructs?

### 3. Consistency
- Do similar Temporal concepts use similar syntactic patterns in the DSL?
- Are there constructs that use inconsistent naming, structure, or semantics relative to each other?
- Does the DSL's terminology align with Temporal's terminology, or does it introduce unnecessary divergence?

### 4. Expressiveness vs. Complexity
- Where does the DSL require too many lines to express a simple concept?
- Where does the DSL collapse too much into a single construct, hiding important distinctions?
- Is the level of abstraction appropriate — design-time clarity without losing semantic precision?

## Workflow

**Follow this phased approach strictly. Do not skip or combine phases.**

### Phase 1: Explore

Use sub-agents in parallel:
- **Temporal primitives agent**: Use the Temporal docs MCP server to enumerate Temporal's primitives, patterns, and SDK concepts. Build an inventory: for each concept, note its name, purpose, and any important nuances (e.g., update vs. signal semantics).
- **DSL agent**: Read `tools/lsp/LANGUAGE_SPEC.md` in full. Build a parallel inventory: for each DSL construct, note its syntax, semantics, and what Temporal concept it maps to.

### Phase 2: Catalog

Map the Temporal inventory against the DSL inventory. For each Temporal concept:
- **Coverage**: `full` | `partial` | `absent` | `misrepresented`
- **DSL construct**: which notation covers it (if any)
- **Gap**: what's missing, awkward, or semantically imprecise
- **Severity**: `critical` (common concept, no representation) | `moderate` (expressible but awkward) | `minor` (edge case or minor inconsistency)

Also flag DSL constructs with no clear Temporal mapping.

Cross-reference against `AST_REVISIONS.md`. Drop findings that are already planned.

### Phase 3: Evaluate Possible Features

Read `POSSIBLE_DSL_FEATURES.md`. For each proposed feature:
- Does it address a gap identified in Phase 2? If so, does the proposed approach resolve it well?
- Is it motivated by a real Temporal primitive or pattern, or is it speculative?
- Does it fit consistently with the existing DSL grammar and style?
- Assign: `validated` (addresses a real gap well) | `reconsidered` (addresses a gap but the approach needs work) | `deferred` (no gap found, premature) | `superseded` (the Phase 2 review suggests a better approach)

### Phase 4: Group & Prioritize

Group findings into **language change proposals**, ordered by:
1. Coverage gaps for high-frequency Temporal patterns first
2. Consistency fixes before new additions
3. Representation improvements before edge cases

Each group should have:
- Theme name
- List of gaps or inconsistencies it addresses
- Whether it requires a grammar change, a rename, or a new construct
- Whether it would be a breaking change to existing `.twf` files
- Any validated `POSSIBLE_DSL_FEATURES.md` proposals that belong here

**STOP. Present the grouped plan and wait for approval. To execute language changes, invoke `/project:address-review`.**

## Constraints
- **Spec lens only.** Don't review parser implementation, AST structure, or resolver behavior — those belong in `/project:review-parser-internals`.
- **Use the Temporal docs MCP server.** Don't evaluate coverage from memory alone. Search for each concept.
- **Possible features come last.** The Phase 2 review drives findings; Phase 3 validates proposals against those findings. Don't let the possible features list anchor the review.
- **No backwards compatibility.** Pre-v1. If a better representation exists, propose it. Note which changes would break existing `.twf` files.
- **Language changes require parser work.** Flag this in each group but don't plan the parser implementation here.
