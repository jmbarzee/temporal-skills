# Determinism & Idempotency

## Determinism: Workflows Must Replay Identically

Temporal replays workflow code to reconstruct state. Different replay results = non-determinism errors. See [Temporal: Deterministic Constraints](https://docs.temporal.io/workflows#deterministic-constraints) for the authoritative reference.

| Safe in Workflows | Must Be in Activities |
|-------------------|----------------------|
| Logic on activity results | Current time, dates |
| Deterministic loops/conditionals | Random numbers, UUIDs |
| Child workflows | HTTP/API calls |
| Temporal timers | Database operations |
| Local variables | File I/O |
| Signal waits | External service calls |
| Deterministic iteration (arrays, slices) | Map/dictionary iteration (order varies) |
| Temporal SDK concurrency (promises, await all) | Language-level threads, goroutines, async |
| Workflow-local state | Mutable global/shared state |

**Workflows = pure orchestration. Activities = side effects.**

## Idempotency: Activities May Run Multiple Times

Retries happen (network failures, crashes, timeouts). Activities must be **idempotent**: same inputs → same result regardless of execution count.

| Pattern | Example |
|---------|---------|
| **Create-or-get** — when entity has a natural unique key | Check existence before creating |
| **Idempotency keys** — when external system supports them | Workflow ID + activity name as operation key |
| **Upsert** — when database supports atomic upsert | Prefer over insert-then-update |
| **Deduplication** — last resort when no built-in mechanism | Query before mutating |

**Think through retries:** CreateUser → return existing if exists. SendEmail → provider idempotency key. DeployResource → verify state, return success if deployed.
