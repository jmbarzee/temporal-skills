# TWF Language Changelog

## v0.6.0 - Namespace Blocks & Worker Refactor

**Breaking change** — Workers are now reusable type sets. Deployment configuration (namespace, task_queue) has been moved to namespace blocks.

### What changed

Workers no longer contain `namespace` or `task_queue` entries. Instead, workers are pure type sets (workflows + activities), and a new `namespace` top-level definition instantiates workers with deployment options.

### Before (v0.5.0)

```twf
worker orderWorker:
    namespace orders
    task_queue orderProcessing
    workflow ProcessOrder
    activity ChargePayment
```

### After (v0.6.0)

```twf
worker orderWorker:
    workflow ProcessOrder
    activity ChargePayment

namespace orders:
    worker orderWorker
        options:
            task_queue: "orderProcessing"
```

### Benefits

- **Reusable type sets** — Same worker definition can be instantiated in multiple namespaces (e.g., production vs staging)
- **Richer deployment config** — Worker instantiation options include concurrency limits, rate limits, and other Temporal worker settings
- **Clearer separation** — Type grouping (worker) is separate from deployment topology (namespace)

### AST

- `WorkerDef` no longer has `Namespace` or `TaskQueue` fields — only `Name`, `Workflows`, and `Activities`
- New `NamespaceDef` node with `Name` and `Workers` (list of `NamespaceWorker`)
- `NamespaceWorker` has `WorkerName` and optional `Options` block
- JSON output: `workerDef` no longer includes `namespace`/`taskQueue`; new `namespaceDef` type added

### Resolver validation

- Worker type set refs to undefined workflows/activities produce errors
- Namespace refs to undefined workers produce errors
- Worker instantiation missing `task_queue` option produces error
- Workers on same task queue with different type sets produce errors
- Workers not instantiated in any namespace produce warnings
- Workflows/activities not on any instantiated worker produce warnings

### Worker instantiation options

Worker options (all snake_case): `task_queue`, `worker_activity_rate_limit`, `task_queue_activity_rate_limit`, `worker_local_activity_rate_limit`, `max_concurrent_activity_executions`, `max_concurrent_workflow_task_executions`, `max_concurrent_local_activity_executions`, `max_concurrent_workflow_task_pollers`, `max_concurrent_activity_task_pollers`, `max_cached_workflows`, `sticky_schedule_to_start_timeout`, `heartbeat_throttle_interval`, `worker_identity`, `worker_shutdown_timeout`, `local_activity_only_mode`

### Semantic tokens

- `namespace` keyword now colored as `type` (same as `workflow`/`activity`/`worker`) instead of `property`
- Namespace name colored as `function` with declaration modifier (same as worker/workflow/activity names)

---

## v0.5.0 - Worker Blocks

New top-level `worker` definition that connects workflows and activities to a task queue and namespace, enabling deployment topology validation at design time.

### New syntax: `worker`

```twf
worker orderWorker:
    namespace orders
    task_queue orderProcessing
    workflow ProcessOrder
    workflow CancelOrder
    activity ChargePayment
    activity SendNotification
```

### New tokens

- `worker` — top-level definition keyword
- `namespace` — worker namespace declaration
- `task_queue` — worker task queue declaration

### AST

- New `WorkerDef` and `WorkerRef` AST nodes
- `WorkerDef` includes `Name`, `Namespace`, `TaskQueue`, `Workflows`, and `Activities` fields
- JSON output uses type `"workerDef"` with `workflows` and `activities` arrays

### Resolver validation

- Worker references to undefined workflows/activities produce errors
- Duplicate worker names produce errors
- Defined workflows/activities not registered on any worker produce warnings
- Workers on the same task queue with different type sets produce errors
- Workers on the same task queue with identical type sets produce warnings (redundant)

### Semantic tokens

- `worker` keyword colored as `type` (same as `workflow`/`activity`)
- `namespace` and `task_queue` keywords colored as `property` (muted, like `options`)
- Worker name after `worker` keyword colored as `function` with declaration modifier
- Namespace/queue values colored as `variable`

### Options parser fix

The `task_queue` keyword is now accepted as a valid option key in `options:` blocks (previously it tokenized as IDENT, now it tokenizes as TASK_QUEUE).

---

## v0.4.0 - Options Restricted to Calls Only

**Breaking change** — options blocks are no longer allowed on activity or workflow definitions. Options are now only valid on call sites (`activity Name(args)` and `workflow Name(args)` statements).

### What changed

- `Options` field removed from `WorkflowDef` and `ActivityDef` AST nodes
- Parser no longer accepts `options:` blocks inside definition bodies
- Options remain fully supported on `ActivityCall` and `WorkflowCall` nodes (suffix-style, indented after the call)

### Why

Temporal SDK patterns always apply options at the call site. Definition-level defaults added language complexity without matching real API usage — the caller always controls timeouts, retry policies, and task queue routing.

### Migration

Move any definition-level `options:` blocks to the call sites. For example:

```twf
# Before (no longer valid)
activity Foo(x: int) -> (Result):
    options:
        start_to_close_timeout: 10s
    return x

# After
workflow Bar():
    activity Foo(x) -> result
        options:
            start_to_close_timeout: 10s
```

### Visualizer impact

- JSON output for `workflowDef` and `activityDef` nodes no longer includes `"options"` field
- `activityCall` and `workflowCall` JSON nodes still include `"options"` when present
- No changes to options block structure or schema validation

---

## v0.3.0 - Structured Options Blocks

**Breaking change** - replaces the old `options(key: value, ...)` parenthesized syntax.

### New syntax: `options:`

Options are now indentation-based blocks with one key-value pair per line. Nested blocks (e.g. `retry_policy:`) use indentation without braces.

```twf
activity ChargePayment(order) -> payment
    options:
        task_queue: "payment-workers"
        start_to_close_timeout: 60s
        retry_policy:
            maximum_attempts: 3
            initial_interval: 1s
            backoff_coefficient: 2.0
```

### Key naming

All option keys use `snake_case`, matching the Temporal proto field names directly.

### Schema validation

Option keys are validated per context — activity calls and workflow calls each have a defined set of allowed keys. Unrecognized keys produce parse errors. Values are type-checked against expected types (string, duration, number, bool/enum).

**Activity options:** `task_queue`, `schedule_to_close_timeout`, `schedule_to_start_timeout`, `start_to_close_timeout`, `heartbeat_timeout`, `request_eager_execution`, `retry_policy`, `priority`

**Workflow options:** `task_queue`, `workflow_execution_timeout`, `workflow_run_timeout`, `workflow_task_timeout`, `parent_close_policy`, `workflow_id_reuse_policy`, `cron_schedule`, `retry_policy`, `priority`

**Retry policy (nested):** `initial_interval`, `backoff_coefficient`, `maximum_interval`, `maximum_attempts`, `non_retryable_error_types`

### New value literals

- **Duration** - `30s`, `5m`, `1h`, `500ms`, `2d` (numeric value with time suffix)
- **Number** - `3`, `2.0` (integer or float)
- **Enum** - validated identifiers like `TERMINATE`, `ABANDON`, `REQUEST_CANCEL`

Duration and number literals are recognized inside `options:` blocks.

### Coloring

Options render with reduced visual prominence compared to execution logic:
- `options` keyword, option keys, and enum values to `property` (muted)
- Duration and number values to `number`
