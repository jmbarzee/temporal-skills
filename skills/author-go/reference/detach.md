# detach

## DSL

```twf
detach workflow NotifyCustomer(order.customer)
```

## Go

```go
childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
    ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON,
})
workflow.ExecuteChildWorkflow(childCtx, NotifyCustomer, order.Customer)
// No .Get() — fire-and-forget
```

## Detach with nexus

### DSL

```twf
detach nexus NotificationsEndpoint NotificationsService.SendConfirmation(order.customer, paymentResult)
```

### Go

```go
workflow.ExecuteNexusOperation(ctx, "NotificationsEndpoint", notificationsservice.SendConfirmation, sendConfirmationInput, workflow.NexusOperationOptions{})
// No .Get() — fire-and-forget
```

## Notes

- `detach` = start the child but never wait for its result
- Set `ParentClosePolicy` to `ABANDON` so the child survives if the parent completes
- Do not call `.Get()` on the returned future
- The child workflow runs independently and its success/failure does not affect the parent
