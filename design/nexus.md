# Nexus: Cross-Namespace Communication

> **Example:** [`examples/nexus.twf`](./examples/nexus.twf)

Nexus enables workflows in one Temporal namespace to call workflows in another namespace, with proper authorization and abstraction.

## When to Use Nexus

| Use Nexus | Use Child Workflow Instead |
|-----------|---------------------------|
| Cross-namespace calls | Same namespace |
| Cross-team boundaries | Same team |
| Different security contexts | Same security context |
| Service abstraction needed | Direct coupling acceptable |
| Multi-tenant architectures | Single-tenant |

---

## Nexus Concepts

### Architecture

```
┌─────────────────────────────────────┐
│           Caller Namespace          │
│  ┌─────────────────────────────┐   │
│  │       Caller Workflow       │   │
│  │  nexus target/Operation()   │───┼──┐
│  └─────────────────────────────┘   │  │
└─────────────────────────────────────┘  │
                                         │ Nexus
┌─────────────────────────────────────┐  │
│          Target Namespace           │  │
│  ┌─────────────────────────────┐   │  │
│  │    Nexus Endpoint Handler   │◄──┼──┘
│  │  → starts target workflow   │   │
│  └─────────────────────────────┘   │
│  ┌─────────────────────────────┐   │
│  │      Target Workflow        │   │
│  └─────────────────────────────┘   │
└─────────────────────────────────────┘
```

### Components

| Component | Description |
|-----------|-------------|
| **Nexus Endpoint** | Named entry point in target namespace |
| **Nexus Operation** | Specific operation exposed by endpoint |
| **Nexus Service** | Collection of operations (like an API) |
| **Caller** | Workflow invoking the Nexus operation |
| **Handler** | Code that receives and processes Nexus call |

---

## Basic Nexus Pattern

### Caller Side

```
workflow OrderWorkflow(order: Order) -> OrderResult:
    # Validate locally
    activity ValidateOrder(order)
    
    # Call into payments namespace via Nexus
    paymentResult = nexus payments/ProcessPayment(order.payment):
        timeout: 5m
    
    # Call into notifications namespace
    nexus notifications/SendConfirmation(order.customer, paymentResult)
    
    return OrderResult{paymentId: paymentResult.id}
```

### Target Side (Handler)

```
nexus_service PaymentsService:
    
    operation ProcessPayment(payment: Payment) -> PaymentResult:
        # Start a workflow in this namespace to handle the request
        return workflow ProcessPaymentWorkflow(payment):
            workflow_id: "payment-{payment.id}"

workflow ProcessPaymentWorkflow(payment: Payment) -> PaymentResult:
    activity ValidatePayment(payment)
    result = activity ChargeCard(payment)
    activity RecordTransaction(result)
    return result
```

---

## Authorization

### Configuring Allowed Callers

```
nexus_endpoint PaymentsEndpoint:
    target_namespace: payments
    target_task_queue: payments-worker
    allowed_caller_namespaces:
        - orders
        - subscriptions
        - admin
```

### Request Validation in Handler

```
nexus_service PaymentsService:
    
    operation ProcessPayment(payment: Payment) -> PaymentResult:
        # Validate caller context
        caller = get_nexus_caller_info()
        
        if caller.namespace not in allowed_namespaces:
            raise UnauthorizedError("Caller not allowed")
        
        # Validate request
        if payment.amount <= 0:
            raise InvalidArgumentError("Amount must be positive")
        
        return workflow ProcessPaymentWorkflow(payment)
```

---

## Synchronous vs Asynchronous Operations

### Synchronous (Default)

Caller waits for operation to complete:

```
workflow Caller() -> Result:
    # Blocks until ProcessPayment completes
    result = nexus payments/ProcessPayment(payment)
    return Result{paymentId: result.id}
```

### Asynchronous

Start operation, continue without waiting:

```
workflow Caller() -> Result:
    # Start operation, get handle
    handle = nexus_async payments/ProcessPayment(payment)
    
    # Do other work
    activity DoOtherWork()
    
    # Wait for result when needed
    result = await handle
    return Result{paymentId: result.id}
```

### Fire-and-Forget

Start operation, never wait:

```
workflow Caller() -> Result:
    # Start and don't wait
    nexus_async notifications/SendEmail(email):
        wait: false
    
    return Result{status: "initiated"}
```

---

## Error Handling

### Error Types

| Error | Meaning | Handling |
|-------|---------|----------|
| `NexusOperationError` | Operation failed (application error) | Handle based on error type |
| `NexusTimeoutError` | Operation timed out | Retry or fail |
| `NexusUnauthorizedError` | Caller not allowed | Configuration issue |
| `NexusNotFoundError` | Endpoint/operation doesn't exist | Configuration issue |

### Error Handling Pattern

```
workflow Caller(data: Data) -> Result:
    try:
        result = nexus target/Operation(data):
            timeout: 5m
        return Result{success: true, data: result}
    
    catch NexusOperationError as e:
        # Application-level failure from target
        if e.type == "ValidationError":
            return Result{success: false, error: "invalid input"}
        elif e.type == "ResourceNotFound":
            return Result{success: false, error: "not found"}
        else:
            raise  # Unexpected error, let workflow fail
    
    catch NexusTimeoutError:
        # Operation took too long
        activity AlertTimeout(data)
        return Result{success: false, error: "timeout"}
```

---

## Nexus vs Alternatives

### Nexus vs Direct Activity Call

```
# Direct activity: tight coupling, same namespace
workflow OrderWorkflow(order: Order):
    activity ProcessPayment(order.payment)  # Activity in same worker

# Nexus: loose coupling, cross-namespace
workflow OrderWorkflow(order: Order):
    nexus payments/ProcessPayment(order.payment)  # Separate namespace/team
```

### Nexus vs Signal

```
# Signal: fire-and-forget to known workflow
workflow A():
    temporal.signal("workflow-b-id", DoSomething, data)

# Nexus: request-response to service endpoint
workflow A():
    result = nexus service/Operation(data)  # Get response back
```

### Nexus vs HTTP

```
# HTTP from activity: loses Temporal guarantees
workflow A():
    result = activity CallExternalAPI(data)  # You handle retries, failures

# Nexus: Temporal-native, durable, retryable
workflow A():
    result = nexus external/Operation(data)  # Temporal handles failures
```

---

## Design Patterns

### Service Gateway Pattern

Expose multiple operations through a single Nexus endpoint:

```
nexus_service OrdersGateway:
    operation CreateOrder(order: Order) -> OrderResult
    operation GetOrder(orderId: string) -> Order
    operation CancelOrder(orderId: string) -> CancelResult
    operation UpdateOrder(orderId: string, updates: Updates) -> Order
```

### Request Routing

Route to different workflows based on input:

```
nexus_service ProcessingService:
    
    operation Process(request: Request) -> Result:
        # Route based on request type
        if request.type == "fast":
            return workflow FastProcessing(request)
        elif request.type == "batch":
            return workflow BatchProcessing(request)
        else:
            return workflow StandardProcessing(request)
```

### Multi-Tenant Routing

```
nexus_service TenantService:
    
    operation ProcessForTenant(tenantId: string, data: Data) -> Result:
        # Route to tenant-specific task queue
        return workflow TenantWorkflow(data):
            task_queue: "tenant-{tenantId}"
            workflow_id: "tenant-{tenantId}-{data.id}"
```

---

## Anti-Patterns

### Nexus for Same-Namespace Calls

```
# BAD: Nexus overhead for local calls
workflow A():
    nexus local/Operation(data)  # Same namespace!

# GOOD: Child workflow for same namespace
workflow A():
    child OperationWorkflow(data)
```

### Ignoring Authorization

```
# BAD: No caller validation
nexus_service OpenService:
    operation SensitiveOperation(data):
        return workflow DoSensitiveThing(data)

# GOOD: Validate caller
nexus_service SecureService:
    operation SensitiveOperation(data):
        if not authorized(get_caller_info()):
            raise UnauthorizedError()
        return workflow DoSensitiveThing(data)
```

### Missing Timeout Configuration

```
# BAD: Default/no timeout
workflow A():
    nexus target/SlowOperation(data)  # May hang indefinitely

# GOOD: Explicit timeout
workflow A():
    nexus target/SlowOperation(data):
        timeout: 5m
```

---

## Monitoring and Debugging

### Tracing Across Namespaces

Nexus preserves trace context across namespace boundaries:

```
Caller workflow (namespace A)
└─ nexus call
   └─ Handler (namespace B)
      └─ Target workflow (namespace B)
         └─ activities
```

### Debugging Tips

| Issue | Check |
|-------|-------|
| "Operation not found" | Endpoint exists? Operation registered? |
| "Unauthorized" | Caller namespace in allowed list? |
| Timeout | Target workflow stuck? Task queue has workers? |
| Unexpected error | Check target workflow history |
