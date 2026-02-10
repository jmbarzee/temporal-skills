# TWF Spec Changes: Unified Await Model

## Summary

Replaced `parallel` and `select` keywords with unified `await all` and `await one` syntax. Removed standalone `await` statement. This creates a clearer, more consistent async model.

## Motivation

The previous model had three overlapping async primitives:
- `await signal X` - wait indefinitely for signal
- `select` with timer/hints - wait for first with timeout
- `parallel` - wait for all operations

The new model unifies these into two clear patterns:
- `await all` - wait for ALL operations to complete
- `await one` - wait for FIRST condition (timer or operations via hints)

## Changes

### 1. Removed `await` Statement

**Old:**
```twf
# Wait indefinitely for signal
await signal PaymentReceived

# Wait for any of multiple
await signal Approved or update UpdateAddress
```

**New:**
```twf
# Wait indefinitely: use await one with no timer
hint signal PaymentReceived
await one:
    # Blocks until signal arrives

# Wait for any of multiple: use multiple hints
hint signal Approved
hint signal Rejected
await one:
    # Blocks until either arrives
```

### 2. Renamed `parallel` → `await all`

**Old:**
```twf
parallel:
    activity ReserveInventory(order) -> inventory
    activity ProcessPayment(order) -> payment
```

**New:**
```twf
await all:
    activity ReserveInventory(order) -> inventory
    activity ProcessPayment(order) -> payment
```

**Semantics:** Wait for ALL contained operations to complete before continuing.

### 3. Renamed `select` → `await one`

**Old:**
```twf
hint signal PaymentReceived
select:
    timer 24h:
        return timeout
```

**New:**
```twf
hint signal PaymentReceived
await one:
    timer 24h:
        return timeout
```

**Semantics:** Wait for FIRST case to complete (race between timer and hints).

### 4. `await all` Can Be an `await one` Case

**Pattern: Parallel with Timeout**
```twf
await one:
    await all:
        activity SlowOp1() -> r1
        activity SlowOp2() -> r2
        # Completes when BOTH finish
    timer 1h:
        # Timeout: parallel operations abandoned
        return Result{status: "timeout"}

# After await one: either all completed or timeout
```

This enables "wait for all operations BUT with a global timeout" pattern.

## Clarification on Hints

**Critical:** Signals, queries, and updates are **ONLY** referenced via `hint` statements. They cannot be called or awaited directly.

```twf
workflow Example(input: Input) -> (Result):
    signal Done():
        status = "complete"

    # CORRECT: Use hint to mark where signal may arrive
    hint signal Done
    await one:
        timer 1h:
            return timeout

    # INCORRECT (removed from language):
    # await signal Done  ❌
```

**Hints can appear anywhere** in the workflow - they mark points where signal/query/update handlers may execute.

## Pattern Comparison

### Wait Indefinitely for Signal
**Old:**
```twf
await signal X
```

**New:**
```twf
hint signal X
await one:
    # No timer = wait indefinitely
```

### Wait with Timeout
**Old:**
```twf
hint signal X
select:
    timer 1h: ...
```

**New:**
```twf
hint signal X
await one:
    timer 1h: ...
```

### Wait for Multiple Signals (Any-Of)
**Old:**
```twf
await signal Approved or signal Rejected
```

**New:**
```twf
hint signal Approved
hint signal Rejected
await one:
    # Whichever arrives first
```

### Concurrent Operations
**Old:**
```twf
parallel:
    activity A() -> a
    activity B() -> b
```

**New:**
```twf
await all:
    activity A() -> a
    activity B() -> b
```

### Best-Effort Pattern (Google Query)
**Pattern:** Start many operations, use whatever completes within timeout.

```twf
# Start 10 searches, get results within 1s
await one:
    await all:
        activity Search1() -> r1
        activity Search2() -> r2
        # ... 10 total
        # Each activity has its own timeout (e.g., 500ms)
    timer 1s:
        # Global timeout: proceed with partial results
        pass

# Use whatever completed
# Note: Individual activity timeouts prevent hanging
# Comment in code: "Best effort - using partial results"
```

## Grammar Changes

### Removed
```
await_stmt ::= 'await' await_target ('or' await_target)* NEWLINE
await_target ::= ('signal' | 'update') IDENT
```

### Added/Changed
```
# Wait for all operations
await_all_block ::= 'await' 'all' ':' NEWLINE
                    INDENT statement* DEDENT

# Wait for first condition
await_one_block ::= 'await' 'one' ':' NEWLINE
                    INDENT await_one_case+ DEDENT

await_one_case ::= timer_case | await_all_case

timer_case ::= 'timer' duration ':' NEWLINE
               INDENT statement* DEDENT

await_all_case ::= 'await' 'all' ':' NEWLINE
                   INDENT statement* DEDENT
```

### Hint Statement Updated
```
hint_stmt ::= 'hint' ('signal' | 'update' | 'query') IDENT NEWLINE
```

Now explicitly includes `query` and clarifies that hints are the ONLY way to reference signals/queries/updates.

## Keywords Updated

**Added:**
- `await` - Base keyword for waiting operations
- `all` - Modifier for `await` (wait for all)
- `one` - Modifier for `await` (wait for first)

**Removed:**
- Standalone `await` statement syntax

**Repurposed:**
- `parallel` → replaced by `await all`
- `select` → replaced by `await one`

## Benefits

1. **Clearer Semantics** - `await all` and `await one` are self-documenting
2. **Unified Model** - All waiting uses `await`, just with different strategies
3. **Composability** - `await all` can be nested in `await one` for timeout patterns
4. **Consistency** - Signals/queries/updates ONLY via hints (no special await syntax)
5. **Familiar** - `await` is widely understood in async programming (JS, C#, Python)

## Migration Guide

### For Spec Readers
- Mentally replace `parallel` with `await all`
- Mentally replace `select` with `await one`
- Remember: signals/queries/updates only via `hint` + `await one`

### For Implementers
- Update lexer: Add `await`, `all`, `one` keywords
- Update parser: Replace `parallel_block` → `await_all_block`, `select_block` → `await_one_block`
- Update AST node names to match
- Remove `await_stmt` parsing

## Examples Updated

All example files (`.twf`) have been updated to use new keywords:
- `parallel:` → `await all:`
- `select:` → `await one:`
- `await signal X` → commented out as REMOVED

## Next Steps

1. ✅ Update LANGUAGE.md spec
2. ✅ Update all examples
3. ⏳ Update Go parser/lexer code
4. ⏳ Update Go AST node types
5. ⏳ Update LSP server
6. ⏳ Test with existing .twf files
