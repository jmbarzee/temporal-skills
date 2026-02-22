# Temporal Primitives Reference

## Workflow Execution

| Primitive | Purpose | Details |
|-----------|---------|---------|
| `activity` | Side-effecting operation | Core primitive |
| `workflow` | Child workflow | [child-workflows.md](../topics/child-workflows.md) |
| `nexus` | Cross-namespace call | [nexus.md](../topics/nexus.md) |
| `promise` | Async operation, await later | [promises-conditions.md](../topics/promises-conditions.md) |
| `detach` | Fire-and-forget child/nexus | [child-workflows.md](../topics/child-workflows.md), [nexus.md](../topics/nexus.md) |
| `close continue_as_new` | Reset history, continue | [long-running.md](../topics/long-running.md) |

**Selection:** `activity` for single side-effecting operations. `workflow` for multi-step orchestration needing its own retry/failure boundary. `nexus` when crossing namespace or team boundaries. `promise` when you need the result later, not immediately. `detach` for fire-and-forget — you cannot observe the result. `close continue_as_new` when history grows unbounded (long-running or entity workflows).

## Timing

| Primitive | Purpose | Details |
|-----------|---------|---------|
| `timer` | Durable sleep (survives restarts) | [timers-scheduling.md](../topics/timers-scheduling.md) |
| `schedule` | Cron-like recurring execution | [timers-scheduling.md](../topics/timers-scheduling.md) |
| `timeout` | Deadline for operations | [timers-scheduling.md](../topics/timers-scheduling.md) |

**Selection:** `timer` for durable waits inside workflow logic (survives replay). Activity-level timeouts (`heartbeat_timeout`, `start_to_close_timeout` in `options:`) for bounding activity execution. `schedule` for cron-like recurring workflow starts — this is platform configuration, not TWF notation.

## External Communication

Read, write, or read-write interaction with a running workflow:

| Primitive | I/O | Purpose | Details |
|-----------|-----|---------|---------|
| `query` | Read | Sync read of workflow state | [signals-queries-updates.md](../topics/signals-queries-updates.md) |
| `signal` | Write | Async fire-and-forget into workflow | [signals-queries-updates.md](../topics/signals-queries-updates.md) |
| `update` | Read-write | Sync mutation with result | [signals-queries-updates.md](../topics/signals-queries-updates.md) |

**Selection:** Use I/O direction as the decision rule. **Do not** use `query` to modify state — queries must be pure reads. **Do not** use `signal` when you need confirmation — signals have no return value. Prefer `update` over signal-then-query when the caller needs to know the mutation succeeded.

## State and Conditions

| Primitive | Purpose | Details |
|-----------|---------|---------|
| `state` | Workflow state declaration block | [promises-conditions.md](../topics/promises-conditions.md) |
| `condition` | Named boolean awaitable | [promises-conditions.md](../topics/promises-conditions.md) |
| `set` / `unset` | Set condition to true / false | [promises-conditions.md](../topics/promises-conditions.md) |

**Selection:** Use `condition` when handlers and the main workflow body need to coordinate on a boolean flag (e.g., "payment received", "approved"). Use local variables for workflow-scoped state that doesn't need cross-handler coordination.

## Activity Options

| Primitive | Purpose | Details |
|-----------|---------|---------|
| `heartbeat` | Report progress, detect worker death | [activities-advanced.md](../topics/activities-advanced.md) |
| `async_complete` | Complete from external system | [activities-advanced.md](../topics/activities-advanced.md) |

## Infrastructure

| Primitive | Purpose | Details |
|-----------|---------|---------|
| `task_queue` | Route work to specific workers | [task-queues.md](../topics/task-queues.md) |
| `worker` | Defines a reusable type set (which workflows, activities, nexus services run together) | [task-queues.md](../topics/task-queues.md) |
| `namespace` | Deployment topology — instantiates workers with `task_queue` and options | [task-queues.md](../topics/task-queues.md) |
| `nexus service` | Typed operation group for cross-namespace calls | [nexus.md](../topics/nexus.md) |
| `nexus endpoint` | Routes nexus calls to a target task queue | [nexus.md](../topics/nexus.md) |
| `search_attribute` | Index workflow for queries | Core primitive |
| `memo` | Attach metadata to workflow | Core primitive |
