---
name: temporal-go-author
description: Generate Go code from .twf workflow designs using the Temporal Go SDK. Use when implementing workflows defined in TWF, producing compilable Go packages from DSL specifications.
---

# Temporal Go Author

Generate functioning Go code from `.twf` (Temporal Workflow Format) files using the Temporal Go SDK. The primary goal is producing code that compiles, runs, and correctly implements the workflow design. Always produce `.go` files as deliverables.

---

## Core Principles

**Root-down generation.** Start from root workflows (no parent), work down to children, then activities, then types. Each layer is constrained by what the layer above needs. This is the general direction, not a strict sequence — when a decision can't be made confidently, mark it as open, continue with what you know, and revisit when context is clearer.

**Write only what is needed.** No speculative fields, no extra types, no over-generation. The minimum bridge between DSL intent and working Go.

**Prefer imports over generation.** Check `go.mod` and existing project code first. Use well-known libraries when types match. Only generate stubs for application-specific types.

**Iterative type resolution.** Work from certainty outward. Explicit signatures first, then derive from constructors/field access, then defer the rest. Revisit deferred types as surrounding code solidifies. See [types.md](./reference/types.md).

**User as decision-maker.** The skill owns execution; the user owns consequential choices. Handle mechanical mappings, SDK boilerplate, and compilation fixes autonomously. Surface dependency choices, ambiguous domain logic, and architectural direction to the user — present specific options with tradeoffs, not open-ended questions. Revising a previously confirmed decision (type, interface, dependency) always requires user approval: present what was decided, what new information conflicts, and proposed alternatives.

---

## Process

### 1. Context Gathering

- Read `.twf` files in scope
- Examine `go.mod`, existing project code, and conventions
- Ask the user about project context, domain, key dependencies — brief, targeted questions, not an interrogation

### 2. Dependency Resolution

Scan all activities in the `.twf` and identify external integration points. Most activities are thin wrappers around calls to external systems — the activity itself is simple, but the dependency behind it is not.

For each activity:
- **Categorize:** external API call, storage operation, protocol client, or pure logic
- **Check existing code:** does the project already have a client/library for this?
- **Check `go.mod`:** is a relevant SDK already imported?
- **If unresolved:** suggest specific options with tradeoffs to the user
- **Read the chosen dependency's API:** identify the method the activity will call, then trace its signature to concrete types. See [types.md](./reference/types.md) for the full resolution strategy.

This step does not need to resolve everything. If a dependency choice is unclear or blocked on another decision, defer it and continue. But resolve as many as possible early — it prevents expensive rework later.

**Deliverable:** a dependency map, presented to the user for confirmation before Layer 1 begins.

A dependency is resolved when you can write the call expression with verified types. The map should include the method, every parameter type, and the return type — all confirmed from `go doc` or source, not inferred from names.

```
Example dependency map:
  ProcessPayment → stripe-go
    paymentintent.New(params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error)
  SendEmail → sendgrid-go
    client.SendWithContext(ctx, mail *sgmail.SGMailV3) (*rest.Response, error)
  GetOrder → database/sql (no external dependency)
  CalculateTotal → pure logic (no dependency)
```

### 3. Planning

- Identify root workflows (not called as children by other workflows in the `.twf` file)
- Outline the generation order: roots → children → activities → types
- Review the dependency map against planned type signatures — flag any conflicts
- If the dependency map is incomplete, note which decisions are deferred
- Surface ambiguities to the user

### 4. Generate + Verify Incrementally

General direction is root-down, with build checks between layers:

```
  ┌──────────────────────────────────────┐
  │ 1. Types + signatures  → go build    │
  │ 2. Workflow bodies     → go build    │
  │ 3. Activity impl       → go build    │
  │ 4. Worker wiring       → go build    │
  │ 5. Final               → go vet      │
  └──────────────────────────────────────┘
```

**Layer 1 — Types + signatures:** Generate type structs, interfaces, and workflow/activity function signatures (empty bodies returning zero values). Use the dependency map to inform interface shapes — they should reflect real SDK capabilities, not guesses. Run `go build`.

**Layer 2 — Workflow bodies:** Fill in orchestration logic (activity calls, child workflows, signals, timers, selectors). Run `go build`.

**Layer 3 — Activity implementations:** See [Activity Implementation Pattern](#activity-implementation-pattern). Produce the thin activity methods and concrete implementations behind the interfaces. Run `go build`.

**Layer 4 — Worker wiring:** Generate the `cmd/` entry point: construct dependencies, wire into activity struct, register workflows and activities, start the worker. Run `go build`.

**Layer 5 — Final:** Run `go vet` for correctness.

At each layer: consult [reference files](#reference-index) for DSL→Go mapping.

Revising a confirmed decision requires user approval. Present: what was decided, what new information conflicts, and proposed options.

After generation: present the code to the user for review. Incorporate feedback before considering the task done.

### When to Ask the User

**Handle yourself:** mechanical DSL→Go mappings, SDK boilerplate, type derivation from explicit signatures, standard options translation, compilation fixes.

**Ask the user:** dependency/library choices, activity implementation details beyond what the SDK dictates, ambiguous type resolution, domain logic not captured in the DSL, any revision to a previously confirmed decision.

**How to ask:** present specific options with tradeoffs, not open-ended questions.
> Example: "For ProcessPayment, should I use stripe-go (official SDK, matches your existing stripe dependency) or a generic HTTP client (more flexible, but more boilerplate)?"

---

## Activity Implementation Pattern

Activities follow a thin-wrapper pattern with dependency injection:

1. **Activity struct** holds interfaces as fields (one per external dependency)
2. **Activity methods** are thin translation layers: validate inputs, call the interface, translate the output
3. **Interfaces** are shaped by what activities need, informed by real SDK capabilities from the dependency map
4. **Concrete implementations** of those interfaces contain the real SDK integration — client construction, request/response conversion, error handling

The skill generates all four pieces. Activity methods and interfaces are mechanical. Concrete implementations require SDK knowledge from the dependency resolution step.

---

## Output Conventions

- Go files live alongside `.twf` sources or in a location the user specifies
- Each `.twf` maps to a Go package
- Within a package: one file per workflow, shared types file if needed, activity files grouped logically
- Package names derived from `.twf` filename (snake_case)
- Follow existing project conventions for naming, error handling, and imports
- Worker entry point goes in `cmd/` following Go conventions

---

## Reference Index

Read only what the current generation step requires.

### Definitions

| DSL Construct | Go Mapping | File |
|---------------|------------|------|
| `workflow Name(...)` | Workflow function | [workflow-def.md](./reference/workflow-def.md) |
| `activity Name(...)` | Activity function | [activity-def.md](./reference/activity-def.md) |

### Calls

| DSL Construct | Go Mapping | File |
|---------------|------------|------|
| `activity Name(args) -> result` | `workflow.ExecuteActivity` | [activity-call.md](./reference/activity-call.md) |
| `workflow Name(args) -> result` | `workflow.ExecuteChildWorkflow` | [workflow-call.md](./reference/workflow-call.md) |

### Handlers

| DSL Construct | Go Mapping | File |
|---------------|------------|------|
| `signal Name(params):` | Signal channel + selector | [signal-handler.md](./reference/signal-handler.md) |
| `query Name(params) -> (Type):` | `workflow.SetQueryHandler` | [query-handler.md](./reference/query-handler.md) |
| `update Name(params) -> (Type):` | `workflow.SetUpdateHandler` | [update-handler.md](./reference/update-handler.md) |

### Async Primitives

| DSL Construct | Go Mapping | File |
|---------------|------------|------|
| `await timer(duration)` | `workflow.Sleep` | [await-timer.md](./reference/await-timer.md) |
| `promise p <- ...` | Future (deferred `.Get`) | [promise.md](./reference/promise.md) |
| `state:` / `condition` / `set` / `unset` | `bool` + `workflow.Await` | [condition.md](./reference/condition.md) |

### Compound Async

| DSL Construct | Go Mapping | File |
|---------------|------------|------|
| `await all:` | `workflow.Go` + futures | [await-all.md](./reference/await-all.md) |
| `await one:` | `workflow.NewSelector` | [await-one.md](./reference/await-one.md) |

### Modifiers & Control Flow

| DSL Construct | Go Mapping | File |
|---------------|------------|------|
| `options: ...` | `ActivityOptions` / `ChildWorkflowOptions` | [options.md](./reference/options.md) |
| `nexus Endpoint Service.Op(args) -> result` | Nexus operation | [nexus.md](./reference/nexus.md) |
| `detach workflow ...` | Fire-and-forget (no `.Get`) | [detach.md](./reference/detach.md) |
| `if`/`for`/`switch`/`break`/`continue` | Go equivalents | [control-flow.md](./reference/control-flow.md) |
| `close complete`/`fail`/`continue_as_new` | `return` / `workflow.NewContinueAsNewError` | [close.md](./reference/close.md) |
| `x = expr` | Variable declaration/assignment | [assignment.md](./reference/assignment.md) |
| `heartbeat(details)` | `activity.RecordHeartbeat` | [heartbeat.md](./reference/heartbeat.md) |

### Types

| Topic | File |
|-------|------|
| Type resolution strategy | [types.md](./reference/types.md) |
