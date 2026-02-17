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

## Notes

- No return value: omit the `Get` target variable, still check `err`
- The activity function is passed by reference (not a string) — `workflow.ExecuteActivity(ctx, FuncName, args...)`
- When using the struct method pattern (see [Activity Implementation Pattern](../SKILL.md#activity-implementation-pattern)), pass a method reference via nil pointer: `var a *Activities; workflow.ExecuteActivity(ctx, a.ValidateOrder, order)`
- `ctx` must carry activity options; see [options.md](./options.md) for setting `ActivityOptions`
