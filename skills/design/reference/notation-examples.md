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

    # Loops
    for (item in order.items):
        activity ProcessItem(item)

    # Parallel execution
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
