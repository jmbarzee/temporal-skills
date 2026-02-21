# Temporal Workflow (.twf)

A language-agnostic DSL for Temporal workflows — capturing workflow structure, activity boundaries, and Temporal primitives before writing SDK code.

`.twf` serves two goals:

1. **Document Temporal Architectures** — Describe production-scale systems with namespaces, workers, workflows, activities, and Nexus services in a single readable notation.
2. **Facilitate AI-Driven Development** — Give AI agents a structured, parseable representation they can design against and translate into SDK code.

## Features

### Language Server

Full language server with real-time diagnostics:

- **Parse & resolve errors** — undefined activities, duplicate definitions, temporal keywords in wrong context
- **Symbol resolution** — activity calls, workflow calls, signals, queries, updates, promises, and conditions are all cross-referenced
- **Syntax highlighting** — keywords, types, operators, durations, and comments
- **Bracket matching and code folding**
- **Completions, hover, go-to-definition, references, and rename**

### Workflow Visualizer

Interactive visualization of `.twf` files, accessible from the editor title bar or command palette:

- **Visualize file** — parses the current `.twf` file and renders workflows, activities, and their relationships
- **Visualize folder** — renders all workflows across multiple `.twf` files in a folder
- **Live refresh** — updates automatically when you save a `.twf` file
- **Focused view** — follows the active editor, highlighting the workflows defined in the current file

Two complementary views:

- **Tree View** — Every definition rendered as a collapsible, color-coded block. Expand a workflow call to see the target workflow's body inline. Filter by file, definition type, or search by name.
- **Graph View** — A force-directed graph showing relationships across namespaces, workers, and workflows. Semantic zoom lets you switch between abstraction levels. Interactive force-tuning controls and animated transitions.

### AI Design Skills

Installs Temporal design skills for Cursor's AI agent:

- **Workflow design** — guides the agent through designing workflows with proper determinism, idempotency, and decomposition
- **Go authoring** — translates `.twf` designs into Temporal Go SDK code

### `twf` CLI

The bundled `twf` binary is also available as a standalone CLI:

| Command | Description |
|---------|-------------|
| `twf check <file...>` | Parse and validate `.twf` files |
| `twf parse <file...>` | Output the AST as JSON |
| `twf symbols <file...>` | List workflows and activities with signatures |
| `twf lsp` | Start the language server (stdio) |

## Temporal Features

The TWF notation covers the core Temporal feature set:

| Feature | TWF Construct | Purpose |
|---------|---------------|---------|
| Namespaces | `namespace` | Define deployment topology — workers and nexus endpoints |
| Workers | `worker` | Group workflows, activities, and nexus services into deployment units |
| Workflows | `workflow` (definition) | Deterministic orchestration with signals, queries, and updates |
| Activities | `activity` | Side-effecting operations with retry and timeout options |
| Child Workflows | `workflow` (call) | Decompose into independent sub-workflows |
| Signals | `signal` | Async fire-and-forget input to a running workflow |
| Queries | `query` | Synchronous read of workflow state |
| Updates | `update` | Synchronous mutation with a return value |
| Timers | `timer` | Durable sleep that survives restarts |
| Promises | `promise` | Non-blocking async operations, awaited later |
| Conditions | `condition` / `set` / `unset` | Named boolean awaitables for coordination |
| Fan-out / Fan-in | `await all` | Run operations concurrently, wait for all |
| Racing / Select | `await one` | Race between signals, timers, activities, and more |
| Control Flow | `if` / `for` / `switch` | Conditional logic, iteration, and branching |
| Detach | `detach workflow` / `detach nexus` | Fire-and-forget child workflows or nexus calls |
| Nexus Services | `nexus service` | Define sync and async service operation APIs |
| Nexus Endpoints | `nexus endpoint` | Route cross-namespace calls to workers within a namespace |
| Nexus Calls | `nexus` | Invoke operations across namespace boundaries |
| Continue-as-New | `close continue_as_new` | Reset history for long-running workflows |
| Heartbeats | `heartbeat` | Report activity progress, detect worker death |
| Options | `options:` | Task queues, timeouts, retry policies, priority |
| Workflow Termination | `close complete` / `close fail` | Explicit workflow exit with status |

## Installation

Search for **"Temporal Workflow"** in the VS Code or Cursor extension marketplace.

The extension bundles the `twf` binary (language server + parser). No additional setup required.

## Commands

| Command | Description |
|---------|-------------|
| **TWF: Visualize Workflow** | Open the interactive visualizer for the current `.twf` file |
| **TWF: Visualize All Workflows in Folder** | Visualize all `.twf` files in a folder |
| **TWF: Install Temporal Skills** | Re-install AI design skills to `~/.cursor/skills/` |

## License

MIT
