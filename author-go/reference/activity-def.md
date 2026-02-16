# activity definition

## DSL

```twf
activity ValidateOrder(order: Order) -> (ValidateResult):
    result = validate(order)
    return result
```

## Go

```go
func ValidateOrder(ctx context.Context, order Order) (ValidateResult, error) {
    result, err := validate(ctx, order)
    if err != nil {
        return ValidateResult{}, err
    }
    return result, nil
}
```

## Notes

- Activities use `context.Context` (stdlib), not `workflow.Context`
- Every activity returns `error` as the last return value
- No return type in DSL → `func Name(ctx context.Context, params...) error`
- Activity bodies in `.twf` are pseudocode — ask the user about real implementation logic when it's ambiguous
- In practice, activities are methods on a struct with injected dependencies — the signature becomes `func (a *Activities) ValidateOrder(ctx context.Context, order Order) (ValidateResult, error)`. See the [Activity Implementation Pattern](../SKILL.md#activity-implementation-pattern) in SKILL.md.
