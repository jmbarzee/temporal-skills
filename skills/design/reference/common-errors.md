# Common Errors

Common `twf check` errors and how to fix them.

## Resolve Errors

| Error | Cause | Fix |
|-------|-------|-----|
| `undefined activity: Foo` | Activity `Foo` is called but not defined | Add `activity Foo(...):` definition to the file |
| `undefined workflow: Foo` | Child workflow `Foo` is called but not defined | Add `workflow Foo(...):` definition to the file |
| `undefined signal: Foo` | `await signal Foo` or `signal Foo:` case but no signal handler declared | Add `signal Foo(...):` declaration inside the workflow, before the body |
| `undefined update: Foo` | `await update Foo` or `update Foo:` case but no update handler declared | Add `update Foo(...) -> (Type):` declaration inside the workflow, before the body |
| `undefined condition: Foo` | `set Foo`, `unset Foo`, or `await Foo` but no condition declared | Add `condition Foo` inside the workflow's `state:` block |
| `undefined promise or condition: Foo` | `await Foo` or `Foo:` case in `await one` but `Foo` is not a promise or condition | Add `promise Foo <- ...` in the workflow body or `condition Foo` in the `state:` block |
| `duplicate workflow definition: Foo` | Two `workflow Foo` definitions in the same file | Remove or rename the duplicate |
| `duplicate activity definition: Foo` | Two `activity Foo` definitions in the same file | Remove or rename the duplicate |
| `condition "Foo" cannot have a result binding` | `await Foo -> result` where `Foo` is a condition | Conditions are boolean — remove the `-> result` binding |

## Parse Errors

| Error | Cause | Fix |
|-------|-------|-----|
| `<keyword> is not allowed in activity body` | Using a temporal primitive (`workflow`, `activity`, `timer`, `signal`, `await`, etc.) inside an activity definition or query handler | Move the temporal primitive to a workflow. Activities run outside the replay-safe workflow context as normal side-effecting code — temporal primitives require deterministic replay and cannot function in activities. |
| `expected ( after return type ->` | Return type not parenthesized: `-> Result` | Use `-> (Result)` — return types must be wrapped in parentheses |
| `expected ( after if` / `expected ( after for` | Missing parentheses around condition/iterator | Use `if (expr):` / `for (x in items):` |
| `unexpected token <tok> at top level` | Statement or keyword that doesn't start a workflow or activity definition | Ensure all top-level items are `workflow`, `activity`, `worker`, `namespace`, or `nexus service` definitions |
| `unexpected token <tok> in await one case` | Invalid case type inside `await one:` block | Cases must be `signal`, `update`, `timer`, `activity`, `workflow`, an identifier, or `await all` |

## Worker / Namespace / Nexus Errors

| Error | Cause | Fix |
|-------|-------|-----|
| `worker X references undefined ...` | Worker lists a name that doesn't exist | Add the definition or fix the name |
| `namespace X references undefined worker: Y` | Namespace uses unknown worker | Add worker block or fix name |
| `worker instantiation missing required task_queue` | No `task_queue` in options | Add `options: task_queue: "..."` |
| `duplicate nexus endpoint name "X"` | Same endpoint name in multiple namespaces | Use unique endpoint names |
| `workers on same task_queue with different type sets` | Two workers share a task queue but register different workflows/activities | Ensure all workers on the same task queue have identical type sets, or use separate task queues |
| `nexus endpoint routes to task_queue with no worker registering the service` | Endpoint's task queue has no worker that registers the nexus service | Add the nexus service to a worker on that task queue |
| `explicit task_queue routing: target not on any worker polling that queue` | Activity/workflow call specifies a `task_queue` option, but no worker on that queue registers the target | Add the target to a worker on the specified task queue, or fix the task queue name |
| `implicit task_queue routing: target not on calling workflow's task queue` | Activity/workflow is called without explicit `task_queue`, but no worker on the caller's task queue registers it | Add the target to a worker on the same task queue, or add an explicit `task_queue` option to route correctly |
| `unknown option key` | Unrecognized key in an `options:` block | Check spelling against allowed option keys for the context (activity call, workflow call, worker instantiation, etc.) |
| `wrong value type for option key` | Option value doesn't match expected type (e.g., number where duration expected) | Check the expected type for the option key |

## Warnings

These indicate coverage gaps or potential issues, not hard errors:

| Warning | Meaning |
|---------|---------|
| Workflow/activity not registered on any instantiated worker | Definition exists but won't be reachable at runtime |
| Nexus service not referenced by any worker | Service defined but no worker registers it |
| Worker not instantiated in any namespace | Worker defined but never deployed |
| Empty worker (no registrations) | Worker block has no workflow/activity/nexus service entries |
| Empty namespace (no instantiations) | Namespace block has no worker or endpoint instantiations |
| Empty workflow body | Workflow has no statements |
| Empty activity body | Activity has no statements |
| Unresolved nexus endpoint (no endpoints defined) | Nexus call references an endpoint that may be external |
| Unresolved nexus service (no services defined) | Nexus call references a service that may be external |
