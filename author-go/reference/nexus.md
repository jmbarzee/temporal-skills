# nexus

## DSL

```twf
nexus "payments" workflow ProcessPayment(order.payment) -> paymentResult
```

## Go

```go
// Define the Nexus operation reference
processPaymentOp := nexus.NewWorkflowRunOperation(
    "ProcessPayment",
    ProcessPayment,
    func(ctx context.Context, input Payment, opts nexus.StartWorkflowOptions) (client.StartWorkflowOptions, error) {
        return client.StartWorkflowOptions{}, nil
    },
)

// In the calling workflow:
nexusClient := workflow.NewNexusClient("payments", "ProcessPayment")
paymentFuture := nexusClient.ExecuteOperation(ctx, processPaymentOp, order.Payment, workflow.NexusOperationOptions{})
var paymentResult PaymentResult
err := paymentFuture.Get(ctx, &paymentResult)
if err != nil {
    return Result{}, err
}
```

## Notes

- Nexus requires endpoint configuration on the Temporal server — the namespace string maps to a registered Nexus endpoint
- The operation reference and client setup are boilerplate; the calling pattern mirrors child workflows
- For fire-and-forget nexus: see [detach.md](./detach.md)
- Nexus API is evolving — check the SDK version in `go.mod` for the exact API surface
