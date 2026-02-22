# Visualizer Spec Review

Review the visualizer against its view docs and product evaluation plan. Translate spec gaps into grouped TypeScript changes.

This review asks "does the visualizer do what it should?" — not "is the code well-written?" Code quality belongs in `/project:review-visualizer`.

## Context

Before starting, read:
- `tools/visualizer/spec/TREE_VIEW.md` — Tree View requirements and current behavior
- `tools/visualizer/spec/GRAPH_VIEW.md` — Graph View requirements and planned behavior
- `tools/visualizer/spec/VIZUALIZER_PRODUCT_REVIEW_PLAN.md` — product evaluation: framing, user question hierarchy, gap analysis, product patterns, priority tiers

The product eval plan is the primary input. It defines the user questions the visualizer must answer, the gaps between current and needed behavior, and the priority tiers. This review translates that into concrete work.

## Review Rubric

### 1. Spec Coverage
- For each gap in the product eval plan, does the TypeScript implement it, partially implement it, or not at all?
- Are there view doc requirements that the current components don't address?
- Are there components that implement behavior not described in any spec?

### 2. Feasibility
- For gaps that require new data, does the parser's current JSON output already provide it? If not, note the missing data shape — don't trace it into the parser's internal roadmap.
- For gaps that require new components, do existing patterns in the codebase support them? Or is new architecture needed?

### 3. Interaction Design
- For each product pattern in the eval plan (glanceable summaries, progressive disclosure, focus+context, etc.), what would the TypeScript change look like?
- Which existing components need modification vs. which are new?
- How do proposed interactions compose with existing behavior?

## Workflow

### Phase 1: Explore

Use sub-agents in parallel:
- **Spec agent**: Read all three context documents. Build a checklist: every requirement, gap, and product pattern application from the docs.
- **Implementation agent**: Read all TypeScript in `tools/visualizer/src/`. For each component, note what it renders, what data it consumes, and what interactions it supports.

### Phase 2: Catalog

Match the spec checklist against the implementation inventory. For each spec item:
- **Status**: `implemented` | `partial` | `missing` | `data unavailable` (parser doesn't emit needed fields today)
- **Location**: which component(s) are relevant
- **Gap**: what's missing or different from the spec
- **Tier**: which priority tier from the product eval plan

Drop items that are fully implemented. Focus on gaps.

### Phase 3: Group & Prioritize

Group findings into **actionable TypeScript change groups**, ordered by:
1. Priority tier (Tier 1 before Tier 2, etc.)
2. Dependency order (foundational changes before features that build on them)
3. Parallelism (independent changes can be separate groups)

Each group should have:
- Theme name
- List of spec gaps it addresses
- Files affected
- Whether it depends on data the parser doesn't yet provide (describe the missing data shape, not the parser work)

**STOP. Present the grouped plan and wait for approval. To execute, invoke `/project:address-review`.**

## Constraints
- **Spec lens only.** Don't audit code quality, type safety, or architecture — that's `/project:review-visualizer`.
- **The product eval plan is authoritative.** Don't re-derive the user question hierarchy or priority tiers. Use what's in the plan. If you disagree with a tier assignment, flag it but don't override it.
- **Flag missing data, not parser work.** If a spec gap requires JSON fields the parser doesn't currently emit, note the missing data shape and move on. Don't trace blockers into parser internals or reference parser roadmap documents.
- **No backwards compatibility.** Pre-v1. Propose clean implementations.
