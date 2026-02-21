# await all

## DSL

```twf
await all:
    activity ReserveInventory(order) -> inventory
    activity ProcessPayment(order) -> payment
```

## Go

```go
var inventory Inventory
var payment Payment

var inventoryErr, paymentErr error

workflow.Go(ctx, func(gCtx workflow.Context) {
    inventoryErr = workflow.ExecuteActivity(gCtx, ReserveInventory, order).Get(gCtx, &inventory)
})
workflow.Go(ctx, func(gCtx workflow.Context) {
    paymentErr = workflow.ExecuteActivity(gCtx, ProcessPayment, order).Get(gCtx, &payment)
})

// Wait for all goroutines — use workflow.Await to block until both complete
err := workflow.Await(ctx, func() bool {
    return inventoryErr != nil || paymentErr != nil ||
        (inventory != Inventory{} && payment != Payment{})
})
```

## Fan-out pattern

### DSL

```twf
await all:
    for (item in items):
        activity ProcessBatchItem(item)
```

### Go

```go
futures := make([]workflow.Future, len(items))
for i, item := range items {
    futures[i] = workflow.ExecuteActivity(ctx, ProcessBatchItem, item)
}
for _, f := range futures {
    if err := f.Get(ctx, nil); err != nil {
        return Result{}, err
    }
}
```

## Mixed activity + nexus

### DSL

```twf
await all:
    activity ReserveInventory(order) -> inventory
    nexus PaymentsEndpoint PaymentsService.ProcessPayment(order.payment) -> payment
```

### Go

```go
var inventory Inventory
var payment PaymentResult
var inventoryErr, paymentErr error

workflow.Go(ctx, func(gCtx workflow.Context) {
    inventoryErr = workflow.ExecuteActivity(gCtx, ReserveInventory, order).Get(gCtx, &inventory)
})
workflow.Go(ctx, func(gCtx workflow.Context) {
    c := workflow.NewNexusClient("PaymentsEndpoint", "PaymentsService")
    paymentErr = c.ExecuteOperation(gCtx, "ProcessPayment", order.Payment, workflow.NexusOperationOptions{}).Get(gCtx, &payment)
})

err := workflow.Await(ctx, func() bool {
    return inventoryErr != nil || paymentErr != nil ||
        (inventory != Inventory{} && payment != PaymentResult{})
})
```

## Notes

- Each statement in `await all:` runs in its own `workflow.Go` goroutine
- Use a `workflow.WaitGroup` (if available in SDK version) or `workflow.Await` with a completion predicate to join
- Fan-out with `for`: start all futures first, then `.Get` all — no goroutines needed since `ExecuteActivity` returns immediately
- Nexus operations in `await all:` follow the same goroutine pattern — `ExecuteOperation` returns a future, `.Get()` blocks in the goroutine
- Errors: check each goroutine's error after joining; propagation strategy depends on workflow requirements
