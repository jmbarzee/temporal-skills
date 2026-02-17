# Common Anti-Patterns

## Non-Determinism in Workflows

```twf
# BAD: Time check in workflow (non-deterministic on replay)
# if (current_time() > deadline):
#     cancel()

# GOOD: Timer-based deadline
# await one:
#     activity DoWork() -> result:
#         close complete(Result{status: "success"})
#     timer(deadline):
#         close fail(Result{status: "timeout"})
```

## Non-Idempotent Activities

```pseudo
# BAD: Assumes fresh state — fails on retry
activity CreateUser(name):
    db.insert(User(name))

# GOOD: Create-or-get
activity CreateUser(name):
    existing = db.get_by_name(name)
    if existing: return existing
    return db.insert(User(name))
```

## Orchestration in Activities

```pseudo
# BAD: Loop in activity — partial failure unrecoverable
# activity DeployAll(specs):
#     for spec in specs:
#         deploy(spec)
#         wait_healthy(spec)

# GOOD: Workflow orchestrates, each step retryable
# workflow DeployAll(specs: []Spec):
#     for (spec in specs):
#         activity Deploy(spec)
#         activity WaitHealthy(spec)
```
