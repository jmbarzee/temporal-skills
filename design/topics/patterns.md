# Workflow Patterns

> **Example:** [`patterns.twf`](./patterns.twf)

Common patterns for structuring Temporal workflows. Choose based on your use case characteristics.

## Pattern Overview

| Pattern | Use When | Example |
|---------|----------|---------|
| **Process** | Discrete operation with start and end | Order fulfillment |
| **Entity** | Long-lived, represents a thing | User account |
| **Saga** | Distributed transaction with compensation | Multi-service booking |
| **Fan-Out/Fan-In** | Parallel processing with aggregation | Batch processing |
| **Pipeline** | Sequential stages of transformation | Data processing |
| **State Machine** | Explicit state transitions | Document approval |
| **Polling** | Wait for external condition | Resource provisioning |

---

## Process Workflow

A discrete operation that drives toward completion.

### Characteristics
- Has a clear start and end
- Progresses through defined steps
- Returns a result when complete
- Relatively short-lived (minutes to hours)

### Pattern

```twf
workflow OrderFulfillment(order: Order) -> OrderResult:
    # Step 1: Validate
    activity ValidateOrder(order) -> validated
    if not validated.success:
        close failed OrderResult{status: "invalid", error: validated.error}
    
    # Step 2: Reserve
    activity ReserveInventory(order.items) -> reservation
    
    # Step 3: Charge
    activity ProcessPayment(order.payment) -> payment
    
    # Step 4: Fulfill
    activity ShipOrder(order, reservation)
    
    # Step 5: Notify
    activity SendConfirmation(order.customer)
    
    close OrderResult{status: "completed", trackingId: reservation.trackingId}
```

### When to Use
- Order processing
- User registration
- Deployment pipelines
- Report generation

---

## Entity Workflow

A long-running workflow representing a business entity.

### Characteristics
- Long-lived (days, months, indefinitely)
- Reacts to external events (signals)
- Maintains state over time
- Uses continue-as-new to manage history

### Pattern

```twf
workflow AccountEntity(accountId: string, state: AccountState) -> void:
    if state == null:
        activity LoadAccount(accountId) -> state

    for:
        await one:
            signal Deposit:
                state.balance += signal.amount
                activity RecordTransaction(accountId, "deposit", signal.amount)

            signal Withdraw:
                if state.balance >= signal.amount:
                    state.balance -= signal.amount
                    activity RecordTransaction(accountId, "withdraw", signal.amount)

            signal Close:
                activity CloseAccount(accountId)
                close

            timer(24h):
                activity DailyReconciliation(accountId, state)

        if history_size() > 1000:
            continue_as_new(accountId, state)

query GetBalance() -> decimal:
    return state.balance

update Transfer(amount: decimal, toAccount: string) -> TransferResult:
    if state.balance < amount:
        return TransferResult{success: false, error: "insufficient funds"}
    state.balance -= amount
    activity InitiateTransfer(toAccount, amount)
    return TransferResult{success: true}
```

### When to Use
- User accounts
- Subscriptions
- Shopping carts
- IoT device state
- Game sessions

---

## Saga Pattern

Distributed transaction with compensation for failures.

### Characteristics
- Multiple services/steps that must all succeed or all roll back
- Each step has a compensating action
- Compensation runs in reverse order on failure
- Provides eventual consistency

### Pattern

> Note: The saga pattern requires error-handling constructs (try/catch, compensation stacks) that are expressed here as conceptual pseudo-code. See [`patterns.twf`](./patterns.twf) for the TWF syntax version.

```pseudo
workflow BookingWorkflow(booking: Booking) -> BookingResult:
    # Step 1: Reserve flight
    activity ReserveFlight(booking.flight) -> flight
    
    # Step 2: Reserve hotel (compensate flight on failure)
    activity ReserveHotel(booking.hotel) -> hotel
    # on failure: activity CancelFlight(flight.id)
    
    # Step 3: Reserve car (compensate hotel + flight on failure)
    activity ReserveCar(booking.car) -> car
    # on failure: activity CancelHotel(hotel.id), activity CancelFlight(flight.id)
    
    # Step 4: Charge payment (compensate all on failure)
    activity ChargePayment(booking.payment) -> payment
    # on failure: activity CancelCar(car.id), CancelHotel(...), CancelFlight(...)
    
    # All succeeded
    close BookingResult{status: "confirmed", flight, hotel, car, payment}
    
    # On any step failure, compensations run in reverse order
    # SDK-level error handling drives the compensation logic
```

### Compensation Design

| Step | Forward Action | Compensation |
|------|---------------|--------------|
| Reserve | Create pending reservation | Cancel reservation |
| Charge | Process payment | Refund payment |
| Ship | Create shipment | Cancel shipment |
| Provision | Create resource | Delete resource |

### When to Use
- Multi-service transactions
- Booking systems (travel, events)
- Financial operations
- Resource provisioning

---

## Fan-Out/Fan-In Pattern

Process items in parallel, aggregate results.

### Characteristics
- Split work into parallel tasks
- Each task executes independently
- Aggregate results when all complete
- Handle partial failures

### Pattern

> Note: The TWF DSL currently re-binds the result variable on each iteration of `await all: for`. The aggregation step below is expressed as conceptual pseudo-code. See [`patterns.twf`](./patterns.twf) for the TWF syntax version.

```twf
workflow BatchProcessor(items: []Item) -> BatchResult:
    # Fan-out: start all processing in parallel
    await all:
        for (item in items):
            activity ProcessItem(item) -> result
    
    # Fan-in: aggregate results (conceptual -- SDK collects results)
    activity AggregateResults(items) -> aggregated
    
    close BatchResult{results: aggregated}
```

### Variations

**With Concurrency Limit:**
```twf
workflow RateLimitedBatch(items: []Item) -> BatchResult:
    # Process in batches of 10
    for (batch in chunk(items, 10)):
        await all:
            for (item in batch):
                activity ProcessItem(item) -> result
    
    close BatchResult{}
```

**With Early Exit:**
```twf
workflow FirstSuccessful(sources: []Source) -> Data:
    await all:
        for (source in sources):
            activity TryFetch(source) -> result
    
    # SDK-level: find first successful result
    activity FindFirstSuccess(sources) -> data
    close data
```

### When to Use
- Batch processing
- Parallel API calls
- Distributed computation
- Report aggregation

---

## Pipeline Pattern

Sequential transformation stages.

### Characteristics
- Data flows through ordered stages
- Each stage transforms and passes to next
- Clear separation of concerns
- Easy to add/remove stages

### Pattern

```twf
workflow DataPipeline(rawData: RawData) -> ProcessedData:
    # Stage 1: Ingest
    activity Ingest(rawData) -> ingested
    
    # Stage 2: Validate
    activity Validate(ingested) -> validated
    if not validated.valid:
        close failed ProcessedData{status: "invalid", errors: validated.errors}
    
    # Stage 3: Transform
    activity Transform(validated.data) -> transformed
    
    # Stage 4: Enrich
    activity Enrich(transformed) -> enriched
    
    # Stage 5: Load
    activity Load(enriched)
    
    close ProcessedData{status: "complete", recordCount: enriched.count}
```

### With Conditional Stages

```twf
workflow AdaptivePipeline(data: Data) -> Result:
    activity Parse(data) -> processed
    
    if processed.needsEnrichment:
        activity Enrich(processed) -> processed
    
    if processed.format == "legacy":
        activity ConvertLegacy(processed) -> processed
    
    activity Finalize(processed) -> result
    close result
```

### When to Use
- ETL processes
- Document processing
- Media transcoding
- Data migrations

---

## State Machine Pattern

Explicit states and transitions.

### Characteristics
- Well-defined states
- Explicit transition rules
- Events trigger transitions
- Clear audit trail

### Pattern

```twf
workflow DocumentApproval(doc: Document) -> ApprovalResult:
    state = "draft"

    for:
        if state == "draft":
            await signal Submit
            activity NotifyReviewers(doc)
            state = "pending_review"

        elif state == "pending_review":
            await one:
                signal Approve:
                    state = "approved"
                signal Reject:
                    state = "rejected"
                signal RequestChanges:
                    state = "changes_requested"

        elif state == "changes_requested":
            await one:
                signal Submit:
                    state = "pending_review"
                signal Withdraw:
                    state = "withdrawn"

        elif state == "approved":
            activity PublishDocument(doc)
            close ApprovalResult{status: "approved"}

        elif state == "rejected":
            activity ArchiveDocument(doc)
            close ApprovalResult{status: "rejected"}

        elif state == "withdrawn":
            close ApprovalResult{status: "withdrawn"}

query GetState() -> string:
    return state
```

### State Transition Table

| From State | Event | To State | Action |
|------------|-------|----------|--------|
| draft | Submit | pending_review | Notify reviewers |
| pending_review | Approve | approved | None |
| pending_review | Reject | rejected | None |
| pending_review | RequestChanges | changes_requested | None |
| changes_requested | Submit | pending_review | None |
| changes_requested | Withdraw | withdrawn | None |

### When to Use
- Approval workflows
- Order status tracking
- Support tickets
- Insurance claims

---

## Polling Pattern

Wait for external condition to be met.

### Characteristics
- External system doesn't push updates
- Must poll periodically
- Need backoff strategy
- Has timeout/deadline

### Pattern

```twf
workflow WaitForResource(resourceId: string) -> Resource:
    backoff = 5s
    maxBackoff = 60s

    for:
        activity CheckResourceStatus(resourceId) -> status

        if status.ready:
            activity GetResource(resourceId) -> resource
            close resource

        if status.failed:
            activity CancelProvisioning(resourceId)
            close failed ProvisioningError{error: status.error}

        # Wait with backoff, deadline via await one + timer
        await one:
            timer(backoff):
                backoff = min(backoff * 2, maxBackoff)
            timer(30m):
                activity CancelProvisioning(resourceId)
                close failed ProvisioningTimeout{}
```

### With Progress Updates

> Note: `upsert_search_attributes` is an SDK-level call, not TWF notation.

```twf
workflow MonitorJob(jobId: string) -> JobResult:
    for:
        activity GetJobStatus(jobId) -> status

        # Update search attributes for visibility (SDK call)
        # upsert_search_attributes({JobProgress: status.percentComplete, JobStage: status.currentStage})

        if status.complete:
            activity GetJobResult(jobId) -> result
            close result

        await timer(30s)
```

### When to Use
- Resource provisioning
- External job monitoring
- Third-party integrations
- CI/CD pipelines

---

## Pattern Selection Guide

```text
Start
  │
  ├─ Is it a long-lived entity? ──────────────► Entity Pattern
  │
  ├─ Does it need distributed rollback? ──────► Saga Pattern
  │
  ├─ Can items be processed in parallel? ─────► Fan-Out/Fan-In
  │
  ├─ Is it a series of transformations? ──────► Pipeline Pattern
  │
  ├─ Are there explicit states/transitions? ──► State Machine
  │
  ├─ Need to wait for external condition? ────► Polling Pattern
  │
  └─ Simple start-to-finish process? ─────────► Process Workflow
```
