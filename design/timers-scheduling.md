# Timers and Scheduling

> **Example:** [`examples/timers-scheduling.twf`](./examples/timers-scheduling.twf)

Durable timing primitives for delays, deadlines, and recurring execution.

## Timers

Durable sleep that survives worker restarts, deployments, and failures.

### Basic Timer

```
workflow DelayedNotification(userId: string, delay: duration) -> void:
    # Durable sleep - workflow pauses but state is preserved
    timer delay
    
    activity SendNotification(userId)
```

### Timer Considerations

| Aspect | Guidance |
|--------|----------|
| **Durability** | Timer survives worker restarts; workflow resumes when timer fires |
| **Precision** | Not precise to the millisecond; expect seconds of variance |
| **History** | Each timer adds to workflow history; avoid very frequent short timers |
| **Cancellation** | Timers can be cancelled if workflow is cancelled |

---

## Deadlines and Timeouts

### Workflow-Level Deadline

```
workflow OrderFulfillment(order: Order) -> OrderResult:
    # Entire workflow must complete within deadline
    workflow_timeout: 7d
    
    activity ValidateOrder(order)
    await signal PaymentReceived
    activity ShipOrder(order)
    return OrderResult{status: "completed"}
```

### Operation Deadline Pattern

```
workflow ProcessWithDeadline(data: Data) -> Result:
    # Race between operation and deadline
    select:
        result = activity LongOperation(data):
            return Result{success: true, data: result}
        timer 1h:
            activity Cleanup(data)
            return Result{success: false, error: "deadline exceeded"}
```

### Timeout on Signal Wait

```
workflow ApprovalWorkflow(request: Request) -> Decision:
    activity NotifyApprovers(request)
    
    await signal Approved or signal Rejected:
        timeout: 7d
        on_timeout:
            activity NotifyExpired(request)
            return Decision{status: "expired"}
    
    if received Approved:
        return Decision{status: "approved"}
    else:
        return Decision{status: "rejected"}
```

---

## Scheduling Patterns

### Periodic Execution Within Workflow

```
workflow Heartbeat(resourceId: string) -> void:
    loop:
        activity CheckHealth(resourceId)
        timer 5m
```

### Polling with Backoff

```
workflow WaitForResource(resourceId: string) -> Resource:
    backoff = 1s
    max_backoff = 5m
    
    loop:
        resource = activity CheckResource(resourceId)
        if resource.ready:
            return resource
        
        timer backoff
        backoff = min(backoff * 2, max_backoff)
```

### Deadline with Periodic Check

```
workflow WaitForCompletion(jobId: string) -> JobResult:
    deadline = now() + 2h
    
    loop:
        status = activity GetJobStatus(jobId)
        if status.complete:
            return JobResult{status: "complete", data: status.data}
        
        select:
            timer 30s:
                continue  # Check again
            timer until deadline:
                return JobResult{status: "timeout"}
```

---

## Schedules (Cron Workflows)

Temporal Schedules execute workflows on a recurring basis, like cron.

### Schedule Concepts

| Concept | Description |
|---------|-------------|
| **Schedule** | Configuration defining when/how to run workflow |
| **Action** | What to do (start workflow with specific input) |
| **Spec** | When to run (cron expression, intervals, calendars) |
| **Policy** | How to handle overlaps, catch-ups, pauses |

### Schedule Specification

```
schedule DailyReport:
    action: workflow GenerateReport(type: "daily")
    spec:
        cron: "0 9 * * *"  # 9 AM daily
    policy:
        overlap: SKIP           # Skip if previous still running
        catchup_window: 1h      # Catch up missed runs within 1 hour
```

### Schedule Patterns

**Simple Interval:**
```
schedule HealthCheck:
    action: workflow CheckSystemHealth()
    spec:
        interval: 5m
```

**Cron Expression:**
```
schedule WeeklyBackup:
    action: workflow BackupDatabase(type: "full")
    spec:
        cron: "0 2 * * 0"  # 2 AM every Sunday
```

**Business Hours:**
```
schedule BusinessHoursCheck:
    action: workflow CheckQueues()
    spec:
        calendars:
            - hour: 9-17
              day_of_week: MON-FRI
        interval: 15m
```

### Overlap Policies

| Policy | Behavior |
|--------|----------|
| `SKIP` | Don't start if previous execution still running |
| `BUFFER_ONE` | Queue one execution if previous running |
| `BUFFER_ALL` | Queue all scheduled executions |
| `CANCEL_OTHER` | Cancel running execution, start new |
| `TERMINATE_OTHER` | Terminate running execution, start new |
| `ALLOW_ALL` | Start regardless of running executions |

### Catchup Policy

When schedule is paused or system is down:

```
schedule DailyReport:
    action: workflow GenerateReport(type: "daily")
    spec:
        cron: "0 9 * * *"
    policy:
        catchup_window: 24h     # Run missed executions within 24h
        # catchup_window: 0     # Never catch up (skip missed)
```

---

## Timer Anti-Patterns

### Very Short Timers in Loops

```
# BAD: Adds massive workflow history
workflow PollForever(resourceId: string):
    loop:
        activity Check(resourceId)
        timer 100ms  # 10 timers/second = huge history

# GOOD: Reasonable interval or use continue-as-new
workflow PollForever(resourceId: string):
    count = 0
    loop:
        activity Check(resourceId)
        timer 5s
        count += 1
        if count > 1000:
            continue_as_new(resourceId)  # Reset history
```

### Non-Deterministic Time Checks

```
# BAD: Non-deterministic
workflow Process(data: Data):
    if current_time() > some_deadline:  # Different on replay!
        cancel()

# GOOD: Use timer
workflow Process(data: Data):
    select:
        activity DoWork(data): return success
        timer some_deadline: return timeout
```

### Timer for Immediate Execution

```
# BAD: Unnecessary timer
workflow Process(data: Data):
    timer 0s  # Why?
    activity DoWork(data)

# GOOD: Just execute
workflow Process(data: Data):
    activity DoWork(data)
```

---

## Design Considerations

### Timer Duration Selection

| Duration | Use Case | Considerations |
|----------|----------|----------------|
| Seconds | Retry backoff, quick polls | Fine, but avoid sub-second |
| Minutes | Health checks, progress updates | Common and reasonable |
| Hours | Batch processing, cleanup | Consider continue-as-new for very long |
| Days | Approval workflows, SLA tracking | Watch workflow history growth |

### Timezone Handling

Schedules can be timezone-aware:

```
schedule DailyReport:
    action: workflow GenerateReport()
    spec:
        cron: "0 9 * * *"
        timezone: "America/New_York"  # 9 AM Eastern
```

### Jitter for Load Distribution

```
schedule DistributedHealthCheck:
    action: workflow CheckHealth()
    spec:
        interval: 5m
        jitter: 30s  # Random offset 0-30s to avoid thundering herd
```
