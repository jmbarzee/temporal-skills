# Testing Temporal Workflows

> **Example:** [`testing.twf`](./testing.twf)

Strategies for testing workflows, activities, and ensuring determinism.

## Testing Pyramid for Temporal

```text
         ┌─────────────────┐
         │   End-to-End    │  Few, slow, high confidence
         │    Tests        │
         ├─────────────────┤
         │  Integration    │  Test with real services
         │    Tests        │
         ├─────────────────┤
         │   Workflow      │  Mock activities
         │  Unit Tests     │
         ├─────────────────┤
         │   Activity      │  Mock clients/SDKs
         │  Unit Tests     │
         └─────────────────┘  Many, fast, isolated
```

---

## Activity Unit Testing

Test activities in isolation by mocking external dependencies.

### Pattern

> Note: Activity body implementations and test assertions are SDK-level code, not TWF notation.

```pseudo
# Activity implementation
activity ProcessPayment(payment: Payment) -> PaymentResult:
    # Validate
    if payment.amount <= 0:
        raise InvalidPaymentError("Amount must be positive")
    
    # Call external service
    result = paymentClient.charge(payment)
    
    return PaymentResult{
        transactionId: result.id,
        status: result.status
    }

# Test
test "ProcessPayment succeeds with valid payment":
    mockPaymentClient = Mock()
    mockPaymentClient.charge.returns({id: "txn-123", status: "success"})
    
    result = ProcessPayment(Payment{amount: 100, cardId: "card-1"})
    
    assert result.transactionId == "txn-123"
    assert result.status == "success"
    assert mockPaymentClient.charge.called_with(Payment{amount: 100})

test "ProcessPayment fails with invalid amount":
    expect_error InvalidPaymentError:
        ProcessPayment(Payment{amount: -50})
```

### What to Test in Activities
- Input validation
- Error handling for external failures
- Correct client/SDK usage
- Return value construction

---

## Workflow Unit Testing

Test workflow logic by mocking activities. Use Temporal's test framework.

### Pattern

```twf
workflow OrderWorkflow(order: Order) -> OrderResult:
    activity ValidateOrder(order) -> validated
    if not validated.success:
        close failed OrderResult{status: "invalid"}
    
    activity ProcessPayment(order.payment) -> payment
    activity ShipOrder(order)
    
    close OrderResult{status: "completed", paymentId: payment.id}
```

```pseudo
# Test
test "OrderWorkflow completes successfully":
    env = TestWorkflowEnvironment()
    
    # Mock activities
    env.mock_activity(ValidateOrder, returns: {success: true})
    env.mock_activity(ProcessPayment, returns: {id: "pay-123"})
    env.mock_activity(ShipOrder, returns: {})
    
    # Execute workflow
    result = env.execute_workflow(OrderWorkflow, Order{id: "order-1"})
    
    # Assert result
    assert result.status == "completed"
    assert result.paymentId == "pay-123"
    
    # Assert activity calls
    assert env.activity_called(ValidateOrder, with: Order{id: "order-1"})
    assert env.activity_called(ProcessPayment)
    assert env.activity_called(ShipOrder)

test "OrderWorkflow returns invalid for failed validation":
    env = TestWorkflowEnvironment()
    
    env.mock_activity(ValidateOrder, returns: {success: false, error: "bad order"})
    
    result = env.execute_workflow(OrderWorkflow, Order{id: "order-1"})
    
    assert result.status == "invalid"
    assert not env.activity_called(ProcessPayment)  # Not called
    assert not env.activity_called(ShipOrder)       # Not called
```

### What to Test in Workflows
- Activity call ordering
- Conditional logic (correct branch taken)
- Error handling (activity failures)
- Signal/query handlers
- Timeout behavior

---

## Replay Testing (Determinism Verification)

Verify workflows are deterministic by replaying against recorded history.

### Why Replay Testing

```text
Version 1: Workflow runs, generates history
Version 2: Workflow code changes
Replay Test: Run version 2 against version 1 history
Result: PASS (deterministic) or FAIL (non-determinism detected)
```

### Pattern

```pseudo
# Record workflow history
test "record OrderWorkflow history":
    env = TestWorkflowEnvironment()
    # ... execute workflow ...
    history = env.get_workflow_history()
    save_to_file("order_workflow_v1.history", history)

# Replay test
test "OrderWorkflow replays deterministically":
    history = load_from_file("order_workflow_v1.history")
    
    env = TestWorkflowEnvironment()
    result = env.replay_workflow(OrderWorkflow, history)
    
    assert result.replay_successful
    assert not result.non_determinism_errors
```

### Maintaining History Files
- Store history files in version control
- Update when workflow signature changes
- Keep multiple versions for migration testing

---

## Testing Signals and Queries

### Signal Testing

```twf
workflow ApprovalWorkflow(request: Request) -> Decision:
    await one:
        signal Approved:
            close Decision{status: "approved"}
        signal Rejected:
            close Decision{status: "rejected"}
        timer(1h):
            close Decision{status: "timeout"}
```

```pseudo
# Test
test "ApprovalWorkflow handles Approved signal":
    env = TestWorkflowEnvironment()
    
    # Start workflow
    handle = env.start_workflow(ApprovalWorkflow, Request{id: "req-1"})
    
    # Send signal
    env.signal_workflow(handle, Approved, {approver: "alice"})
    
    # Get result
    result = env.get_workflow_result(handle)
    assert result.status == "approved"

test "ApprovalWorkflow handles timeout":
    env = TestWorkflowEnvironment()
    
    handle = env.start_workflow(ApprovalWorkflow, Request{id: "req-1"})
    
    # Skip time forward
    env.skip_time(2h)
    
    result = env.get_workflow_result(handle)
    assert result.status == "timeout"  # or however timeout is handled
```

### Query Testing

```twf
workflow OrderWorkflow(order: Order) -> OrderResult:
    status = "pending"
    
    status = "processing"
    activity ProcessOrder(order)
    
    status = "completed"
    close OrderResult{status: status}

query GetStatus() -> string:
    return status
```

```pseudo
# Test
test "GetStatus query returns current status":
    env = TestWorkflowEnvironment()
    
    # Mock activity to block
    blocker = env.mock_activity(ProcessOrder, blocks: true)
    
    handle = env.start_workflow(OrderWorkflow, Order{id: "1"})
    
    # Query while processing
    status = env.query_workflow(handle, GetStatus)
    assert status == "processing"
    
    # Unblock activity
    blocker.unblock({})
    
    # Query after completion
    env.wait_for_workflow(handle)
    status = env.query_workflow(handle, GetStatus)
    assert status == "completed"
```

---

## Testing Timers

Use time-skipping to test timer behavior without waiting.

```twf
workflow ReminderWorkflow(userId: string) -> void:
    activity SendFirstReminder(userId)

    await timer(24h)

    activity SendSecondReminder(userId)

    await timer(48h)

    activity SendFinalReminder(userId)
```

```pseudo
# Test
test "ReminderWorkflow sends reminders at correct intervals":
    env = TestWorkflowEnvironment()
    
    env.mock_activity(SendFirstReminder, returns: {})
    env.mock_activity(SendSecondReminder, returns: {})
    env.mock_activity(SendFinalReminder, returns: {})
    
    handle = env.start_workflow(ReminderWorkflow, "user-1")
    
    # First reminder sent immediately
    assert env.activity_called(SendFirstReminder)
    assert not env.activity_called(SendSecondReminder)
    
    # Skip 24 hours
    env.skip_time(24h)
    
    assert env.activity_called(SendSecondReminder)
    assert not env.activity_called(SendFinalReminder)
    
    # Skip 48 more hours
    env.skip_time(48h)
    
    assert env.activity_called(SendFinalReminder)
```

---

## Testing Child Workflows

```twf
workflow ParentWorkflow(data: Data) -> Result:
    workflow ChildWorkflow(data.item) -> childResult
    close Result{childData: childResult}

workflow ChildWorkflow(item: Item) -> ChildResult:
    activity ProcessItem(item)
    close ChildResult{processed: true}
```

```pseudo
# Test parent in isolation
test "ParentWorkflow calls child correctly":
    env = TestWorkflowEnvironment()
    
    env.mock_child_workflow(ChildWorkflow, returns: {processed: true})
    
    result = env.execute_workflow(ParentWorkflow, Data{item: Item{id: "1"}})
    
    assert result.childData.processed == true
    assert env.child_workflow_called(ChildWorkflow, with: Item{id: "1"})

# Test parent and child together
test "ParentWorkflow integration with ChildWorkflow":
    env = TestWorkflowEnvironment()
    
    # Mock only activities, let child workflow run
    env.mock_activity(ProcessItem, returns: {})
    
    result = env.execute_workflow(ParentWorkflow, Data{item: Item{id: "1"}})
    
    assert result.childData.processed == true
```

---

## Integration Testing

Test with real Temporal server (local or test cluster).

### Setup

```bash
# Start local Temporal for testing
temporal server start-dev

# Or use testcontainers
test_environment = TemporalTestContainer()
test_environment.start()
```

### Pattern

```pseudo
test "OrderWorkflow end-to-end":
    # Use real Temporal client
    client = TemporalClient(address: "localhost:7233")
    
    # Start real worker
    worker = Worker(
        client: client,
        task_queue: "test-queue",
        workflows: [OrderWorkflow],
        activities: [ValidateOrder, ProcessPayment, ShipOrder]
    )
    worker.start_async()
    
    # Execute workflow
    handle = client.start_workflow(
        OrderWorkflow,
        Order{id: "test-order-1"},
        workflow_id: "test-order-1"
    )
    
    # Wait for result
    result = handle.result(timeout: 30s)
    
    assert result.status == "completed"
    
    # Cleanup
    worker.stop()
```

---

## Testing Best Practices

### Do's

| Practice | Rationale |
|----------|-----------|
| Mock activities in workflow tests | Isolate workflow logic |
| Use replay tests | Catch non-determinism early |
| Test failure paths | Verify error handling |
| Use time-skipping | Fast timer tests |
| Test signal ordering | Validate async behavior |

### Don'ts

| Anti-Pattern | Problem |
|--------------|---------|
| Testing deterministic logic via integration tests | Slow, flaky |
| Skipping replay tests | Non-determinism in production |
| Mocking workflow internals | Brittle tests |
| Real time waits | Slow tests |
| Testing Temporal internals | Not your responsibility |

---

## Test Coverage Checklist

### Activity
- [ ] Valid inputs produce correct output
- [ ] Invalid inputs produce appropriate errors
- [ ] External service failures handled
- [ ] Retryable vs non-retryable errors classified

### Workflow
- [ ] Happy path completes successfully
- [ ] Each conditional branch tested
- [ ] Activity failures handled correctly
- [ ] Signals processed correctly
- [ ] Queries return correct state
- [ ] Timeouts handled appropriately
- [ ] Replay test passes

### Integration
- [ ] End-to-end happy path
- [ ] Cross-service communication
- [ ] Failure recovery
