# condition

## DSL

```twf
workflow ClusterManager(config: Config):
    state:
        condition clusterStarted

    # ... later in body:
    set clusterStarted

    # ... elsewhere:
    await clusterStarted
```

## Go

```go
func ClusterManager(ctx workflow.Context, config Config) error {
    clusterStarted := false

    // ... later in body:
    clusterStarted = true

    // ... elsewhere:
    err := workflow.Await(ctx, func() bool { return clusterStarted })
    if err != nil {
        return err
    }
}
```

## Notes

- `condition name` in `state:` → `name := false` (a `bool` variable)
- `set name` → `name = true`
- `unset name` → `name = false`
- `await name` → `workflow.Await(ctx, func() bool { return name })`
- Conditions in `await one:` become selector cases via `workflow.AwaitWithTimeout` or a separate goroutine — see [await-one.md](./await-one.md)
