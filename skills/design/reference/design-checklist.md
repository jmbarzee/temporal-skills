# Design Checklist

## TWF Validation
- [ ] `twf check` passes (`✓ OK`)
- [ ] `twf symbols` lists all expected definitions
- [ ] No undefined references
- [ ] No SDK-specific code in `.twf`

## Determinism
- [ ] All I/O, time, randomness in activities
- [ ] No external calls in workflow code
- [ ] Loops have deterministic bounds
- [ ] Timers use Temporal primitives

## Idempotency
- [ ] Activities handle "already exists" gracefully
- [ ] Retries produce same end state
- [ ] No duplicate side effects on replay

## Failure Handling
- [ ] Each failure mode identified
- [ ] Recovery strategy defined (retry, compensate, fail)
- [ ] Partial success handled
- [ ] Timeouts configured

## Decomposition
- [ ] Each workflow has single clear purpose
- [ ] Child workflow vs activity choice justified
- [ ] Workflow names describe outcomes, not steps

## Deployment Topology
- [ ] Workers group all workflows/activities into type sets
- [ ] Each worker instantiated in a namespace with `task_queue`
- [ ] Cross-namespace calls have nexus endpoints
- [ ] `twf check` passes topology validation
