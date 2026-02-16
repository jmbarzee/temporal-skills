---
name: temporal-workflow-design
description: Design Temporal workflows with proper determinism, idempotency, and decomposition. Use when designing new workflows, planning workflow-activity boundaries, or reviewing workflow architecture.
---

# Temporal Workflow Design

Design Temporal workflows using `.twf` (Temporal Workflow Format) — a language-agnostic DSL capturing workflow structure, activity boundaries, and Temporal primitives. Always produce `.twf` files as deliverables, never SDK code.

---

## Design Flow

Core loop: **write TWF → `twf check` → fix/consult → repeat**. Parser errors are design feedback — validate early and often.

```
  ┌────────────┐
  │ Write/Edit │◄──────────────┐
  │ .twf file  │               │
  └─────┬──────┘               │
        ▼                      │
  ┌────────────┐               │
  │ twf check  │               │
  └─────┬──────┘               │
        ▼                      │
   ┌─────────┐                 │
   │ Error?  │──No──► Done     │
   └────┬────┘                 │
        │Yes                   │
        ▼                      │
  ┌─────────────┐              │
  │ Can fix     │──Yes─────────┘
  │ confidently?│
  └──────┬──────┘
         │No
         ▼
  ┌─────────────┐
  │ Consult     │──────────────┘
  │ user        │
  └─────────────┘
```

### Worked Example

```twf
workflow Greet(name: string) -> (Greeting):
    activity BuildGreeting(name) -> greeting
    close complete(Greeting{greeting})
```

`twf check` → `resolve error at 2:5: undefined activity "BuildGreeting"`

Fix — add the definition:

```twf
workflow Greet(name: string) -> (Greeting):
    activity BuildGreeting(name) -> greeting
    close complete(Greeting{greeting})

activity BuildGreeting(name: string) -> (Greeting):
    return format("Hello, {name}")
```

`twf check` → `✓ OK`

### When to Consult the User

**Fix yourself:** clear syntax mistakes, unambiguous errors (undefined → add definition), pattern exists in docs.

**Ask the user:** multiple valid approaches, requirements gap, unclear architectural choice, workaround feels wrong.

**Cost of asking < Cost of wrong design.**

---

## `twf` CLI

**Run `twf check` after every `.twf` edit.** Fix all errors before presenting to user.

| Command | Purpose |
|---------|---------|
| `twf check <file...>` | Parse + resolve — run after every edit |
| `twf symbols <file...>` | List all workflow/activity signatures |
| `twf symbols --json <file...>` | Machine-readable symbol output |
| `twf check --lenient <file...>` | Partial tolerance for incomplete designs |

**Error format** (stderr): `parse error at <line>:<col>: <message>` / `resolve error at <line>:<col>: <message>`

---

## TWF Syntax

Full grammar: [`LANGUAGE.md`](./lsp/LANGUAGE.md). All `.twf` must pass `twf check` before presenting to user.

Activity bodies are intentionally free-form (`raw_stmt`) — they represent SDK-level implementation, not orchestration. Use pseudocode or descriptive text.

### Rules (enforced by `twf check`)

| Rule | Correct | Wrong |
|------|---------|-------|
| Return types parenthesized | `-> (Result)` | `-> Result` |
| `if`/`for` require parentheses | `if (expr):` / `for (x in items):` | `if expr:` / `for x in items:` |
| Handlers inside workflows, before body | `signal`/`query`/`update` in workflow | At top level |
| All calls need matching definitions | `activity Foo()` requires `activity Foo(...):` | Calling undefined |
| Activities: no temporal primitives | — | `timer`, `signal`, `await` in activity body |
| Files self-contained | All referenced definitions present | — |

### Basic Structure

```twf
workflow WorkflowName(input: InputType) -> (OutputType):
    activity ActivityName(input) -> result
    workflow ChildWorkflowName(input) -> childResult
    close complete(OutputType{result, childResult})

workflow ChildWorkflowName(input: InputType) -> (ChildResult):
    activity DoWork(input) -> result
    close complete(ChildResult{result})

activity ActivityName(input: InputType) -> (Result):
    return process(input)

activity DoWork(input: InputType) -> (WorkResult):
    return work(input)
```

---

## Reference Index

Read only what the current design requires.

### Design Essentials

| Topic | When to Consult | File |
|-------|-----------------|------|
| Core Principles | Determinism/idempotency review | [core-principles.md](./reference/core-principles.md) |
| Workflow Boundaries | Activity vs child workflow decision | [workflow-boundaries.md](./reference/workflow-boundaries.md) |
| Notation Examples | Control flow, handlers, timers, nexus in TWF | [notation-examples.md](./reference/notation-examples.md) |
| Notation Reference | All TWF syntax constructs | [notation-reference.md](./reference/notation-reference.md) |
| Design Checklist | Final verification before presenting | [design-checklist.md](./reference/design-checklist.md) |
| Anti-Patterns | Common Temporal design mistakes | [anti-patterns.md](./reference/anti-patterns.md) |
| Editor Setup | VS Code/Cursor extension | [editor-setup.md](./reference/editor-setup.md) |
| Primitives Reference | Temporal primitive lookup | [primitives-reference.md](./reference/primitives-reference.md) |

### Topic Deep-Dives

| Topic | When to Consult | File |
|-------|-----------------|------|
| Signals, Queries, Updates | External communication with running workflows | [signals-queries-updates.md](./topics/signals-queries-updates.md) |
| Promises and Conditions | Async operations, named conditions | [promises-conditions.md](./topics/promises-conditions.md) |
| Child Workflows | Parent/child decomposition | [child-workflows.md](./topics/child-workflows.md) |
| Timers and Scheduling | Durable timers, schedules, deadlines | [timers-scheduling.md](./topics/timers-scheduling.md) |
| Advanced Activities | Heartbeats, async completion | [activities-advanced.md](./topics/activities-advanced.md) |
| Long-Running Workflows | Continue-as-new, history management | [long-running.md](./topics/long-running.md) |
| Nexus | Cross-namespace calls | [nexus.md](./topics/nexus.md) |
| Task Queues | Worker routing, scaling | [task-queues.md](./topics/task-queues.md) |
| Workflow Patterns | Saga, pipeline, fan-out/fan-in | [patterns.md](./topics/patterns.md) |
| Testing | Test strategy for workflows | [testing.md](./topics/testing.md) |
| Versioning | Evolving existing workflows safely | [versioning.md](./topics/versioning.md) |
| Common Errors | Troubleshooting `twf check` | [common-errors.md](./reference/common-errors.md) |
| Full Grammar | Exact grammar rules | [lsp/LANGUAGE.md](./lsp/LANGUAGE.md) |
