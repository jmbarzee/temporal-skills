# TWF Language Changelog

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
