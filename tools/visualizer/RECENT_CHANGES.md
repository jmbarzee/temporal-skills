# TWF Visualizer

## Recent Language Changes: Promises, Conditions, State Blocks

### Removed

- **`spawn` keyword** — Removed entirely. `WorkflowCallMode` no longer has a `'spawn'` variant. The only remaining modes are `'child'` and `'detach'`.

### New Keywords

| Keyword | Purpose |
|---------|---------|
| `promise` | Non-blocking async declaration (`promise p <- activity Foo(x)`) |
| `condition` | Named boolean awaitable (declared in `state:` block) |
| `set` | Set a condition to true |
| `unset` | Set a condition to false |
| `state` | Workflow state declaration block |

### New Symbol

- `<-` (LEFT_ARROW) — Promise binding operator, used between promise name and async target

### New AST Nodes

| Node | Fields | Context |
|------|--------|---------|
| `PromiseStmt` | `name`, async target fields (same shape as activity/workflow/timer/signal/update calls) | Workflow body statement |
| `SetStmt` | `name` (condition name) | Workflow body statement |
| `UnsetStmt` | `name` (condition name) | Workflow body statement |
| `StateBlock` | `conditions []ConditionDecl`, `rawStmts []RawStmt` | Inside `WorkflowDef`, before handlers |
| `ConditionDecl` | `name` | Inside `StateBlock` |

### Changes to Existing AST Nodes

- **`WorkflowDef`** — New `state` field (`StateBlock | null`), appears before signals/queries/updates
- **`AwaitStmt`** — New `ident` and `identResult` fields for awaiting promises/conditions by name (e.g., `await myPromise -> result`)
- **`AwaitOneCase`** — New `ident` and `identResult` fields; case kind `"ident"` for promise/condition cases in `await one:` blocks

### JSON Serialization

The JSON AST output (used by the visualizer) reflects these changes:
- `WorkflowCall.mode` values are now only `"child"` or `"detach"` (no `"spawn"`)
- New statement types: `"promise"`, `"set"`, `"unset"`
- `WorkflowDef` has a `"state"` field
- `AwaitStmt` and `AwaitOneCase` may have `"ident"` / `"identResult"` fields

### Migration from `spawn`

| Before | After |
|--------|-------|
| `spawn workflow Child(x) -> handle` | `promise handle <- workflow Child(x)` |
| `spawn nexus "ns" workflow Foo(x) -> handle` | `promise handle <- nexus "ns" workflow Foo(x)` |
| `await one:` with `handle -> result:` | Same, but case kind is now `"ident"` |
