# Workflow vs Activity Boundary

## Use Activities When

- Single atomic operation
- External system interaction (API, DB, file)
- Short, predictable completion time (single timeout period)
- No orchestration logic

## Use Child Workflows When

- Multiple steps with independent retry/timeout policies
- Reusable across parent workflows
- Separate failure boundary needed
- Very long operations (separate history)
- Complex enough to warrant own tests

**Rule of thumb:** Loops or conditionals inside an activity → should be a workflow.

## Use Nexus When

- Crosses namespace or team boundaries (separate deployment lifecycle)
- Different team owns the target service
- Target needs independent scaling, versioning, or failure isolation at the organizational level
- You want a typed API contract between services

**Child workflow vs nexus:** Child workflows share a namespace and are tightly coupled to the parent's lifecycle. Nexus calls are loosely coupled — the target is an independent service that may be owned by another team, deployed on a different schedule, or running in a different namespace.

## Common Mistakes

**Wrapper workflow:** A child workflow containing a single activity call adds orchestration overhead with no benefit. If there's only one step, use an activity directly.

**Monolithic workflow:** All logic in one workflow with hundreds of history events. If a workflow has more than ~10 sequential activity calls, consider decomposing into child workflows.

**Activity with orchestration:** If an activity contains retry logic, conditional branching, or calls to other services, it should be a workflow — these are orchestration concerns that benefit from Temporal's durability.

See [child-workflows.md](../topics/child-workflows.md) for detailed child workflow patterns and the full decision table.

For deployment topology and task queue routing, see [task-queues.md](../topics/task-queues.md).
