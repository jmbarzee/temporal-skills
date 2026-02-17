# Signals, Queries, and Updates

> **Example:** [`signals-queries-updates.twf`](./signals-queries-updates.twf)

External communication with running workflows. These three primitives let code outside the workflow interact with it during execution — as a **read**, a **write**, or a **read-write**.

## Overview

| Primitive | I/O | Direction | Execution | Use Case |
|-----------|-----|-----------|-----------|----------|
| **Query** | Read | External → Workflow → External | Sync (request-response) | Read current state |
| **Signal** | Write | External → Workflow | Async (fire-and-forget) | Events, notifications, data injection |
| **Update** | Read-write | External → Workflow → External | Sync (request-response) | Mutate state with confirmation |

---

## Signals

Asynchronous messages sent to a running workflow. The sender doesn't wait for processing.

### When to Use

- External events that workflow should react to
- Injecting data into a running workflow
- Triggering state transitions
- Human approval/rejection flows

### Signal Handler Bodies

Signals are declared with handler body blocks that execute when the signal arrives. Handler bodies have access to the full workflow statement set (activities, child workflows, timers, etc.).

```twf
signal PaymentReceived(transactionId: string, amount: decimal):
    paymentStatus = "received"
    lastTransactionId = transactionId
```

The handler body executes when the signal arrives, whether via `await signal` or as a case in `await one`/`await all`. Handler bodies should primarily update workflow state. Heavy side effects (activity calls, child workflows) belong in the main workflow body after the signal is awaited, since handlers fire even when not being actively awaited and can execute between any two deterministic steps.

### Handler Execution Semantics

When a signal is awaited (via `await signal` or as an `await one` case), the execution order is:

1. **Signal arrives** → handler body runs first (updates state)
2. **Await resolves** → case body runs (reacts to updated state, calls activities, etc.)

This two-phase execution means:

```twf
workflow OrderWorkflow(orderId: string):
    signal PaymentReceived(transactionId: string, amount: decimal):
        # Phase 1: Handler runs immediately on signal arrival
        paymentStatus = "received"
        lastTransactionId = transactionId

    # Phase 2: Case body runs after handler
    await one:
        signal PaymentReceived:
            # paymentStatus is already "received" here
            activity FulfillOrder(orderId, lastTransactionId)
            close complete(OrderResult{status: "completed"})
        timer(24h):
            close fail(OrderResult{status: "timeout"})
```

**Key implications:**
- Handler bodies run on every signal arrival, even if the workflow isn't actively awaiting that signal
- Keep handler bodies lightweight (state updates only, no activities)
- Place activity calls and side effects in the `await one` case body, not the handler body

### Signal Considerations

| Consideration | Guidance |
|---------------|----------|
| **Ordering** | Signals are processed in order received, but arrival order isn't guaranteed |
| **Buffering** | Signals queue if workflow is busy; consider signal coalescing for high-volume |
| **Idempotency** | Signal handlers should be idempotent (same signal twice = same result) |
| **Validation** | Validate signal payload; invalid signals can corrupt workflow state |
| **Ambient arrival** | Signals can arrive between any two deterministic steps and are buffered until handled by `await signal` or `await one`/`await all` |

---

## Queries

Synchronous, read-only access to workflow state. The caller blocks until the query returns.

### When to Use

- UI needs to display current workflow state
- Monitoring/debugging workflow progress
- External system needs workflow data
- Building workflow dashboards

### Query Handler Bodies

Queries are declared with handler body blocks that execute when queried. Query handlers are restricted to activity-style statements (no temporal primitives like timers, signals, or child workflows).

```twf
query GetStatus() -> (string):
    return status

query GetProgress() -> (Progress):
    return Progress{status: status, processed: itemCount}
```

### Query Considerations

| Consideration | Guidance |
|---------------|----------|
| **Read-only** | Queries MUST NOT modify workflow state |
| **Determinism** | Query handlers run during replay; must be deterministic |
| **Performance** | Queries replay workflow history; expensive for long histories |
| **Consistency** | Returns point-in-time state; may be stale by the time caller uses it |
| **Restrictions** | Query handlers use activity-restricted statement set (no timers, signals, workflows) |

### Anti-Patterns

```twf
# BAD: Query modifies state
query GetAndIncrementCounter() -> (int):
    counter = counter + 1  # NOT ALLOWED
    return counter

# GOOD: Pure read
query GetStatus() -> (string):
    return status
```

---

## Updates

Synchronous read-write operations. The caller sends data, the workflow processes it, and the caller blocks until it receives a result (or error) back.

### When to Use

- Need confirmation that change was applied
- Validating input before accepting
- Returning computed result from mutation
- Request-response pattern with workflow

### Update Handler Bodies

Updates are declared with handler body blocks that execute when the update is received. Handler bodies have access to the full workflow statement set (activities, child workflows, timers, etc.) and **must return a value** to the caller.

Simple state mutation with immediate return:

```twf
update ChangePlan(newPlan: string) -> (ChangeResult):
    plan = newPlan
    return ChangeResult{success: true, plan: plan}
```

Validation via activity before accepting mutation:

```twf
update ChangePlan(newPlan: string) -> (ChangeResult):
    activity ValidatePlan(newPlan) -> validation
    if (validation.valid):
        plan = newPlan
        return ChangeResult{success: true, plan: plan}
    else:
        return ChangeResult{success: false, error: validation.reason}
```

The caller blocks until the handler returns — including any time spent waiting on activities, child workflows, or timers within the handler.

### Handler Execution Semantics

Signal and update handlers run as coroutines alongside the main workflow body, but **only one piece of workflow code runs at a time** (cooperative scheduling). When a workflow wakes up, it processes pending messages (signals/updates) in order, then makes progress in the main workflow body.

This means:
1. The update handler runs as part of the workflow execution loop
2. If the handler blocks (on an activity, timer, etc.), the main workflow body can make progress while it waits
3. The handler reads from and writes to the same shared workflow state as the main body and signal handlers
4. The caller only receives a response after the handler has completed and returned

**Update handlers cannot call `close`** — they can mutate state and return values, but only the main workflow body can terminate the workflow.

### Awaiting Updates

Updates can be awaited in the workflow body, similar to signals. This is useful when the main workflow body needs to wait for an external mutation before continuing:

```twf
await update ChangeAddress
```

Updates can also race against other operations in `await one`:

```twf
await one:
    update ChangeAddress -> (newAddress):
        activity NotifyShipping(orderId, newAddress)
    timer(1h):
        activity FinalizeShipping(orderId)
```

When the update wins the race, its handler body runs and returns a value to the caller, then the case body executes.

### Update Handlers with Conditions

A common pattern has an update handler wait on workflow state using `condition`. The caller blocks until the condition becomes true, then receives a result reflecting the current state:

```twf
workflow ClusterManager(config: Config):
    state:
        condition clusterStarted

    signal Shutdown():
        shutdownRequested = true

    update WaitUntilStarted() -> (ClusterState):
        await clusterStarted
        return ClusterState{started: true}

    # Main body provisions and starts the cluster
    activity ProvisionCluster(config)
    activity StartCluster(config)
    set clusterStarted

    await signal Shutdown
    close complete
```

In this pattern:
1. The client calls the update and blocks waiting for a result
2. The update handler starts running but yields on `await clusterStarted`
3. The main workflow body mutates the condition via `set clusterStarted`
4. The update handler resumes and returns a value
5. The client receives the result

See [promises-conditions.md](./promises-conditions.md) for more on conditions and the `state:` block.

### Updates vs Signals

| Aspect | Signal (write) | Update (read-write) |
|--------|--------|--------|
| **Response** | None (fire-and-forget) | Returns result to caller |
| **Validation** | In handler, but caller doesn't know | Caller receives validation errors |
| **Confirmation** | No guarantee processing happened | Caller knows when complete |
| **Handler can block** | Yes, but caller doesn't wait | Yes, and caller blocks until done |
| **Use when** | "Notify workflow of X" | "Change X and tell me if it worked" |

### Update Considerations

| Consideration | Guidance |
|---------------|----------|
| **Atomicity** | Update handlers should be atomic; don't leave partial state |
| **Validation** | Validate before mutating; return errors, don't throw |
| **Idempotency** | Consider idempotency keys for critical updates |
| **Timeouts** | Caller should set appropriate timeout; handler may block on activities or state |
| **Shared state** | Handler reads/writes the same state as the main workflow body and signal handlers |
| **Ambient arrival** | Like signals, updates can arrive between any two deterministic steps and are buffered until handled by `await update` or `await one`/`await all` |

---

## Choosing Between Primitives

**Use QUERY (read) when:**
- Need to read current state
- Building UI/dashboard
- Debugging/monitoring

**Use SIGNAL (write) when:**
- Fire-and-forget is acceptable
- External event notification
- No response needed

**Use UPDATE (read-write) when:**
- Need confirmation that a change was applied
- Validating input before accepting
- Returning a computed result from a mutation

---

## Signal/Query/Update Naming Conventions

| Type | Convention | Examples |
|------|------------|----------|
| Signals | Event-style, past tense or imperative | `PaymentReceived`, `Cancel`, `AddItem` |
| Queries | Getter-style, "Get" prefix | `GetStatus`, `GetProgress`, `GetItems` |
| Updates | Action-style, verb phrase | `ChangePlan`, `AddCredits`, `UpdateAddress` |

