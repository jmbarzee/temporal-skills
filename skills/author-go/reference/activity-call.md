# activity call

## DSL

```twf
activity ValidateOrder(order) -> validated
```

## Go

```go
var validated ValidateResult
err := workflow.ExecuteActivity(ctx, ValidateOrder, order).Get(ctx, &validated)
if err != nil {
    return Result{}, err
}
```

## With inline options

### DSL

```twf
activity ValidateOrder(order) -> validated
    options:
        start_to_close_timeout: 30s
```

### Go

```go
actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Second,
})
var validated ValidateResult
err := workflow.ExecuteActivity(actCtx, ValidateOrder, order).Get(ctx, &validated)
if err != nil {
    return Result{}, err
}
```

For the full options reference, see [options.md](./options.md).

## Notes

- No return value: omit the `Get` target variable, still check `err`
- The activity function is passed by reference (not a string) — `workflow.ExecuteActivity(ctx, FuncName, args...)`
- When using the struct method pattern (see [Activity Implementation Pattern](../SKILL.md#activity-implementation-pattern)), pass a method reference via nil pointer: `var a *Activities; workflow.ExecuteActivity(ctx, a.ValidateOrder, order)`
- `ctx` must carry activity options; see [options.md](./options.md) for setting `ActivityOptions`
