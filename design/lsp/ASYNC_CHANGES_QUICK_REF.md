# TWF Async Changes - Quick Reference

## What Changed?

| Feature | Old Syntax | New Syntax |
|---------|-----------|------------|
| **Standalone Timer** | `timer 5m` | `await timer(5m)` |
| **Timer Case** | `timer (5m):` | `timer(5m):` |
| **Signal Awaiting** | `watch (var): hint signal X` | `signal X:` |
| **Update Awaiting** | `watch (var): hint update X` | `update X:` |
| **Activity Racing** | ❌ Not supported | `activity Foo() -> result:` |
| **Workflow Racing** | ❌ Not supported | `workflow Bar() -> result:` |
| **Hint Statements** | `hint signal X` | ❌ Removed |
| **Watch Cases** | `watch (variable):` | ❌ Removed |

---

## Removed
- ❌ `hint` statements
- ❌ `watch` cases
- ❌ Standalone `timer` without `await`

## Added
- ✅ Single `await` statements
- ✅ Direct signal/update cases in `await one`
- ✅ Activity/workflow cases in `await one`
- ✅ Optional/empty case bodies

---

## Examples

### Before & After

#### Timer
```twf
# OLD
timer 5m

# NEW
await timer(5m)
```

#### Signal with Timeout
```twf
# OLD
signal Approved():
    approved = true
await one:
    watch (approved):
        hint signal Approved
        close
    timer (7d):
        close failed "timeout"

# NEW
signal Approved():
    approved = true
await one:
    signal Approved:
        close
    timer(7d):
        close failed "timeout"
```

#### Multiple Signals → One Condition
```twf
# OLD
await one:
    watch (balance > 100):
        hint signal Deposit
        hint signal Withdraw
        close

# NEW
for:
    await one:
        signal Deposit:
        signal Withdraw:
    if (balance > 100):
        break
```

#### Racing Activities
```twf
# NEW (not possible before)
await one:
    activity FastAPI(data) -> result:
        return result
    activity SlowAPI(data) -> result:
        return result
    timer(30s):
        close failed "timeout"
```

---

## AST Changes

### Removed Nodes
- `HintStmt`
- `TimerStmt`

### New Nodes
- `AwaitStmt` (single await)

### Updated Fields in AwaitOneCase
- ❌ Removed: `WatchVariable`, `TimerDuration`
- ✅ Added: `Signal`, `SignalParams`, `Update`, `UpdateParams`, `Timer`, `Activity`, `ActivityArgs`, `ActivityResult`, `Workflow`, `WorkflowMode`, etc.
- ✅ Changed: `Body` is now optional (can be empty)

---

## Common Patterns

### Pattern: Signal or Timeout
```twf
await one:
    signal Ready:
        # proceed
    timer(1h):
        close failed "timeout"
```

### Pattern: Multiple Signals
```twf
for:
    await one:
        signal A:
        signal B:
    if (done):
        break
```

### Pattern: Entity with Periodic Task
```twf
for:
    await one:
        signal Command:
            # handle command
        timer(24h):
            activity DailyTask()
```

### Pattern: Racing APIs
```twf
await one:
    activity PrimaryAPI() -> result:
    activity BackupAPI() -> result:
    timer(30s):
        close failed "timeout"
# result from whichever completed first
```

---

## Migration Checklist

- [ ] Replace `timer X` with `await timer(X)`
- [ ] Replace `timer (X):` with `timer(X):`
- [ ] Remove all `hint` statements
- [ ] Convert `watch` cases to direct signal/update cases
- [ ] Update code that references `TimerDuration` → `Timer`
- [ ] Update code that references `WatchVariable` (no replacement)
- [ ] Add support for signal/update/activity/workflow cases
- [ ] Add support for empty case bodies
- [ ] Add support for `AwaitStmt` node

---

## Keywords

### Removed
- `watch`
- `hint`

### Modified Usage
- `timer` - Now requires parens and `await` for standalone use

### New Usage
- `await` - Can now precede timer/signal/update/activity/workflow

---

## For More Details

See `DSL_ASYNC_CHANGES.md` for complete implementation guide.
