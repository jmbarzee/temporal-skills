# Temporal Workflow (.twf)

Design Temporal workflows using a language-agnostic notation that captures workflow structure, activity boundaries, and Temporal primitives — before writing SDK code.

## Features

### Language Server

Full language server with real-time diagnostics:

- **Parse & resolve errors** — undefined activities, duplicate definitions, temporal keywords in wrong context
- **Symbol resolution** — activity calls, workflow calls, signals, queries, updates, promises, and conditions are all cross-referenced
- **Syntax highlighting** — keywords, types, operators, durations, and comments
- **Bracket matching and code folding**

### Workflow Visualizer

Interactive graph visualization of `.twf` files, accessible from the editor title bar or command palette:

- **Visualize file** — parses the current `.twf` file and renders workflows, activities, and their relationships
- **Visualize folder** — renders all workflows across multiple `.twf` files in a folder
- **Live refresh** — updates automatically when you save a `.twf` file
- **Focused view** — follows the active editor, highlighting the workflows defined in the current file

### AI Design Skills

Installs Temporal design skills for Cursor's AI agent:

- **Workflow design** — guides the agent through designing workflows with proper determinism, idempotency, and decomposition
- **Go authoring** — translates `.twf` designs into Temporal Go SDK code

## Temporal Features

The TWF notation covers the core Temporal feature set:

| Feature | TWF Construct | Purpose |
|---------|---------------|---------|
| Activities | `activity` | Side-effecting operations with retry and timeout options |
| Child Workflows | `workflow` | Decompose into independent sub-workflows |
| Signals | `signal` | Async fire-and-forget input to a running workflow |
| Queries | `query` | Synchronous read of workflow state |
| Updates | `update` | Synchronous mutation with a return value |
| Timers | `timer` | Durable sleep that survives restarts |
| Promises | `promise` | Non-blocking async operations, awaited later |
| Conditions | `condition` / `set` / `unset` | Named boolean awaitables for coordination |
| Fan-out / Fan-in | `await all` | Run operations concurrently, wait for all |
| Racing / Select | `await one` | Race between signals, timers, activities, and more |
| Detach | `detach workflow` | Fire-and-forget child workflows |
| Nexus | `nexus` | Cross-namespace workflow calls |
| Continue-as-New | `close continue_as_new` | Reset history for long-running workflows |
| Heartbeats | `heartbeat` | Report activity progress, detect worker death |
| Options | `options { }` | Task queues, timeouts, retry policies, priority |
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
