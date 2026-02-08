# Workflow Patterns

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

```
workflow OrderFulfillment(order: Order) -> OrderResult:
    # Step 1: Validate
    validated = activity ValidateOrder(order)
    if not validated.success:
        return OrderResult{status: "invalid", error: validated.error}
    
    # Step 2: Reserve
    reservation = activity ReserveInventory(order.items)
    
    # Step 3: Charge
    payment = activity ProcessPayment(order.payment)
    
    # Step 4: Fulfill
    activity ShipOrder(order, reservation)
    
    # Step 5: Notify
    activity SendConfirmation(order.customer)
    
    return OrderResult{status: "completed", trackingId: reservation.trackingId}
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

```
workflow AccountEntity(accountId: string, state: AccountState) -> void:
    if state == null:
        state = activity LoadAccount(accountId)
    
    loop:
        select:
            signal Deposit:
                state.balance += signal.amount
                activity RecordTransaction(accountId, "deposit", signal.amount)
            
            signal Withdraw:
                if state.balance >= signal.amount:
                    state.balance -= signal.amount
                    activity RecordTransaction(accountId, "withdraw", signal.amount)
            
            signal Close:
                activity CloseAccount(accountId)
                return
            
            timer 24h:
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

```
workflow BookingWorkflow(booking: Booking) -> BookingResult:
    compensations = []
    
    try:
        # Step 1: Reserve flight
        flight = activity ReserveFlight(booking.flight)
        compensations.push(() => activity CancelFlight(flight.id))
        
        # Step 2: Reserve hotel
        hotel = activity ReserveHotel(booking.hotel)
        compensations.push(() => activity CancelHotel(hotel.id))
        
        # Step 3: Reserve car
        car = activity ReserveCar(booking.car)
        compensations.push(() => activity CancelCar(car.id))
        
        # Step 4: Charge payment
        payment = activity ChargePayment(booking.payment)
        compensations.push(() => activity RefundPayment(payment.id))
        
        # All succeeded
        return BookingResult{status: "confirmed", flight, hotel, car, payment}
    
    catch as error:
        # Run compensations in reverse
        for compensation in reversed(compensations):
            try:
                compensation()
            catch as compError:
                activity AlertCompensationFailure(compError)
        
        return BookingResult{status: "failed", error: error.message}
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

```
workflow BatchProcessor(items: []Item) -> BatchResult:
    # Fan-out: start all processing in parallel
    parallel:
        for item in items:
            results[item.id] = activity ProcessItem(item)
    
    # Fan-in: aggregate results
    successful = filter(results, r => r.success)
    failed = filter(results, r => not r.success)
    
    # Handle based on results
    if len(failed) > 0:
        activity AlertPartialFailure(failed)
    
    return BatchResult{
        processed: len(successful),
        failed: len(failed),
        results: results
    }
```

### Variations

**With Concurrency Limit:**
```
workflow RateLimitedBatch(items: []Item) -> BatchResult:
    results = []
    
    # Process in batches of 10
    for batch in chunk(items, 10):
        parallel:
            for item in batch:
                results.append(activity ProcessItem(item))
    
    return BatchResult{results}
```

**With Early Exit:**
```
workflow FirstSuccessful(sources: []Source) -> Data:
    parallel:
        for source in sources:
            results[source.id] = activity TryFetch(source)
    
    # Return first successful result
    for result in results:
        if result.success:
            return result.data
    
    raise AllSourcesFailed()
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

```
workflow DataPipeline(rawData: RawData) -> ProcessedData:
    # Stage 1: Ingest
    ingested = activity Ingest(rawData)
    
    # Stage 2: Validate
    validated = activity Validate(ingested)
    if not validated.valid:
        return ProcessedData{status: "invalid", errors: validated.errors}
    
    # Stage 3: Transform
    transformed = activity Transform(validated.data)
    
    # Stage 4: Enrich
    enriched = activity Enrich(transformed)
    
    # Stage 5: Load
    activity Load(enriched)
    
    return ProcessedData{status: "complete", recordCount: enriched.count}
```

### With Conditional Stages

```
workflow AdaptivePipeline(data: Data) -> Result:
    processed = activity Parse(data)
    
    if processed.needsEnrichment:
        processed = activity Enrich(processed)
    
    if processed.format == "legacy":
        processed = activity ConvertLegacy(processed)
    
    return activity Finalize(processed)
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

```
workflow DocumentApproval(doc: Document) -> ApprovalResult:
    state = "draft"
    
    loop:
        select state:
            case "draft":
                await signal Submit:
                    activity NotifyReviewers(doc)
                    state = "pending_review"
            
            case "pending_review":
                await signal Approve:
                    state = "approved"
                await signal Reject:
                    state = "rejected"
                await signal RequestChanges:
                    state = "changes_requested"
            
            case "changes_requested":
                await signal Submit:
                    state = "pending_review"
                await signal Withdraw:
                    state = "withdrawn"
            
            case "approved":
                activity PublishDocument(doc)
                return ApprovalResult{status: "approved"}
            
            case "rejected":
                activity ArchiveDocument(doc)
                return ApprovalResult{status: "rejected"}
            
            case "withdrawn":
                return ApprovalResult{status: "withdrawn"}

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

```
workflow WaitForResource(resourceId: string) -> Resource:
    deadline = now() + 30m
    backoff = 5s
    maxBackoff = 60s
    
    loop:
        status = activity CheckResourceStatus(resourceId)
        
        if status.ready:
            return activity GetResource(resourceId)
        
        if status.failed:
            raise ResourceProvisioningFailed(status.error)
        
        # Check deadline
        if now() > deadline:
            activity CancelProvisioning(resourceId)
            raise ProvisioningTimeout()
        
        # Wait with backoff
        timer backoff
        backoff = min(backoff * 2, maxBackoff)
```

### With Progress Updates

```
workflow MonitorJob(jobId: string) -> JobResult:
    loop:
        status = activity GetJobStatus(jobId)
        
        # Update search attributes for visibility
        upsert_search_attributes({
            JobProgress: status.percentComplete,
            JobStage: status.currentStage
        })
        
        if status.complete:
            return activity GetJobResult(jobId)
        
        timer 30s
```

### When to Use
- Resource provisioning
- External job monitoring
- Third-party integrations
- CI/CD pipelines

---

## Pattern Selection Guide

```
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
