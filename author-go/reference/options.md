# options

## Activity options

### DSL

```twf
activity QuickLookup(data.id) -> result
    options(startToCloseTimeout: 30s)
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
    options(startToCloseTimeout: 2m, retryPolicy: {maxAttempts: 5, initialInterval: 1s, backoffCoefficient: 2.0, maxInterval: 60s})
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
    options(workflowExecutionTimeout: 1h, retryPolicy: {maxAttempts: 3})
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

## Notes

- When no `options(...)` is specified, set a default `ActivityOptions` with `StartToCloseTimeout` on `ctx` near the top of the workflow function
- Option keys map: `startToCloseTimeout` → `StartToCloseTimeout`, `scheduleToCloseTimeout` → `ScheduleToCloseTimeout`, `heartbeatTimeout` → `HeartbeatTimeout`
- `retryPolicy` → `&temporal.RetryPolicy{...}` (pointer)
- Activity definition-level `options(...)` sets defaults; call-site options override
