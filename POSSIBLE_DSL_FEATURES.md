# Possible Future Features

Unimplemented features, consolidated from design testing and ongoing use.

---

## Naming Conventions

### UpperCamelCase for All Top-Level Primitives

Worker type set names and namespace names currently use lowerCamelCase, while workflows/activities use UpperCamelCase. All top-level definitions should consistently use UpperCamelCase.

```twf
# Current:
worker orderTypes:
namespace orders:

# Proposed:
worker OrderTypes:
namespace Orders:
```

**Why deferred:** Requires updating all existing examples, test data, topic files, and the LANGUAGE.md spec. Should be done as a standalone cleanup.

---

## Nexus Extensions

### List Workflows in Sync Operations

Sync nexus operation handlers can list/query workflows as part of their implementation. No current syntax for representing a "list workflows" primitive in the DSL.

```twf
nexus service OrderService:
    sync ListActiveOrders(filter: Filter) -> (OrderList):
        list ProcessOrder(filter) -> orders
        close complete(orders)
```

**Open questions:** What does the syntax look like? `list WorkflowType(filter)` as a primitive? How does it relate to Temporal's visibility/list APIs? Is this a workflow-body primitive or nexus-operation-only?

---

## Workflow Semantics

### Signal/Query/Update Send Statements

Explicit DSL syntax for sending signals, queries, and updates to other workflows. Currently the DSL declares handlers on the receiving side but has no way to express the send side.

```twf
workflow OrderSaga(order: Order) -> (Result):
    workflow ProcessPayment(order) -> payment

    # Signal a running workflow
    signal ProcessPayment.PaymentReceived(payment)

    # Query a running workflow
    query ProcessPayment.Status() -> status

    # Update a running workflow
    update ProcessPayment.AdjustAmount(newAmount) -> confirmation
```

**Why needed:** The visualizer's graph view models dependency edges between workflows. Currently only call/await edges exist. Signal/query/update sends create real dependencies — "WorkflowB sends a signal to WorkflowA" — but these are invisible without send-side data in the AST. Adding typed send statements would enable message flow edges in the graph view (see GRAPH_VIEW.md future section on message flow edges).

**Open questions:** What is the syntax? `signal TargetWorkflow.HandlerName(args)` vs `send signal HandlerName to TargetWorkflow(args)`? Should sends target a specific workflow instance (by ID) or a workflow type? How does the resolver validate that the target workflow actually handles the named signal/query/update? Should sends appear as statements (in workflow body) or as part of await expressions?

---

### Workflow Cancellation Handler

`await one:` documents auto-cancellation of race losers, but there's no way to express what happens when an *entire workflow* is cancelled externally.

```twf
workflow ProcessOrder(order: Order) -> (Result):
    on_cancel:
        activity RefundPayment(order)
        activity NotifyCustomer(order, "cancelled")

    activity ChargePayment(order) -> payment
    activity FulfillOrder(order) -> fulfillment
    close Result{payment, fulfillment}
```

**Why needed:** Cancellation is a first-class Temporal concept. Cleanup/compensation on cancel is a common pattern with no current TWF representation.

### Async Activity Completion

Activity that starts, then completes from an external system (human approval, webhook callback). Referenced in activities-advanced.md topic but no language syntax.

```twf
activity RequestApproval(order: Order) -> (Approval):
    async_complete
    send_approval_request(order)
```

**Why needed:** Common pattern for human-in-the-loop workflows. `heartbeat` has syntax; `async_complete` does not.

### Explicit Type Definitions

Types are bare identifiers — no `type Foo: ...` struct syntax. Type structure only exists in implementation code.

```twf
# Can't do this:
type OrderResult:
    status: string
    total: decimal
    items: Item[]

# Must do this:
workflow ProcessOrder(order: Order) -> (OrderResult):
    # OrderResult structure lives in implementation
```

**Impact:** No single source of truth for data structures at design time. Can't validate field names/types.

**Trade-off:** Adding types moves the DSL toward a full IDL. May conflict with "skeleton, not meat" principle — or may be exactly what's needed for design clarity.

### SDK Built-in Functions

Deterministic SDK utilities like `workflow.history_length()` have no formal syntax. Currently shown as raw expressions in examples.

```twf
# Used in practice but not formalized:
historyBytes = sdk.HistorySize()
if (historyBytes > 40_000_000):
    continue_as_new(data)
```

**Open questions:** Should the DSL formalize a set of SDK intrinsics? Or are these implementation details that belong in `raw_stmt`?

### Workflow ID Call Option

`workflow_id` as a workflow call option for specifying deterministic workflow IDs at call sites. This is a core Temporal SDK pattern (e.g., deriving a child workflow ID from a business entity) but is not currently in the allowed workflow call options defined in LANGUAGE.md.

```twf
workflow ProcessOrder(order: Order) -> (Result):
    workflow ShipOrder(order) -> shipment
        options:
            workflow_id: "ship-order-" + order.id
            parent_close_policy: TERMINATE

    # Idempotent fan-out via deterministic IDs
    for (item in order.items):
        workflow ProcessItem(item)
            options:
                workflow_id: "process-item-" + item.id
                workflow_id_reuse_policy: ALLOW_DUPLICATE_FAILED_ONLY
```

**Why deferred:** The concept is already used in topic docs (child-workflows.md shows `workflow_id` in options blocks), but the allowed workflow call options list in LANGUAGE.md does not include it. Adding it requires deciding whether the value is a plain string, a template expression (`"order-{order.id}"`), or a concatenation expression (`"order-" + order.id`) — which ties into the broader question of expression syntax in option values. The current `value` grammar (`STRING | DURATION | NUMBER | IDENT`) has no expression support.

**Open questions:** Should `workflow_id` values support string interpolation (template syntax), concatenation expressions, or just static strings? Should `workflow_id_reuse_policy` also be added as a formal option (it appears in child-workflows.md alongside `workflow_id`)? How does this interact with the resolver — should it warn on non-unique IDs inside loops?

---

## Syntax Extensions

### Bare Promise Declaration

Declare a promise without immediate `<-` binding.

```twf
promise myPromise
# ... later assign it
myPromise <- activity ProcessItem(input)
```

**Why deferred:** No clear use case. `promise p <- ...` covers all known patterns.

### Condition Declarations Outside `state:`

Allow `condition` directly in workflow body without `state:` block.

```twf
workflow Example():
    condition ready
    activity Setup() -> config
    set ready
```

**Why deferred:** `state:` block provides clear separation between declarations and execution. Conditions anywhere complicates parsing and readability.

### Expression-Based Conditions

Arbitrary boolean expressions as await targets, not just named conditions.

```twf
await condition (balance > threshold and not suspended)
```

**Why deferred:** Requires the DSL to understand expression evaluation, conflicting with "skeleton, not meat" principle. Named conditions achieve the same thing with explicit state management:

```twf
state:
    condition thresholdReached

signal Deposit(amount: decimal):
    balance = balance + amount
    if (balance >= 1000):
        set thresholdReached

await thresholdReached
```

### SDK Language Specification

Optional declaration of which Temporal SDK language a worker, workflow, activity, or other definition targets. Useful for polyglot codebases where different services or teams own different languages.

```twf
# At the worker level — applies to all contained definitions
worker OrderTypes (go):
    workflow ProcessOrder
    activity ChargePayment

# At the individual definition level
workflow ProcessOrder(order: Order) -> (Result) (go):
    activity ChargePayment(order) -> payment
    activity RunFraudModel(order) -> score (python)
    activity NotifyCustomer(order, payment) (typescript)

# At the namespace level
namespace Orders (go):
    workflow ProcessOrder
```

**Why needed:** Polyglot Temporal deployments are common — a Go workflow may call Python ML activities and TypeScript frontend services. Making the SDK language a first-class part of the syntax (rather than an annotation) enables clearer design-time intent, code generation targeting the correct SDK, ownership boundaries, and better onboarding context.

**Open questions:** Should the language be a fixed enum of supported SDKs (`go`, `python`, `typescript`, `java`, `dotnet`, `php`) or freeform? Should a parent declaration (e.g. on a `worker`) propagate as a default to children, with per-definition overrides? How does it interact with Nexus boundaries where language differences are already implicit? How does this relate to `@lang` annotations — should both exist, or should one replace the other?

### Local Activity Option

`local: true` option on activity calls to route execution to the local worker, avoiding the task queue round-trip. The Temporal SDK supports this natively via local activity APIs. The concept is referenced in activities-advanced.md topic docs but no formal TWF syntax or parser support exists.

```twf
workflow ProcessOrder(order: Order) -> (Result):
    # Local activity: runs in-process on the same worker
    activity ValidateInput(order) -> validated
        options:
            local: true
            start_to_close_timeout: 5s

    # Local activity with retry threshold
    activity EnrichData(validated) -> enriched
        options:
            local: true
            start_to_close_timeout: 10s
            local_retry_threshold: 5s
            retry_policy:
                maximum_attempts: 3
                initial_interval: 100ms

    # Regular activity: goes through task queue as normal
    activity ChargePayment(enriched) -> payment
        options:
            start_to_close_timeout: 60s
```

**Why deferred:** Local activities are an SDK-level execution optimization, not a workflow design concern. TWF's "skeleton, not meat" principle suggests this may be too implementation-specific. However, the choice between local and regular activities has design implications — local activities should be short, deterministic, and avoid network calls, which is valuable to capture at design time.

**Open questions:** Should `local` be a boolean option in the options block, or a modifier keyword (e.g., `local activity ValidateInput(...)`)? If it is an option, should it be in the activity call options list alongside `task_queue`? Does `local: true` conflict with an explicit `task_queue` option (local activities bypass the task queue)? Should `local_retry_threshold` and `schedule_to_close_timeout` (which local activities do not support) be validated contextually?

### Non-Retryable Error Types List Syntax

The retry policy option `non_retryable_error_types` is listed in the grammar spec (LANGUAGE.md) as a valid retry policy key, but the option value grammar (`value ::= STRING | DURATION | NUMBER | IDENT`) has no list literal type. Specifying a list of error type strings requires a list value syntax that the lexer, parser, and AST do not currently support.

```twf
activity ChargePayment(order: Order) -> (Payment):
    options:
        start_to_close_timeout: 60s
        retry_policy:
            maximum_attempts: 5
            initial_interval: 1s
            non_retryable_error_types: ["InvalidInput", "NotFound", "Unauthorized"]

activity SendNotification(user: User, message: Message):
    options:
        retry_policy:
            maximum_attempts: 3
            non_retryable_error_types: ["InvalidRecipient"]
```

**Why deferred:** Adding list literals requires changes across the entire parser pipeline. The lexer needs `[` and `]` tokens (or at least recognition within option value context). The AST needs a list value node type. The parser needs a production rule for `list_value ::= '[' value (',' value)* ']'`. The resolver needs to validate that `non_retryable_error_types` specifically accepts a list of strings. This is a meaningful grammar extension for a single option key.

**Open questions:** Should list syntax be general-purpose (`value ::= ... | list_value`) or restricted to specific option keys? If general-purpose, what other options could benefit from list values in the future? Should the list elements be restricted to strings, or allow mixed types? Could an alternative syntax avoid brackets entirely (e.g., multi-line list under the key, one entry per line)?

### Promise Composition

Dynamic promise collection for batch awaiting.

```twf
promises = []
for (item in items):
    promise p <- activity Process(item)
    promises.append(p)

await all promises -> results
```

**Why deferred:** `await all:` with inline operations covers most parallel patterns. Dynamic collection adds significant type system and resolver complexity.

---

## Annotations

Annotations support the DSL as living documentation for a project, bridging the gap between design-level orchestration and the underlying implementation.

### Language Annotations

Declare what implementation language a workflow, activity, or block should be written in. Useful for polyglot codebases where different teams or services own different languages.

```twf
workflow ProcessOrder(order: Order) -> (Result):
    @lang("go")
    activity ChargePayment(order) -> payment

    @lang("python")
    activity RunFraudModel(order) -> score

    @lang("typescript")
    activity NotifyCustomer(order, payment)
```

**Why needed:** Polyglot Temporal deployments are common - a Go workflow may call Python ML activities and TypeScript frontend services. Language annotations make this explicit at design time, enabling code generation targeting the correct SDK, clearer ownership boundaries, and better onboarding context.

**Open questions:** Should `@lang` apply at the block level (workflow, activity) or also at the file level as a default? Should it be a fixed enum of supported SDKs (`go`, `python`, `typescript`, `java`, `dotnet`, `php`) or freeform? How does it interact with Nexus boundaries where language is already implicit?

### Reference Annotations

Point to where an existing implementation lives in the codebase. Turns TWF files into navigable maps of a running system.

```twf
workflow ProcessOrder(order: Order) -> (Result):
    @ref("order-service/workflows/process_order.go:17")
    activity ChargePayment(order) -> payment

@ref("payments/activities/charge.go")
activity ChargePayment(order: Order) -> (Payment):
    heartbeat 30s
    options:
        start_to_close_timeout: 60s
```

**Why needed:** As a design DSL, TWF captures intent — but teams also need to find the real code. Reference annotations close that loop, making `.twf` files a living index of the project. LSP features like go-to-definition could resolve `@ref` paths to open the actual source file.

**Open questions:** Should paths be relative to the repo root, or allow URLs for multi-repo setups? Should the LSP validate that `@ref` targets exist? Could references be auto-populated by scanning the codebase for matching workflow/activity registrations?

---

## Design Quality Linting

### `twf lint` Command

A design-quality pass beyond syntax/resolution validation (`twf check`) to catch common anti-patterns and missing considerations.

**Potential checks:**

| Check | Category | Description |
|-------|----------|-------------|
| Unbounded loops | History | `for:` without `continue_as_new` — history grows forever |
| Missing continue-as-new | History | Signal-driven loops with no history reset strategy |
| Missing error handling | Resilience | Activities with no timeout/retry configuration |
| Signal vs update choice | Design | Signal used where update semantics (validation, confirmation) seem more appropriate |
| Unbounded tool/retry loops | Safety | Loops calling activities with no iteration bound |
| Missing queries | Observability | Stateful workflow with no query handlers for inspection |
| Large activity fan-out | Performance | Many parallel activities without task queue routing |

**Why difficult:** TWF is intentionally high-level. Many checks require understanding *intent*, not just structure. For example, a `for:` loop without `continue_as_new` might be intentional for short-lived workflows. Linting would need heuristics or configurable rules, not hard errors.

**Possible approach:** Advisory warnings (not errors) with suppression comments, e.g.:
```twf
# twf:lint-ignore unbounded-loop
for:
    await signal Event -> event
    activity ProcessEvent(event)
```

**Open questions:** Should lint rules be configurable per-project? Should they run as part of `twf check` (with a `--strict` flag) or as a separate command? How to avoid false positives on intentionally simple designs?
