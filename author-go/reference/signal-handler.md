# signal handler

## DSL

```twf
workflow OrderWorkflow(orderId: string) -> (OrderResult):
    signal PaymentReceived(transactionId: string, amount: decimal):
        status = "payment_received"
        lastTransactionId = transactionId

    # ... body uses await signal PaymentReceived or await one with signal case
```

## Go

```go
func OrderWorkflow(ctx workflow.Context, orderId string) (OrderResult, error) {
    var status string
    var lastTransactionId string

    // Signal struct
    type PaymentReceivedSignal struct {
        TransactionId string
        Amount        float64
    }

    // Signal channel
    paymentReceivedCh := workflow.GetSignalChannel(ctx, "PaymentReceived")

    // Register handler via goroutine that loops on the channel
    workflow.Go(ctx, func(gCtx workflow.Context) {
        for {
            var sig PaymentReceivedSignal
            paymentReceivedCh.Receive(gCtx, &sig)
            status = "payment_received"
            lastTransactionId = sig.TransactionId
        }
    })

    // ... workflow body
}
```

## Notes

- Signal params become a struct; signal name becomes the channel name string
- The handler goroutine loops forever — it processes every signal arrival, not just the first
- Handler body mutates workflow-scoped variables (closures over workflow state)
- Signals with no params: use `paymentReceivedCh.Receive(gCtx, nil)`
- When a signal is also used in `await one:`, the selector reads from the same channel — see [await-one.md](./await-one.md)
