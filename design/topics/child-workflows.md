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
workflow ParentWorkflow(input: Input) -> Result:
    # Simple child workflow call
    workflow ChildWorkflow(input.data) -> childResult
    
    # Child workflow with options
    options(workflow_id: "child-{input.id}", timeout: 1h, retry_policy: {max_attempts: 3})
    workflow ChildWorkflow(input.data) -> childResult
    
    close Result{childResult}

workflow ChildWorkflow(data: Data) -> ChildResult:
    activity Step1(data)
    activity Step2(data)
    close ChildResult{success: true}
```

---

## Execution Modes: `workflow`, `spawn`, `detach`

Child workflows support three execution modes:

| Mode | Syntax | Behavior | Result |
|------|--------|----------|--------|
| **Synchronous** | `workflow Name(args) -> result` | Parent blocks until child completes | Child result bound to variable |
| **Async (spawn)** | `spawn workflow Name(args) -> handle` | Parent continues immediately | Handle for later awaiting |
| **Fire-and-forget (detach)** | `detach workflow Name(args)` | Parent continues, never waits | No result binding |

### Synchronous (Default)

Parent blocks until child workflow completes and receives the result:

```twf
workflow Parent(input: Input) -> Result:
    workflow ChildWorkflow(input.data) -> childResult
    close Result{childResult}
```

### Async with `spawn`

Start a child workflow and get a handle. Continue with other work, then await the handle later:

```twf
workflow Parent(input: Input) -> Result:
    # Start child without blocking
    spawn workflow SlowChild(input.data) -> handle
    
    # Do other work in parallel
    activity QuickTask(input)
    
    # Await child result when needed
    await one:
        handle -> childResult:
    
    close Result{childResult}
```

### Fire-and-forget with `detach`

Start a child workflow that runs independently. The parent never waits for it and cannot receive its result. Detached children survive parent completion, cancellation, or failure.

```twf
workflow Parent(input: Input) -> Result:
    activity ProcessOrder(input) -> result
    
    # Fire-and-forget notification - runs independently
    detach workflow SendNotification(input.customer, result)
    
    close Result{result}
```

> **Note:** `detach` implies `ABANDON` parent close policy. The detached child continues even if the parent is cancelled or terminated.

### `spawn` and `detach` with Nexus

Both modifiers also work with nexus calls for cross-namespace workflows:

```twf
workflow Parent(input: Input) -> Result:
    # Async nexus call
    spawn nexus "payments" workflow ProcessPayment(input.payment) -> handle
    
    # Fire-and-forget nexus call
    detach nexus "notifications" workflow SendEmail(input.customer)
    
    # Await the async handle
    await one:
        handle -> paymentResult:
    
    close Result{paymentResult}
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
workflow Parent(input: Input) -> Result:
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
workflow Parent(input: Input) -> Result:
    # Child workflow call -- if it fails, the parent workflow fails
    workflow ChildWorkflow(input) -> result
    close Result{success: true, data: result}
```

### Retry Policies for Children

```twf
workflow Parent(input: Input) -> Result:
    # Child with custom retry policy
    options(retry_policy: {initial_interval: 1s, backoff_coefficient: 2.0, max_interval: 60s, max_attempts: 5})
    workflow ProcessOrder(input.order)
```

---

## Workflow ID Design

Child workflow IDs determine uniqueness and idempotency.

### Patterns

```twf
workflow Parent(input: Input) -> Result:
    # Pattern 1: Derived from parent + child identifier
    options(workflow_id: "{workflow.id}-child-{item.id}")
    workflow ChildWorkflow(item)
    
    # Pattern 2: Deterministic from business entity
    options(workflow_id: "order-{order.id}")
    workflow ProcessOrder(order)
    
    # Pattern 3: With attempt counter for retries
    options(workflow_id: "op-{data.id}-attempt-{attemptCount}")
    workflow RetryableOperation(data)
```

### Idempotency via Workflow ID

```twf
workflow Parent(items: []Item) -> Result:
    # Same workflow ID = same workflow execution
    # If child already exists and completed, returns cached result
    for (item in items):
        options(workflow_id: "process-item-{item.id}", workflow_id_reuse_policy: ALLOW_DUPLICATE_FAILED_ONLY)
        workflow ProcessItem(item)
```

### Workflow ID Reuse Policies

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
workflow DeployApplication(app: App) -> DeployResult:
    # Each child has own failure boundary
    workflow DeployDatabase(app.database)
    workflow DeployBackend(app.backend)
    workflow DeployFrontend(app.frontend)
    workflow ConfigureRouting(app)
    
    close DeployResult{status: "deployed"}
```

### Parallel Children

```twf
workflow ProcessBatch(items: []Item) -> BatchResult:
    # Start all children in parallel
    await all:
        for (item in items):
            workflow ProcessItem(item) -> result
    
    close BatchResult{}
```

### Conditional Children

```twf
workflow Onboarding(user: User) -> OnboardingResult:
    workflow CreateAccount(user)
    
    if user.type == "enterprise":
        workflow EnterpriseSetup(user)
    else:
        workflow StandardSetup(user)
    
    workflow SendWelcomeEmail(user)
    close OnboardingResult{success: true}
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

### Hardcoded Workflow IDs

```twf
# BAD: Collision risk
workflow Parent(order: Order):
    options(workflow_id: "process-order")  # Same ID for all orders!
    workflow ProcessOrder(order)

# GOOD: Unique per entity
workflow Parent(order: Order):
    options(workflow_id: "process-order-{order.id}")
    workflow ProcessOrder(order)
```
