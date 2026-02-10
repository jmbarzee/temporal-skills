# TWF DSL Async Changes - Implementation Guide

## Overview

The TWF (Temporal Workflow Format) DSL has been updated to provide a more direct and SDK-aligned async model. This document summarizes the changes for implementation teams working on features, code generation, or tooling.

## Summary of Changes

### Removed Features
1. **`watch` cases** - Indirect state variable watching
2. **`hint` statements** - Signal/update documentation annotations
3. **Standalone `timer` statements** - Blocking timer without `await`

### Added Features
1. **Single `await` statements** - Direct await for individual operations
2. **Direct signal/update cases** - Signals and updates can be awaited directly in `await one`
3. **Activity/workflow cases** - Activities and workflows can be raced in `await one`
4. **Optional case bodies** - Cases can have empty bodies (just colon, no statements)

---

## Detailed Changes

### 1. Timer Syntax Unification

**OLD:**
```twf
# Standalone timer (blocking)
timer 5m

# Timer case in await one
await one:
    timer (5m):
        activity HandleTimeout()
```

**NEW:**
```twf
# Single await timer
await timer(5m)

# Timer case in await one (no space before paren)
await one:
    timer(5m):
        activity HandleTimeout()
```

**Key Changes:**
- All timers now use parentheses: `timer(duration)`
- Standalone timers require `await` keyword
- Timer case syntax: `timer(duration):` (no space)

---

### 2. Signal/Update Awaiting

**OLD (Indirect via watch):**
```twf
signal Approved(approver: string):
    approved = true
    approver_name = approver

approved = false

await one:
    watch (approved):
        hint signal Approved
        close Decision{status: "approved", approver: approver_name}
    timer (7d):
        close failed "timeout"
```

**NEW (Direct awaiting):**
```twf
signal Approved(approver: string):
    approved = true
    approver_name = approver

await one:
    signal Approved:
        close Decision{status: "approved"}
    timer(7d):
        close failed "timeout"
```

**Key Changes:**
- Signals and updates are directly awaitable in `await one` blocks
- Signal case syntax: `signal Name [-> params]:` with body
- Update case syntax: `update Name [-> params]:` with body
- `hint` statements removed (no longer needed)
- `watch` cases removed (replaced by direct awaiting)

**Signal/Update Execution Model:**
- Handler body ALWAYS runs when signal/update arrives
- If the signal/update is being awaited, the case body runs AFTER the handler
- This provides consistent behavior whether or not the signal is being awaited

---

### 3. Activity/Workflow Cases in Await One

**NEW:**
```twf
await one:
    activity SlowOperation(data) -> result:
        activity ProcessResult(result)
    activity FastOperation(data) -> result:
        activity ProcessResult(result)
    timer(1h):
        close failed "all operations timed out"

# Result from whichever activity completes first
```

**Key Changes:**
- Activities can be raced in `await one` blocks
- Workflow calls (including `spawn`/`detach`) can be raced
- Syntax: `activity Name(args) [-> result]:`
- Syntax: `workflow Name(args) [-> result]:`

---

### 4. Empty Case Bodies

**NEW:**
```twf
for:
    await one:
        signal Deposit:      # Empty body - just consume signal
        signal Withdraw:     # Empty body - just consume signal

    if (balance > threshold):
        break
```

**Key Changes:**
- Case bodies are now optional
- Colon is still required even for empty bodies
- Useful for consuming signals without additional processing
- Common pattern: multiple signals affecting one condition

---

### 5. Single Await Statements

**NEW:**
```twf
# Wait for a timer
await timer(5m)

# Wait for a signal
await signal Approved

# Wait for a signal with parameter binding
await signal Approved -> (approver, timestamp)

# Wait for an update
await update ChangePlan -> (result)

# Wait for an activity
await activity Process(data) -> result

# Wait for a workflow
await workflow Child(input) -> output
```

**Key Changes:**
- New `await` statement for waiting on single operations
- Cleaner syntax for common case of waiting on one thing
- Parameter/result binding uses `->` syntax

---

## Common Patterns

### Pattern 1: Signal with Timeout

**OLD:**
```twf
signal Approved():
    approved = true

await one:
    watch (approved):
        hint signal Approved
        close
    timer (7d):
        close failed "timeout"
```

**NEW:**
```twf
signal Approved():
    approved = true

await one:
    signal Approved:
        close
    timer(7d):
        close failed "timeout"
```

---

### Pattern 2: Multiple Signals Affecting One Condition

**OLD:**
```twf
signal Deposit(amount: decimal):
    balance = balance + amount

signal Withdraw(amount: decimal):
    balance = balance - amount

await one:
    watch (balance > threshold):
        hint signal Deposit
        hint signal Withdraw
        close
```

**NEW:**
```twf
signal Deposit(amount: decimal):
    balance = balance + amount

signal Withdraw(amount: decimal):
    balance = balance - amount

# Loop until condition is met
for:
    await one:
        signal Deposit:
        signal Withdraw:

    if (balance > threshold):
        break
```

**Note:** The loop + await pattern replaces condition watching. Signal handlers update state, then the loop checks the condition.

---

### Pattern 3: Entity Workflow with Signals and Timers

**OLD:**
```twf
workflow AccountEntity(accountId: string):
    signal Deposit():
        balance = balance + amount

    signal Close():
        shouldClose = true

    for:
        await one:
            watch (shouldClose):
                hint signal Close
                close
            timer (24h):
                activity DailyReconciliation()
```

**NEW:**
```twf
workflow AccountEntity(accountId: string):
    signal Deposit(amount: decimal):
        balance = balance + amount

    signal Close():
        shouldClose = true

    for:
        await one:
            signal Close:
                close
            timer(24h):
                activity DailyReconciliation()
```

---

### Pattern 4: Racing Activities

**NEW:**
```twf
# Try multiple approaches, use whichever succeeds first
await one:
    activity TryPrimaryAPI(data) -> result:
        return result
    activity TryBackupAPI(data) -> result:
        return result
    timer(30s):
        close failed "all attempts timed out"
```

---

## Grammar Changes

### New Productions

```
await_stmt ::= 'await' await_target NEWLINE

await_target ::= timer_target
               | signal_target
               | update_target
               | activity_target
               | workflow_target

timer_target ::= 'timer' '(' duration ')'
signal_target ::= 'signal' IDENT ['->' params]
update_target ::= 'update' IDENT ['->' params]
```

### Updated Productions

```
# Await one case (added signal/update/activity/workflow)
await_one_case ::= signal_case
                 | update_case
                 | timer_case
                 | activity_case
                 | workflow_case
                 | await_all_case

signal_case ::= 'signal' IDENT ['->' params] ':' NEWLINE
                [INDENT statement+ DEDENT]

update_case ::= 'update' IDENT ['->' params] ':' NEWLINE
                [INDENT statement+ DEDENT]

timer_case ::= 'timer' '(' duration ')' ':' NEWLINE
               [INDENT statement+ DEDENT]

# Case bodies are now optional (empty body allowed)
```

### Removed Productions

```
# These no longer exist:
watch_case ::= 'watch' '(' IDENT ')' ':' ...
hint_stmt ::= 'hint' ('signal' | 'update') IDENT NEWLINE
timer_stmt ::= 'timer' duration NEWLINE
```

---

## AST Node Changes

### Removed Nodes
- `HintStmt` - No longer exists
- `TimerStmt` - Replaced by `AwaitStmt`

### New Nodes
```go
type AwaitStmt struct {
    Pos
    // Timer await
    Timer string

    // Signal await
    Signal       string
    SignalParams string
    SignalResolved *SignalDecl

    // Update await
    Update       string
    UpdateParams string
    UpdateResolved *UpdateDecl

    // Activity await
    Activity       string
    ActivityArgs   string
    ActivityResult string
    ActivityResolved *ActivityDef

    // Workflow await
    Workflow          string
    WorkflowMode      WorkflowCallMode
    WorkflowNamespace string
    WorkflowArgs      string
    WorkflowResult    string
    WorkflowResolved  *WorkflowDef
}
```

### Updated Nodes
```go
type AwaitOneCase struct {
    Pos
    // Signal case (NEW)
    Signal         string
    SignalParams   string
    SignalResolved *SignalDecl

    // Update case (NEW)
    Update         string
    UpdateParams   string
    UpdateResolved *UpdateDecl

    // Timer case (CHANGED: was TimerDuration)
    Timer string

    // Activity case (NEW)
    Activity       string
    ActivityArgs   string
    ActivityResult string
    ActivityResolved *ActivityDef

    // Workflow case (NEW)
    Workflow          string
    WorkflowMode      WorkflowCallMode
    WorkflowNamespace string
    WorkflowArgs      string
    WorkflowResult    string
    WorkflowResolved  *WorkflowDef

    // Await all case
    AwaitAll *AwaitAllBlock

    // Body (now optional)
    Body []Statement
}
```

**Field Changes:**
- Removed: `WatchVariable` (watch cases removed)
- Removed: `TimerDuration` (renamed to `Timer` for consistency)
- Added: All signal/update/activity/workflow fields

---

## Migration Guide

### For Code Generators

1. **Remove support for:**
   - `HintStmt` AST nodes
   - `TimerStmt` AST nodes
   - `watch` cases in `AwaitOneCase`

2. **Add support for:**
   - `AwaitStmt` for single await operations
   - Signal/update cases in `AwaitOneCase`
   - Activity/workflow cases in `AwaitOneCase`
   - Empty case bodies (Body field can be empty)

3. **Update field references:**
   - `AwaitOneCase.TimerDuration` → `AwaitOneCase.Timer`
   - Check for new fields: `Signal`, `Update`, `Activity`, `Workflow`

### For Language Tooling (LSP, Formatters, etc.)

1. **Remove keywords:**
   - `watch` (no longer valid)
   - `hint` (no longer valid)

2. **Update completions:**
   - Suggest `await timer()` instead of `timer`
   - In `await one` blocks, suggest signal/update/activity/workflow cases
   - Remove watch/hint suggestions

3. **Update syntax highlighting:**
   - `await` keyword can now appear before timer/signal/update/activity/workflow
   - Remove special handling for `watch` and `hint`

### For Documentation/Examples

1. **Replace watch patterns** with direct signal awaiting
2. **Add `await` to standalone timers**
3. **Remove all `hint` statements**
4. **Use `for` + `await one`** for multiple-signal-one-condition patterns

---

## Semantic Guarantees

### Signal/Update Handler Execution
- Signal/update handlers ALWAYS execute when the message arrives
- If the signal/update is being awaited, the case body runs AFTER the handler
- This ensures consistent state updates regardless of await status

### Cancellation in Await One
- When one case completes, all other pending operations are cancelled
- Activities receive cancellation signals
- Child workflows are cancelled
- Timers are stopped

### Empty Case Bodies
- Empty bodies are semantically valid
- Useful for consuming signals without processing
- Handler bodies still execute (for signals/updates)

---

## Implementation Notes

### Parser Changes
- Main entry point: `parseAwaitStmt` in `statements.go`
- Handles both single await and await blocks
- New helper: `parseOptionalCaseBody` for optional bodies

### Resolver Changes
- `AwaitStmt` needs resolution for signal/update/activity/workflow references
- `AwaitOneCase` needs resolution for all case types
- Same resolution logic as before, just more case types

### LSP Changes
- Hover: Show await statement details
- References: Track signal/update/activity/workflow references in await statements
- Definitions: Jump to definition from await statements
- Code actions: Remove hint-related actions

---

## Testing Considerations

### Test Coverage Needed
1. Single await statements (all types)
2. Signal/update cases in await one
3. Activity/workflow cases in await one
4. Empty case bodies
5. Parameter binding in signal/update awaits
6. Result binding in activity/workflow awaits

### Edge Cases
1. Signal arrives when NOT being awaited (handler still runs)
2. Multiple await one blocks awaiting same signal
3. Empty await one block (should be error)
4. Detach workflow with result binding (should be error)

---

## Timeline and Status

**Status:** ✅ Complete
- Spec updated
- Parser implemented
- Tests passing
- Examples updated
- Documentation updated

**Breaking Change:** Yes - this is a breaking change to the DSL syntax.

---

## Questions?

For implementation questions or clarifications, refer to:
- `design/lsp/LANGUAGE.md` - Complete language specification
- `design/lsp/PROPOSED_ASYNC_CHANGES.md` - Detailed design rationale
- `design/examples/*.twf` - Updated examples
- `design/*.md` - Updated pattern documentation
