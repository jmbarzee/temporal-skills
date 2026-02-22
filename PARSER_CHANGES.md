# Parser Changes

Breaking changes to the `twf parse` JSON contract and new CLI capabilities that need propagation to the TypeScript visualizer.

## Breaking JSON Changes

### 1. Resolved refs on calls

`activityCall` and `workflowCall` statements gain an optional `resolved` field:

```json
{
  "type": "activityCall",
  "name": "GetOrder",
  "args": "orderId",
  "result": "order",
  "resolved": { "name": "GetOrder", "line": 10, "column": 1 }
}
```

### 2. Always-present handler arrays

`signals`, `queries`, `updates` on `workflowDef` are always emitted as `[]`, never omitted. TS consumers no longer need null-checks for these fields.

### 3. Async target restructured

Flat fields on `await`, `awaitOne` cases, and `promise` replaced by a nested `"target"` object with a discriminated union by `kind`. Each kind has its own sub-object with only relevant fields.

**Before (flat, 22 fields all omitempty):**
```json
{
  "type": "await",
  "kind": "activity",
  "activity": "GetOrder",
  "activityArgs": "orderId",
  "activityResult": "order"
}
```

**After (nested, per-kind):**
```json
{
  "type": "await",
  "target": {
    "kind": "activity",
    "activity": {
      "name": "GetOrder",
      "args": "orderId",
      "result": "order",
      "resolved": { "name": "GetOrder", "line": 10, "column": 1 }
    }
  }
}
```

Target kinds: `timer`, `signal`, `update`, `activity`, `workflow`, `nexus`, `ident`.

### 4. Summary metadata

Top-level JSON gains a `summary` object before `definitions`:

```json
{
  "summary": {
    "namespaces": 1,
    "workers": 3,
    "workflows": 6,
    "activities": 8,
    "nexusServices": 2
  },
  "definitions": [...]
}
```

### 5. sourceFile field

All definitions gain a `sourceFile` string (basename of the source `.twf` file):

```json
{ "type": "workflowDef", "name": "ProcessOrder", "sourceFile": "orders.twf", ... }
```

### 6. Per-file line numbers

Line numbers are now per-file, not global offsets into concatenated input. Each file is parsed independently — line 1 is always the first line of that file.

## New CLI Capabilities

### twf symbols expansion

`twf symbols` now outputs `worker`, `namespace`, and `nexusService` kinds with sub-symbols:

```
worker specWorker()
  workflow SpecCompliance
  activity ValidateInput
namespace specTests()
  worker specWorker
nexusService ExternalService()
  operation ExternalTask (async)
```

### twf deps

New subcommand for dependency graph extraction:

```
twf deps [--json] [--lenient] <file...>
```

Text output shows containment, edges, cross-worker dependencies, and unresolved references. JSON output provides the full graph structure with nodes, edges, containment hierarchy, coarsened projections, and summary.

## Files to Update

| File | What to change |
|------|---------------|
| `tools/visualizer/src/types/ast.ts` | Update `ActivityCall`, `WorkflowCall` (add `resolved`), `AwaitStmt`, `AwaitOneCase`, `PromiseStmt` (nested `target` object), all definition types (add `sourceFile`) |
| `tools/visualizer/src/components/blocks/AwaitBlocks.tsx` | Update `getAwaitTargetDisplay` to read nested `target` object by `kind` |
| `tools/visualizer/src/components/blocks/LeafBlocks.tsx` | Update `PromiseBlock` to read nested `target` object |
