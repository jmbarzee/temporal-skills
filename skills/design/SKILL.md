---
name: temporal-workflow-design
description: Design Temporal workflows with proper determinism, idempotency, and decomposition. Use when designing new workflows, planning workflow-activity boundaries, or reviewing workflow architecture.
---

# Temporal Workflow Design

Design Temporal workflows using `.twf` (Temporal Workflow Format) — a language-agnostic DSL capturing workflow structure, activity boundaries, and Temporal primitives. Always produce `.twf` files as deliverables, never SDK code.

---

## Design Flow

Core loop: **write TWF → `twf check` → fix/consult → repeat**. Parser errors are design feedback — validate early and often.

**Write before you read.** Draft TWF from the workflow description even if you're unsure — use `twf check --lenient` for incomplete designs. Consult references only to fix specific errors, not to prepare.

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

**Draft** — write from the description, don't worry about completeness:

```twf
workflow ProcessOrder(order: Order) -> (OrderResult):
    activity ValidateOrder(order) -> validated
    activity ChargePayment(order.payment) -> payment
    activity ShipOrder(order, payment) -> shipment
    close complete(OrderResult{shipment})
```

**Iteration 1** — `twf check` finds errors:

```
resolve error at 2:5: undefined activity "ValidateOrder"
resolve error at 3:5: undefined activity "ChargePayment"
resolve error at 4:5: undefined activity "ShipOrder"
```

Fix — add the missing definitions. `twf check` → `✓ OK`

**Iteration 2** — design review. Shipping involves creating a shipment, waiting for carrier pickup, and tracking — multiple steps with independent retry. Consult [workflow-boundaries.md](./reference/workflow-boundaries.md): multi-step orchestration with its own lifecycle → child workflow.

Revise — extract `ShipOrder` as a child workflow:

```twf
workflow ProcessOrder(order: Order) -> (OrderResult):
    activity ValidateOrder(order) -> validated
    activity ChargePayment(order.payment) -> payment
    workflow ShipOrder(order, payment) -> shipment
    close complete(OrderResult{shipment})
```

`twf check` → `✓ OK`. Design is structurally sound.

### Revising an Existing Design

To revise an existing `.twf` file: run `twf symbols` to understand current structure, make edits, then re-enter the core loop (`twf check` → fix → repeat). Treat user feedback as new requirements — ask clarifying questions before editing if the feedback is ambiguous.

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

Full grammar: [`LANGUAGE_SPEC.md`](../../../tools/lsp/LANGUAGE_SPEC.md). All `.twf` must pass `twf check` before presenting to user.

Activity bodies are intentionally free-form (`raw_stmt`) — they represent SDK-level implementation, not orchestration. Use pseudocode or descriptive text. The right level of detail depends on how obvious the behavior is from the name and signature:

- **Obvious** — minimal body. `activity SendEmail(to: string, body: string)` doesn't need elaboration.
- **Non-obvious** — describe key operations and external systems:

```twf
activity ExecuteToolCalls(toolCalls: ToolCalls) -> (ToolResults):
    # Look up each tool by name in the tool registry
    # Execute calls in parallel where possible
    # If a tool is not found, return an error result (don't fail the activity)
```

- **Complex contract** — describe error conditions, ordering, and idempotency requirements:

```twf
activity ReconcileInventory(warehouseId: string, expected: Inventory) -> (ReconcileResult):
    # Fetch current inventory, diff against expected, flag discrepancies
    # Must be idempotent — running twice with same input produces same flags
    # Warehouse API is rate-limited: max 10 requests/second
```

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

worker mainWorker:
    workflow WorkflowName
    workflow ChildWorkflowName
    activity ActivityName
    activity DoWork

namespace default:
    worker mainWorker
        options:
            task_queue: "main"
```

---

## Completion

The design is ready to present when:

1. `twf check` passes with no errors
2. `twf symbols` lists all expected workflows and activities
3. Worker/namespace topology validates (when present)
4. All I/O, time, and randomness live in activities (determinism)
5. Activities are idempotent (retries produce same result)
6. Failure modes have recovery strategies

For the full checklist: [design-checklist.md](./reference/design-checklist.md). Present a summary alongside the `.twf` file: key workflows, activity purposes, and notable design decisions.

---

## Handoff

The deliverable is the `.twf` file. Do not implement SDK code within this skill. If an authoring skill is available (e.g. `author-go`, `author-ts`), suggest it. Alongside the `.twf` file, note: target SDK/language, external system assumptions, and design decisions not captured in the notation.

---

## Reference Index

Read only what the current design requires.

| Topic | When to Consult | File |
|-------|-----------------|------|
| Core Principles | Determinism/idempotency review | [core-principles.md](./reference/core-principles.md) |
| Workflow Boundaries | Activity vs child workflow decision | [workflow-boundaries.md](./reference/workflow-boundaries.md) |
| Signal vs Update | Choosing between signal and update for external input | [signals-queries-updates.md](./topics/signals-queries-updates.md) |
| Notation Examples | Control flow, handlers, timers, nexus in TWF | [notation-examples.md](./reference/notation-examples.md) |
| Notation Reference | All TWF syntax constructs | [notation-reference.md](./reference/notation-reference.md) |
| Design Checklist | Final verification before presenting | [design-checklist.md](./reference/design-checklist.md) |
| Anti-Patterns | Common Temporal design mistakes | [anti-patterns.md](./reference/anti-patterns.md) |
| Common Errors | Troubleshooting `twf check` | [common-errors.md](./reference/common-errors.md) |
| Primitives Reference | Temporal primitive lookup | [primitives-reference.md](./reference/primitives-reference.md) |
| Workers & Task Queues | Worker grouping, task queue routing, deployment | [task-queues.md](./topics/task-queues.md) |
| Nexus | Cross-namespace communication | [nexus.md](./topics/nexus.md) |
| Editor Setup | VS Code/Cursor extension | [editor-setup.md](./reference/editor-setup.md) |

Topic deep-dives are in `reference/` and `topics/` — consult as needed during design.
