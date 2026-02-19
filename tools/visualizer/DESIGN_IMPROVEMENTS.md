# Design Improvements

Architecture and design issues to address before expanding visualizer complexity.

## 1. Extract shared block rendering primitives

**Problem:** Every expandable block reimplements the same pattern — `useState(false)`, `useRefocus()`, toggle handler, and identical JSX structure (toggle + icon + keyword + signature). This appears in 15+ components across `StatementBlock.tsx` and `DefinitionBlock.tsx`. Adding universal behavior (keyboard nav, animation, source location links) requires touching every component.

**Changes:**
- A. Create a `useToggle` hook that encapsulates `useState` + `useRefocus` + toggle handler
- B. Create a `<BlockHeader>` component for the shared toggle/icon/keyword/signature layout
- C. Create an `<ExpandableBlock>` wrapper that composes the above

## 2. Unify handler declaration components

**Problem:** `SignalDeclBlock`, `QueryDeclBlock`, and `UpdateDeclBlock` in `DefinitionBlock.tsx` are ~95% identical. They differ only in icon, CSS class, and whether `returnType` is shown.

**Changes:**
- D. Replace the three components with a single `<HandlerDeclBlock>` parameterized by handler type

## 3. Extract shared workflow content rendering

**Problem:** `WorkflowCallBlock` (StatementBlock.tsx:116-239) re-implements the signal/query/update/body rendering from `WorkflowDefBlock` (DefinitionBlock.tsx:35-179). Changes to handler display must be applied in two places.

**Changes:**
- E. Extract a `<WorkflowContent>` component for the handler groups + body, used by both `WorkflowDefBlock` and `WorkflowCallBlock`

## 4. Unify await display helpers

**Problem:** `getAwaitStmtDisplay` and `getAwaitOneCaseDisplay` are structurally near-identical switch statements over overlapping kind unions. The `AwaitStmt` and `AwaitOneCase` types share the same optional fields.

**Changes:**
- F. Extract a shared `getAwaitTargetDisplay` helper that both functions delegate to

## 5. Split StatementBlock.tsx

**Problem:** Single 750-line file contains 17 renderer components, 2 display helpers, and 2 signature formatters.

**Changes:**
- G. Split into separate files by concern (e.g. calls, control-flow, await, leaf statements). The top-level `StatementBlock` dispatcher stays as the entry point.

## 6. Fix context naming

**Problem:** `DefinitionContextProvider` and `HandlerContextProvider` are `React.createContext()` values, not provider components. Using `.Provider` on them reads as `DefinitionContextProvider.Provider`.

**Changes:**
- H. Rename to `DefinitionContext` / `HandlerContext` (the context values), optionally export wrapper provider components

## 7. Use stable keys for statement lists

**Problem:** All statement lists use `key={i}` (array index). This causes incorrect state preservation if statements are reordered. The AST provides `Position` (line, column) on every statement.

**Changes:**
- I. Switch to `key={`${stmt.line}:${stmt.column}`}` for statement lists

## 8. Clean up dead CSS

**Problem:** Several CSS classes are defined but never applied in JSX:
- `.block-timer` (standalone) — timer blocks use `.block-await-stmt-timer`
- `.block-signal`, `.block-update` (standalone) — same pattern
- `.block-continue-as-new` — `CloseBlock` uses `close-continue-as-new`
- `.close-value` — not referenced in any component

**Changes:**
- J. Audit and remove unused CSS classes

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
