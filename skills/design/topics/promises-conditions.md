# Promises and Conditions

> **Example:** [`promises-conditions.twf`](./promises-conditions.twf)

Deferred async operations and named boolean awaitables for workflow state.

## Overview

| Primitive | Purpose | Syntax |
|-----------|---------|--------|
| **Promise** | Start async operation, await later | `promise p <- activity Foo(x)` |
| **Condition** | Named boolean awaitable | `condition myCondition` (in `state:` block) |
| **Set / Unset** | Mutate a condition | `set myCondition` / `unset myCondition` |
| **State block** | Declare workflow state | `state:` (before handlers) |

---

## Promises

A `promise` wraps any async operation (activity, workflow, timer, signal, update) for non-blocking execution. The `<-` operator visually distinguishes async declaration from sync result binding (`->`).

### Declaration

```twf
promise p <- activity ProcessItem(input)
promise report <- workflow BuildReport(data)
promise timeout <- timer(5m)
promise approved <- signal Approved
promise addr <- update ChangeAddress
promise pay <- nexus PaymentsEndpoint PaymentsService.Charge(card)
```

### Awaiting a Promise

Block until the promise resolves and bind the result:

```twf
await p -> result
await timeout
```

### Promises in `await one` (race)

```twf
await one:
    p -> result:
        close complete(Result{data: result})
    timeout:
        close fail("timed out")
```

### Start-Now, Wait-Later Pattern

The primary use case for promises is starting operations without blocking, doing other work, then collecting results:

```twf
workflow ParallelProcessing(items: Items) -> (Result):
    promise handleA <- activity ProcessA(items.a)
    promise handleB <- activity ProcessB(items.b)

    activity QuickSetup(items)

    await handleA -> resultA
    await handleB -> resultB

    close complete(Result{a: resultA, b: resultB})
```

### Async Duality

Every async operation has two forms:

| Form | Syntax | Behavior |
|------|--------|----------|
| **Blocking** | `activity Process(item) -> result` | Start and wait immediately |
| **Non-blocking** | `promise p <- activity Process(item)` | Start, continue, wait later |

This applies uniformly to: `activity`, `workflow`, `timer`, `signal`, `update`.

---

## Conditions

A `condition` is a named boolean temporal primitive declared in the workflow `state:` block. It can be set, unset, and awaited.

### Declaration (in `state:` block only)

```twf
workflow Example():
    state:
        condition clusterStarted
        condition thresholdReached
```

### Mutation

```twf
set clusterStarted
unset clusterStarted
```

### Awaiting a Condition

```twf
await clusterStarted
```

### Conditions in `await one` (race)

```twf
await one:
    clusterStarted:
        close complete(ClusterState{started: true})
    timer(30d):
        close fail("timeout")
```

### Update Handler + Condition Pattern

The primary motivator for conditions: update handlers that wait on workflow state. The client blocks until the handler returns, and the handler waits for a condition that the main workflow body sets:

```twf
workflow ClusterManager(config: Config):
    state:
        condition clusterStarted

    update WaitUntilStarted() -> (ClusterState):
        await clusterStarted
        return ClusterState{started: true}

    activity ProvisionCluster(config)
    activity StartCluster(config)
    set clusterStarted

    await signal Shutdown
    close complete
```

---

## State Block

The `state:` block declares workflow state including conditions and variable initializations. It must appear first in a workflow definition, before signal/query/update handlers.

```twf
workflow Example():
    state:
        condition myCondition
        balance = 0
        status = "pending"

    signal Deposit(amount: decimal):
        balance = balance + amount
        if (balance >= 1000):
            set myCondition

    await myCondition
    close complete
```

### Restrictions

- No temporal primitives inside `state:` block (it is purely declarative)
- `condition` declarations can only appear inside `state:` blocks
- `set`/`unset` targets must refer to conditions declared in the `state:` block

---

## Condition Considerations

| Consideration | Guidance |
|---------------|----------|
| **Boolean only** | Conditions are simple true/false values |
| **Declarative** | Must be declared in `state:` block before use |
| **Signal-driven** | Typically set/unset in signal or update handlers |
| **Reactive** | `await condition` unblocks when condition becomes true |
| **No expressions** | Conditions are named booleans, not arbitrary predicates |
