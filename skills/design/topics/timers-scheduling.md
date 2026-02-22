# Timers and Scheduling

> **Example:** [`timers-scheduling.twf`](./timers-scheduling.twf)

Durable timing primitives for delays, deadlines, and recurring execution.

## Timers

Durable sleep that survives worker restarts, deployments, and failures.

### Basic Timer

```twf
workflow DelayedNotification(userId: string, delay: duration):
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
workflow OrderFulfillment(order: Order) -> (OrderResult):
    # Entire workflow must complete within deadline (SDK-level config)
    # workflow_timeout: 7d

    activity ValidateOrder(order)
    await signal PaymentReceived
    activity ShipOrder(order)
    close complete(OrderResult{status: "completed"})
```

### Operation Deadline Pattern

```twf
workflow ProcessWithDeadline(data: Data) -> (Result):
    # Race between operation and deadline
    await one:
        activity LongOperation(data) -> result:
            close complete(Result{success: true, data: result})
        timer(1h):
            activity Cleanup(data)
            close fail(Result{success: false, error: "deadline exceeded"})
```

### Timeout on Signal Wait

```twf
workflow ApprovalWorkflow(request: Request) -> (Decision):
    activity NotifyApprovers(request)

    await one:
        signal Approved:
            close complete(Decision{status: "approved"})
        signal Rejected:
            close complete(Decision{status: "rejected"})
        timer(7d):
            activity NotifyExpired(request)
            close complete(Decision{status: "expired"})
```

---

## Scheduling Patterns

### Periodic Execution Within Workflow

```twf
workflow Heartbeat(resourceId: string):
    for:
        activity CheckHealth(resourceId)
        await timer(5m)
```

### Polling with Backoff

```twf
workflow WaitForResource(resourceId: string) -> (Resource):
    backoff = 1s
    max_backoff = 5m

    for:
        activity CheckResource(resourceId) -> resource
        if resource.ready:
            close complete(resource)

        await timer(backoff)
        backoff = min(backoff * 2, max_backoff)
```

### Deadline with Periodic Check

```twf
workflow WaitForCompletion(jobId: string) -> (JobResult):
    for:
        activity GetJobStatus(jobId) -> status
        if status.complete:
            close complete(JobResult{status: "complete", data: status.data})

        await one:
            timer(30s):
                # Continue polling
            timer(2h):
                close fail(JobResult{status: "timeout"})
```

---

## Schedules (Cron Workflows)

Temporal Schedules execute workflows on a recurring basis (cron expressions, intervals, calendars). Schedules are **platform configuration**, not workflow design — they define *when* to start a workflow, not *how* it runs.

Schedule configuration (specs, overlap policies, catchup windows, timezones) is managed through the Temporal CLI or SDK, not TWF notation. See [Temporal Schedules documentation](https://docs.temporal.io/workflows#schedule) for details.

**Design implication:** A scheduled workflow should be designed like any other workflow — idempotent, with continue-as-new if long-running. The schedule itself is an external trigger, not part of the workflow's logic.
