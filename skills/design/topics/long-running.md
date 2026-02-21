# Long-Running Workflows

> **Example:** [`long-running.twf`](./long-running.twf)

Patterns for workflows that run for extended periods: continue-as-new, history management, and entity workflows.

## The History Problem

Temporal replays the full event history to reconstruct workflow state after any restart. This is the **primary constraint** on long-running workflows — history size directly determines replay cost, and Temporal enforces a hard limit (~50MB / ~50K events).

| Issue | Impact |
|-------|--------|
| Replay cost | **Entire history replayed on every recovery** — this is the main bottleneck |
| Hard limit | ~50MB event history / ~50K events — workflow terminates if exceeded |
| Memory | Full history loaded into worker memory during replay |
| Latency | Longer history = slower recovery after worker restart |

**Solution:** Reset history periodically with `continue_as_new`.

---

## Continue-As-New

Atomically complete current workflow and start a new execution with fresh history, preserving logical continuity.

### Basic Pattern

```twf
workflow LongRunningProcessor(processor: Processor):
    eventCount = 0
    
    for:
        await signal NewEvent -> (event)
        activity ProcessEvent(event)
        processor.processed += 1
        eventCount += 1
        
        # Reset history before it gets too large
        if eventCount >= 1000:
            close continue_as_new(processor)  # Fresh history, same logical workflow
```

### Continue-As-New Semantics

| Aspect | Behavior |
|--------|----------|
| Workflow ID | Same (logical continuity) |
| Run ID | New (fresh execution) |
| History | Reset to zero |
| Pending signals | Carried over (configurable) |
| State | Passed as input to new execution |

### When to Continue-As-New

| Trigger | Example |
|---------|---------|
| Event count | After processing N events |
| Time-based | Every 24 hours |
| History size | Approaching limit |
| Periodic reset | End of billing cycle |

### SDK Intrinsics for History Tracking

These deterministic SDK functions are available in workflow code (not activities) for deciding when to continue-as-new:

| Function | Returns | Use |
|----------|---------|-----|
| `workflow.history_length()` | Event count | Compare against threshold (e.g., `>= 1000`) |
| `workflow.history_size()` | Bytes | Compare against limit (e.g., `> 40_000_000`) |

These appear in TWF as raw expressions since they're SDK-level calls, not TWF keywords.

### Data Serialization

```twf
workflow EntityWorkflow(entity: Entity, data: EntityData):
    for:
        await signal Command -> (command)
        data = applyCommand(data, command)
        
        # Periodic continuation with current data
        if should_continue():
            close continue_as_new(entity, data)
```

> Note: Data structs are defined at the SDK level, not in TWF notation.

```pseudo
# Data must be serializable!
struct EntityData:
    balance: decimal
    lastUpdated: timestamp
    pendingOperations: []Operation
```

---

## Entity Workflow Pattern

Long-lived workflow representing a business entity (user, order, account, subscription).

### Structure

```twf
workflow UserEntity(userId: string, user: User):
    # Initialize user if new
    if user == null:
        activity LoadUser(userId) -> (user)

    for:
        # Wait for commands or periodic triggers
        await one:
            signal UpdateProfile:
                user.profile = signal.data

            signal AddCredits:
                user.credits += signal.amount

            signal Deactivate:
                user.active = false
                close complete  # End entity lifecycle

            timer(24h):
                # Periodic maintenance

        # Persist after any change
        activity PersistUser(user)

        # Continue-as-new periodically
        if eventCount > 500:
            close continue_as_new(userId, user)

query GetUser() -> (User):
    return user

update UpdateSettings(settings: Settings) -> (Result):
    user.settings = settings
    return Result{success: true}
```

### Entity Lifecycle

> Note: Entity lifecycle management uses SDK-level API calls, not TWF notation.

```pseudo
# Create entity (start workflow)
temporal.start_workflow(
    workflow: UserEntity,
    id: "user-{userId}",
    input: {userId: userId, user: null}
)

# Interact with entity (signals, queries, updates)
temporal.signal("user-{userId}", UpdateProfile, {name: "Alice"})
user = temporal.query("user-{userId}", GetUser)
result = temporal.update("user-{userId}", AddCredits, {amount: 100})

# Entity continues until explicit termination
temporal.signal("user-{userId}", Deactivate, {})
```

### Entity vs Process Workflows

| Entity Workflow | Process Workflow |
|-----------------|------------------|
| Long-lived (days, months, years) | Short-lived (minutes, hours) |
| Represents a thing | Represents a process |
| Reacts to external events | Drives toward completion |
| No natural end state | Has completion state |
| Examples: User, Account, Subscription | Examples: Order, Deployment, Migration |

---

## History Management Strategies

### Fixed Event Count

```twf
workflow Processor(data: ProcessorData):
    MAX_EVENTS = 1000
    eventCount = 0
    
    for:
        doWork()
        eventCount += 1
        
        if eventCount >= MAX_EVENTS:
            close continue_as_new(data)
```

### Time-Based

```twf
workflow DailyProcessor(data: ProcessorData, startTime: timestamp):
    for:
        doWork()
        
        # Continue every 24 hours
        if now() - startTime > 24h:
            close continue_as_new(data, now())
```

### History Size Estimation

```twf
workflow AdaptiveProcessor(data: ProcessorData):
    heavyEventCount = 0
    lightEventCount = 0
    
    for:
        event = receiveEvent()
        
        if event.type == "heavy":
            heavyEventCount += 1
        else:
            lightEventCount += 1
        
        # Weight heavy events more
        estimatedSize = heavyEventCount * 10 + lightEventCount
        if estimatedSize > 5000:
            close continue_as_new(data)
```

---

## Signal Handling Across Continue-As-New

### Default Behavior

Signals sent during continue-as-new transition are preserved:

```text
Execution 1: receives signal A, B, C
continue_as_new()
Execution 2: starts with signals A, B, C in buffer (if pending)
```

### Explicit Signal Draining

> Note: Signal draining logic is SDK-specific. Conceptual pseudo-code below.

```pseudo
workflow Processor(data: ProcessorData):
    for:
        # Process all pending signals before continue
        while has_pending_signals():
            signal = receive_signal()
            data = process(signal, data)
        
        if should_continue():
            continue_as_new(data)
```

---

## Querying Long-Running Workflows

### Query Across Continuations

> Note: Query API calls are SDK-level, not TWF notation.

```pseudo
# Same workflow ID, query works across continue-as-new
temporal.query("entity-123", GetState)

# Each continuation is a separate run
# Query always goes to latest run
```

### Search Attributes for Discovery

> Note: `upsert_search_attributes` is an SDK-level call, not TWF notation.

```pseudo
workflow EntityWorkflow(entityId: string, entity: Entity):
    # Set search attributes for discovery (SDK call)
    upsert_search_attributes({
        EntityId: entityId,
        EntityType: entity.type,
        Status: entity.status,
        LastUpdated: now()
    })
    
    for:
        # ... workflow logic ...
        
        # Update search attributes on change (SDK call)
        upsert_search_attributes({
            Status: entity.status,
            LastUpdated: now()
        })
```

---

## Anti-Patterns

### Never Continuing

```twf
# BAD: Unbounded history growth
workflow InfiniteLoop(data: LoopData):
    for:
        await signal Event -> (event)
        process(event)
        # Never continues - history grows forever!

# GOOD: Periodic continuation
workflow InfiniteLoop(data: LoopData):
    count = 0
    for:
        await signal Event -> (event)
        process(event)
        count += 1
        if count > 1000:
            close continue_as_new(data)
```

### Losing Data on Continue

```twf
# BAD: Data not passed to continuation
workflow Processor(data: ProcessorData):
    modifiedData = transform(data)
    close continue_as_new()  # Lost modifiedData!

# GOOD: Pass current data
workflow Processor(data: ProcessorData):
    modifiedData = transform(data)
    close continue_as_new(modifiedData)
```

### Continue-As-New in Wrong Place

```twf
# BAD: Continue in middle of operation
workflow Processor(data: ProcessorData):
    activity Step1()
    if shouldContinue:
        close continue_as_new(data)  # Step2 never runs!
    activity Step2()

# GOOD: Continue at natural boundary
workflow Processor(data: ProcessorData):
    activity Step1()
    activity Step2()
    if shouldContinue:
        close continue_as_new(data)
```

### Too Frequent Continuation

```twf
# BAD: Continue every event
workflow Processor(data: ProcessorData):
    event = await signal Event
    process(event)
    close continue_as_new(data)  # Unnecessary overhead!

# GOOD: Batch before continuing
workflow Processor(data: ProcessorData):
    count = 0
    for:
        event = await signal Event
        process(event)
        count += 1
        if count >= 1000:
            close continue_as_new(data)
```

---

## Monitoring Long-Running Workflows

### Key Metrics

| Metric | Why It Matters |
|--------|----------------|
| History event count | Approaching limits? |
| Continuation frequency | Too often? Too rare? |
| State size | Growing unbounded? |
| Signal backlog | Processing fast enough? |

### Health Checks

```pseudo
query GetHealth() -> (HealthStatus):
    return HealthStatus{
        eventCount: workflow.history_length(),  # SDK-specific API
        uptime: now() - startTime,
        dataSize: sizeof(data),
        pendingSignals: pending_signal_count()  # SDK-specific API
    }
```
