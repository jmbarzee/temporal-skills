# Async Handles and Conditions (Future Work)

## Overview

This document outlines two major async primitives that appear across Temporal SDKs but are not yet fully represented in the TWF DSL. These are deferred for future design work.

## 1. Async Handles (Futures/Promises)

### SDK Pattern

All Temporal SDKs support a **start-now, wait-later** pattern for async operations:

```java
// Java: Start operation, get handle
Promise<String> handle = Async.function(activities::process, input);
// ... do other work ...
String result = handle.get(); // Wait later
```

```go
// Go: Start operation, get Future
fut := workflow.ExecuteActivity(ctx, MyActivity, input)
// ... do other work ...
var result string
fut.Get(ctx, &result) // Wait later
```

```typescript
// TypeScript: Start operation, get Promise
const promise = activityCall(input);
// ... do other work ...
const result = await promise; // Wait later
```

### Key Properties

1. **Explicit handles**: Operations return a value representing "future result"
2. **Composable**: Handles can be stored, passed around, collected in arrays
3. **Flexible waiting**: Can wait immediately or later, individually or in groups
4. **First-class values**: Handles are values like any other

### Current DSL Gap

The DSL currently uses statement-based async:
```twf
# This starts AND waits immediately
activity ProcessItem(item) -> result
```

No way to:
- Start operation and continue
- Store handle for later
- Manually compose multiple handles

### Future Design Questions

1. Should TWF have explicit handle syntax?
   ```twf
   handle = spawn activity ProcessItem(item)
   # ... other work ...
   result = await handle
   ```

2. Or remain statement-based with implicit handle management?
   ```twf
   await all:
       activity ProcessItem(item1) -> result1
       activity ProcessItem(item2) -> result2
   ```

3. Trade-offs:
   - **Explicit handles**: More flexible, more verbose, closer to SDK reality
   - **Implicit handles**: Cleaner syntax, less control, more opinionated
   - **Hybrid**: Statement-based by default, explicit handles when needed?

## 2. Condition Waiting

### SDK Pattern

All SDKs support waiting on arbitrary boolean predicates:

```typescript
// TypeScript
await condition(() => approvedForRelease);
await condition(() => balance > threshold && !suspended);
```

```python
# Python
await workflow.wait_condition(lambda: self.approved_for_release)
await workflow.wait_condition(lambda: self.balance > self.threshold)
```

```go
// Go (via Selector + custom logic)
selector.AddFuture(
    workflow.NewTimer(ctx, pollInterval),
    func(f workflow.Future) {
        if balanceExceedsThreshold() {
            // condition met
        }
    },
)
```

### Key Properties

1. **Arbitrary expressions**: Not just simple boolean variables
2. **Reactive**: Condition checked whenever workflow state changes
3. **Signal-driven**: Typically used with signals that modify state
4. **Composable**: Can combine multiple conditions

### Current DSL Approach

The `watch` statement (to be removed) provided limited condition waiting:
```twf
signal Approved():
    approved = true

await one:
    watch (approved):  # Only truthiness checks
        close
```

Limitations:
- Only simple variable truthiness
- No expressions like `watch (balance > 100)`
- Only works in `await one` blocks
- Indirect (via state variables)

### Future Design Questions

1. Should TWF support general condition expressions?
   ```twf
   await condition (balance > threshold and not suspended)
   ```

2. Should conditions be first-class cases in await one?
   ```twf
   await one:
       condition (balance > threshold):
           close "threshold reached"
       timer (1h):
           close failed "timeout"
   ```

3. How do conditions interact with signal handlers?
   ```twf
   signal Deposit(amount: decimal):
       balance = balance + amount

   # Condition implicitly re-evaluated after signal?
   await condition (balance >= 1000)
   ```

4. Syntax considerations:
   - `await condition (expr)` - explicit condition keyword
   - `await (expr)` - implicit condition from expression
   - `watch (expr)` - keep watch keyword but allow expressions
   - Integration with handles: `await handle and (balance > 100)`?

## Relationship Between Handles and Conditions

These two primitives interact in SDK patterns:

```typescript
// TypeScript: Condition waiting for async operation result
const processPromise = longRunningActivity();
await condition(() => userCancelled || processPromise.isResolved());
```

```go
// Go: Condition within selector
selector.AddFuture(activityFuture, func(f workflow.Future) {
    if satisfiesCondition(result) {
        // proceed
    }
})
```

A complete async model may need both:
- **Handles** for explicit async operation management
- **Conditions** for waiting on derived state/predicates

## Decision Deferred

Both features are complex and interact with the core DSL design. We defer concrete proposals until:

1. The basic await one/all semantics are finalized
2. Direct signal/update awaiting is working
3. We have real-world use cases that demonstrate the need

## Notes

- Handles are common in imperative SDKs (Java, Go, C#)
- Conditions appear across all SDKs but with varying syntax
- TypeScript/JavaScript developers expect Promise-based patterns
- Go developers expect Future + Selector patterns
- Python developers expect async/await with conditions

The DSL should abstract these patterns into a language-agnostic form that feels natural to workflow authors while mapping cleanly to SDK implementations.
