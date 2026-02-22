# Parser Internals Code Review

Review the Go implementation in `tools/lsp/` — the parser pipeline (`parser/`), LSP server (`internal/server/`), and CLI (`cmd/twf/`).

We value clear, elegant, functional code. The parser is the foundation of the developer experience — it must be exemplary.

## Context

Before starting, read `AST_REVISIONS.md` for the current revision plan. Check which groups have been completed. Your review should focus on **what remains** and **what's new** — don't re-report issues that are already tracked.

If prior review documents exist in `tools/lsp/` (e.g., `OLD_CODE_REVIEW.md`), read them to understand what was already found and what was fixed. Surface any regressions or findings that were missed.

## Review Rubric

Analyze the code through these lenses, in this order:

### 1. Architecture & Organization
- Package boundaries: loosely coupled, tightly cohesive
- File sizing: no god-files, no micro-files — each file earns its existence
- Dependency direction: clean layer separation (token <- lexer <- parser -> ast, resolver -> ast, server -> all)
- Public API surface: minimal exports, clear contracts between layers

### 2. Go Idioms & Best Practices
- Error handling: wrapping, propagation, sentinel vs typed errors — is it consistent?
- Interface usage: accept interfaces, return structs
- Naming: follows Go conventions (receivers, package names, exported identifiers)
- Zero-value usefulness, constructor patterns

### 3. Parser & Compiler Design Quality
- Recursive-descent patterns: clean, consistent, no layer bleeding
- AST node design: consistent structure, position-tracked, visitable
- `Ref[T]` usage: every cross-definition reference should use `Ref[T]`, never bare string + resolved pointer pairs
- Error recovery: graceful degradation, useful messages, synchronization points
- Walker completeness: does `WalkStatements` cover all reference-carrying nodes?

### 4. Code Health
- DRY without premature abstraction — repeated patterns that should be unified
- Functions that do one thing well
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

Each agent reads every file in its assigned package and evaluates against all four rubric lenses. The agent should return a **structured list of findings** — not prose. Each finding has: location, lens, severity, theme, and a 1-2 sentence description.

**Critical:** Agents must be skeptical. "Well-designed" is not a finding. Verify claims by checking how the code is actually *used* by downstream consumers (the LSP server, the CLI, the JSON output). A pattern that looks clean in isolation may be broken in context.

### Phase 2: Catalog

Synthesize all sub-agent findings into a single catalog. Each finding must include:
- **Location**: `file:function` or `file:line`
- **Lens**: which rubric section (1-4)
- **Severity**: `critical` | `moderate` | `minor`
- **Theme**: a short grouping label
- **Finding**: 1-2 sentences describing the issue and why it matters

Cross-reference against `AST_REVISIONS.md`. Drop findings that are already tracked there. Flag any finding that *contradicts* a planned revision.

### Phase 3: Group & Prioritize

- Group findings by **theme**, not by file
- Order theme-groups by:
  1. Critical severity first
  2. Foundational changes before downstream
  3. Smaller, lower-risk groups first within the same priority tier
- Present the grouped plan as a numbered list with:
  - Theme name and severity summary
  - List of findings in that group
  - Estimated scope (which files are touched)
  - What can be parallelized within the group

**STOP here. Present the plan and wait for approval. To execute, invoke `/project:address-review`.**

## Constraints

- **Never mix discovery and fixing.** Complete Phase 2 fully before proposing any changes.
- **Prefer surgical changes over sweeping rewrites.** Refactors should be incremental and reviewable.
- **Preserve test coverage.** If you change behavior, update or add tests.
- **Flag ambiguity.** If a finding has multiple valid approaches, present the tradeoffs and ask.
- **Stay in scope.** Review `tools/lsp/` only. Don't wander into skills, docs, or the VS Code extension.
- **No backwards compatibility.** This is pre-v1. If a better design exists, propose it. Document breaking changes in `AST_REVISIONS.md`.
