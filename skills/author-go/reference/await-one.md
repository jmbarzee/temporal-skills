# await one

## DSL

```twf
await one:
    signal PaymentReceived:
        status = "processing"
    timer(24h):
        activity CancelOrder(orderId)
        close fail(OrderResult{status: "cancelled"})
```

## Go

```go
sel := workflow.NewSelector(ctx)

sel.AddReceive(paymentReceivedCh, func(ch workflow.ReceiveChannel, more bool) {
    var sig PaymentReceivedSignal
    ch.Receive(ctx, &sig)
    status = "processing"
})

timerFuture := workflow.NewTimer(ctx, 24*time.Hour)
sel.AddFuture(timerFuture, func(f workflow.Future) {
    if err := f.Get(ctx, nil); err != nil {
        // timer cancelled
        return
    }
    _ = workflow.ExecuteActivity(ctx, CancelOrder, orderId).Get(ctx, nil)
    // close fail handled after selector
})

sel.Select(ctx)
```

## Case types

**Signal case** — `sel.AddReceive(channel, handler)`
```go
sel.AddReceive(signalCh, func(ch workflow.ReceiveChannel, more bool) {
    var sig SignalType
    ch.Receive(ctx, &sig)
    // case body
})
```

**Timer case** — `sel.AddFuture(workflow.NewTimer(...), handler)`
```go
sel.AddFuture(workflow.NewTimer(ctx, duration), func(f workflow.Future) {
    _ = f.Get(ctx, nil)
    // case body
})
```

**Activity case** — `sel.AddFuture(workflow.ExecuteActivity(...), handler)`
```go
sel.AddFuture(workflow.ExecuteActivity(ctx, DoWork, args), func(f workflow.Future) {
    var result ResultType
    _ = f.Get(ctx, &result)
    // case body
})
```

**Workflow case** — `sel.AddFuture(workflow.ExecuteChildWorkflow(...), handler)`
```go
sel.AddFuture(workflow.ExecuteChildWorkflow(ctx, Child, args), func(f workflow.Future) {
    var result ResultType
    _ = f.Get(ctx, &result)
    // case body
})
```

**Promise (ident) case** — add the existing future or channel to the selector
```go
// future promise
sel.AddFuture(myFuture, func(f workflow.Future) {
    var result ResultType
    _ = f.Get(ctx, &result)
    // case body
})

// condition promise — use a goroutine that signals a channel
condCh := workflow.NewChannel(ctx)
workflow.Go(ctx, func(gCtx workflow.Context) {
    _ = workflow.Await(gCtx, func() bool { return myCondition })
    condCh.Send(gCtx, true)
})
sel.AddReceive(condCh, func(ch workflow.ReceiveChannel, more bool) {
    ch.Receive(ctx, nil)
    // case body
})
```

## Notes

- `sel.Select(ctx)` blocks until exactly one case fires — the first to complete wins
- Cancellation of losing cases is automatic for timers and child workflows when the parent context is cancelled
- Empty case bodies (just the colon in DSL) → handler function with only the `Receive`/`Get` call, no additional logic
- For `close` inside a case body: set a variable in the handler, check it after `sel.Select`, then return
- Nested `await all:` inside `await one:` → wrap the `await all` logic in a future via `workflow.Go` + channel, add as `AddReceive`
