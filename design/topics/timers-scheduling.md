# Timers and Scheduling

> **Example:** [`timers-scheduling.twf`](./timers-scheduling.twf)

Durable timing primitives for delays, deadlines, and recurring execution.

## Timers

Durable sleep that survives worker restarts, deployments, and failures.

### Basic Timer

```twf
workflow DelayedNotification(userId: string, delay: duration) -> void:
    # Durable sleep - workflow pauses but state is preserved
    await timer(delay)

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

```twf
workflow OrderFulfillment(order: Order) -> OrderResult:
    # Entire workflow must complete within deadline (SDK-level config)
    # workflow_timeout: 7d

    activity ValidateOrder(order)
    await signal PaymentReceived
    activity ShipOrder(order)
    close OrderResult{status: "completed"}
```

### Operation Deadline Pattern

```twf
workflow ProcessWithDeadline(data: Data) -> Result:
    # Race between operation and deadline
    await one:
        activity LongOperation(data) -> result:
            close Result{success: true, data: result}
        timer(1h):
            activity Cleanup(data)
            close failed Result{success: false, error: "deadline exceeded"}
```

### Timeout on Signal Wait

```twf
workflow ApprovalWorkflow(request: Request) -> Decision:
    activity NotifyApprovers(request)

    await one:
        signal Approved:
            close Decision{status: "approved"}
        signal Rejected:
            close Decision{status: "rejected"}
        timer(7d):
            activity NotifyExpired(request)
            close Decision{status: "expired"}
```

---

## Scheduling Patterns

### Periodic Execution Within Workflow

```twf
workflow Heartbeat(resourceId: string) -> void:
    for:
        activity CheckHealth(resourceId)
        await timer(5m)
```

### Polling with Backoff

```twf
workflow WaitForResource(resourceId: string) -> Resource:
    backoff = 1s
    max_backoff = 5m

    for:
        activity CheckResource(resourceId) -> resource
        if resource.ready:
            close resource

        await timer(backoff)
        backoff = min(backoff * 2, max_backoff)
```

### Deadline with Periodic Check

```twf
workflow WaitForCompletion(jobId: string) -> JobResult:
    for:
        activity GetJobStatus(jobId) -> status
        if status.complete:
            close JobResult{status: "complete", data: status.data}

        await one:
            timer(30s):
                # Continue polling
            timer(2h):
                close failed JobResult{status: "timeout"}
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

> Note: Schedule definitions are Temporal platform configuration, not TWF notation.

```yaml
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
```yaml
schedule HealthCheck:
    action: workflow CheckSystemHealth()
    spec:
        interval: 5m
```

**Cron Expression:**
```yaml
schedule WeeklyBackup:
    action: workflow BackupDatabase(type: "full")
    spec:
        cron: "0 2 * * 0"  # 2 AM every Sunday
```

**Business Hours:**
```yaml
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

```yaml
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

```twf
# BAD: Adds massive workflow history
workflow PollForever(resourceId: string):
    for:
        activity Check(resourceId)
        await timer(100ms)  # 10 timers/second = huge history

# GOOD: Reasonable interval or use continue-as-new
workflow PollForever(resourceId: string):
    count = 0
    for:
        activity Check(resourceId)
        await timer(5s)
        count += 1
        if count > 1000:
            continue_as_new(resourceId)  # Reset history
```

### Non-Deterministic Time Checks

```twf
# BAD: Non-deterministic
workflow Process(data: Data):
    if current_time() > some_deadline:  # Different on replay!
        cancel()

# GOOD: Use timer
workflow Process(data: Data):
    await one:
        activity DoWork(data) -> result:
            close Result{status: "success"}
        timer(some_deadline):
            close failed Result{status: "timeout"}
```

### Timer for Immediate Execution

```twf
# BAD: Unnecessary timer
workflow Process(data: Data):
    await timer(0s)  # Why?
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

```yaml
schedule DailyReport:
    action: workflow GenerateReport()
    spec:
        cron: "0 9 * * *"
        timezone: "America/New_York"  # 9 AM Eastern
```

### Jitter for Load Distribution

```yaml
schedule DistributedHealthCheck:
    action: workflow CheckHealth()
    spec:
        interval: 5m
        jitter: 30s  # Random offset 0-30s to avoid thundering herd
```
