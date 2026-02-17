# Workflow vs Activity Boundary

## Use Activities When

- Single atomic operation
- External system interaction (API, DB, file)
- Bounded completion time
- No orchestration logic

## Use Child Workflows When

- Multiple steps with independent retry/timeout policies
- Reusable across parent workflows
- Separate failure boundary needed
- Very long operations (separate history)
- Complex enough to warrant own tests

**Rule of thumb:** Loops or conditionals inside an activity → should be a workflow.

See [child-workflows.md](../topics/child-workflows.md) for detailed patterns.
