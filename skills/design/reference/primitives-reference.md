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

## Timing

| Primitive | Purpose | Details |
|-----------|---------|---------|
| `timer` | Durable sleep (survives restarts) | [timers-scheduling.md](../topics/timers-scheduling.md) |
| `schedule` | Cron-like recurring execution | [timers-scheduling.md](../topics/timers-scheduling.md) |
| `timeout` | Deadline for operations | [timers-scheduling.md](../topics/timers-scheduling.md) |

## External Communication

Read, write, or read-write interaction with a running workflow:

| Primitive | I/O | Purpose | Details |
|-----------|-----|---------|---------|
| `query` | Read | Sync read of workflow state | [signals-queries-updates.md](../topics/signals-queries-updates.md) |
| `signal` | Write | Async fire-and-forget into workflow | [signals-queries-updates.md](../topics/signals-queries-updates.md) |
| `update` | Read-write | Sync mutation with result | [signals-queries-updates.md](../topics/signals-queries-updates.md) |

## State and Conditions

| Primitive | Purpose | Details |
|-----------|---------|---------|
| `state` | Workflow state declaration block | [promises-conditions.md](../topics/promises-conditions.md) |
| `condition` | Named boolean awaitable | [promises-conditions.md](../topics/promises-conditions.md) |
| `set` / `unset` | Set or clear a condition | [promises-conditions.md](../topics/promises-conditions.md) |

## Activity Options

| Primitive | Purpose | Details |
|-----------|---------|---------|
| `heartbeat` | Report progress, detect worker death | [activities-advanced.md](../topics/activities-advanced.md) |
| `async_complete` | Complete from external system | [activities-advanced.md](../topics/activities-advanced.md) |

## Infrastructure

| Primitive | Purpose | Details |
|-----------|---------|---------|
| `task_queue` | Route work to specific workers | [task-queues.md](../topics/task-queues.md) |
| `worker` | Reusable type set (workflows, activities, nexus services) | [task-queues.md](../topics/task-queues.md) |
| `namespace` | Instantiates workers with deployment options | [task-queues.md](../topics/task-queues.md) |
| `nexus service` | Typed operation group for cross-namespace calls | [nexus.md](../topics/nexus.md) |
| `nexus endpoint` | Routes nexus calls to a target task queue | [nexus.md](../topics/nexus.md) |
| `search_attribute` | Index workflow for queries | Core primitive |
| `memo` | Attach metadata to workflow | Core primitive |
