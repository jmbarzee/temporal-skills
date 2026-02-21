# Design Improvements

Architecture and design issues to address before expanding visualizer complexity.

## ~~1. Extract shared block rendering primitives~~ (Partial — A done, B+C dropped)

- A. ~~`useToggle` hook~~ — Done. Extracted to `useToggle.ts`, applied to all 21 toggle sites.
- B. `<BlockHeader>` — Dropped. 4 distinct header CSS structures with different child class names; a unified component would need 6-7 props to replace 4 lines of self-documenting JSX.
- C. `<ExpandableBlock>` — Dropped. Body rendering varies too much across components.

## ~~2. Unify handler declaration components~~ (Done)

- D. ~~Replace the three components with a single `<HandlerDeclBlock>` parameterized by handler type~~ — Done. Config map + `'returnType' in decl` check.

## ~~3. Extract shared workflow content rendering~~ (Done)

- E. ~~Extract a `<WorkflowContent>` component for the handler groups + body, used by both `WorkflowDefBlock` and `WorkflowCallBlock`~~ — Done. New `WorkflowContent.tsx` with `HandlerDeclBlock` + `WorkflowContent`.

## ~~4. Unify await display helpers~~ (Done)

- F. ~~Extract a shared `getAwaitTargetDisplay` helper that both functions delegate to~~ — Done. Shared helper handles the 6 common kinds; `getAwaitStmtDisplay` adds `await` prefix and overrides icons for activity/workflow; `getAwaitOneCaseDisplay` handles `await_all` separately then delegates.

## 5. Split StatementBlock.tsx

**Problem:** Single 750-line file contains 17 renderer components, 2 display helpers, and 2 signature formatters.

**Changes:**
- G. Split into separate files by concern (e.g. calls, control-flow, await, leaf statements). The top-level `StatementBlock` dispatcher stays as the entry point.

## ~~6. Fix context naming~~ (Done)

Contexts already named `DefinitionContext` / `HandlerContext`.

## ~~7. Use stable keys for statement lists~~ (Done)

All statement lists use `key={`${stmt.line}:${stmt.column}`}`.

## ~~8. Clean up dead CSS~~ (Done)

Audited and cleaned up:
- **Dead CSS removed:** `.tagged-query` (blocks.css), `.app`/`.toolbar`/`.file-upload` (index.css)
- **Missing CSS added:** `.block-close.close-continue-as-new` rule (CSS variables existed but had no rule)
- **Phantom JSX classes removed:** `.close-completed`, `.close-args`, `.block-statements`, `.block-then`, `.block-else`, `.has-body`

## Change Dependencies

Changes are labeled A-J. Dependencies:

```
A (useToggle) ← B (BlockHeader) ← C (ExpandableBlock)
D (HandlerDeclBlock) — independent
E (WorkflowContent) — benefits from D
F (await helpers) — independent
G (split StatementBlock) — benefits from A/B/C, E, F being done first
H (context naming) — independent
I (stable keys) — independent
J (dead CSS) — independent
```

Suggested ordering:
1. Independent cleanups: H, I, J (low risk, immediate value)
2. Shared primitives: A → B → C
3. Deduplication: D, then E, then F
4. File reorganization: G (after the above reduce file size naturally)
