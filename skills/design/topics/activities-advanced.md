# Advanced Activity Patterns

> **Example:** [`activities-advanced.twf`](./activities-advanced.twf)

Heartbeats, async completion, local activities, and timeout configuration for sophisticated activity designs.

## Activity Heartbeats

Long-running activities should periodically heartbeat to report progress and allow early failure detection.

### Why Heartbeat

| Without Heartbeat | With Heartbeat |
|-------------------|----------------|
| Worker crash detected only at activity timeout | Worker crash detected within heartbeat timeout |
| No visibility into progress | Progress visible during execution |
| Full retry on any failure | Can resume from last heartbeat |

### Basic Heartbeat Pattern

> Note: Activity body implementations are SDK-level code, not TWF notation.

```pseudo
activity ProcessLargeFile(fileId: string) -> ProcessResult:
    file = download(fileId)
    total = len(file.chunks)
    
    for i, chunk in enumerate(file.chunks):
        process(chunk)
        
        # Report progress every chunk
        heartbeat(progress: {
            current: i + 1,
            total: total,
            percentage: (i + 1) / total * 100
        })
    
    return ProcessResult{success: true}
```

### Heartbeat with Cancellation Check

```pseudo
activity LongRunningTask(taskId: string) -> TaskResult:
    for step in steps:
        # Check if workflow requested cancellation
        if heartbeat_and_check_cancelled():
            cleanup(taskId)
            raise CancelledException("Task cancelled")
        
        execute(step)
    
    return TaskResult{complete: true}
```

### Heartbeat Timeout Configuration

```twf
workflow Parent(data: Data) -> (Result):
    activity LongProcess(data) -> result
        options:
            start_to_close_timeout: 1h
            heartbeat_timeout: 30s

    close complete(result)
```

### Resume from Heartbeat Details

```pseudo
activity ResumableProcess(batchId: string) -> BatchResult:
    # Get last heartbeat details (if resuming after failure)
    lastProgress = get_heartbeat_details()
    startIndex = lastProgress?.index ?? 0
    
    items = getItems(batchId)
    
    for i in range(startIndex, len(items)):
        process(items[i])
        heartbeat(index: i + 1)  # Save progress
    
    return BatchResult{processed: len(items)}
```

---

## Async Activity Completion

Complete an activity from outside the activity execution context. Useful for human tasks, external callbacks, and webhook-driven flows.

### When to Use

| Use Case | Description |
|----------|-------------|
| Human tasks | Activity waits for human action in external system |
| External callbacks | Third-party API will callback when done |
| Long polling avoidance | External system notifies completion |
| Multi-system coordination | Completion triggered by external event |

### Async Completion Pattern

```pseudo
activity RequestHumanApproval(request: ApprovalRequest) -> ApprovalResult:
    # Get task token for external completion
    taskToken = get_activity_task_token()
    
    # Send task to external system (e.g., ticketing system)
    createTicket(
        title: request.title,
        callback_token: taskToken,  # External system uses this to complete
        callback_url: "https://temporal/complete-activity"
    )
    
    # Activity does NOT complete here
    # It will be completed externally via API call
    do_not_complete()

# External system calls Temporal API:
# POST /complete-activity
# {
#   "task_token": "...",
#   "result": {"approved": true, "approver": "alice"}
# }
```

### Workflow Using Async Activity

```twf
workflow ApprovalWorkflow(request: Request) -> (Decision):
    activity NotifyRequestCreated(request)

    # This activity blocks until external completion
    activity RequestHumanApproval(request) -> result
        options:
            start_to_close_timeout: 7d

    if (result.approved):
        activity ExecuteApprovedAction(request)

    close complete(Decision{approved: result.approved})
```

### External Completion API

> Note: Temporal API calls are SDK-level, not TWF notation.

```pseudo
# Complete activity successfully
temporal.complete_activity(
    task_token: "...",
    result: {approved: true}
)

# Fail activity
temporal.fail_activity(
    task_token: "...",
    error: {message: "Rejected by compliance"}
)

# Report activity cancelled
temporal.cancel_activity(
    task_token: "..."
)
```

---

## Local Activities

Lightweight activities that execute in the workflow worker process without task queue round-trip.

### When to Use Local Activities

| Use Local Activity | Use Regular Activity |
|--------------------|---------------------|
| Very short operations (< 10s) | Longer operations |
| Low latency required | Normal latency acceptable |
| Simple operations | Complex operations |
| Tight retry needed | Standard retry policies |
| Same worker has required resources | May need different worker |

### Local Activity Pattern

> Note: Local activities are an SDK-level concept. In TWF notation, use `activity` with an `options:` block specifying local execution. The syntax below is conceptual.

```twf
workflow ProcessOrder(order: Order) -> (Result):
    # Local activity: fast, in-process (SDK: use local activity API)
    activity ValidateInput(order) -> validated
        options:
            local: true
            start_to_close_timeout: 5s

    # Regular activity: goes through task queue
    activity ProcessPayment(order) -> result

    close complete(result)
```

### Local Activity Limitations

| Limitation | Implication |
|------------|-------------|
| No task queue routing | Must execute on workflow worker |
| Limited retries | Short retry window |
| Worker restart = retry | Not persisted across restarts |
| No heartbeat | Not for long operations |

### Local Activity Configuration

> Note: Local activity configuration is SDK-specific.

```twf
workflow Parent(data: Data) -> (Result):
    activity QuickValidation(data) -> result
        options:
            local: true
            start_to_close_timeout: 10s
            local_retry_threshold: 5s
            retry_policy:
                maximum_attempts: 3
                initial_interval: 100ms
```

---

## Activity Timeout Configuration

### Timeout Types

| Timeout | Description | Default |
|---------|-------------|---------|
| `schedule_to_start` | Time from scheduled to worker pickup | None |
| `start_to_close` | Time from worker pickup to completion | Required |
| `schedule_to_close` | Total time from scheduled to completion | None |
| `heartbeat` | Max time between heartbeats | None |

### Timeout Relationships

```text
|-------- schedule_to_close --------|
|-- schedule_to_start --|-- start_to_close --|

schedule_to_close >= schedule_to_start + start_to_close
```

### Configuration Examples

```twf
workflow Parent(data: Data) -> (Result):
    # Short operation, tight timeout
    activity QuickLookup(data.id) -> result
        options:
            start_to_close_timeout: 30s

    # Long operation with heartbeat
    activity ProcessBatch(data) -> result
        options:
            start_to_close_timeout: 2h
            heartbeat_timeout: 60s

    # Operation with queue wait tolerance
    activity LowPriorityTask(data) -> result
        options:
            schedule_to_start_timeout: 5m
            start_to_close_timeout: 10m

    close complete(result)
```

### Timeout Selection Guidelines

| Operation Type | schedule_to_start | start_to_close | heartbeat |
|----------------|-------------------|----------------|-----------|
| Quick lookup | None | 10-30s | None |
| API call | None | 30s-2m | None |
| Batch processing | None | Minutes-hours | 30-60s |
| Human task | Minutes-hours | Days | None |
| External callback | None | Hours-days | None |

---

## Retry Policies

### Retry Configuration

```twf
workflow Parent(data: Data) -> (Result):
    activity UnreliableService(data) -> result
        options:
            retry_policy:
                initial_interval: 1s
                backoff_coefficient: 2.0
                maximum_interval: 60s
                maximum_attempts: 5
                non_retryable_errors: ["InvalidInput", "NotFound"]
```

### Error Classification

```pseudo
activity CallExternalAPI(request: Request) -> Response:
    try:
        return api.call(request)
    catch RateLimitError:
        # Retryable - throw and let Temporal retry
        raise
    catch InvalidInputError:
        # Non-retryable - application error
        raise NonRetryableError("Invalid input: " + error.message)
    catch NotFoundError:
        # Business logic - handle gracefully
        return Response{found: false}
```

---

## Anti-Patterns

### Missing Heartbeat on Long Operations

```pseudo
# BAD: 2-hour activity with no heartbeat
activity ProcessHugeFile(fileId: string):
    for chunk in file.chunks:  # Takes 2 hours
        process(chunk)
    # Worker crash at 1h59m = full retry from start

# GOOD: Heartbeat with resumable progress
activity ProcessHugeFile(fileId: string):
    lastIndex = get_heartbeat_details()?.index ?? 0
    for i in range(lastIndex, len(chunks)):
        process(chunks[i])
        heartbeat(index: i + 1)
```

### Wrong Timeout for Operation Type

```pseudo
# BAD: 1-minute timeout for human task
activity GetHumanApproval():
    start_to_close_timeout: 1m  # Humans are slower than this!

# GOOD: Appropriate timeout for human task
activity GetHumanApproval():
    start_to_close_timeout: 7d
    # Or use async completion
```

### Local Activity for External Calls

```pseudo
# BAD: Network call in local activity (SDK-level concept)
# Local activities should not make network calls

# GOOD: Regular activity for external calls
activity CallExternalAPI(data):
    return http.post(external_url, data)
```
