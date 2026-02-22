# Design Checklist

## TWF Validation
- [ ] `twf check` passes (`✓ OK`)
- [ ] `twf symbols` lists all expected definitions
- [ ] No undefined references
- [ ] No SDK-specific code in `.twf`
→ See [common-errors.md](./common-errors.md) for error troubleshooting

## Determinism
- [ ] All I/O, time, randomness in activities
- [ ] No external calls in workflow code
- [ ] Loops have deterministic bounds
- [ ] Timers use Temporal primitives
- [ ] No non-deterministic data structure iteration (maps, sets)
- [ ] Version-specific branching uses proper versioning pattern
→ See [core-principles.md](./core-principles.md) for determinism rules

## Idempotency
- [ ] Activities handle "already exists" gracefully
- [ ] Retries produce same end state
- [ ] No duplicate side effects on replay
→ See [core-principles.md](./core-principles.md) for idempotency patterns

## Failure Handling
- [ ] Each failure mode identified
- [ ] Recovery strategy defined (retry, compensate, fail)
- [ ] Partial success handled
- [ ] Timeouts configured
→ See [anti-patterns.md](./anti-patterns.md) for common failure handling mistakes

## Decomposition
- [ ] Each workflow has single clear purpose
- [ ] Child workflow vs activity choice justified
- [ ] Workflow names describe outcomes, not steps
→ See [workflow-boundaries.md](./workflow-boundaries.md) for boundary decisions

## Deployment Topology (design review — `twf check` validates syntax)
- [ ] Worker groupings reflect actual deployment needs (not just "one worker for everything")
- [ ] Task queue separation matches scaling and isolation requirements
- [ ] Cross-namespace calls have nexus endpoints
- [ ] `twf check` passes topology validation
→ See [task-queues.md](../topics/task-queues.md) for task queue design
