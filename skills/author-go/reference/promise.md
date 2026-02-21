# promise

## DSL

```twf
promise handleA <- activity ProcessA(items.a)
# ... do other work ...
await handleA -> resultA
```

## Go

```go
futureA := workflow.ExecuteActivity(ctx, ProcessA, items.A)
// ... do other work ...
var resultA ResultA
err := futureA.Get(ctx, &resultA)
if err != nil {
    return Result{}, err
}
```

## Variants

```twf
promise childHandle <- workflow SlowChild(input.data)
```

```go
childFuture := workflow.ExecuteChildWorkflow(ctx, SlowChild, input.Data)
```

```twf
promise timeout <- timer(5m)
```

```go
timerFuture := workflow.NewTimer(ctx, 5*time.Minute)
```

```twf
promise approved <- signal Approved
```

```go
approvedCh := workflow.GetSignalChannel(ctx, "Approved")
// Use approvedCh.Receive later, or add to selector
```

```twf
promise payHandle <- nexus PaymentsEndpoint PaymentsService.ProcessPayment(payment)
```

```go
c := workflow.NewNexusClient("PaymentsEndpoint", "PaymentsService")
payFuture := c.ExecuteOperation(ctx, "ProcessPayment", payment, workflow.NexusOperationOptions{})
```

## Notes

- A promise is just a future — the call starts immediately, `.Get` defers the blocking
- Activity/workflow promises → `workflow.Future` from `ExecuteActivity`/`ExecuteChildWorkflow`
- Timer promises → `workflow.Future` from `workflow.NewTimer`
- Signal promises → `workflow.ReceiveChannel` from `workflow.GetSignalChannel`; await with `.Receive` or add to a selector
- Nexus promises → `workflow.Future` from `NexusClient.ExecuteOperation`; same `.Get` pattern as activity/workflow futures
- Updates are handler-driven, not future-driven — they don't produce futures directly. To race an update completion, use a channel set by the update handler (see [update-handler.md](./update-handler.md))
- Promises used in `await one:` are added as selector cases — see [await-one.md](./await-one.md)
