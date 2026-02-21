# nexus

## DSL

```twf
nexus PaymentsEndpoint PaymentsService.ProcessPayment(order.payment) -> paymentResult
```

## Go

```go
c := workflow.NewNexusClient("PaymentsEndpoint", "PaymentsService")
var paymentResult PaymentResult
fut := c.ExecuteOperation(ctx, "ProcessPayment", order.Payment, workflow.NexusOperationOptions{})
if err := fut.Get(ctx, &paymentResult); err != nil {
    return Result{}, err
}
```

## Notes

- `workflow.NewNexusClient(endpoint, service)` creates a typed client scoped to one endpoint + service pair
- `ExecuteOperation(ctx, operationName, input, options)` starts the operation and returns a `NexusOperationFuture`
- The calling pattern mirrors child workflows — execute and `.Get()`
- For nexus options (timeouts), see [options.md](./options.md)
- For fire-and-forget nexus: see [detach.md](./detach.md)
- For nexus service definitions and handler registration: see [nexus-service-def.md](./nexus-service-def.md)
