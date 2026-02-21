# options

## Activity options

### DSL

```twf
activity QuickLookup(data.id) -> result
    options:
        start_to_close_timeout: 30s
```

### Go

```go
actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Second,
})
var result LookupResult
err := workflow.ExecuteActivity(actCtx, QuickLookup, data.Id).Get(ctx, &result)
```

## Activity options with retry policy

### DSL

```twf
activity UnreliableService(data) -> result
    options:
        start_to_close_timeout: 2m
        retry_policy:
            maximum_attempts: 5
            initial_interval: 1s
            backoff_coefficient: 2.0
            maximum_interval: 60s
```

### Go

```go
actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    StartToCloseTimeout: 2 * time.Minute,
    RetryPolicy: &temporal.RetryPolicy{
        MaximumAttempts:        5,
        InitialInterval:       1 * time.Second,
        BackoffCoefficient:    2.0,
        MaximumInterval:       60 * time.Second,
    },
})
var result ServiceResult
err := workflow.ExecuteActivity(actCtx, UnreliableService, data).Get(ctx, &result)
```

## Child workflow options

### DSL

```twf
workflow ChildWorkflow(input.data) -> childResult
    options:
        workflow_execution_timeout: 1h
        retry_policy:
            maximum_attempts: 3
```

### Go

```go
childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
    WorkflowExecutionTimeout: 1 * time.Hour,
    RetryPolicy: &temporal.RetryPolicy{
        MaximumAttempts: 3,
    },
})
var childResult ChildResult
err := workflow.ExecuteChildWorkflow(childCtx, ChildWorkflow, input.Data).Get(ctx, &childResult)
```

## Nexus operation options

### DSL

```twf
nexus PaymentsEndpoint PaymentsService.ProcessPayment(payment) -> paymentResult
    options:
        schedule_to_close_timeout: 1h
```

### Go

```go
c := workflow.NewNexusClient("PaymentsEndpoint", "PaymentsService")
var paymentResult PaymentResult
fut := c.ExecuteOperation(ctx, "ProcessPayment", payment, workflow.NexusOperationOptions{
    ScheduleToCloseTimeout: 1 * time.Hour,
})
err := fut.Get(ctx, &paymentResult)
```

## Notes

- When no `options:` block is specified, set a default `ActivityOptions` with `StartToCloseTimeout` on `ctx` near the top of the workflow function
- Option keys map: `start_to_close_timeout` → `StartToCloseTimeout`, `schedule_to_close_timeout` → `ScheduleToCloseTimeout`, `heartbeat_timeout` → `HeartbeatTimeout`
- `retry_policy:` → `&temporal.RetryPolicy{...}` (pointer)
- `NexusOperationOptions` fields: `ScheduleToCloseTimeout` (primary) and `CancellationType` (experimental). Options are passed inline — no context wrapping like activities
