---
name: temporal-workflow-design
description: Design Temporal workflows with proper determinism, idempotency, and decomposition. Use when designing new workflows, planning workflow-activity boundaries, or reviewing workflow architecture.
---

# Temporal Workflow Design

Language-agnostic guide for designing Temporal workflows with correct separation of concerns, determinism, and failure handling.

## Core Principles

### 1. Determinism: Workflows Must Replay Identically

Temporal replays workflow code to reconstruct state. If replay produces different results than original execution, the workflow fails with non-determinism errors.

| Safe in Workflows | Must Be in Activities |
|-------------------|----------------------|
| Logic based on activity results | Current time, dates |
| Deterministic loops and conditionals | Random numbers, UUIDs |
| Child workflows | HTTP/API calls |
| Timers (Temporal-provided) | Database operations |
| Local variable manipulation | File I/O |
| Waiting on signals | External service calls |

**Mental model:** Workflows are pure orchestration. Activities are where side effects happen.

### 2. Idempotency: Activities May Execute Multiple Times

Network failures, worker crashes, or timeouts cause retries. Design activities to be **idempotent**: same inputs → same result, no matter how many times executed.

| Pattern | Example |
|---------|---------|
| **Create-or-get** | Check if resource exists before creating |
| **Idempotency keys** | Use workflow ID + activity name as operation key |
| **Upsert** | Prefer upserts over insert-then-update |
| **Deduplication** | Query for existing before mutating |

**Think through retries:**
- "CreateUser" → What if user exists? Return existing.
- "SendEmail" → Use provider's idempotency key.
- "DeployResource" → Verify state, return success if already deployed.

---

## Workflow vs Activity Boundary

### When to Use Activities

- Single atomic operation
- External system interaction (API, DB, file)
- Operation completes in bounded time
- No orchestration logic needed

### When to Use Child Workflows

- Multiple steps with independent retry/timeout policies
- Reusable logic across parent workflows
- Need separate failure boundaries
- Very long operations (separate history)
- Logic is complex enough to warrant its own tests

**Rule of thumb:** If you're tempted to put loops or conditionals inside an activity, it should probably be a workflow.

For detailed patterns, see [child-workflows.md](./topics/child-workflows.md).

---

## Temporal Primitives Reference

Quick reference for all Temporal primitives. Each links to detailed sub-skill documentation.

### Workflow Execution

| Primitive | Purpose | Details |
|-----------|---------|---------|
| `activity` | Execute side-effecting operation | Core primitive |
| `workflow` | Execute child workflow | [child-workflows.md](./topics/child-workflows.md) |
| `nexus` | Cross-namespace workflow call | [nexus.md](./topics/nexus.md) |
| `spawn` | Start async child workflow or nexus call | [child-workflows.md](./topics/child-workflows.md), [nexus.md](./topics/nexus.md) |
| `detach` | Fire-and-forget child workflow or nexus call | [child-workflows.md](./topics/child-workflows.md), [nexus.md](./topics/nexus.md) |
| `continue_as_new` | Reset history, continue with new input | [long-running.md](./topics/long-running.md) |

### Timing

| Primitive | Purpose | Details |
|-----------|---------|---------|
| `timer` | Durable sleep (survives restarts) | [timers-scheduling.md](./topics/timers-scheduling.md) |
| `schedule` | Cron-like recurring execution | [timers-scheduling.md](./topics/timers-scheduling.md) |
| `timeout` | Deadline for operations | [timers-scheduling.md](./topics/timers-scheduling.md) |

### External Communication

| Primitive | Purpose | Details |
|-----------|---------|---------|
| `signal` | Async event sent to workflow | [signals-queries-updates.md](./topics/signals-queries-updates.md) |
| `query` | Synchronous read of workflow state | [signals-queries-updates.md](./topics/signals-queries-updates.md) |
| `update` | Synchronous mutation of workflow state | [signals-queries-updates.md](./topics/signals-queries-updates.md) |

### Activity Options

| Primitive | Purpose | Details |
|-----------|---------|---------|
| `heartbeat` | Report progress, detect worker death | [activities-advanced.md](./topics/activities-advanced.md) |
| `async_complete` | Complete activity from external system | [activities-advanced.md](./topics/activities-advanced.md) |

### Infrastructure

| Primitive | Purpose | Details |
|-----------|---------|---------|
| `task_queue` | Route work to specific workers | [task-queues.md](./topics/task-queues.md) |
| `search_attribute` | Index workflow for queries | Core primitive |
| `memo` | Attach metadata to workflow | Core primitive |

---

## Workflow Notation (DSL)

Use this pseudo-code notation to document workflow designs. It's language-agnostic and captures Temporal semantics.

### Basic Structure

```twf
workflow WorkflowName(input: InputType) -> OutputType:
    activity ActivityName(args) -> result
    workflow ChildWorkflowName(args) -> childResult
    close OutputType{result, childResult}
```

### Control Flow

```twf
workflow ProcessOrder(order: Order) -> Result:
    activity ValidateOrder(order) -> validated
    
    # Conditionals
    if validated.priority == "high":
        activity ExpediteOrder(order)
    else:
        activity StandardProcessing(order)
    
    # Loops
    for item in order.items:
        activity ProcessItem(item)
    
    # Parallel execution
    await all:
        activity ReserveInventory(order) -> inventory
        activity ProcessPayment(order) -> payment
    
    close Result{inventory, payment}
```

### Temporal Primitives in Notation

```twf
workflow OrderFulfillment(orderId: string) -> OrderResult:
    activity GetOrder(orderId) -> order

    # Durable timer
    await timer(1h)

    # Wait for signal with timeout
    await one:
        signal PaymentReceived:
        timer(24h):
            activity CancelOrder(orderId)
            close failed OrderResult{status: "cancelled"}

    # Child workflow
    workflow ShipOrder(order)

    # Cross-namespace call
    nexus "notifications" workflow SendNotification(order.customer, "shipped")

    close OrderResult{status: "completed"}

# Signal definitions
signal PaymentReceived(transactionId: string, amount: decimal):

# Query definitions
query GetOrderStatus() -> OrderStatus:
    return currentStatus

# Update definitions
update UpdateShippingAddress(address: Address) -> Result:
    order.shippingAddress = address
    return Result{success: true}
```

### Notation Reference

| Syntax | Meaning |
|--------|---------|
| `activity Name(args) -> result` | Activity call with result binding |
| `workflow Name(args) -> result` | Child workflow call with result binding |
| `nexus "namespace" workflow Name(args) -> result` | Nexus cross-namespace call |
| `spawn workflow Name(args) -> handle` | Start async child workflow, get handle |
| `spawn nexus "ns" workflow Name(args) -> handle` | Start async nexus call, get handle |
| `detach workflow Name(args)` | Fire-and-forget child workflow |
| `detach nexus "ns" workflow Name(args)` | Fire-and-forget nexus call |
| `await timer(duration)` | Durable sleep |
| `await signal Name` | Wait for signal |
| `await one:` | Wait for first of multiple operations |
| `await all:` | Wait for all operations |
| `options(key: value)` | Set options on next statement |
| `-> result` | Bind result of preceding operation |
| `close [completed\|failed] Value` | End workflow with result or failure |
| `if/else` | Conditional (deterministic) |
| `for x in collection:` | Bounded loop |
| `for:` | Infinite loop (use with `continue_as_new` or `close`) |
| `switch/case` | Multi-branch conditional |
| `continue_as_new(args)` | Reset and continue |
| `signal Name:` | Signal handler definition |
| `query Name():` | Query handler definition |
| `update Name():` | Update handler definition |

---

## Design Checklist

Before implementing, verify:

### Determinism
- [ ] All I/O, time, randomness is in activities
- [ ] No external calls in workflow code
- [ ] Loops have deterministic bounds
- [ ] Timer waits use Temporal primitives

### Idempotency
- [ ] Each activity handles "already exists" gracefully
- [ ] Retries produce same end state
- [ ] No duplicate side effects on replay

### Failure Handling
- [ ] Each failure mode identified
- [ ] Recovery strategy defined (retry, compensate, fail)
- [ ] Partial success handling specified
- [ ] Timeouts configured appropriately

### Decomposition
- [ ] Each workflow has single clear purpose
- [ ] Child workflow vs activity choice justified
- [ ] Workflow names describe outcomes, not steps

---

## Common Anti-Patterns

### Non-Determinism in Workflows

```twf
# BAD: Time check in workflow
if current_time() > deadline:
    cancel()

# GOOD: Timer-based deadline
await one:
    activity DoWork() -> result:
        close Result{status: "success"}
    timer(deadline):
        close failed Result{status: "timeout"}
```

### Non-Idempotent Activities

> Note: Activity bodies contain SDK-level implementation code, not TWF notation.

```python
# BAD: Assumes fresh state
def CreateUser(name):
    db.insert(User(name))  # Fails on retry

# GOOD: Handles existing state
def CreateUser(name):
    existing = db.get_by_name(name)
    if existing: return existing
    return db.insert(User(name))
```

### Orchestration in Activities

```twf
# BAD: Loop inside activity (activity bodies are SDK code, not TWF)
# activity DeployAll(specs):
#     for spec in specs:
#         deploy(spec)           # What if fails mid-loop?
#         wait_healthy(spec)

# GOOD: Workflow handles orchestration
workflow DeployAll(specs):
    for spec in specs:
        activity Deploy(spec)
        activity WaitHealthy(spec)
```

---

## Sub-Skills Reference

For detailed coverage of specific topics:

| Topic | File |
|-------|------|
| Signals, Queries, Updates | [signals-queries-updates.md](./topics/signals-queries-updates.md) |
| Child Workflows | [child-workflows.md](./topics/child-workflows.md) |
| Timers and Scheduling | [timers-scheduling.md](./topics/timers-scheduling.md) |
| Advanced Activities | [activities-advanced.md](./topics/activities-advanced.md) |
| Long-Running Workflows | [long-running.md](./topics/long-running.md) |
| Nexus (Cross-Namespace) | [nexus.md](./topics/nexus.md) |
| Task Queues and Scaling | [task-queues.md](./topics/task-queues.md) |
| Workflow Patterns | [patterns.md](./topics/patterns.md) |
| Testing Workflows | [testing.md](./topics/testing.md) |
| Versioning and Evolution | [versioning.md](./topics/versioning.md) |
