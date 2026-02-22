# Visualizer Spec Review

Review the visualizer from a product and user experience perspective. Identify feature gaps, UX improvements, and missing spec coverage.

This review asks "does the visualizer serve its users well?" — not "is the code well-written?" Code quality belongs in `/project:review-visualizer`. Implementation planning happens in `/project:address-review`.

## Context

Before starting, read:
- `tools/visualizer/spec/TREE_VIEW.md` — Tree View requirements and current behavior
- `tools/visualizer/spec/GRAPH_VIEW.md` — Graph View requirements and planned behavior
- `tools/visualizer/spec/VIZUALIZER_PRODUCT_REVIEW_PLAN.md` — product evaluation: framing, user question hierarchy, gap analysis, product patterns, priority tiers

The product eval plan is the primary input. It defines the user questions the visualizer must answer, the UX patterns that should be applied, and the priority tiers. This review evaluates whether the spec and current behavior deliver on those goals.

## Review Rubric

### 1. User Questions
- For each user question in the product eval plan, can the visualizer currently answer it? What's missing?
- Are there common Temporal workflow questions a developer would ask that no current or specced feature addresses?

### 2. Spec Completeness
- Are there behaviors in the view docs that have no current implementation, and no clear path in the spec?
- Are there current behaviors with no spec coverage (undocumented features or accidental behavior)?
- Where the spec is silent, what should it say?

### 3. UX Patterns
- For each product pattern in the eval plan (glanceable summaries, progressive disclosure, focus+context, etc.), what experiences does it imply for each view?
- Are those experiences described in the spec? Do they compose coherently with each other?
- Are there UX patterns missing from the spec that would obviously serve the user questions?

### 4. Feature Coherence
- Do the specced features form a coherent product, or are there gaps between them?
- Are there features that conflict, overlap, or undermine each other?
- What's the simplest set of features that would answer all Tier 1 user questions?

## Workflow

### Phase 1: Explore

Use sub-agents in parallel:
- **Spec agent**: Read all three context documents. Build an inventory: every user question, every specced feature, every gap, every product pattern from the docs.
- **Current state agent**: Read all TypeScript in `tools/visualizer/src/`. For each component, note only: what it currently renders and what interactions it supports. Do not evaluate code quality.

### Phase 2: Catalog

For each user question and specced feature, assess:
- **Status**: `answered` | `partial` | `unanswered` | `data unavailable` (parser doesn't emit needed fields)
- **Gap**: what experience or information is missing
- **Tier**: which priority tier from the product eval plan

For each gap, describe it in product terms — what the user experiences, not what the code does. Flag any gaps where the missing piece is upstream data from the parser (describe the data shape needed, nothing more).

Drop fully-answered items. Focus on gaps.

### Phase 3: Group & Prioritize

Group findings into **product feature sets**, ordered by:
1. Priority tier (Tier 1 before Tier 2, etc.)
2. Dependency order (foundational UX before features that build on it)
3. Coherence (group features that compose into a single user experience together)

Each group should have:
- Feature set name
- User questions it answers
- Description of the target experience (what the user sees and does)
- Feasibility note: is the data available from the parser? Does it require new spec additions?

### Phase 4: Write to VISUALIZER_REVISIONS.md

Write the grouped plan to `VISUALIZER_REVISIONS.md` at the repo root:
- One `## Group N: Title` section per feature set
- Each group has: User questions addressed, Target experience, Data requirements, Spec additions needed
- Include a summary at the top: what's working, what's missing, what's blocked on parser data

**STOP after writing the file. Present a summary and wait for approval. To execute, invoke `/project:address-review`.**

## Constraints
- **Product lens only.** Describe gaps in terms of user experience and features, not TypeScript components or file changes. Implementation is `/project:address-review`'s job.
- **The product eval plan is authoritative.** Don't re-derive the user question hierarchy or priority tiers. Use what's in the plan. If you disagree with a tier assignment, flag it but don't override it.
- **TypeScript is for current-state only.** Consult the implementation to understand what already exists. Don't use it to plan what should be built.
- **Flag missing data, not parser work.** If a gap requires JSON fields the parser doesn't emit, note the missing data shape and move on.
- **No backwards compatibility.** Pre-v1. Propose the right experience.
