# Plan: Workers, Options, and Namespaces

Infrastructure additions to the TWF DSL: structured options, worker blocks, and namespace blocks.

These three concepts form the **deployment topology** layer — the "who runs what, where" on top of the existing orchestration logic.

## Motivation

The DSL currently defines workflows and activities in isolation (the *what*). There is no way to express:
- Which workers run which workflows/activities
- What task queue a worker polls
- What namespace a worker belongs to
- Routing constraints (explicit or implicit task queue inheritance)

Adding this layer enables the resolver to validate reachability, routing correctness, and deployment coherence.

## Temporal Background

**Worker** — A process that connects to one namespace, polls one task queue, and registers specific workflow/activity types.

**Task Queue** — A named routing destination. Workers poll them; dispatches target them. No pre-creation needed. **All workers polling the same task queue must handle identical workflow/activity type sets** (Temporal requirement).

**Namespace** — Isolation boundary. All workflows, activities, and task queues are scoped within a namespace. Cross-namespace communication uses Nexus.

**Implicit task queue inheritance** — When an activity or child workflow call has no explicit `task_queue` option, it inherits the calling workflow's task queue. This is transitive through child workflows.

---

## Stage 1: Options Rework

### Goal

Replace the current opaque `options(...)` syntax with a structured `options { }` block that the parser understands, type-checks, and the resolver can query (especially for `task_queue`).

### Current State

Options are structured and validated on call sites only:
- **Lexer**: `OPTIONS` token exists, indentation-based block scanning
- **AST**: `Options *OptionsBlock` field on `ActivityCall` and `WorkflowCall` (not on definitions)
- **Parser**: `options:` parsed as structured key-value block with schema validation
- **Resolver**: Ignores options (structured access available for future stages)

### New Syntax

```
options_block ::= 'options' '{' NEWLINE option_entry* '}'
option_entry  ::= IDENT ':' option_value NEWLINE
                 | IDENT '{' NEWLINE option_entry* '}'    # nested (e.g. retry_policy)
option_value  ::= STRING | DURATION | NUMBER | BOOL | IDENT  # IDENT for enums
```

`options` keyword required. Always on the next line after the statement it modifies. Curly braces delimit the block.

```twf
activity ChargePayment(order) -> payment
    options {
        task_queue: "payment-workers"
        start_to_close: 60s
        retry_policy {
            max_attempts: 3
            initial_interval: 1s
            backoff_coefficient: 2.0
        }
    }

workflow SendNotifications(order) -> (notified)
    options {
        parent_close_policy: ABANDON
    }
```

### Allowed Option Keys

Derived from the Temporal API proto definitions. Each context has its own set — do not cross-pollinate.

**Activity call** (from `ScheduleActivityTaskCommandAttributes`):
| Key | Value Type | Description |
|-----|-----------|-------------|
| `task_queue` | string | Route to specific worker pool |
| `schedule_to_close_timeout` | duration | Total time from scheduled to complete |
| `schedule_to_start_timeout` | duration | Time waiting for a worker |
| `start_to_close_timeout` | duration | Time for execution after pickup |
| `heartbeat_timeout` | duration | Max gap between heartbeats |
| `request_eager_execution` | bool | Request eager dispatch |
| `retry_policy` | nested block | Retry configuration (see below) |
| `priority` | number | Execution priority |

**Child workflow call** (from `StartChildWorkflowExecutionCommandAttributes`):
| Key | Value Type | Description |
|-----|-----------|-------------|
| `task_queue` | string | Route to specific worker pool |
| `workflow_execution_timeout` | duration | Total across all runs |
| `workflow_run_timeout` | duration | Single run timeout |
| `workflow_task_timeout` | duration | Task processing timeout |
| `parent_close_policy` | enum | TERMINATE, REQUEST_CANCEL, ABANDON |
| `workflow_id_reuse_policy` | enum | ALLOW_DUPLICATE, etc. |
| `cron_schedule` | string | Cron expression |
| `retry_policy` | nested block | Retry configuration (see below) |
| `priority` | number | Execution priority |

**Retry policy** (nested, from `RetryPolicy`):
| Key | Value Type | Description |
|-----|-----------|-------------|
| `initial_interval` | duration | First retry delay |
| `backoff_coefficient` | number | Exponential multiplier |
| `maximum_interval` | duration | Cap on backoff |
| `maximum_attempts` | number | Max retries |
| `non_retryable_error_types` | string | Error type names |

Options are only supported on call sites (activity calls and workflow calls), not on definitions.

### Value Types

| Type | Examples | Token |
|------|---------|-------|
| duration | `60s`, `5m`, `1h`, `500ms`, `2h30m` | `DURATION` |
| number | `3`, `2.0` | `NUMBER` |
| bool | `true`, `false` | `BOOL` (or IDENT) |
| string | `"payment-workers"` | `STRING` (already exists) |
| enum | `TERMINATE`, `ABANDON` | `IDENT` (validated by schema) |

### Token Changes

New tokens to add:
- `LBRACE` — `{`
- `RBRACE` — `}`
- `DURATION` — numeric value with time suffix
- `NUMBER` — integer or float literal

`BOOL` can reuse `IDENT` (the schema validator distinguishes `true`/`false`).

**Important**: `LBRACE`/`RBRACE` are ONLY valid after `OPTIONS` or inside an options block. The lexer should not emit them in other contexts — this prevents braces from affecting the rest of the indentation-based language. A modal approach (the lexer enters "options mode" when it sees `OPTIONS` followed by `{`) keeps the change scoped.

### AST Changes

Replace `Options string` with a structured type:

```go
// OptionsBlock represents a parsed options { ... } block.
type OptionsBlock struct {
    Pos     Position
    Entries []*OptionEntry
}

// OptionEntry is a single key-value pair or nested block within options.
type OptionEntry struct {
    Pos      Position
    Key      string
    Value    string   // literal value for flat entries
    Nested   []*OptionEntry // for nested blocks like retry_policy
}
```

Fields on `WorkflowDef`, `ActivityDef`, `ActivityCall`, `WorkflowCall` change from `Options string` to `Options *OptionsBlock`.

### Parser Changes

- New file: `parser/parser/options.go` — self-contained options block parser
- The options parser handles everything between `options {` and `}`
- Existing statement parsers (`parseActivityCall`, `parseWorkflowCall`) delegate to the options parser when they see an indented `OPTIONS` token
- The options parser validates keys against the allowed set for the context (activity call, workflow call, definition) and emits errors for unrecognized keys
- Value type validation: check that each key's value matches its expected type

### Semantic Token / Coloring Changes

New semantic token types added to the legend:
- `property` (index 10) — for option keys (`task_queue`, `max_attempts`)
- `number` (index 11) — for duration and numeric values

Coloring intent: options should be **more visible than comments, less prominent than execution logic**. `property` in most VS Code themes renders as a muted, desaturated color — readable but not eye-catching.

| Token | Semantic type | Visual intent |
|-------|--------------|---------------|
| `options` keyword | `property` | Muted (not `type` like other keywords) |
| `{`, `}` | `operator` | Structural, same as `:` and `->` |
| Option keys | `property` | Muted but readable |
| String values | `string` | Existing string color |
| Duration/number values | `number` | Subtle literal |
| Enum values (IDENT) | `property` | Muted |

### Resolver Changes

- Extract `task_queue` from options blocks for later use by worker validation (stage 2)
- No new cross-reference validation in this stage — just structured access to option data

### Migration

The existing opaque `options(...)` syntax stops being recognized. Existing `.twf` files using it need to be updated to the new `options { }` syntax. This is a breaking change to the language.

Update:
- `LANGUAGE.md` — new grammar rules for options blocks
- Test files in `parser/testdata/` — update any files using old options syntax
- Topic docs under `design/topics/` — update examples

---

## Stage 2: Worker Blocks

### Goal

Add `worker` as a new top-level definition that connects workflows and activities to a task queue and namespace. Enable the resolver to validate deployment topology.

### Syntax

```twf
worker "order-worker":
    namespace "orders"
    task_queue "order-processing"
    workflow ProcessOrder
    workflow CancelOrder
    activity ChargePayment
    activity SendNotification
```

Worker is a top-level definition alongside `workflow` and `activity`. It declares:
- A name (string literal)
- A namespace (string literal) — required
- A task queue (string literal) — required
- A list of workflow and activity registrations (by name)

### Token Changes

- `WORKER` keyword — new token
- `NAMESPACE` keyword — new token
- `TASK_QUEUE` keyword — new token (or two tokens `TASK` + `QUEUE`; single keyword preferred)

### AST Changes

```go
type WorkerDef struct {
    Pos        Position
    Name       string
    Namespace  string
    TaskQueue  string
    Workflows  []string   // names of registered workflows
    Activities []string   // names of registered activities
}
```

`WorkerDef` implements `Definition` and is added to the file's definition list.

### Parser Changes

- Register `WORKER` in `topLevelParsers`
- New parser function for worker blocks — straightforward keyword-value pairs
- Self-contained; does not affect workflow/activity parsing

### Resolver Changes — Validation Rules

**Direct checks:**
| Check | Severity |
|-------|----------|
| Worker references undefined workflow | Error |
| Worker references undefined activity | Error |
| Defined workflow not on any worker | Warning (lenient OK) |
| Defined activity not on any worker | Warning (lenient OK) |
| Orphaned worker (references nothing defined) | Error |

**Task queue coherence:**
| Check | Severity |
|-------|----------|
| Workers on same task queue with different type sets | Error |
| Workers on same task queue with identical type sets | Warning (redundant) |

**Routing reachability (uses task_queue from stage 1 options):**

Build a lookup: `task_queue name → worker → registered types`.

| Check | Severity |
|-------|----------|
| Explicit `task_queue` option: target type not on any worker polling that queue | Error |
| Implicit task queue: called activity/workflow not on all workers that have the calling workflow | Error |
| Implicit task queue: calling workflow has no worker (can't determine queue) | Warning (lenient OK) |

**Implicit task queue inheritance** — When a call has no `task_queue` option, it inherits the calling workflow's task queue. Find all workers that register the calling workflow → those workers' task queues → the called type must be registered on workers polling each of those queues. This is transitive through child workflow calls.

---

## Stage 3: Namespace Blocks

### Goal

Add `namespace` as a top-level declaration for validating worker namespace references and nexus call targets.

### Syntax (tentative)

```twf
namespace "orders"
namespace "payments"
```

Initially just a named declaration — no body. Validates that:
- Workers reference existing namespaces
- `nexus "name"` calls target declared namespaces

### Future Expansion

Namespace blocks may eventually contain:
- Nexus endpoint declarations (which workflows/activities are exposed)
- Default configuration (retention, timeouts)
- Search attribute schemas

This is deferred. Stage 3 focuses on the declaration + reference validation.

### Resolver Changes

| Check | Severity |
|-------|----------|
| Worker references undeclared namespace | Error |
| `nexus "name"` targets undeclared namespace | Error |
| Declared namespace with no workers | Warning |

---

## Execution Notes

Each stage is independently shippable. Stage 1 is prerequisite to stage 2 (structured options needed for `task_queue` extraction). Stage 2 is prerequisite to stage 3 (worker-namespace references).

Parser changes should be scoped to new files where possible:
- Stage 1: `parser/parser/options.go` for the options sub-parser
- Stage 2: `parser/parser/definitions.go` extended for worker parsing (or a new file)
- Stage 3: Minimal — namespace is a simple declaration

The lexer changes (new tokens, options-mode scanning) should be carefully bounded so they don't affect the existing indentation-based scanning for the rest of the language.
