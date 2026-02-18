# types

## Strategy

Types in generated code come from two sources. Each has a different resolution method.

**Types you define** — derived from the `.twf` file. You control the shape; the TWF tells you what fields exist.

**Dependency types** — from external packages. The shape is fixed; you discover it at the call site.

### Defined types

Resolve in priority order:

1. **Explicit signatures** — workflow/activity params and return types name the type directly
2. **Constructor usage** — `Result{field: value}` reveals struct fields and their types (inferred from the values assigned)
3. **Field access** — `order.items` implies `Order` has an `items` field; the accessed type propagates from usage context
4. **Generate** — only for application-specific types with no existing match

### Dependency types

The ground truth is the call site — the method you will call in the activity body.

1. **Identify the method** — which function or method on the dependency will the activity call?
2. **Read the method signature** — `go doc <package>.Method` gives you the exact parameter and return types
3. **Read each parameter type** — `go doc <package>.ParamType` gives you the fields you need to populate. Verify every field type the same way — the field name alone can be misleading (e.g., a `Tools` field may accept a union wrapper type, not the type you'd guess from the name)
4. **Follow the chain until you reach primitives or types you recognize** — stop when every type in the call is verified

A dependency is resolved when you can write the full call expression with concrete types:
```
client.Messages.New(ctx, anthropic.MessageNewParams{
    Model:    anthropic.Model(model),     // verified: Model is a string typedef
    Messages: []anthropic.MessageParam{}, // verified: not []Message
    Tools:    []anthropic.ToolUnionParam{}, // verified: not []ToolParam
})
```

### Serialization boundaries

Workflow and activity parameters pass through Temporal's data converter (JSON by default). Types you define are safe — you control their fields and they round-trip cleanly. Dependency types may have custom marshaling, unexported fields, or non-JSON-safe constructs.

**Keep dependency types inside activity bodies.** Expose your own types in workflow/activity signatures and convert to dependency types within the activity implementation. This keeps the serialization boundary clean and decouples your workflow logic from any single dependency.

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
