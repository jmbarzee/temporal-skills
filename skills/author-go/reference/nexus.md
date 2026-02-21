# nexus

## DSL

```twf
nexus PaymentsEndpoint PaymentsService.ProcessPayment(order.payment) -> paymentResult
```

## Go

```go
// Execute a Nexus operation from within a workflow.
// The endpoint and operation are typed references — no string-based client construction.
var paymentResult PaymentResult
err := workflow.ExecuteNexusOperation(ctx, "PaymentsEndpoint", paymentsservice.ProcessPayment, order.Payment, workflow.NexusOperationOptions{})
if err != nil {
    return Result{}, err
}
```

## Notes

- Endpoints and services are first-class DSL constructs: endpoints are declared inside `namespace` blocks (`nexus endpoint EndpointName`) and services are defined as top-level `nexus service Name:` blocks with typed operations
- The Go code uses typed operation references (generated from `nexus service` definitions), not string-based clients
- The calling pattern mirrors child workflows — execute and `.Get()` (or assign inline)
- For fire-and-forget nexus: see [detach.md](./detach.md)
- Nexus API is evolving — check the SDK version in `go.mod` for the exact API surface
