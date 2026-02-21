# TWF Notation Examples

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
