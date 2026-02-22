# Visualizer Code Review

Review the React + TypeScript implementation in `tools/visualizer/src/`.

This review is about the *consumer* of the parser's JSON contract — the webview that renders Tree View and Graph View panels in VS Code / Cursor.

## Context

Before starting, read:
- `AST_REVISIONS.md` — understand the current parser contract, what's changing
- `tools/visualizer/spec/TREE_VIEW.md` — Tree View requirements and current implementation
- `tools/visualizer/spec/GRAPH_VIEW.md` — Graph View requirements and planned behavior

## Review Rubric

### 1. Contract Consumption
- Does the TypeScript correctly reflect the current JSON schema from `twf parse`?
- Are there fields being accessed that no longer exist, or fields being ignored that now carry data?
- Are discriminated unions (`kind`/`type` fields) handled exhaustively?

### 2. Architecture & Organization
- Component boundaries: is each component responsible for one thing?
- Data flow: is there a clear separation between data fetching, transformation, and rendering?
- Are there TypeScript types that duplicate or diverge from the Go JSON schema?

### 3. TypeScript Quality
- Type safety: are there `any` casts or unsafe assertions that hide real type errors?
- Null/undefined handling: are optional JSON fields handled defensively?
- Error boundaries: what happens when the parser returns errors or malformed output?

### 4. Node Coverage
- Does the rendering code handle every definition type in the JSON schema? Every statement type?
- Are there `kind`/`type` values that fall through to a default or are silently ignored?
- Are there JSON fields that are parsed but never used in either view?

## Workflow

**Follow this phased approach strictly. Do not skip or combine phases.**

### Phase 1: Explore

Use sub-agents to read all source in parallel:
- One agent for `tools/visualizer/src/` — all TypeScript files, evaluating against all four rubric lenses
- One agent for the context files: `tools/visualizer/spec/TREE_VIEW.md`, `tools/visualizer/spec/GRAPH_VIEW.md`, `AST_REVISIONS.md`

Agents should work from source code only. Do not run the visualizer or evaluate visual output — that is the domain of `/project:review-visualizer-spec`.

### Phase 2: Catalog

Each finding must include:
- **Location**: `file:function` or `file:line`
- **Lens**: which rubric section (1–4)
- **Severity**: `critical` | `moderate` | `minor`
- **Theme**: a short grouping label (e.g., "contract drift", "unsafe casts", "unhandled node types")
- **Finding**: 1–2 sentences describing the issue and why it matters

Cross-reference against `AST_REVISIONS.md`. Note any TypeScript code that will break due to planned parser changes.

### Phase 3: Group & Prioritize
**STOP. Present plan and wait for approval. To execute, invoke `/project:address-review`.**

## Constraints
- **Source code only.** Don't run the visualizer or evaluate visual output — stay in `tools/visualizer/src/`.
- **Focus on the consumer side.** Don't re-review the Go parser internals.
- **Flag contract mismatches explicitly.** If the TS expects a field the parser no longer emits, that is a blocker.
- **No backwards compatibility.** Pre-v1. Propose clean fixes.
- **Visual design and UX are out of scope.** Those belong in `/project:review-visualizer-spec`.
