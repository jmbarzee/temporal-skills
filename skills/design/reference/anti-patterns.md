# Common Anti-Patterns

## Structural

### Unbounded History

A workflow that runs indefinitely without resetting accumulates unbounded event history, eventually degrading performance.

```twf
# BAD: Infinite loop with no history reset
# workflow EventProcessor(config: Config):
#     for:
#         activity PollEvents(config) -> events
#         activity ProcessBatch(events)

# GOOD: Continue-as-new resets history periodically
workflow EventProcessor(config: Config):
    state:
        condition shutdownRequested
    signal Shutdown():
        set shutdownRequested
    for:
        if (shutdownRequested):
            close complete
        activity PollEvents(config) -> events
        activity ProcessBatch(events)
        close continue_as_new(config)

activity PollEvents(config: Config) -> (Events):
    return poll(config)

activity ProcessBatch(events: Events):
    process(events)
```

**Why:** Temporal stores every event in workflow history. Long-running workflows without `close continue_as_new` grow history without bound, causing slow replays and eventual failure. See [long-running.md](../topics/long-running.md).

### Wrapper Workflow

A child workflow containing a single activity call adds orchestration overhead with no benefit.

```pseudo
# BAD: Unnecessary child workflow wrapper
workflow Parent():
    workflow SendEmailWorkflow(to, body)

workflow SendEmailWorkflow(to, body):
    activity SendEmail(to, body)
    close complete

# GOOD: Call the activity directly
workflow Parent():
    activity SendEmail(to, body)
```

**Why:** Child workflows create separate history, require their own task queue routing, and add latency. Use them only when you need independent retry policies, a separate failure boundary, or multi-step orchestration.

### Monolithic Workflow

All business logic in a single workflow with dozens of sequential steps.

```pseudo
# BAD: One workflow doing everything
workflow ProcessOrder(order):
    activity Validate(order)
    activity CheckInventory(order)
    activity ReserveInventory(order)
    activity ChargePayment(order)
    activity CreateShipment(order)
    activity NotifyWarehouse(order)
    activity UpdateCRM(order)
    activity SendConfirmation(order)
    activity ScheduleFollowUp(order)
    # ... 20 more steps

# GOOD: Decompose into child workflows with clear boundaries
workflow ProcessOrder(order):
    activity ValidateOrder(order) -> validated
    workflow FulfillOrder(validated) -> fulfillment
    workflow NotifyStakeholders(order, fulfillment)
    close complete(OrderResult{fulfillment})
```

**Why:** Large workflows have large histories (slow replay), make failure recovery coarse-grained (one failure may require re-running unrelated steps), and are hard to test. Decompose when a group of steps has its own lifecycle, retry needs, or failure boundary.

### Large Payloads in Workflow State

Storing large data (files, full database results, images) in workflow variables or signal/update payloads.

```pseudo
# BAD: Entire dataset in workflow state
workflow AnalyzeData(datasetId):
    activity FetchDataset(datasetId) -> dataset  # 500MB result stored in history
    activity Analyze(dataset) -> results

# GOOD: Pass references, not data
workflow AnalyzeData(datasetId):
    activity FetchAndStore(datasetId) -> dataRef  # Returns S3 key, not data
    activity Analyze(dataRef) -> results
```

**Why:** Every activity input and result is persisted in workflow history. Large payloads bloat history size, slow down replay, and may exceed Temporal's payload size limit. Pass references (IDs, URLs, keys) instead of data.

## Primitive Misuse

### Signal for Request-Response

Using a signal when the caller needs confirmation or a return value.

```pseudo
# BAD: Signal has no return value — caller doesn't know if it worked
signal ApproveOrder(orderId):
    approved = true

# GOOD: Update returns a result to the caller
update ApproveOrder(orderId: string) -> (ApprovalResult):
    activity ValidateApproval(orderId) -> validation
    if (validation.ok):
        approved = true
        return ApprovalResult{accepted: true}
    else:
        return ApprovalResult{accepted: false, reason: validation.error}
```

**Why:** Signals are fire-and-forget — the sender gets no acknowledgment, no validation, and no result. Use `update` when the caller needs to know the mutation was accepted.

### Query That Modifies State

Using a query handler to change workflow state.

```pseudo
# BAD: Query with side effects
query GetOrderStatus():
    accessCount = accessCount + 1  # Modifies state!
    return OrderStatus{status, accessCount}

# GOOD: Query is a pure read
query GetOrderStatus():
    return OrderStatus{status}
```

**Why:** Queries are read-only by contract. They may be called multiple times during replay without the workflow's knowledge. State modifications in queries produce unpredictable behavior and violate Temporal's execution model.

### Update Without Validation

Accepting an update without checking whether the mutation is valid.

```pseudo
# BAD: Blindly applies the update
update SetShippingAddress(address):
    shippingAddress = address
    return Result{ok: true}

# GOOD: Validate before committing
update SetShippingAddress(address: Address) -> (Result):
    activity ValidateAddress(address) -> validation
    if (validation.valid):
        shippingAddress = address
        return Result{ok: true}
    else:
        return Result{ok: false, error: validation.reason}
```

**Why:** Updates execute inside the workflow — invalid data corrupts workflow state. Always validate before committing. The caller receives the validation result, so they can react to rejection.

### Detach When You Need the Result

Using `detach` on a child workflow or nexus call when the parent needs the outcome.

```pseudo
# BAD: Detached — parent can't observe success or failure
detach workflow ProcessPayment(order)
# ... parent continues, has no idea if payment succeeded

# GOOD: Synchronous call or promise when result matters
workflow ProcessPayment(order) -> paymentResult
# or: promise p <- workflow ProcessPayment(order) ... await p -> paymentResult
```

**Why:** `detach` is fire-and-forget — the parent cannot await the result, check for errors, or compensate on failure. Use `detach` only when you genuinely don't care about the outcome (audit logs, analytics, notifications where failure is acceptable).

## Activity Anti-Patterns

### Non-Determinism in Workflows

Using non-deterministic operations directly in workflow code.

```pseudo
# BAD: Current time varies on replay
# if (current_time() > deadline):
#     cancel()

# BAD: Map iteration order varies across replays
# for (key in map.keys()):
#     activity Process(key)

# BAD: Goroutines/threads — execution order not deterministic
# go func() { activity DoWork() }

# GOOD: Use Temporal primitives for time
# await one:
#     activity DoWork() -> result:
#         close complete(Result{result})
#     timer(deadline):
#         close fail(Result{status: "timeout"})

# GOOD: Sort before iterating
# for (key in sorted(map.keys())):
#     activity Process(key)

# GOOD: Use promises for concurrency
# promise a <- activity DoWorkA()
# promise b <- activity DoWorkB()
# await a -> resultA
# await b -> resultB
```

**Why:** Temporal replays workflow code to reconstruct state. Any operation that produces different results on replay — time, random numbers, non-deterministic iteration, language-level threading — causes non-determinism errors. See [core-principles.md](./core-principles.md).

### Non-Idempotent Activities

Activities that fail or produce incorrect results on retry.

```pseudo
# BAD: Assumes fresh state — duplicate user on retry
activity CreateUser(name):
    db.insert(User(name))

# GOOD: Create-or-get — idempotent
activity CreateUser(name):
    existing = db.get_by_name(name)
    if existing: return existing
    return db.insert(User(name))
```

**Why:** Activities may be retried on network failures, worker crashes, or timeouts. An activity that isn't idempotent (same inputs → same result) will produce duplicate records, double charges, or inconsistent state. See [core-principles.md](./core-principles.md) for idempotency patterns.

### Orchestration in Activities

Putting multi-step logic, retry loops, or conditional branching inside an activity.

```pseudo
# BAD: Multi-step orchestration in activity — partial failure unrecoverable
activity DeployAll(specs):
    for spec in specs:
        deploy(spec)          # If this fails on spec #5 of 10,
        wait_healthy(spec)    # specs 1-4 deployed but no rollback
```

```twf
# GOOD: Workflow orchestrates, each step independently retryable
workflow DeployAll(specs: Specs):
    for (spec in specs.items):
        activity Deploy(spec)
        activity WaitHealthy(spec)
    close complete

activity Deploy(spec: Spec):
    deploy(spec)

activity WaitHealthy(spec: Spec):
    wait_healthy(spec)
```

**Why:** Activities run outside Temporal's durable execution model — if an activity fails mid-way through a loop, there's no replay, no history, and no way to resume from the last successful step. Workflows provide exactly this: durable, retryable orchestration with full visibility into progress.
