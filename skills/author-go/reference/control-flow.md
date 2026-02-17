# control flow

## if / else

### DSL

```twf
if (validated.priority == "high"):
    activity ExpediteOrder(order)
else:
    activity StandardProcessing(order)
```

### Go

```go
if validated.Priority == "high" {
    err := workflow.ExecuteActivity(ctx, ExpediteOrder, order).Get(ctx, nil)
    if err != nil {
        return Result{}, err
    }
} else {
    err := workflow.ExecuteActivity(ctx, StandardProcessing, order).Get(ctx, nil)
    if err != nil {
        return Result{}, err
    }
}
```

## for (iteration)

### DSL

```twf
for (item in order.items):
    activity ProcessItem(item)
```

### Go

```go
for _, item := range order.Items {
    err := workflow.ExecuteActivity(ctx, ProcessItem, item).Get(ctx, nil)
    if err != nil {
        return Result{}, err
    }
}
```

## for (conditional)

### DSL

```twf
for (retries < maxRetries):
    activity Attempt(data)
    retries = retries + 1
```

### Go

```go
for retries < maxRetries {
    err := workflow.ExecuteActivity(ctx, Attempt, data).Get(ctx, nil)
    if err != nil {
        return Result{}, err
    }
    retries++
}
```

## for (infinite loop)

### DSL

```twf
for:
    # body
    if (done):
        break
```

### Go

```go
for {
    // body
    if done {
        break
    }
}
```

## switch

### DSL

```twf
switch (phase):
    case "draft":
        # ...
    case "approved":
        # ...
    else:
        # ...
```

### Go

```go
switch phase {
case "draft":
    // ...
case "approved":
    // ...
default:
    // ...
}
```

## Notes

- `break` → `break`, `continue` → `continue` — direct mapping
- DSL `else:` in switch → Go `default:`
- DSL boolean operators: `and` → `&&`, `or` → `||`, `not` → `!`
