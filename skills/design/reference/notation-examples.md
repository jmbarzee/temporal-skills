# TWF Notation Examples

## Basic Structure

A complete `.twf` file: workflows, activities, worker registration, and namespace deployment.

```twf
workflow WorkflowName(input: InputType) -> (OutputType):
    activity ActivityName(input) -> result
    workflow ChildWorkflowName(input) -> childResult
    close complete(OutputType{result, childResult})

workflow ChildWorkflowName(input: InputType) -> (ChildResult):
    activity DoWork(input) -> result
    close complete(ChildResult{result})

activity ActivityName(input: InputType) -> (Result):
    return process(input)

activity DoWork(input: InputType) -> (WorkResult):
    return work(input)

worker mainWorker:
    workflow WorkflowName
    workflow ChildWorkflowName
    activity ActivityName
    activity DoWork

namespace default:
    worker mainWorker
        options:
            task_queue: "main"
```

## Activity Body Detail

Activity bodies are intentionally free-form (`raw_stmt`) — pseudocode or descriptive text representing SDK-level implementation. Detail level depends on how obvious the behavior is from name and signature:

**Obvious** — minimal body:

```twf
activity SendEmail(to: string, body: string):
    send(to, body)
```

**Non-obvious** — describe key operations and external systems:

```twf
activity ExecuteToolCalls(toolCalls: ToolCalls) -> (ToolResults):
    # Look up each tool by name in the tool registry
    # Execute calls in parallel where possible
    # If a tool is not found, return an error result (don't fail the activity)
```

**Complex contract** — describe error conditions, ordering, and idempotency requirements:

```twf
activity ReconcileInventory(warehouseId: string, expected: Inventory) -> (ReconcileResult):
    # Fetch current inventory, diff against expected, flag discrepancies
    # Must be idempotent — running twice with same input produces same flags
    # Warehouse API is rate-limited: max 10 requests/second
```

## Control Flow

```twf
workflow ProcessOrder(order: Order) -> (Result):
    activity ValidateOrder(order) -> validated

    # Conditionals
    if (validated.priority == "high"):
        activity ExpediteOrder(order)
    else:
        activity StandardProcessing(order)

    # Sequential loop — use for when each iteration depends on order or shared state
    # For independent iterations, consider await all with parallel activities instead
    for (item in order.items):
        activity ProcessItem(item)

    # Parallel execution — use await all when tasks are independent and all results needed
    await all:
        activity ReserveInventory(order) -> inventory
        activity ProcessPayment(order) -> payment

    close complete(Result{inventory, payment})

# Every referenced activity must be defined
activity ValidateOrder(order: Order) -> (ValidateResult):
    return validate(order)

activity ExpediteOrder(order: Order):
    expedite(order)

activity StandardProcessing(order: Order):
    process(order)

activity ProcessItem(item: Item) -> (ItemResult):
    return process(item)

activity ReserveInventory(order: Order) -> (Inventory):
    return reserve(order)

activity ProcessPayment(order: Order) -> (Payment):
    return charge(order)
```

## Temporal Primitives in Notation

```twf
workflow OrderFulfillment(orderId: string) -> (OrderResult):
    # Handlers go before body
    # signal = write (fire-and-forget)
    signal PaymentReceived(transactionId: string, amount: decimal):
        paymentStatus = "received"
        lastTransactionId = transactionId

    # query = read (caller gets result)
    query GetOrderStatus() -> (OrderStatus):
        return OrderStatus{status: status, payment: paymentStatus}

    # update = read-write (caller sends data, gets result)
    update UpdateShippingAddress(address: Address) -> (Result):
        activity ValidateAddress(address) -> validation
        if (validation.valid):
            shippingAddress = address
            return Result{success: true}
        else:
            return Result{success: false, error: validation.reason}

    # Workflow body starts after handlers
    activity GetOrder(orderId) -> order
    paymentStatus = "pending"
    status = "awaiting_payment"

    # Durable timer
    await timer(1h)

    # Wait for signal with timeout
    await one:
        signal PaymentReceived:
            status = "processing"
        timer(24h):
            activity CancelOrder(orderId)
            close fail(OrderResult{status: "cancelled"})

    # Child workflow
    workflow ShipOrder(order) -> shipResult

    # Cross-namespace nexus call
    nexus NotificationsEndpoint NotificationsService.SendNotification(order.customer, "shipped")

    close complete(OrderResult{status: "completed"})

# Supporting definitions
activity GetOrder(orderId: string) -> (Order):
    return db.get(orderId)

activity ValidateAddress(address: Address) -> (Validation):
    return validate(address)

activity CancelOrder(orderId: string):
    cancel(orderId)

workflow ShipOrder(order: Order) -> (ShipResult):
    activity CreateShipment(order) -> shipment
    close complete(ShipResult{shipment})

activity CreateShipment(order: Order) -> (Shipment):
    return ship(order)

workflow SendNotification(customer: Customer, message: string):
    activity Notify(customer, message)
    close complete

activity Notify(customer: Customer, message: string):
    send(customer, message)

nexus service NotificationsService:
    async SendNotification workflow SendNotification

worker orderFulfillmentWorker:
    workflow OrderFulfillment
    workflow ShipOrder
    workflow SendNotification
    activity GetOrder
    activity ValidateAddress
    activity CancelOrder
    activity CreateShipment
    activity Notify
    nexus service NotificationsService

namespace default:
    worker orderFulfillmentWorker
        options:
            task_queue: "orderFulfillment"
    nexus endpoint NotificationsEndpoint
        options:
            task_queue: "orderFulfillment"
```

## Async Patterns

```twf
workflow OrderPipeline(order: Order) -> (PipelineResult):
    state:
        condition paymentConfirmed

    # Update handler — validates and confirms payment
    update ConfirmPayment(txn: Transaction) -> (ConfirmResult):
        activity ValidateTxn(txn) -> validation
        if (validation.ok):
            set paymentConfirmed
            return ConfirmResult{accepted: true}
        else:
            return ConfirmResult{accepted: false, reason: validation.error}

    # Promise — start async, await later
    promise inventory <- activity CheckInventory(order)

    # Detach — fire-and-forget, no result observation
    detach workflow AuditLog(order)

    # Await condition — blocks until handler sets it
    await paymentConfirmed

    # Await promise — get the result started earlier
    await inventory -> stock

    # Switch — multi-branch dispatch
    switch (stock.level):
        case "high":
            activity ShipStandard(order) -> shipment
        case "low":
            activity ShipFromWarehouse(order, stock.warehouseId) -> shipment
        case "none":
            close fail(PipelineResult{error: "out of stock"})

    close complete(PipelineResult{shipment})

# Heartbeat — report progress from long-running activity
activity ProcessLargeDataset(datasetId: string) -> (ProcessResult):
    # Call heartbeat() periodically to report progress
    # If worker dies, Temporal detects missed heartbeat and retries on another worker
    heartbeat()
    return process(datasetId)

# Call-level options — override timeout, routing, retry for a specific call
workflow DeployService(config: DeployConfig) -> (DeployResult):
    activity BuildArtifact(config) -> artifact
    activity Deploy(artifact) -> result
        options:
            start_to_close_timeout: 30m
            heartbeat_timeout: 5m
            retry_policy:
                maximum_attempts: 3
    close complete(DeployResult{result})

# Supporting definitions
activity CheckInventory(order: Order) -> (InventoryStatus):
    return inventory.check(order)

activity ValidateTxn(txn: Transaction) -> (TxnValidation):
    return payments.validate(txn)

workflow AuditLog(order: Order):
    activity RecordAudit(order)
    close complete

activity RecordAudit(order: Order):
    audit.record(order)

activity ShipStandard(order: Order) -> (Shipment):
    return shipping.standard(order)

activity ShipFromWarehouse(order: Order, warehouseId: string) -> (Shipment):
    return shipping.fromWarehouse(order, warehouseId)

activity BuildArtifact(config: DeployConfig) -> (Artifact):
    return build(config)

activity Deploy(artifact: Artifact) -> (DeployStatus):
    return deploy(artifact)

worker pipelineWorker:
    workflow OrderPipeline
    workflow AuditLog
    workflow DeployService
    activity CheckInventory
    activity ValidateTxn
    activity RecordAudit
    activity ShipStandard
    activity ShipFromWarehouse
    activity ProcessLargeDataset
    activity BuildArtifact
    activity Deploy

namespace default:
    worker pipelineWorker
        options:
            task_queue: "pipeline"
```
