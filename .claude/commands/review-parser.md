# Parser & LSP Code Review

Review the Go implementation in `tools/lsp/` — the parser pipeline (`parser/`), LSP server (`internal/server/`), and CLI (`cmd/twf/`).

We value clear, elegant, functional code. The parser is the foundation of the developer experience — it must be exemplary.

## Review Rubric

Analyze the code through these lenses, in this order:

### 1. Architecture & Organization
- Package boundaries: loosely coupled, tightly cohesive
- File sizing: no god-files, no micro-files — each file earns its existence
- Dependency direction: clean layer separation (token ← lexer ← parser → ast, resolver → ast, server → all)
- Public API surface: minimal exports, clear contracts between layers
- Are packages doing too many jobs? Could anything be split or merged?

### 2. Go Idioms & Best Practices
- Error handling: wrapping, propagation, sentinel vs typed errors — is it consistent?
- Interface usage: accept interfaces, return structs
- Naming: follows Go conventions (receivers, package names, exported identifiers)
- Zero-value usefulness, constructor patterns, option patterns
- Goroutine safety where applicable

### 3. Parser & Compiler Design Quality
- Does the recursive-descent parser follow established patterns?
- AST node design: consistent structure, position-tracked, visitable
- Error recovery: graceful degradation, useful messages, synchronization points
- Clean separation: lexing vs parsing vs semantic analysis — no layer bleeding
- Is the resolver doing work that belongs in the parser, or vice versa?

### 4. Code Health
- DRY without premature abstraction — repeated patterns that should be unified
- Functions that do one thing well
- No split-brain logic (related behavior scattered across unrelated files)
- Testability: can components be tested in isolation?
- Dead code, vestigial patterns, or remnants of previous designs

## Workflow

**Follow this phased approach strictly. Do not skip or combine phases.**

### Phase 1: Explore

Use sub-agents to read ALL Go files in `tools/lsp/` in parallel. One agent per package:
- `parser/token/`
- `parser/lexer/`
- `parser/parser/`
- `parser/ast/`
- `parser/resolver/`
- `internal/server/`
- `cmd/twf/`

Each agent reads every file in its assigned package and evaluates against all four rubric lenses. The agent should return a structured list of findings — not prose.

### Phase 2: Catalog

Synthesize all sub-agent findings into a single catalog. Each finding must include:
- **Location**: `file:function` or `file:line`
- **Lens**: which rubric section (1–4)
- **Severity**: `critical` | `moderate` | `minor`
- **Theme**: a short grouping label (e.g., "error propagation", "AST node consistency", "naming drift", "layer bleeding")
- **Finding**: 1–2 sentences describing the issue and why it matters

### Phase 3: Group & Prioritize

- Group findings by **theme**, not by file — a theme like "error handling consistency" may span multiple packages
- Order theme-groups by:
  1. Critical severity first
  2. Foundational changes before downstream (e.g., fix AST node design before fixing resolver usage of those nodes)
  3. Smaller, lower-risk groups first within the same priority tier
- Present the grouped plan as a numbered list of reform sets, each with:
  - Theme name
  - Severity summary
  - List of findings in that group
  - Estimated scope (which files are touched)

**STOP here. Present the plan and wait for approval before any code changes.**

### Phase 4: Execute (only after approval)

- Work through one theme-group at a time
- After each group: run `go build ./...` and `go test ./...` from `tools/lsp/` to validate
- Present a brief summary of what changed before moving to the next group
- If a change in one group conflicts with a planned change in another, flag it

## Constraints

- **Never mix discovery and fixing.** Complete Phase 2 fully before proposing any changes.
- **Prefer surgical changes over sweeping rewrites.** Refactors should be incremental and reviewable.
- **Preserve test coverage.** If you change behavior, update or add tests.
- **Flag ambiguity.** If a finding has multiple valid approaches, present the tradeoffs and ask — don't choose silently.
- **Stay in scope.** Review `tools/lsp/` only. Don't wander into skills, docs, or the VS Code extension.
