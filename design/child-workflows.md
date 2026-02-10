# Child Workflows

> **Example:** [`examples/child-workflows.twf`](./examples/child-workflows.twf)

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

```
workflow ParentWorkflow(input: Input) -> Result:
    # Simple child workflow call
    childResult = child ChildWorkflow(input.data)
    
    # Child workflow with options
    childResult = child ChildWorkflow(input.data):
        workflow_id: "child-{input.id}"
        timeout: 1h
        retry_policy: {max_attempts: 3}
    
    return Result{childResult}

workflow ChildWorkflow(data: Data) -> ChildResult:
    activity Step1(data)
    activity Step2(data)
    return ChildResult{success: true}
```

---

## Parent-Child Lifecycle

### Execution Relationship

```
Parent starts
├─ Child starts
│  ├─ Child activities execute
│  └─ Child completes/fails
└─ Parent continues with child result
```

### Cancellation Propagation

By default, cancelling a parent cancels its children.

```
workflow Parent(input: Input) -> Result:
    # Default: child cancelled if parent cancelled
    child ChildWorkflow(input)
    
    # Detached: child continues even if parent cancelled
    child ChildWorkflow(input):
        parent_close_policy: ABANDON
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

```
workflow Parent(input: Input) -> Result:
    try:
        result = child ChildWorkflow(input)
        return Result{success: true, data: result}
    catch ChildWorkflowFailure as e:
        # Child workflow failed after all retries
        return Result{success: false, error: e.message}
    catch ChildWorkflowTimeout as e:
        # Child workflow exceeded timeout
        activity Cleanup(input)
        return Result{success: false, error: "timeout"}
    catch ChildWorkflowCancelled as e:
        # Child was cancelled
        return Result{success: false, error: "cancelled"}
```

### Retry Policies for Children

```
workflow Parent(input: Input) -> Result:
    # Child with custom retry policy
    child ProcessOrder(input.order):
        retry_policy:
            initial_interval: 1s
            backoff_coefficient: 2.0
            max_interval: 60s
            max_attempts: 5
```

---

## Workflow ID Design

Child workflow IDs determine uniqueness and idempotency.

### Patterns

```
workflow Parent(input: Input) -> Result:
    # Pattern 1: Derived from parent + child identifier
    child ChildWorkflow(item):
        workflow_id: "{workflow.id}-child-{item.id}"
    
    # Pattern 2: Deterministic from business entity
    child ProcessOrder(order):
        workflow_id: "order-{order.id}"
    
    # Pattern 3: With attempt counter for retries
    child RetryableOperation(data):
        workflow_id: "op-{data.id}-attempt-{attemptCount}"
```

### Idempotency via Workflow ID

```
workflow Parent(items: []Item) -> Result:
    # Same workflow ID = same workflow execution
    # If child already exists and completed, returns cached result
    for item in items:
        child ProcessItem(item):
            workflow_id: "process-item-{item.id}"
            workflow_id_reuse_policy: ALLOW_DUPLICATE_FAILED_ONLY
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

```
workflow DeployApplication(app: App) -> DeployResult:
    # Each child has own failure boundary
    child DeployDatabase(app.database)
    child DeployBackend(app.backend)
    child DeployFrontend(app.frontend)
    child ConfigureRouting(app)
    
    return DeployResult{status: "deployed"}
```

### Parallel Children

```
workflow ProcessBatch(items: []Item) -> BatchResult:
    # Start all children in parallel
    parallel:
        for item in items:
            results[item.id] = child ProcessItem(item)
    
    return BatchResult{results}
```

### Conditional Children

```
workflow Onboarding(user: User) -> OnboardingResult:
    child CreateAccount(user)
    
    if user.type == "enterprise":
        child EnterpriseSetup(user)
    else:
        child StandardSetup(user)
    
    child SendWelcomeEmail(user)
    return OnboardingResult{success: true}
```

### Hierarchical Decomposition

```
workflow DeployShard(shard: Shard) -> ShardResult:
    # Level 1: Shard deployment
    ├─ child DeployOrganization(shard.org1)
    │      # Level 2: Org deployment
    │      ├─ child DeployPeer(org1.peer1)
    │      │      # Level 3: Component deployment
    │      │      ├─ activity CreateCertificates(peer1)
    │      │      ├─ activity DeployContainer(peer1)
    │      │      └─ activity ConfigureNetwork(peer1)
    │      └─ child DeployPeer(org1.peer2)
    └─ child DeployOrganization(shard.org2)
```

---

## Testing Child Workflows

### Unit Testing Parent

Mock child workflows to test parent orchestration logic:

```
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

```
test "full workflow execution":
    # Real child workflow implementations
    result = execute Parent(input)
    
    assert result.status == "success"
    assert external_system.has(expected_state)
```

---

## Anti-Patterns

### Too Many Children

```
# BAD: Every operation is a child workflow
workflow Parent(data: Data):
    child Step1(data)      # Just calls one activity
    child Step2(data)      # Just calls one activity  
    child Step3(data)      # Just calls one activity

# GOOD: Children for meaningful decomposition
workflow Parent(data: Data):
    activity Step1(data)
    activity Step2(data)
    child ComplexOperation(data)  # Has multiple steps, own retry logic
```

### Ignoring Child Failures

```
# BAD: Silent failure
workflow Parent(items: []Item):
    for item in items:
        try:
            child ProcessItem(item)
        catch:
            pass  # Item silently skipped

# GOOD: Explicit failure handling
workflow Parent(items: []Item):
    results = []
    for item in items:
        try:
            results.append(child ProcessItem(item))
        catch as e:
            results.append({item: item, error: e, status: "failed"})
    
    if any_failed(results):
        activity AlertOnPartialFailure(results)
    
    return results
```

### Hardcoded Workflow IDs

```
# BAD: Collision risk
workflow Parent(order: Order):
    child ProcessOrder(order):
        workflow_id: "process-order"  # Same ID for all orders!

# GOOD: Unique per entity
workflow Parent(order: Order):
    child ProcessOrder(order):
        workflow_id: "process-order-{order.id}"
```
