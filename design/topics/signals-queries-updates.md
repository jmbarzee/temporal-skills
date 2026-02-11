# Signals, Queries, and Updates

> **Example:** [`signals-queries-updates.twf`](./signals-queries-updates.twf)

External communication with running workflows. These primitives let code outside the workflow interact with it during execution.

## Overview

| Primitive | Direction | Execution | Use Case |
|-----------|-----------|-----------|----------|
| **Signal** | External → Workflow | Async (fire-and-forget) | Events, notifications, data injection |
| **Query** | External → Workflow → External | Sync (request-response) | Read current state |
| **Update** | External → Workflow → External | Sync (request-response) | Mutate state with confirmation |

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
            close OrderResult{status: "completed"}
        timer(24h):
            close failed OrderResult{status: "timeout"}
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

Synchronous mutations with confirmation. Caller sends data, workflow processes it, caller receives result.

### When to Use

- Need confirmation that change was applied
- Validating input before accepting
- Returning computed result from mutation
- Request-response pattern with workflow

### Update Handler Bodies

Updates are declared with handler body blocks that execute when the update is received. Handler bodies have access to the full workflow statement set (activities, child workflows, timers, etc.) and can return values to the caller.

```twf
update ChangePlan(newPlan: string) -> (ChangeResult):
    plan = newPlan
    activity PersistChange(newPlan)
    return ChangeResult{success: true, plan: plan}
```

### Updates vs Signals

| Aspect | Signal | Update |
|--------|--------|--------|
| **Response** | None (fire-and-forget) | Returns result to caller |
| **Validation** | In handler, but caller doesn't know | Caller receives validation errors |
| **Confirmation** | No guarantee processing happened | Caller knows when complete |
| **Use when** | "Notify workflow of X" | "Change X and tell me if it worked" |

### Update Considerations

| Consideration | Guidance |
|---------------|----------|
| **Atomicity** | Update handlers should be atomic; don't leave partial state |
| **Validation** | Validate before mutating; return errors, don't throw |
| **Idempotency** | Consider idempotency keys for critical updates |
| **Timeouts** | Caller should set appropriate timeout; update may wait for workflow |
| **Ambient arrival** | Like signals, updates can arrive between any two deterministic steps and are buffered until handled by `await update` or `await one`/`await all` |

---

## Choosing Between Primitives

**Use SIGNAL when:**
- Fire-and-forget is acceptable
- External event notification
- No response needed

**Use QUERY when:**
- Need to read current state
- Building UI/dashboard
- Debugging/monitoring

**Use UPDATE when:**
- Need confirmation of change
- Validating input before accepting
- Request-response mutation

---

## Signal/Query/Update Naming Conventions

| Type | Convention | Examples |
|------|------------|----------|
| Signals | Event-style, past tense or imperative | `PaymentReceived`, `Cancel`, `AddItem` |
| Queries | Getter-style, "Get" prefix | `GetStatus`, `GetProgress`, `GetItems` |
| Updates | Action-style, verb phrase | `ChangePlan`, `AddCredits`, `UpdateAddress` |

