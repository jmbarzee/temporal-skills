# temporal-skills

A DSL (`.twf`) and toolchain for designing, visualizing, and code-generating Temporal workflows. See [README](./README.md) for full details.

## Project Layout

```
tools/lsp/              Go parser, resolver, validator, LSP server
  parser/token/         Token types and lexer vocabulary
  parser/lexer/         Indentation-aware lexer
  parser/parser/        Recursive-descent parser → AST
  parser/ast/           AST node types, JSON serialization, walker
  parser/resolver/      Name resolution (string refs → pointers)
  internal/server/      LSP server (hover, completions, diagnostics, etc.)
  cmd/twf/              CLI binary (check, parse, symbols, lsp)
tools/visualizer/       React + TypeScript webview (Tree View, Graph View)
packages/               VS Code / Cursor extension
skills/                 AI skill definitions (design, author-go)
```

## Project Status

This project is **pre-v1 and in active greenfield development**. The priority is elegant, correct representation — not stability.

**Breaking changes are expected and welcome.** Do not waste effort on backwards compatibility shims, deprecated field aliases, or migration paths. When a better design emerges, adopt it directly.

**Document breaking changes** in `AST_REVISIONS.md` so the visualizer team can propagate changes to the TypeScript layer. The parser's JSON output is the contract between Go and TypeScript — when it changes, both sides update together.

The current revision plan is tracked in [`AST_REVISIONS.md`](./AST_REVISIONS.md).

## Dependency Map

Changes propagate downstream along this graph. Each edge has a named contract:

```
DSL grammar (tools/lsp/LANGUAGE_SPEC.md)
  └─► Parser (tools/lsp/)
        │  contract: token types, AST node types, resolver error model
        ├─► LSP Server (tools/lsp/internal/server/)
        │     contract: Go AST types + resolver API (same module)
        ├─► Visualizer (tools/visualizer/)
        │     contract: JSON output of `twf parse` and `twf symbols`
        ├─► Skill: Design (skills/design/)
        │     contract: DSL syntax and semantics as in LANGUAGE_SPEC.md
        │     └─► Skill: Author-Go (skills/author-go/)
        │           contract: Design skill semantics + Temporal Go SDK mapping
        └─► VS Code Extension (packages/vscode/)
              contract: LSP protocol responses + JSON output
```

When a layer changes, the contracts it exposes determine what needs to update downstream. AST field renames and JSON schema changes are the most common sources of cascading work.

## Development Commands

These project commands drive the development loop. Invoke with `/project:<name>`:

| Command | Purpose |
|---------|---------|
| `dev-cycle` | Full orchestrated loop: review → group → execute → document → propagate |
| `review-parser-internals` | Deep review of Go parser, AST, resolver implementation |
| `review-parser-output` | Review JSON contract from the consumer's perspective |
| `review-visualizer` | Review TypeScript visualizer against current JSON contract |
| `review-skills` | Review design and author-go skills for DSL accuracy |
| `propagate-changes` | Assess and plan downstream updates for changes in AST_REVISIONS.md |

**Start here for a new cycle:** `/project:dev-cycle`
**Start here for targeted work:** pick the specific review command for the layer you're focused on.
