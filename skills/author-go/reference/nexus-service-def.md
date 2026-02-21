# nexus service definition

## DSL

```twf
nexus service PaymentsService:
    operation ProcessPayment(PaymentRequest) -> (PaymentResult)
    operation RefundPayment(RefundRequest) -> (RefundResult)
```

## Go — Service contract (shared types)

```go
const PaymentsServiceName = "PaymentsService"
const ProcessPaymentOp = "ProcessPayment"
const RefundPaymentOp = "RefundPayment"
```

## Go — Async operation handler (workflow-backed)

```go
var ProcessPaymentOperation = temporalnexus.NewWorkflowRunOperation(
    ProcessPaymentOp,
    ProcessPaymentWorkflow,
    func(ctx context.Context, input PaymentRequest, options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
        return client.StartWorkflowOptions{}, nil
    },
)
```

## Go — Sync operation handler

```go
var RefundPaymentOperation = nexus.NewSyncOperation(RefundPaymentOp, func(ctx context.Context, input RefundRequest, options nexus.StartOperationOptions) (RefundResult, error) {
    // Direct implementation or use temporalnexus.GetClient(ctx) for Temporal client calls
    return RefundResult{}, nil
})
```

## Go — Registration on worker

```go
service := nexus.NewService(PaymentsServiceName)
err := service.Register(ProcessPaymentOperation, RefundPaymentOperation)
if err != nil {
    log.Fatalln("Unable to register operations", err)
}
w.RegisterNexusService(service)
w.RegisterWorkflow(ProcessPaymentWorkflow) // handler workflows must also be registered
```

## Notes

- Imports: `"github.com/nexus-rpc/sdk-go/nexus"` for `nexus.NewService`, `nexus.NewSyncOperation`; `"go.temporal.io/sdk/temporalnexus"` for `NewWorkflowRunOperation`, `GetClient`
- Async operations (backed by workflows) use `temporalnexus.NewWorkflowRunOperation` — the workflow is started and the nexus operation resolves when the workflow completes
- Sync operations use `nexus.NewSyncOperation` — for direct computation or queries/signals via `temporalnexus.GetClient(ctx)`
- The service is registered on the handler worker (the target namespace's worker), not the caller
- Handler workflows must also be registered with `RegisterWorkflow` on the same worker
