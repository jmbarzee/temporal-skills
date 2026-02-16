# assignment

## DSL

```twf
paymentStatus = "pending"
status = "awaiting_payment"
```

## Go

```go
paymentStatus := "pending"
status := "awaiting_payment"

// Reassignment (variable already declared):
status = "processing"
```

## Notes

- First use of a variable → `:=` (short declaration); subsequent assignments → `=`
- Workflow-scoped variables are declared at the top of the workflow function so signal/update handlers can access them via closure
- DSL assignments in `state:` blocks and signal handler bodies all map to the same workflow-level variables
