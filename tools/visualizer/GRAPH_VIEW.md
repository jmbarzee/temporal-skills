# Graph View: Features Needed

The visualizer needs a second view mode — a graph/topology view where **workers** are the core nodes, contained within **namespace** grouping nodes. This complements the existing tree view which visualizes individual workflow/activity control flow.

## Blocking: AST Type Gaps

### WorkerDef not in TypeScript types

The Go parser already emits `WorkerDef` in JSON:

```json
{
  "type": "workerDef",
  "name": "MyWorker",
  "namespace": "my-namespace",
  "taskQueue": "my-queue",
  "workflows": [{ "name": "Foo", "line": 10, "column": 3 }],
  "activities": [{ "name": "Bar", "line": 11, "column": 3 }]
}
```

But `types/ast.ts` has no `WorkerDef` interface, and the `Definition` union is `WorkflowDef | ActivityDef` only. `WorkflowCanvas` silently drops worker definitions during context building.

**Needed:**
- Add `WorkerDef`, `WorkerRef` types to `ast.ts`
- Extend `Definition` union to include `WorkerDef`
- Add type guard `isWorkerDef()`

### Namespace as a grouping concept

Workers declare a `namespace` string. The graph view needs to group workers by namespace into container nodes. No new parser work needed — the data is already in the JSON — but the visualizer needs a data model for namespace grouping.

## Graph View Data Model

The graph view operates on a different slice of the AST than the tree view:

| Concept | Tree View | Graph View |
|---------|-----------|------------|
| Primary node | Workflow/Activity definition | Worker |
| Grouping | Source file | Namespace |
| Edges | None (nested expansion) | Workflow-to-workflow calls, workflow-to-activity calls, cross-namespace (nexus) calls |
| Detail | Full control flow body | Reference list (which workflows/activities a worker hosts) |

## View Switching

Support swapping between tree and graph views. Low priority for now — the two views are independent enough that a simple tab/toggle in the canvas header should suffice.
