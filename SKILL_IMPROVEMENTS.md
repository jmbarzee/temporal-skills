# Skill Improvement Opportunities

Catalog of enhancement ideas for skill documentation. These are not errors — all current content is functional. Items here represent coverage gaps or areas where newer DSL features (workers, namespaces, nexus) could be better leveraged.

---

## Design Skill (`skills/design/`)

### SKILL.md
- No mention of workers, namespaces, or nexus in syntax summary or basic structure example
- "Rules" table doesn't mention nexus/worker/namespace validation rules from `twf check`
- "Completion" checklist missing deployment topology validation (workers defined, namespaces configured, etc.)
- Reference Index has no direct link to `topics/nexus.md` or `topics/task-queues.md`

### reference/common-errors.md
- Missing all 15+ nexus/worker/namespace error types documented in LANGUAGE.md
- Could be expanded to cover the full `twf check` error catalog

### reference/design-checklist.md
- Missing deployment topology checks (workers defined, namespaces with task_queue, etc.)
- Missing nexus-specific checks (cross-namespace boundaries justified, call timeouts configured)

### reference/anti-patterns.md
- Only 3 anti-patterns covered
- Could add: nexus for same-namespace calls, deployment config in workers instead of namespaces, workers not instantiated in namespaces

### reference/workflow-boundaries.md
- No mention of nexus as a third boundary option alongside activity vs child workflow
- Could add "Use Nexus When" section for cross-namespace/cross-team/different security contexts

### reference/primitives-reference.md
- Missing `worker` and `namespace` as infrastructure primitives

### topics/activities-advanced.md
- Local activity section uses conceptual syntax not supported by parser — consider marking as future/conceptual more clearly
- No worker/namespace context shown for how activities relate to deployment

### topics/child-workflows.md
- `workflow_id` examples are central to this file but not a supported option — needs rethinking once workflow_id support is decided
- No worker/namespace deployment shown for parent/child relationships

### topics/long-running.md
- Entity workflow signal/query/update declaration ordering may confuse readers (appears after body code in examples)
- SDK intrinsics (`history_size()`, `history_length()`) used inconsistently across examples
- Could benefit from worker/namespace blocks showing how entity workflows get deployed

### topics/patterns.md
- No nexus patterns shown — could add cross-namespace pattern using current nexus syntax
- State machine pattern uses unsupported `elif` — needs `else if` support or `switch`/`case` rewrite

### topics/testing.md
- Could add nexus testing section showing how to test cross-namespace calls

### topics/task-queues.md
- Line mentioning "Workers contain only workflow and activity entries" should also mention `nexus service`
- Could demonstrate nexus endpoint deployment alongside worker deployment in namespace blocks

---

## Author-Go Skill (`skills/author-go/`)

### SKILL.md
- Reference Index missing nexus call row (`nexus Endpoint Service.Op(args) -> result`)
- Layer 4 ("Worker wiring") could reference TWF worker/namespace blocks as inputs for generating worker initialization code
- Missing rows for nexus service definitions, worker blocks, and namespace blocks mapped to Go equivalents

### reference/await-one.md
- Missing nexus case example in `await one:` blocks
- Missing update case example

### reference/promise.md
- Missing nexus promise variant (`promise p <- nexus Endpoint Service.Op(args)`)
- Missing update promise variant

### reference/options.md
- Missing nexus call options example (`schedule_to_close_timeout`, `retry_policy`, `priority`)
- Could reference worker/namespace options

### reference/await-all.md
- Could show nexus call inside `await all:` block

### reference/activity-call.md
- Could show inline options block example alongside the basic call

### Missing reference files
- No `worker.md` or `namespace.md` reference file for Go code generation of these constructs
- No `nexus-service-def.md` for generating nexus service handler code
