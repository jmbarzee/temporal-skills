# Long-Running Workflows

> **Example:** [`examples/long-running.twf`](./examples/long-running.twf)

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

```
workflow LongRunningProcessor(state: State) -> void:
    eventCount = 0
    
    loop:
        event = await signal NewEvent
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

```
workflow EntityWorkflow(entity: Entity, state: EntityState) -> void:
    loop:
        await signal Command:
            state = applyCommand(state, signal.command)
        
        # Periodic continuation with current state
        if should_continue():
            continue_as_new(entity, state)

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

```
workflow UserEntity(userId: string, state: UserState) -> void:
    # Initialize state if new
    if state == null:
        state = activity LoadUser(userId)
    
    loop:
        # Wait for commands or periodic triggers
        select:
            signal UpdateProfile:
                state.profile = signal.data
                activity PersistUser(userId, state)
            
            signal AddCredits:
                state.credits += signal.amount
                activity PersistUser(userId, state)
            
            signal Deactivate:
                state.active = false
                activity PersistUser(userId, state)
                return  # End entity lifecycle
            
            timer 24h:
                activity DailyMaintenance(userId, state)
        
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

```
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

```
workflow Processor(state: State) -> void:
    MAX_EVENTS = 1000
    eventCount = 0
    
    loop:
        doWork()
        eventCount += 1
        
        if eventCount >= MAX_EVENTS:
            continue_as_new(state)
```

### Time-Based

```
workflow DailyProcessor(state: State, startTime: timestamp) -> void:
    loop:
        doWork()
        
        # Continue every 24 hours
        if now() - startTime > 24h:
            continue_as_new(state, now())
```

### History Size Estimation

```
workflow AdaptiveProcessor(state: State) -> void:
    heavyEventCount = 0
    lightEventCount = 0
    
    loop:
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

```
Execution 1: receives signal A, B, C
continue_as_new()
Execution 2: starts with signals A, B, C in buffer (if pending)
```

### Explicit Signal Draining

```
workflow Processor(state: State) -> void:
    loop:
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

```
# Same workflow ID, query works across continue-as-new
temporal.query("entity-123", GetState)

# Each continuation is a separate run
# Query always goes to latest run
```

### Search Attributes for Discovery

```
workflow EntityWorkflow(entityId: string, state: State) -> void:
    # Set search attributes for discovery
    upsert_search_attributes({
        EntityId: entityId,
        EntityType: state.type,
        Status: state.status,
        LastUpdated: now()
    })
    
    loop:
        # ... workflow logic ...
        
        # Update search attributes on state change
        upsert_search_attributes({
            Status: state.status,
            LastUpdated: now()
        })
```

---

## Anti-Patterns

### Never Continuing

```
# BAD: Unbounded history growth
workflow InfiniteLoop(state: State) -> void:
    loop:
        event = await signal Event
        process(event)
        # Never continues - history grows forever!

# GOOD: Periodic continuation
workflow InfiniteLoop(state: State) -> void:
    count = 0
    loop:
        event = await signal Event
        process(event)
        count += 1
        if count > 1000:
            continue_as_new(state)
```

### Losing State on Continue

```
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

```
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

```
# BAD: Continue every event
workflow Processor(state: State) -> void:
    event = await signal Event
    process(event)
    continue_as_new(state)  # Unnecessary overhead!

# GOOD: Batch before continuing
workflow Processor(state: State) -> void:
    count = 0
    loop:
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

```
query GetHealth() -> HealthStatus:
    return HealthStatus{
        eventCount: workflow.history_length(),
        uptime: now() - startTime,
        stateSize: sizeof(state),
        pendingSignals: pending_signal_count()
    }
```
