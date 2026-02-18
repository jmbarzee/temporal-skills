# Possible Future Features

Unimplemented features, consolidated from design testing and ongoing use.

---

## Infrastructure Blocks

### Worker Block

Top-level block for worker configuration: interceptors, codecs, worker settings.

```twf
worker "my-worker":
    interceptor LoggingInterceptor
    codec EncryptionCodec(keyId: "key-1")
    task_queue "my-queue"
```

**Open questions:** How much deployment topology belongs in a design DSL? Workers are infrastructure, not workflow logic. Could be out of scope if TWF stays purely orchestration-focused.

### Namespace Block

Formal namespace declaration beyond `nexus "name"` syntax.

```twf
namespace "orders":
    # config? routing? defaults?
```

**Open questions:** What goes here besides the name? Retention policies, search attribute schemas, default timeouts? Or is `nexus "name"` sufficient for cross-namespace references?

### Task Queue Syntax

`task_queue` appears in primitives-reference.md and has a topic file, but no formal language syntax exists. Workers poll task queues, activities/workflows can target specific queues.

```twf
workflow ProcessOrder(order: Order) -> (Result):
    # Route heavy work to GPU workers
    activity RenderImage(order.image)
        options:
            task_queue: "gpu-workers"
```

**Open questions:** Is `options: task_queue: ...` sufficient, or does the DSL need a top-level `task_queue` declaration block?

---

## Workflow Semantics

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
