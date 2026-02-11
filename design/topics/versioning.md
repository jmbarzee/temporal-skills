# Workflow Versioning and Evolution

> **Example:** [`versioning.twf`](./versioning.twf)

Safe strategies for evolving workflows without breaking running executions.

## The Versioning Challenge

Running workflows may execute for hours, days, or months. When you deploy new code:

```text
Problem:
1. Workflow V1 starts, runs step A, B
2. You deploy V2 (changes step B to B')
3. Worker restarts, replays V1 workflow
4. Replay expects B but code has B'
5. Non-determinism error!
```

**Solution:** Version-aware code that handles both old and new execution paths.

---

## Temporal's Versioning Approaches

| Approach | Use Case | Complexity |
|----------|----------|------------|
| **Patching API** | Incremental changes to existing workflows | Low |
| **Worker Versioning** | Major workflow changes, complete rewrites | Medium |
| **Workflow Type Versioning** | Breaking changes, parallel versions | Higher |

---

## Patching API

Add conditional logic to handle old vs new code paths during replay.

### Basic Pattern

```twf
workflow OrderWorkflow(order: Order) -> OrderResult:
    activity ValidateOrder(order)
    
    # Version gate: new code only runs for new executions
    if patched("add-fraud-check"):
        activity FraudCheck(order)  # New step, only for new workflows
    
    activity ProcessPayment(order)
    close OrderResult{status: "complete"}
```

### How Patching Works

```text
New Execution:
1. patched("add-fraud-check") → true (marks in history)
2. FraudCheck runs
3. History: [Validate, Patch:add-fraud-check, FraudCheck, Payment]

Replay of Old Execution (started before patch):
1. History has no patch marker
2. patched("add-fraud-check") → false
3. FraudCheck skipped
4. Replay matches original history

Replay of New Execution (started after patch):
1. History has patch marker
2. patched("add-fraud-check") → true
3. FraudCheck runs
4. Replay matches history
```

### Patching Examples

**Adding a Step:**
```twf
workflow Process(data: Data) -> Result:
    activity Step1(data)
    
    if patched("v2-add-validation"):
        activity NewValidation(data)  # Added in V2
    
    activity Step2(data)
    close Result{}
```

**Removing a Step:**
```twf
workflow Process(data: Data) -> Result:
    activity Step1(data)
    
    if not patched("v3-remove-legacy"):
        activity LegacyStep(data)  # Removed in V3, but runs for old workflows
    
    activity Step2(data)
    close Result{}
```

**Changing a Step:**
```twf
workflow Process(data: Data) -> Result:
    activity Step1(data)
    
    if patched("v4-improved-processing"):
        activity ImprovedProcessing(data)
    else:
        activity OldProcessing(data)
    
    activity Step3(data)
    close Result{}
```

### Deprecating Patches

After all old workflows complete, remove patch:

> Note: Patch lifecycle management uses SDK-specific APIs. The concept is shown as pseudo-code.

```pseudo
# Phase 1: Add patch (both paths exist)
if patched("add-feature"):
    activity NewFeature()

# Phase 2: After all old workflows done, simplify
# (Run deprecate_patch to verify no old workflows)
if deprecated_patch("add-feature"):
    pass  # Old path, will error if any old workflows still running
activity NewFeature()

# Phase 3: Remove patch code entirely
activity NewFeature()
```

---

## Worker Versioning (Build IDs)

Route workflows to workers running compatible code versions.

### Concept

```text
┌─────────────────────────────────────────────────────┐
│                   Task Queue                         │
├─────────────────────────────────────────────────────┤
│  Build ID: 1.0  │  Build ID: 2.0  │  Build ID: 3.0 │
│    (default)    │   (compatible)  │    (latest)    │
└────────┬────────┴────────┬────────┴────────┬───────┘
         │                 │                  │
     ┌───▼───┐        ┌────▼────┐       ┌────▼────┐
     │Worker │        │ Worker  │       │ Worker  │
     │  1.0  │        │   2.0   │       │   3.0   │
     └───────┘        └─────────┘       └─────────┘
```

### Configuration

```bash
# Register build ID with task queue
temporal task-queue update-build-ids add-new-default \
    --task-queue main-queue \
    --build-id "v2.0"
```

> Note: Worker configuration is SDK-level code.

```pseudo
# Worker identifies its build ID
worker = Worker(
    task_queue: "main-queue",
    build_id: "v2.0",
    workflows: [OrderWorkflow],
    activities: [ValidateOrder, ProcessPayment]
)
```

### Version Sets

```bash
# Create version set: v1.0 and v1.1 are compatible
temporal task-queue update-build-ids add-new-compatible \
    --task-queue main-queue \
    --build-id "v1.1" \
    --existing-compatible-build-id "v1.0"

# New version set: v2.0 is NOT compatible with v1.x
temporal task-queue update-build-ids add-new-default \
    --task-queue main-queue \
    --build-id "v2.0"
```

### Routing Behavior

| Workflow State | Routed To |
|----------------|-----------|
| New workflow | Latest default build ID |
| Running workflow | Same build ID (or compatible) |
| Workflow started on v1.0 | v1.0 or v1.1 worker |

---

## Workflow Type Versioning

Create a new workflow type for breaking changes.

### Pattern

```twf
# Version 1
workflow OrderWorkflowV1(order: OrderV1) -> ResultV1:
    # Original implementation
    ...

# Version 2 (breaking changes)
workflow OrderWorkflowV2(order: OrderV2) -> ResultV2:
    # New implementation with different structure
    ...
```

### Migration Strategy

> Note: API routing logic is application-level code, not TWF notation.

```pseudo
# API layer routes to appropriate version
function startOrderWorkflow(order):
    if order.version == 1:
        return client.start(OrderWorkflowV1, convertToV1(order))
    else:
        return client.start(OrderWorkflowV2, convertToV2(order))
```

### When to Use Workflow Type Versioning

| Scenario | Approach |
|----------|----------|
| Adding optional step | Patching |
| Changing activity order | Patching |
| Complete workflow rewrite | New workflow type |
| Input/output schema breaking change | New workflow type |
| Different business logic | New workflow type |

---

## Versioning Best Practices

### 1. Plan for Evolution

```pseudo
# Good: Named constants for versions (SDK-level code)
PATCH_ADD_FRAUD_CHECK = "2024-01-add-fraud-check"
PATCH_IMPROVE_VALIDATION = "2024-02-improve-validation"

workflow Process(data: Data):
    if patched(PATCH_ADD_FRAUD_CHECK):
        ...
```

### 2. Test Both Paths

```pseudo
test "workflow handles both old and new path":
    # Test new execution path
    env = TestEnvironment()
    result = env.execute(Workflow, input)
    assert result.includesFraudCheck
    
    # Test replay of old execution
    old_history = load("workflow_v1.history")
    replay_result = env.replay(Workflow, old_history)
    assert replay_result.success  # No non-determinism
```

### 3. Document Versions

```text
# Workflow: OrderWorkflow
# 
# Version History:
# - 2024-01: Added fraud check (patch: add-fraud-check)
# - 2024-02: Improved validation (patch: improve-validation)
# - 2024-03: Deprecated old validation (patch: remove-legacy-validation)
#
# Active patches: add-fraud-check, improve-validation
# Deprecated patches: remove-legacy-validation (safe to remove after 2024-04)
```

### 4. Monitor Old Workflows

```bash
# Query for workflows started before patch
temporal workflow list \
    --query "StartTime < '2024-01-15' AND ExecutionStatus = 'Running'"
```

---

## Common Versioning Scenarios

### Adding Activity

```twf
workflow Process(data: Data):
    activity Existing1(data)
    
    if patched("add-new-activity"):
        activity NewActivity(data)  # Safe to add
    
    activity Existing2(data)
```

### Removing Activity

```twf
workflow Process(data: Data):
    activity Existing1(data)
    
    if not patched("remove-deprecated"):
        activity DeprecatedActivity(data)  # Removed for new, kept for old
    
    activity Existing2(data)
```

### Reordering Activities

```twf
# Original order: A, B, C
# New order: A, C, B

workflow Process(data: Data):
    activity A(data)
    
    if patched("reorder-bc"):
        activity C(data)
        activity B(data)
    else:
        activity B(data)
        activity C(data)
```

### Changing Activity Parameters

```twf
workflow Process(data: Data):
    if patched("new-activity-params"):
        activity Enhanced(data, extraParam: true)
    else:
        activity Enhanced(data)  # Old signature
```

---

## Anti-Patterns

### Unguarded Changes

```twf
# BAD: Breaking change without version guard
workflow Process(data: Data):
    activity Step1(data)
    # Removed Step2 without patch - breaks replay!
    activity Step3(data)

# GOOD: Version-guarded removal
workflow Process(data: Data):
    activity Step1(data)
    if not patched("remove-step2"):
        activity Step2(data)
    activity Step3(data)
```

### Too Many Active Patches

```twf
# BAD: Accumulated complexity
workflow Process(data: Data):
    if patched("v1"):
        if patched("v2"):
            if patched("v3"):
                ...

# GOOD: Consolidate when safe, or use worker versioning
```

### Forgetting to Deprecate

```pseudo
# BAD: Old patch code lives forever
if patched("feature-from-2020"):  # All workflows with this are done!
    ...

# GOOD: Clean up after old workflows complete
# 1. Verify no running workflows need old path
# 2. Replace with deprecated_patch
# 3. Remove patch code after verification
```

---

## Version Migration Checklist

- [ ] Identify all changes from current version
- [ ] Classify each change (additive, removal, modification)
- [ ] Add appropriate patches for each change
- [ ] Write replay tests against old histories
- [ ] Deploy with both code paths active
- [ ] Monitor for non-determinism errors
- [ ] Wait for old workflows to complete
- [ ] Remove deprecated patch code
- [ ] Update documentation
