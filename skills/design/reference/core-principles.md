# Core Principles

## Determinism: Workflows Must Replay Identically

Temporal replays workflow code to reconstruct state. Different replay results = non-determinism errors.

| Safe in Workflows | Must Be in Activities |
|-------------------|----------------------|
| Logic on activity results | Current time, dates |
| Deterministic loops/conditionals | Random numbers, UUIDs |
| Child workflows | HTTP/API calls |
| Temporal timers | Database operations |
| Local variables | File I/O |
| Signal waits | External service calls |

**Workflows = pure orchestration. Activities = side effects.**

## Idempotency: Activities May Run Multiple Times

Retries happen (network failures, crashes, timeouts). Activities must be **idempotent**: same inputs → same result regardless of execution count.

| Pattern | Example |
|---------|---------|
| **Create-or-get** | Check existence before creating |
| **Idempotency keys** | Workflow ID + activity name as operation key |
| **Upsert** | Prefer over insert-then-update |
| **Deduplication** | Query before mutating |

**Think through retries:** CreateUser → return existing if exists. SendEmail → provider idempotency key. DeployResource → verify state, return success if deployed.
