# Nexus: Cross-Namespace Communication

> **Example:** [`nexus.twf`](./nexus.twf)

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

```text
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

```twf
workflow OrderWorkflow(order: Order) -> OrderResult:
    # Validate locally
    activity ValidateOrder(order)
    
    # Call into payments namespace via Nexus
    options(timeout: 5m)
    nexus "payments" workflow ProcessPayment(order.payment) -> paymentResult
    
    # Call into notifications namespace
    nexus "notifications" workflow SendConfirmation(order.customer, paymentResult)
    
    close OrderResult{paymentId: paymentResult.id}
```

### Target Side (Handler)

> Note: Nexus handler definitions use conceptual pseudo-code, not TWF notation. The handler-side API is SDK-specific.

```pseudo
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

> Note: Endpoint configuration is SDK/platform-specific, not TWF notation.

```yaml
nexus_endpoint PaymentsEndpoint:
    target_namespace: payments
    target_task_queue: payments-worker
    allowed_caller_namespaces:
        - orders
        - subscriptions
        - admin
```

### Request Validation in Handler

```pseudo
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

## Execution Modes: Synchronous, `spawn`, `detach`

Nexus calls support the same three execution modes as child workflows:

| Mode | Syntax | Behavior |
|------|--------|----------|
| **Synchronous** | `nexus "ns" workflow Name(args) -> result` | Caller blocks until operation completes |
| **Async (spawn)** | `spawn nexus "ns" workflow Name(args) -> handle` | Caller continues, awaits handle later |
| **Fire-and-forget (detach)** | `detach nexus "ns" workflow Name(args)` | Caller continues, never waits |

### Synchronous (Default)

Caller waits for operation to complete:

```twf
workflow Caller() -> Result:
    # Blocks until ProcessPayment completes
    nexus "payments" workflow ProcessPayment(payment) -> result
    close Result{paymentId: result.id}
```

### Asynchronous

Start operation with `spawn`, continue without waiting, await the handle later:

```twf
workflow Caller() -> Result:
    # Start operation, get handle
    spawn nexus "payments" workflow ProcessPayment(payment) -> handle
    
    # Do other work
    activity DoOtherWork()
    
    # Wait for result when needed
    await one:
        handle -> result:
    close Result{paymentId: result.id}
```

### Fire-and-Forget

Start operation with `detach`, never wait:

```twf
workflow Caller() -> Result:
    # Start and don't wait
    detach nexus "notifications" workflow SendEmail(email)
    
    close Result{status: "initiated"}
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

> Note: Error handling is SDK-specific. The nexus call will fail the workflow if the operation fails. Use retry policies and timeouts to control failure behavior. A timeout can be expressed with `await one:`.

```twf
workflow Caller(data: Data) -> Result:
    # Race nexus call against a deadline
    await one:
        nexus "target" workflow Operation(data) -> result:
            close Result{success: true, data: result}
        timer(5m):
            activity AlertTimeout(data)
            close failed Result{success: false, error: "timeout"}
```

---

## Nexus vs Alternatives

### Nexus vs Direct Activity Call

```twf
# Direct activity: tight coupling, same namespace
workflow OrderWorkflow(order: Order):
    activity ProcessPayment(order.payment)  # Activity in same worker

# Nexus: loose coupling, cross-namespace
workflow OrderWorkflow(order: Order):
    nexus "payments" workflow ProcessPayment(order.payment)  # Separate namespace/team
```

### Nexus vs Signal

> Note: `temporal.signal()` is an SDK-level API call, not TWF notation.

```pseudo
# Signal: fire-and-forget to known workflow (SDK call)
temporal.signal("workflow-b-id", DoSomething, data)
```

```twf
# Nexus: request-response to service endpoint
workflow A():
    nexus "service" workflow Operation(data) -> result  # Get response back
```

### Nexus vs HTTP

```twf
# HTTP from activity: loses Temporal guarantees
workflow A():
    activity CallExternalAPI(data) -> result  # You handle retries, failures

# Nexus: Temporal-native, durable, retryable
workflow A():
    nexus "external" workflow Operation(data) -> result  # Temporal handles failures
```

---

## Design Patterns

### Service Gateway Pattern

Expose multiple operations through a single Nexus endpoint:

```pseudo
nexus_service OrdersGateway:
    operation CreateOrder(order: Order) -> OrderResult
    operation GetOrder(orderId: string) -> Order
    operation CancelOrder(orderId: string) -> CancelResult
    operation UpdateOrder(orderId: string, updates: Updates) -> Order
```

### Request Routing

Route to different workflows based on input:

```pseudo
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

```pseudo
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

```twf
# BAD: Nexus overhead for local calls
workflow A():
    nexus "local" workflow Operation(data)  # Same namespace!

# GOOD: Child workflow for same namespace
workflow A():
    workflow OperationWorkflow(data)
```

### Ignoring Authorization

```pseudo
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

```twf
# BAD: No deadline
workflow A():
    nexus "target" workflow SlowOperation(data)  # May hang indefinitely

# GOOD: Explicit deadline via await one
workflow A():
    await one:
        nexus "target" workflow SlowOperation(data) -> result:
        timer(5m):
            close failed Result{error: "timeout"}
```

---

## Monitoring and Debugging

### Tracing Across Namespaces

Nexus preserves trace context across namespace boundaries:

```text
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
