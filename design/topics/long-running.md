# Long-Running Workflows

> **Example:** [`long-running.twf`](./long-running.twf)

Patterns for workflows that run for extended periods: continue-as-new, history management, and entity workflows.

## The History Problem

Temporal stores every event in workflow history. Long-running workflows accumulate history that:

| Issue | Impact |
|-------|--------|
| Memory usage | Large history loaded on each replay |
| Replay time | Longer replay = slower recovery |
| Storage costs | More events = more storage |
| Hard limit | Temporal enforces max history size |

**Solution:** Reset history periodically with `continue_as_new`.

---

## Continue-As-New

Atomically complete current workflow and start a new execution with fresh history, preserving logical continuity.

### Basic Pattern

```twf
workflow LongRunningProcessor(state: State) -> void:
    eventCount = 0
    
    for:
        await signal NewEvent -> event
        activity ProcessEvent(event)
        state.processed += 1
        eventCount += 1
        
        # Reset history before it gets too large
        if eventCount >= 1000:
            continue_as_new(state)  # Fresh history, same logical workflow
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

### State Serialization

```twf
workflow EntityWorkflow(entity: Entity, state: EntityState) -> void:
    for:
        await signal Command -> command
        state = applyCommand(state, command)
        
        # Periodic continuation with current state
        if should_continue():
            continue_as_new(entity, state)
```

> Note: State structs are defined at the SDK level, not in TWF notation.

```pseudo
# State must be serializable!
struct EntityState:
    balance: decimal
    lastUpdated: timestamp
    pendingOperations: []Operation
```

---

## Entity Workflow Pattern

Long-lived workflow representing a business entity (user, order, account, subscription).

### Structure

```twf
workflow UserEntity(userId: string, state: UserState) -> void:
    # Initialize state if new
    if state == null:
        activity LoadUser(userId) -> state

    for:
        # Wait for commands or periodic triggers
        await one:
            signal UpdateProfile:
                state.profile = signal.data

            signal AddCredits:
                state.credits += signal.amount

            signal Deactivate:
                state.active = false
                close  # End entity lifecycle

            timer(24h):
                # Periodic maintenance

        # Persist after any state change
        activity PersistUser(userId, state)

        # Continue-as-new periodically
        if eventCount > 500:
            continue_as_new(userId, state)

query GetState() -> UserState:
    return state

update UpdateSettings(settings: Settings) -> Result:
    state.settings = settings
    return Result{success: true}
```

### Entity Lifecycle

> Note: Entity lifecycle management uses SDK-level API calls, not TWF notation.

```pseudo
# Create entity (start workflow)
temporal.start_workflow(
    workflow: UserEntity,
    id: "user-{userId}",
    input: {userId: userId, state: null}
)

# Interact with entity (signals, queries, updates)
temporal.signal("user-{userId}", UpdateProfile, {name: "Alice"})
state = temporal.query("user-{userId}", GetState)
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
workflow Processor(state: State) -> void:
    MAX_EVENTS = 1000
    eventCount = 0
    
    for:
        doWork()
        eventCount += 1
        
        if eventCount >= MAX_EVENTS:
            continue_as_new(state)
```

### Time-Based

```twf
workflow DailyProcessor(state: State, startTime: timestamp) -> void:
    for:
        doWork()
        
        # Continue every 24 hours
        if now() - startTime > 24h:
            continue_as_new(state, now())
```

### History Size Estimation

```twf
workflow AdaptiveProcessor(state: State) -> void:
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
            continue_as_new(state)
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
workflow Processor(state: State) -> void:
    for:
        # Process all pending signals before continue
        while has_pending_signals():
            signal = receive_signal()
            state = process(signal, state)
        
        if should_continue():
            continue_as_new(state)
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
workflow EntityWorkflow(entityId: string, state: State) -> void:
    # Set search attributes for discovery (SDK call)
    upsert_search_attributes({
        EntityId: entityId,
        EntityType: state.type,
        Status: state.status,
        LastUpdated: now()
    })
    
    for:
        # ... workflow logic ...
        
        # Update search attributes on state change (SDK call)
        upsert_search_attributes({
            Status: state.status,
            LastUpdated: now()
        })
```

---

## Anti-Patterns

### Never Continuing

```twf
# BAD: Unbounded history growth
workflow InfiniteLoop(state: State) -> void:
    for:
        await signal Event -> event
        process(event)
        # Never continues - history grows forever!

# GOOD: Periodic continuation
workflow InfiniteLoop(state: State) -> void:
    count = 0
    for:
        await signal Event -> event
        process(event)
        count += 1
        if count > 1000:
            continue_as_new(state)
```

### Losing State on Continue

```twf
# BAD: State not passed to continuation
workflow Processor(state: State) -> void:
    modifiedState = transform(state)
    continue_as_new()  # Lost modifiedState!

# GOOD: Pass current state
workflow Processor(state: State) -> void:
    modifiedState = transform(state)
    continue_as_new(modifiedState)
```

### Continue-As-New in Wrong Place

```twf
# BAD: Continue in middle of operation
workflow Processor(state: State) -> void:
    activity Step1()
    if shouldContinue:
        continue_as_new(state)  # Step2 never runs!
    activity Step2()

# GOOD: Continue at natural boundary
workflow Processor(state: State) -> void:
    activity Step1()
    activity Step2()
    if shouldContinue:
        continue_as_new(state)
```

### Too Frequent Continuation

```twf
# BAD: Continue every event
workflow Processor(state: State) -> void:
    event = await signal Event
    process(event)
    continue_as_new(state)  # Unnecessary overhead!

# GOOD: Batch before continuing
workflow Processor(state: State) -> void:
    count = 0
    for:
        event = await signal Event
        process(event)
        count += 1
        if count >= 1000:
            continue_as_new(state)
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
query GetHealth() -> HealthStatus:
    return HealthStatus{
        eventCount: workflow.history_length(),  # SDK-specific API
        uptime: now() - startTime,
        stateSize: sizeof(state),
        pendingSignals: pending_signal_count()  # SDK-specific API
    }
```
