# Child Workflows

> **Example:** [`child-workflows.twf`](./child-workflows.twf)

Nested workflow execution for decomposition, reusability, and independent failure boundaries.

## When to Use Child Workflows

| Use Child Workflow | Use Activity Instead |
|--------------------|---------------------|
| Multi-step operation with own retry logic | Single atomic operation |
| Reusable across multiple parent workflows | One-off operation |
| Need independent timeout/retry policies | Same policies as parent |
| Operation is complex enough to warrant own tests | Simple request-response |
| Very long operation (separate history) | Completes quickly |
| Different failure semantics needed | Parent handles all failures |

**Rule of thumb:** If the operation has its own "shape" that you'd want to test independently, it's a child workflow.

---

## Basic Pattern

```twf
workflow ParentWorkflow(input: Input) -> (Result):
    # Simple child workflow call
    workflow ChildWorkflow(input.data) -> childResult
    
    # Child workflow with options
    workflow ChildWorkflow(input.data) -> childResult
        options:
            workflow_execution_timeout: 1h
            retry_policy:
                maximum_attempts: 3
    
    close complete(Result{childResult})

workflow ChildWorkflow(data: Data) -> (ChildResult):
    activity Step1(data)
    activity Step2(data)
    close complete(ChildResult{success: true})
```

---

## Execution Modes: `workflow`, `promise`, `detach`

Child workflows support three execution modes:

| Mode | Syntax | Behavior | Result |
|------|--------|----------|--------|
| **Synchronous** | `workflow Name(args) -> result` | Parent blocks until child completes | Child result bound to variable |
| **Async (promise)** | `promise p <- workflow Name(args)` | Parent continues immediately | Promise for later awaiting |
| **Fire-and-forget (detach)** | `detach workflow Name(args)` | Parent continues, never waits | No result binding |

### Synchronous (Default)

Parent blocks until child workflow completes and receives the result:

```twf
workflow Parent(input: Input) -> (Result):
    workflow ChildWorkflow(input.data) -> childResult
    close complete(Result{childResult})
```

### Async with `promise`

Start a child workflow and get a promise. Continue with other work, then await the promise later:

```twf
workflow Parent(input: Input) -> (Result):
    # Start child without blocking
    promise handle <- workflow SlowChild(input.data)

    # Do other work in parallel
    activity QuickTask(input)

    # Await child result when needed
    await handle -> childResult

    close complete(Result{childResult})
```

### Fire-and-forget with `detach`

Start a child workflow that runs independently. The parent never waits for it and cannot receive its result. Detached children survive parent completion, cancellation, or failure.

```twf
workflow Parent(input: Input) -> (Result):
    activity ProcessOrder(input) -> result

    # Fire-and-forget notification - runs independently
    detach workflow SendNotification(input.customer, result)

    close complete(Result{result})
```

> **Note:** `detach` implies `ABANDON` parent close policy. The detached child continues even if the parent is cancelled or terminated.

### `promise` and `detach` with Nexus

Both modifiers also work with nexus calls for cross-namespace workflows:

```twf
nexus service PaymentsService:
    async ProcessPayment workflow ProcessPayment

nexus service NotificationsService:
    async SendEmail workflow SendEmail

workflow Parent(input: Input) -> (Result):
    # Async nexus call
    promise handle <- nexus PaymentsEndpoint PaymentsService.ProcessPayment(input.payment)

    # Fire-and-forget nexus call
    detach nexus NotificationsEndpoint NotificationsService.SendEmail(input.customer)

    # Await the async promise
    await handle -> paymentResult

    close complete(Result{paymentResult})
```

---

## Parent-Child Lifecycle

### Execution Relationship

```text
Parent starts
├─ Child starts
│  ├─ Child activities execute
│  └─ Child completes/fails
└─ Parent continues with child result
```

### Cancellation Propagation

By default, cancelling a parent cancels its children. Use `detach` for fire-and-forget children that survive parent close.

```twf
workflow Parent(input: Input) -> (Result):
    # Default: child cancelled if parent cancelled
    workflow ChildWorkflow(input)
    
    # Detached: child continues even if parent cancelled
    detach workflow ChildWorkflow(input)
```

### Parent Close Policies

| Policy | Behavior |
|--------|----------|
| `TERMINATE` | Child terminated when parent closes (default) |
| `ABANDON` | Child continues running independently |
| `REQUEST_CANCEL` | Cancellation requested but child can handle gracefully |

---

## Error Handling

### Child Failure Modes

> Note: Error handling is SDK-specific. The workflow will fail if a child workflow fails (after retries). Use retry policies and timeouts to control failure behavior.

```twf
workflow Parent(input: Input) -> (Result):
    # Child workflow call -- if it fails, the parent workflow fails
    workflow ChildWorkflow(input) -> result
    close complete(Result{success: true, data: result})
```

### Retry Policies for Children

```twf
workflow Parent(input: Input) -> (Result):
    # Child with custom retry policy
    workflow ProcessOrder(input.order)
        options:
            retry_policy:
                initial_interval: 1s
                backoff_coefficient: 2.0
                maximum_interval: 60s
                maximum_attempts: 5
```

---

## Workflow ID Design

Child workflow IDs determine uniqueness and idempotency. Workflow ID assignment and reuse policies are SDK-level concerns and are not expressed in the TWF DSL. When implementing your workflows, use the SDK to set `workflow_id` and `workflow_id_reuse_policy` on child workflow stubs.

### Common Patterns (SDK-level)

- **Derived from parent + child identifier:** `"{parent_id}-child-{item.id}"`
- **Deterministic from business entity:** `"order-{order.id}"`
- **With attempt counter for retries:** `"op-{data.id}-attempt-{attemptCount}"`

### Idempotency via Workflow ID

Using a deterministic workflow ID ensures idempotency: if a child with the same ID already exists and completed, the SDK returns the cached result. Configure this through workflow ID reuse policies in your SDK code.

### Workflow ID Reuse Policies (SDK-level)

| Policy | Behavior |
|--------|----------|
| `ALLOW_DUPLICATE` | Start new execution even if ID exists |
| `ALLOW_DUPLICATE_FAILED_ONLY` | New execution only if previous failed |
| `REJECT_DUPLICATE` | Error if workflow ID already exists |
| `TERMINATE_IF_RUNNING` | Terminate existing, start new |

---

## Decomposition Patterns

### Sequential Sub-Operations

```twf
workflow DeployApplication(app: App) -> (DeployResult):
    # Each child has own failure boundary
    workflow DeployDatabase(app.database)
    workflow DeployBackend(app.backend)
    workflow DeployFrontend(app.frontend)
    workflow ConfigureRouting(app)
    
    close complete(DeployResult{status: "deployed"})
```

### Parallel Children

```twf
workflow ProcessBatch(items: []Item) -> (BatchResult):
    # Start all children in parallel
    await all:
        for (item in items):
            workflow ProcessItem(item) -> result
    
    close complete(BatchResult{})
```

### Conditional Children

```twf
workflow Onboarding(user: User) -> (OnboardingResult):
    workflow CreateAccount(user)
    
    if user.type == "enterprise":
        workflow EnterpriseSetup(user)
    else:
        workflow StandardSetup(user)
    
    workflow SendWelcomeEmail(user)
    close complete(OnboardingResult{success: true})
```

### Hierarchical Decomposition

```text
workflow DeployShard(shard: Shard) -> ShardResult:
    # Level 1: Shard deployment
    ├─ workflow DeployOrganization(shard.org1)
    │      # Level 2: Org deployment
    │      ├─ workflow DeployPeer(org1.peer1)
    │      │      # Level 3: Component deployment
    │      │      ├─ activity CreateCertificates(peer1)
    │      │      ├─ activity DeployContainer(peer1)
    │      │      └─ activity ConfigureNetwork(peer1)
    │      └─ workflow DeployPeer(org1.peer2)
    └─ workflow DeployOrganization(shard.org2)
```

---

## Testing Child Workflows

### Unit Testing Parent

Mock child workflows to test parent orchestration logic:

> Note: Test examples use conceptual test framework pseudo-code, not TWF notation.

```pseudo
test "parent calls children in correct order":
    mock ChildA -> {result: "a"}
    mock ChildB -> {result: "b"}
    
    result = execute Parent(input)
    
    assert ChildA called with (input.dataA)
    assert ChildB called with (input.dataB)
    assert ChildB called after ChildA
    assert result == {a: "a", b: "b"}
```

### Integration Testing

Test parent and children together:

```pseudo
test "full workflow execution":
    # Real child workflow implementations
    result = execute Parent(input)
    
    assert result.status == "success"
    assert external_system.has(expected_state)
```

---

## Anti-Patterns

### Too Many Children

```twf
# BAD: Every operation is a child workflow
workflow Parent(data: Data):
    workflow Step1(data)      # Just calls one activity
    workflow Step2(data)      # Just calls one activity  
    workflow Step3(data)      # Just calls one activity

# GOOD: Children for meaningful decomposition
workflow Parent(data: Data):
    activity Step1(data)
    activity Step2(data)
    workflow ComplexOperation(data)  # Has multiple steps, own retry logic
```

### Ignoring Child Failures

> Note: Error handling is SDK-specific. This example uses conceptual pseudo-code.

```twf
# BAD: Silent failure (SDK-level: catching and ignoring child errors)
# GOOD: Track failures and alert
workflow Parent(items: []Item):
    for (item in items):
        workflow ProcessItem(item)
    # SDK-level: collect results, alert on partial failures
    activity AlertOnPartialFailure(items)
```

### Hardcoded Workflow IDs (SDK-level)

When setting workflow IDs in your SDK code, always derive them from business entities to avoid collisions. Using a static workflow ID like `"process-order"` for all orders causes failures; use `"process-order-{order.id}"` instead.
