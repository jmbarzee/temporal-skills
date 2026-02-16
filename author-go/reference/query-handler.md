# query handler

## DSL

```twf
query GetStatus() -> (string):
    return status
```

## Go

```go
err := workflow.SetQueryHandler(ctx, "GetStatus", func() (string, error) {
    return status, nil
})
if err != nil {
    return Result{}, err
}
```

## Notes

- Query handlers always return `(ReturnType, error)`
- Query handlers must not modify workflow state — read-only
- Query handlers have no `workflow.Context` parameter in their signature
- With params: `func(param1 Type1, param2 Type2) (ReturnType, error)`
- Register queries early in the workflow function (before any blocking calls) so they're available immediately
