# temporal-skills

A language-agnostic DSL (`.twf`) for Temporal workflows — capturing workflow structure, activity boundaries, and Temporal primitives before writing SDK code.

`.twf` serves two goals:

1. **Document Temporal Architectures** — Describe production-scale systems with namespaces, workers, workflows, activities, and Nexus services in a single readable notation.
2. **Facilitate AI-Driven Development** — Give AI agents a structured, parseable representation they can design against and translate into SDK code.

## Install

Install **Temporal Workflow (.twf)** from the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=jmbarzee.twf-syntax) or [Open VSX](https://open-vsx.org/extension/jmbarzee/twf-syntax). The extension bundles:

- **AI Skills** — automatically installed to `~/.cursor/skills/` so Cursor's agent can use them immediately
- **`twf` CLI** — parser, validator, and language server for `.twf` files, added to your terminal PATH
- **Syntax highlighting** and **workflow visualization** for `.twf` files

## Tools

### `twf` CLI — Parser & Language Server

A Go binary providing parsing, validation, symbol extraction, and a full LSP server.

| Command | Description |
|---------|-------------|
| `twf check <file...>` | Parse and validate `.twf` files, reporting errors |
| `twf parse <file...>` | Output the AST as JSON (always emits partial AST, even with errors) |
| `twf symbols <file...>` | List workflows and activities with their signatures |
| `twf lsp` | Start the language server over stdio |

Options: `--json` (JSON output where applicable), `--lenient` (continue past resolve errors).

The language server provides real-time diagnostics, symbol resolution, completions, hover, go-to-definition, references, rename, code actions, folding, inlay hints, semantic tokens, and signature help.

### Workflow Visualizer

A React + TypeScript webview (Vite-built) that renders parsed `.twf` ASTs. Runs standalone for development or embedded in the VS Code extension.

**[Tree View](./tools/visualizer/spec/TREE_VIEW.md)** — Renders every definition as a collapsible, color-coded block in a vertical list. Supports inline expansion of cross-references (a workflow call expands to show the target workflow's body in place), file filtering, definition type toggles, and search. Full light/dark theme support.

**[Graph View](./tools/visualizer/spec/GRAPH_VIEW.md)** — A force-directed graph showing how definitions relate to each other. Three node levels (Namespace → Worker → Workflow) form a containment hierarchy with dependency edges derived by graph coarsening. Semantic zoom lets you select which abstraction levels are visible. Includes interactive force-tuning controls, animated level transitions, and hover/selection highlighting.

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

## Skills

- **[design](./skills/design/SKILL.md)** — Design Temporal workflows using `.twf`, a language-agnostic DSL with parser/LSP tooling and visualization
- **[author-go](./skills/author-go/SKILL.md)** — Generate Go code from `.twf` workflow designs using the Temporal Go SDK

### Planned

- **Implementers** — More authorship skills (TypeScript, Python, Java, etc.)
- **Translators** — Analyze existing systems (event-based architectures, SQS/Lambda, etc.) and generate equivalent DSL designs
- **Debuggers & Optimizers** — Assist with debugging, profiling, and optimizing existing Temporal workflows

## Repository Structure

```
packages/    VS Code / Cursor extension
tools/
  lsp/       Go parser, resolver, validator, and language server (twf CLI)
  visualizer/React workflow visualizer (tree view + graph view)
skills/      AI skill definitions (SKILL.md + reference docs)
```

## Development

```bash
# Build everything (current platform)
make build

# Run Go tests
make test

# Package a local .vsix
make package

# Package for all platforms (CI)
make package-all

# Publish to marketplaces
VSCE_TOKEN=... OVSX_TOKEN=... make publish
```
