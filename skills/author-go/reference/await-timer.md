# await timer

## DSL

```twf
await timer(5m)
```

## Go

```go
err := workflow.Sleep(ctx, 5*time.Minute)
if err != nil {
    return Result{}, err
}
```

## Notes

- Duration units: `s` → `time.Second`, `m` → `time.Minute`, `h` → `time.Hour`, `d` → `24*time.Hour`
- Variable durations: `await timer(backoff)` → use the variable directly: `workflow.Sleep(ctx, backoff)`
- Inside `await one:`, timers use `workflow.NewTimer` instead — see [await-one.md](./await-one.md)
