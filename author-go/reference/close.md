# close

## close complete

### DSL

```twf
close complete(Result{status: "done"})
```

### Go

```go
return Result{Status: "done"}, nil
```

## close complete (no value)

### DSL

```twf
close complete
```

### Go

```go
return nil
```

## close fail

### DSL

```twf
close fail(OrderResult{status: "cancelled"})
```

### Go

```go
return OrderResult{}, fmt.Errorf("cancelled")
```

## close continue_as_new

### DSL

```twf
close continue_as_new(userId, user)
```

### Go

```go
return workflow.NewContinueAsNewError(ctx, UserEntity, userId, user)
```

## Notes

- `close complete(value)` → `return value, nil`
- `close fail(value)` → return zero value + error; the error message comes from the fail argument. If the DSL passes a struct, extract a meaningful message or use `fmt.Errorf`
- `close continue_as_new` passes args to the same workflow function via `workflow.NewContinueAsNewError`
- `close complete` with no args and no return type → `return nil`
