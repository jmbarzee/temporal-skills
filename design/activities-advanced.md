# Advanced Activity Patterns

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

```
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

```
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

```
workflow Parent(data: Data) -> Result:
    result = activity LongProcess(data):
        start_to_close_timeout: 1h    # Max total execution time
        heartbeat_timeout: 30s         # Must heartbeat within 30s
    
    return result
```

### Resume from Heartbeat Details

```
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

```
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

```
workflow ApprovalWorkflow(request: Request) -> Decision:
    activity NotifyRequestCreated(request)
    
    # This activity blocks until external completion
    result = activity RequestHumanApproval(request):
        start_to_close_timeout: 7d  # Long timeout for human task
        heartbeat_timeout: 0         # No heartbeat needed
    
    if result.approved:
        activity ExecuteApprovedAction(request)
    
    return Decision{approved: result.approved}
```

### External Completion API

```
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

```
workflow ProcessOrder(order: Order) -> Result:
    # Local activity: fast, in-process
    validated = local_activity ValidateInput(order):
        start_to_close_timeout: 5s
    
    # Regular activity: goes through task queue
    result = activity ProcessPayment(order)
    
    return result
```

### Local Activity Limitations

| Limitation | Implication |
|------------|-------------|
| No task queue routing | Must execute on workflow worker |
| Limited retries | Short retry window |
| Worker restart = retry | Not persisted across restarts |
| No heartbeat | Not for long operations |

### Local Activity Configuration

```
workflow Parent(data: Data) -> Result:
    result = local_activity QuickValidation(data):
        start_to_close_timeout: 10s
        local_retry_threshold: 5s      # Retry locally for 5s
        retry_policy:
            max_attempts: 3
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

```
|-------- schedule_to_close --------|
|-- schedule_to_start --|-- start_to_close --|

schedule_to_close >= schedule_to_start + start_to_close
```

### Configuration Examples

```
workflow Parent(data: Data) -> Result:
    # Short operation, tight timeout
    result = activity QuickLookup(data.id):
        start_to_close_timeout: 30s
    
    # Long operation with heartbeat
    result = activity ProcessBatch(data):
        start_to_close_timeout: 2h
        heartbeat_timeout: 60s
    
    # Operation with queue wait tolerance
    result = activity LowPriorityTask(data):
        schedule_to_start_timeout: 5m    # Wait up to 5m in queue
        start_to_close_timeout: 10m
    
    return result
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

```
workflow Parent(data: Data) -> Result:
    result = activity UnreliableService(data):
        retry_policy:
            initial_interval: 1s          # First retry after 1s
            backoff_coefficient: 2.0       # Double each retry
            max_interval: 60s              # Cap at 60s between retries
            max_attempts: 5                # Give up after 5 attempts
            non_retryable_errors: [        # Don't retry these
                "InvalidInput",
                "NotFound"
            ]
```

### Error Classification

```
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

```
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

```
# BAD: 1-minute timeout for human task
activity GetHumanApproval():
    start_to_close_timeout: 1m  # Humans are slower than this!

# GOOD: Appropriate timeout for human task
activity GetHumanApproval():
    start_to_close_timeout: 7d
    # Or use async completion
```

### Local Activity for External Calls

```
# BAD: Network call in local activity
local_activity CallExternalAPI(data):
    return http.post(external_url, data)  # Network latency + failures

# GOOD: Regular activity for external calls
activity CallExternalAPI(data):
    return http.post(external_url, data)
```
