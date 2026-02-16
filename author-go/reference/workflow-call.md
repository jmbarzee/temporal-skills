# workflow call (child)

## DSL

```twf
workflow ShipOrder(order) -> shipResult
```

## Go

```go
var shipResult ShipResult
err := workflow.ExecuteChildWorkflow(ctx, ShipOrder, order).Get(ctx, &shipResult)
if err != nil {
    return Result{}, err
}
```

## Notes

- The child workflow function is passed by reference — `workflow.ExecuteChildWorkflow(ctx, FuncName, args...)`
- `ctx` must carry child workflow options; see [options.md](./options.md) for setting `ChildWorkflowOptions`
- For fire-and-forget, see [detach.md](./detach.md)
- For cross-namespace, see [nexus.md](./nexus.md)
