# workflow definition

## DSL

```twf
workflow ProcessOrder(order: Order) -> (Result):
    # body
    close complete(Result{status: "done"})
```

## Go

```go
func ProcessOrder(ctx workflow.Context, order Order) (Result, error) {
    // body
    return Result{Status: "done"}, nil
}
```

## Notes

- Every workflow returns `error` as the last return value, even if the DSL has no return type (signature becomes `func Name(ctx workflow.Context, params...) error`)
- `workflow.Context` is always the first parameter
- No return type in DSL → `func Name(ctx workflow.Context, params...) error`
- Multiple return types `-> (A, B)` → `func Name(ctx workflow.Context, ...) (A, B, error)`
