# TWF DSL Changes - Watch/Timer/Close Update

## Summary

Updated TWF syntax to better support async coordination patterns and explicit workflow termination.

## Key Changes

### 1. Watch Cases in Await One

**New:** `watch (variable):` blocks wait for state variables to become truthy.

```twf
await one:
    watch (approved):
        hint signal Approved
        close
    timer (7d):
        close failed "timeout"
```

- Watch cases monitor state variables (typically set by signal/update handlers)
- Hints document which signals/updates affect the watched variable
- When variable becomes truthy, watch case body executes

### 2. Timer Cases Now Have Bodies

**Changed:** Timer cases in `await one` now support statement blocks (reverses previous design).

```twf
await one:
    timer (5m):
        activity SendReminder()
        close
```

- Syntax: `timer (duration):` with colon and indented body
- Body executes when timer fires
- Provides deterministic reentry point for workflow execution

### 3. Close Statement for Workflow Termination

**New:** `close` statement explicitly terminates workflow execution with status.

```twf
close                    # Normal completion (default)
close completed          # Explicit successful completion
close failed "reason"    # Failed state with optional message
```

- Replaces ambiguous `return` for workflow termination
- Only valid in workflow context (not activities/queries)
- Signals/updates cannot call `close` - they only mutate state

## Semantic Model

**Signal/Update Execution:**
- Signals and updates **only mutate state**, cannot terminate workflows
- Main workflow body checks state and calls `close` to terminate

**Example Pattern:**
```twf
workflow Approval():
    signal Approved():
        approved = true    # Just set state

    approved = false

    await one:
        watch (approved):
            hint signal Approved
            close          # Main body terminates
        timer (7d):
            close failed "timeout"
```

## Migration Notes

- **Old:** Timers without bodies → **New:** Timers with bodies
- **Old:** `return Result{...}` in workflows → **New:** `close` or `close completed Result{...}`
- **Old:** Signals with `return` → **New:** Signals set state, use `watch` to observe
- `hint` statements now commonly appear inside `watch` blocks to document relationships

## Benefits

1. **Clearer async coordination** - watch blocks show what state changes you're waiting for
2. **Explicit termination** - `close` makes workflow completion intent obvious
3. **Better documentation** - hints inside watch blocks show signal/state relationships
4. **Deterministic execution** - timer/watch bodies provide clear reentry points
