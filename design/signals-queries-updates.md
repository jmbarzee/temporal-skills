# Signals, Queries, and Updates

> **Example:** [`examples/signals-queries-updates.twf`](./examples/signals-queries-updates.twf)

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

```
signal PaymentReceived(transactionId: string, amount: decimal):
    paymentStatus = "received"
    activity FulfillOrder(order)
```

The handler body executes when the signal arrives, whether via `await signal` or `select` with `hint signal` annotations.

### Signal Considerations

| Consideration | Guidance |
|---------------|----------|
| **Ordering** | Signals are processed in order received, but arrival order isn't guaranteed |
| **Buffering** | Signals queue if workflow is busy; consider signal coalescing for high-volume |
| **Idempotency** | Signal handlers should be idempotent (same signal twice = same result) |
| **Validation** | Validate signal payload; invalid signals can corrupt workflow state |
| **Ambient arrival** | Signals can arrive between any two deterministic steps; use `hint` to annotate points where signals may arrive |

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

```
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

```
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

```
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
| **Ambient arrival** | Like signals, updates can arrive between any two deterministic steps; use `hint` to annotate points where updates may arrive |

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

