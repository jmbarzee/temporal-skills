# Signals, Queries, and Updates

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

### Design Pattern

```
workflow OrderWorkflow(orderId: string) -> OrderResult:
    order = activity GetOrder(orderId)
    
    # Wait for payment signal
    await signal PaymentReceived:
        timeout: 24h
        on_timeout:
            return OrderResult{status: "payment_timeout"}
    
    # Continue processing after signal
    activity FulfillOrder(order)
    return OrderResult{status: "completed"}

signal PaymentReceived:
    input: {transactionId: string, amount: decimal}
    handler:
        order.paymentId = transactionId
        order.amountPaid = amount
```

### Signal Considerations

| Consideration | Guidance |
|---------------|----------|
| **Ordering** | Signals are processed in order received, but arrival order isn't guaranteed |
| **Buffering** | Signals queue if workflow is busy; consider signal coalescing for high-volume |
| **Idempotency** | Signal handlers should be idempotent (same signal twice = same result) |
| **Validation** | Validate signal payload; invalid signals can corrupt workflow state |

### Common Patterns

**Approval Flow:**
```
workflow ApprovalWorkflow(request: Request) -> Decision:
    activity NotifyApprovers(request)
    
    await signal Approved or signal Rejected:
        timeout: 7d
        on_timeout: return Decision{status: "expired"}
    
    if received Approved:
        return Decision{status: "approved", approver: signal.approver}
    else:
        return Decision{status: "rejected", reason: signal.reason}

signal Approved:
    input: {approver: string}

signal Rejected:
    input: {approver: string, reason: string}
```

**Data Accumulation:**
```
workflow BatchCollector(batchId: string) -> Batch:
    items = []
    
    # Collect items until deadline or explicit completion
    loop:
        await signal AddItem or signal CompleteBatch or timer 1h:
            on AddItem: items.append(signal.item)
            on CompleteBatch: break
            on timer: break
    
    activity ProcessBatch(items)
    return Batch{items: items}

signal AddItem:
    input: {item: Item}

signal CompleteBatch:
    input: {}
```

---

## Queries

Synchronous, read-only access to workflow state. The caller blocks until the query returns.

### When to Use

- UI needs to display current workflow state
- Monitoring/debugging workflow progress
- External system needs workflow data
- Building workflow dashboards

### Design Pattern

```
workflow OrderWorkflow(orderId: string) -> OrderResult:
    status = "pending"
    items = []
    
    status = "validating"
    activity ValidateOrder(orderId)
    
    status = "processing"
    for item in order.items:
        activity ProcessItem(item)
        items.append(item)
    
    status = "completed"
    return OrderResult{status: status}

query GetStatus() -> string:
    return status

query GetProgress() -> Progress:
    return Progress{
        status: status,
        itemsProcessed: len(items),
        totalItems: len(order.items)
    }
```

### Query Considerations

| Consideration | Guidance |
|---------------|----------|
| **Read-only** | Queries MUST NOT modify workflow state |
| **Determinism** | Query handlers run during replay; must be deterministic |
| **Performance** | Queries replay workflow history; expensive for long histories |
| **Consistency** | Returns point-in-time state; may be stale by the time caller uses it |

### Anti-Patterns

```
# BAD: Query modifies state
query GetAndIncrementCounter() -> int:
    counter += 1  # NOT ALLOWED
    return counter

# BAD: Query has side effects
query GetStatus() -> string:
    log("Query received")  # Non-deterministic side effect
    return status

# GOOD: Pure read
query GetStatus() -> string:
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

### Design Pattern

```
workflow SubscriptionWorkflow(userId: string) -> void:
    plan = "free"
    
    # Long-running workflow
    loop:
        await signal Cancel or timer 30d:
            on Cancel: break
            on timer: activity BillUser(userId, plan)

update ChangePlan(newPlan: string) -> ChangeResult:
    # Validate
    if newPlan not in ["free", "pro", "enterprise"]:
        return ChangeResult{success: false, error: "invalid plan"}
    
    # Apply
    oldPlan = plan
    plan = newPlan
    
    # Confirm
    return ChangeResult{
        success: true,
        oldPlan: oldPlan,
        newPlan: newPlan
    }

update AddCredits(amount: int) -> CreditResult:
    if amount <= 0:
        return CreditResult{success: false, error: "amount must be positive"}
    
    credits += amount
    return CreditResult{success: true, newBalance: credits}
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

---

## Choosing Between Primitives

```
# Use SIGNAL when:
# - Fire-and-forget is acceptable
# - External event notification
# - No response needed
signal OrderShipped:
    input: {trackingNumber: string}

# Use QUERY when:
# - Need to read current state
# - Building UI/dashboard
# - Debugging/monitoring
query GetOrderStatus() -> OrderStatus

# Use UPDATE when:
# - Need confirmation of change
# - Validating input before accepting
# - Request-response mutation
update CancelOrder(reason: string) -> CancelResult
```

---

## Signal/Query/Update Naming Conventions

| Type | Convention | Examples |
|------|------------|----------|
| Signals | Event-style, past tense or imperative | `PaymentReceived`, `Cancel`, `AddItem` |
| Queries | Getter-style, "Get" prefix | `GetStatus`, `GetProgress`, `GetItems` |
| Updates | Action-style, verb phrase | `ChangePlan`, `AddCredits`, `UpdateAddress` |
