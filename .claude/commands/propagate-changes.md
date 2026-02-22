# Propagate Breaking Changes

Given breaking changes documented in `AST_REVISIONS.md`, assess and plan the updates needed across all downstream components.

This command answers: "the parser changed — what else needs to change, and in what order?"

## Dependency Map

Changes propagate along this graph. Each edge has a contract:

```
Parser (tools/lsp/)
  ├─► Visualizer (tools/visualizer/)
  │     contract: JSON output of `twf parse` and `twf symbols`
  ├─► LSP Server (tools/lsp/internal/server/)
  │     contract: Go AST types and resolver API
  ├─► Skill: Design (skills/design/)
  │     contract: DSL syntax and semantics (LANGUAGE_SPEC.md)
  │     └─► Skill: Author-Go (skills/author-go/)
  │           contract: Design skill semantics + Go SDK mapping
  └─► VS Code Extension (packages/vscode/)
        contract: LSP protocol responses
```

## Workflow

### Phase 1: Read Breaking Changes

Read `AST_REVISIONS.md` in full. Identify:
- Each breaking change (AST field removed/renamed, JSON schema change, parser behavior change)
- Each change marked as complete vs. planned
- The type of break: **API** (Go types), **Schema** (JSON output), **Semantic** (behavior), **Grammar** (DSL syntax)

### Phase 2: Impact Assessment

For each breaking change, evaluate impact per downstream layer.
Use sub-agents in parallel — one per downstream:
- **Visualizer agent**: does this break the TypeScript types or rendering logic?
- **LSP Server agent**: does this break any server-side AST traversal or construction?
- **Skills agent**: does this change what the design skill documents or what author-go maps?
- **Extension agent**: does this affect LSP responses the extension depends on?

Each agent returns: [change ID] → [impact: none | minor | breaking] + specific location

### Phase 3: Synthesis

Build a propagation table:

| Change | Visualizer | LSP Server | Skill: Design | Skill: Author-Go | Extension |
|--------|-----------|------------|---------------|------------------|-----------|
| ...    | none      | breaking   | none          | none             | none      |

Identify which downstream layers need work and in what order (upstream fixes before downstream consumers).

### Phase 4: Plan

Present a prioritized list of downstream update tasks.
For each task:
- Which layer is affected
- Which specific files need changes
- Whether this can be done in parallel with other tasks
- Which downstream review command to run afterward to validate

**STOP. Present the propagation plan and wait for approval.**

### Phase 5: Execute

One layer at a time, in dependency order. After each layer:
- Run the relevant build/test (`go build ./...`, `npm run build`, `twf check`)
- Cross-reference against the propagation table to confirm the break is resolved

## Constraints
- **This command is assessment-and-planning only** until approved in Phase 4.
- **Don't modify `AST_REVISIONS.md`** during this process — it's the input, not the output.
- **After completing propagation**, update `AST_REVISIONS.md` to mark changes as fully propagated.
