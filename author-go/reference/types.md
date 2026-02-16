# types

## Strategy

Resolve types in priority order:

1. **Explicit signatures** — workflow/activity params and return types name the type directly
2. **Constructor usage** — `Result{field: value}` reveals struct fields and their types (inferred from the values assigned)
3. **Field access** — `order.items` implies `Order` has an `items` field; the accessed type propagates from usage context
4. **Imports** — check `go.mod` and project code for existing types that match; prefer importing over generating
5. **Generate** — only for application-specific types with no existing match

## Go

```go
// From explicit signature: activity ValidateOrder(order: Order) -> (ValidateResult)
type Order struct { /* fields derived from usage */ }
type ValidateResult struct { /* fields derived from constructor */ }

// From constructor: Result{status: "completed", trackingId: reservation.trackingId}
type Result struct {
    Status     string
    TrackingId string
}

// Primitive mapping
// string  → string
// int     → int
// decimal → float64
// bool    → bool
// []Type  → []Type
```

## Notes

- Collect all constructor sites and field accesses across the `.twf` file before defining a type — a type used in multiple places may reveal different fields in each
- When a field's type is ambiguous (e.g., assigned from an untyped expression), leave a `// TODO` comment and ask the user
- Generic containers: `Map[K]V` → `map[K]V`, `[]Item` → `[]Item`
- Export all struct fields (uppercase) — these cross workflow/activity boundaries via serialization
