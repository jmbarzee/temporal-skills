# AST Revisions

A guide for refactoring the parser pipeline (`tools/lsp/`) in cohesive, reviewable sets. Each group is a checkpoint — complete it, validate with `go build ./...` and `go test ./...`, then move to the next.

Within each group, **parallelizable** work is marked. Sub-agents should not cross group boundaries.

---

## Group 1: Ref[T] Completeness ✅ COMPLETED

**Goal:** Every AST field that references another definition by name uses `Ref[T]`. No more bare string + separate Resolved pointer pairs.

**Why first:** This is the foundational type change. Groups 3, 4, and 9 all depend on a consistent reference model.

### 1a. NexusTarget → Ref[T]

`ast.go` NexusTarget currently stores:
```
Endpoint   string
Service    string
Operation  string
ResolvedEndpoint          *NamespaceEndpoint
ResolvedEndpointNamespace string
ResolvedService           *NexusServiceDef
ResolvedOperation         *NexusOperation
```

Replace with:
```
Endpoint  Ref[*NamespaceEndpoint]
Service   Ref[*NexusServiceDef]
Operation Ref[*NexusOperation]
```

Drop `ResolvedEndpointNamespace` — derive namespace from the resolved endpoint's parent when needed.

### 1b. NexusCall → Ref[T]

Same transformation as 1a. NexusCall has identical string + Resolved fields. After this change, NexusCall and NexusTarget share the same reference pattern as ActivityCall/WorkflowCall.

### 1c. IdentTarget — add Resolved field

`IdentTarget.Name` is validated by the resolver (checked against promises and conditions) but the resolved pointer is discarded. Add:
```
Resolved IdentResolution  // union: *PromiseStmt or *ConditionDecl
```

Design choice: `IdentResolution` could be an interface or a struct with two pointer fields (one nil). Keep it simple — a struct with `Promise *PromiseStmt` and `Condition *ConditionDecl` avoids type assertions.

### 1d. CloseStmt.Reason → enum

Replace `Reason string` with:
```go
type CloseReason int
const (
    CloseComplete CloseReason = iota
    CloseFailWorkflow
    CloseContinueAsNew
)
```

### 1e. Resolver updates

After 1a–1d, update `resolver.go`:
- `resolveNexusRef` returns are assigned to `Ref[T].Resolved` fields instead of separate pointer fields
- `resolveAsyncTarget` NexusTarget case uses `resolveRef` for each of the three Ref fields
- `resolveStatement` NexusCall case uses same pattern
- IdentTarget resolution stores the pointer instead of discarding it

### Parallelism

1a and 1b are independent (different structs, same pattern) — **two agents in parallel**.
1c and 1d are independent of each other and 1a/1b — **can run in parallel** with 1a/1b.
1e depends on all of 1a–1d.

### Files touched
- `parser/ast/ast.go` (struct definitions)
- `parser/resolver/resolver.go` (resolution logic)
- `parser/resolver/resolver_test.go` (test updates)
- `parser/parser/statements.go` (parser constructs these nodes)
- `parser/parser/nexus.go` (nexus node construction)

### Breaking changes
- NexusTarget JSON shape changes (string fields → Ref[T] with name/resolved)
- NexusCall JSON shape changes (same)
- IdentTarget gains resolved info in output
- CloseStmt.Reason changes from string to enum string representation

---

## Group 2: Walker Completeness ✅ COMPLETED

**Goal:** `WalkStatements` provides complete coverage of all reference-carrying nodes. No consumer needs to combine Walk + manual AsyncTarget extraction.

### 2a. Walk visits AsyncTarget nodes ✅

Added functional options pattern to `WalkStatements`:
- `WalkOption` type, `walkConfig` struct, `WithAsyncTargets(func(AsyncTarget, Statement) bool)` option
- `WalkStatements` signature now accepts `...WalkOption` (backward compatible)
- `walkStatement` propagates config and invokes async target callback after visiting each statement
- `references.go` `collectRefsInStmts` migrated from `AsyncTargetOf` default-case pattern to `WithAsyncTargets`

### 2b. SwitchCase is now a Statement ✅

- Added `func (*SwitchCase) stmtNode() {}` for symmetry with `AwaitOneCase`
- Walker now visits `SwitchCase` nodes via `fn` before recursing into their bodies
- `AwaitOneCase` children now recurse via `walkStatement` instead of direct `fn()` call

### Files changed
- `parser/ast/ast.go` — SwitchCase gains stmtNode()
- `parser/ast/walk.go` — WalkOption, WithAsyncTargets, rewritten walkStatement
- `parser/ast/walk_test.go` — Updated SwitchBlock expectations, added WithAsyncTargets tests
- `internal/server/references.go` — Migrated to WithAsyncTargets

### Breaking changes
- Walk API signature change (backward compatible via variadic options)
- SwitchCase is now a Statement (walker visits it — consumers with exhaustive switches may need a case)

---

## Group 3: Resolver Simplification ✅ COMPLETED

**Goal:** Collapse repeated resolution patterns now that all references use `Ref[T]`.

**Depends on:** Group 1

### 3a. Nexus resolution uses resolveRef chains ✅

Extracted `resolveRefWithWarn` generic helper for the "empty map → warning, non-empty → resolve or error" pattern shared by endpoint and service resolution. Extracted `resolveNexusOperation` for operation lookup. `resolveNexusRefs` collapsed from 68 lines to ~10 lines.

### 3b + 3c. Walker-based statement traversal ✅

Replaced `resolveStatements` (loop), `resolveStatement` (12-case switch with manual body recursion), and `resolveAwaitOneCase` (separate method) with a single `resolveStatements` that uses `ast.WalkStatements` + `WithAsyncTargets`. The walker handles all traversal; the resolver switch only contains resolution-specific cases (ActivityCall, WorkflowCall, NexusCall, SetStmt, UnsetStmt). `resolveAsyncTarget` stays as-is — it has genuine per-type logic that isn't reducible.

### 3d. Unified resolveWorkerRefs ✅

`resolveWorkerRefs` now loops `resolveRef` directly. Dropped the `workerName` parameter — error messages use the standard `"undefined <kind>: <name>"` format. Worker error messages changed from `"worker X references undefined Y: Z"` to `"undefined Y: Z"`.

### 3e. Documented ErrorKind constants ✅

Added category headers and per-constant doc comments to all 23 ErrorKind constants, organized into: duplicate definition errors, undefined reference errors, nexus resolution errors, worker reference errors, and namespace reference errors.

### Files changed
- `parser/resolver/resolver.go`

### Breaking changes
- None (internal refactor). Resolver output unchanged except for worker error message wording (pre-v1, acceptable).

---

## Group 4: JSON Serialization Redesign ✅ COMPLETED

**Goal:** JSON output is a clean, complete representation of the resolved AST. No dropped data, no field bleeding, no historical cruft.

**Depends on:** Groups 1, 2, 3

### 4a. Emit resolved refs everywhere ✅

`activityCallJSON` and `workflowCallJSON` now include `Resolved *resolvedRefJSON`. Populated when `Ref[T].Resolved` is non-nil, matching the existing NexusCall and WorkerDef patterns. Also emitted on async target activity/workflow variants.

### 4b. AsyncTarget → discriminated union JSON ✅

Replaced flat 22-field `asyncTargetFieldsJSON` with nested `asyncTargetJSON` discriminated union. Each kind populates exactly one per-kind field (`timer`, `signal`, `update`, `activity`, `workflow`, `nexus`, `ident`). Per-kind types include resolved refs where applicable.

### 4c. Always emit empty arrays ✅

Removed `omitempty` from `signals`, `queries`, `updates` on `WorkflowDefJSON`. Always emits `[]` instead of omitting the key.

### 4d. Break apart marshalStatement ✅

Extracted each case from `marshalStatement` into named per-type functions (`marshalActivityCall`, `marshalWorkflowCall`, etc.). The switch is now a clean dispatch table. Default case returns `fmt.Errorf` instead of silently marshaling.

### 4e. Extract marshalDeclList helper ✅

Generic `marshalDeclList[D, J]` replaces three identical loops in `WorkflowDef.MarshalJSON`. Companion functions: `marshalSignalDecl`, `marshalQueryDecl`, `marshalUpdateDecl`.

### 4f. Add summary metadata ✅

Top-level JSON output now includes a `summary` object counting definitions by type:
```json
{"summary": {"namespaces": 1, "workers": 3, "workflows": 6, "activities": 8, "nexusServices": 2}, "definitions": [...]}
```

### 4g. marshalDefinition + marshalStatement exhaustiveness ✅

Both `marshalDefinition` and `marshalStatement` default cases now return `fmt.Errorf` instead of silently marshaling. New types get an immediate error signal.

### Files changed
- `parser/ast/json.go` (all changes)

### Breaking changes (for TS propagation)
- **`activityCall` and `workflowCall`** gain optional `resolved` field (`{name, line, column}`)
- **`signals`, `queries`, `updates`** always present on `workflowDef` (empty array, never omitted)
- **AsyncTarget JSON restructured**: flat fields replaced by nested `"target"` object with discriminated union by `kind`. Each kind has its own sub-object with only relevant fields. Affects `await`, `awaitOne` cases, and `promise` statements
- **Top-level JSON** gains `summary` object before `definitions`
- **Propagation needed** in:
  - `tools/visualizer/src/types/ast.ts` — update `ActivityCall`, `WorkflowCall`, `AwaitStmt`, `AwaitOneCase`, `PromiseStmt` interfaces
  - `tools/visualizer/src/components/blocks/AwaitBlocks.tsx` — update `getAwaitTargetDisplay` to read nested `target` object
  - `tools/visualizer/src/components/blocks/LeafBlocks.tsx` — update `PromiseBlock` to read nested `target` object

---

## Group 5: Multi-file sourceFile Tracking ✅ COMPLETED

**Goal:** When multiple `.twf` files are parsed together, each definition carries its `sourceFile`. Cross-file resolution works.

### Approach: Parse separately, merge, stamp

Instead of concatenating files and tracking boundaries, each file is parsed independently via `ParseFileAll()`. Definitions are stamped with `SourceFile` (basename) and merged into a single `*ast.File` for cross-file resolution. This gives per-file line numbers (matching editor behavior) instead of global offsets into concatenated input.

### 5a+5b. AST carries sourceFile ✅

Added `SourceFile string` to all 5 Definition types: `WorkflowDef`, `ActivityDef`, `WorkerDef`, `NamespaceDef`, `NexusServiceDef`. Added to each definition JSON struct with `json:"sourceFile,omitempty"` and populated in each `MarshalJSON` method.

### 5c. Per-file parsing in CLI ✅

Rewrote `parseFiles()` in `cmd/twf/files.go` to:
- Parse each file independently (no more concatenation)
- Stamp `SourceFile` on every definition
- Merge definitions into single `*ast.File`
- Resolve + validate on merged file
- Parse error messages prefixed with filename

### Files changed
- `parser/ast/ast.go` — `SourceFile string` on 5 Definition types
- `parser/ast/json.go` — `SourceFile` in 5 JSON structs + 5 MarshalJSON methods
- `cmd/twf/files.go` — rewritten for per-file parsing

### Breaking changes (for TS propagation)
- **All definitions** gain `sourceFile` field (always set to basename of source file)
- **Line numbers** are now per-file, not global offsets into concatenated input
- **Parse error messages** are now prefixed with filename

---

## Group 6: Parser DRY ✅ COMPLETED

**Goal:** Reduce duplication in the parser package. Smaller files, shared helpers, no boolean-flag dispatch.

### 6a. Unify call parsers ✅

Extracted `callParts` struct and `parseCallParts` helper for the shared IDENT→ARGS→ARROW→OPTIONS pattern. `parseActivityCall` and `parseWorkflowCall` are now thin wrappers (~10 lines each) that delegate to `parseCallParts` and construct the appropriate AST node.

### 6b. Refactor parseAsyncTarget — VALIDATED, NO CHANGES NEEDED

The `allowArrows bool, allowDetach bool` parameters are already correct. All call sites pass appropriate values. The flag-based approach is cleaner than separate functions here because the dispatch logic is shared.

### 6c. Merge nexus target entry points ✅

Merged `parseDetachableNexusTarget` and `parseNexusCallTarget` into single `parseNexusTarget(p, detach, allowArrows bool, pos)`. The detach validation is handled uniformly.

### 6d. Split statements.go ✅

Split the 856-line `statements.go` into 4 files:
- `statements_calls.go` — `callParts`, `parseCallParts`, `parseActivityCall`, `parseWorkflowCall`
- `statements_async.go` — await/promise/await-one/await-all parsers, async target parsers, nexus target
- `statements_control.go` — `parseSwitchBlock`, `parseIfStmt`, `parseForStmt`
- `statements_misc.go` — `parseSetStmt`, `parseUnsetStmt`, `parseReturnStmt`, `parseCloseStmt`, `parseBreakStmt`, `parseContinueStmt`, `parseRawStmt`

### 6e. Remove goto in parseWorkflowDef ✅

Replaced `goto parseBody` with labeled `declLoop:` and `break declLoop`. Idiomatic Go.

### 6f. Fix collectRawUntil column math — VALIDATED, NO CHANGES NEEDED

The column-based spacing implementation in `helpers.go` is correct. No issues found.

### Files changed
- `parser/parser/statements.go` — deleted (split into 4 files below)
- `parser/parser/statements_calls.go` — new
- `parser/parser/statements_async.go` — new
- `parser/parser/statements_control.go` — new
- `parser/parser/statements_misc.go` — new
- `parser/parser/definitions.go` — goto removal

### Breaking changes
- None (internal refactor). Parser output unchanged.

---

## Group 7: CLI Improvements ✅ COMPLETED

**Goal:** Clean CLI layer with proper pipeline encapsulation, full symbol coverage, and safe error handling.

### 7a. Pipeline encapsulation — SKIPPED (addressed by prior work)

Group 5 rewrote `parseFiles()` with a clean multi-file pipeline (parse each → stamp → merge → resolve → validate). The LSP's `document.analyze()` is already a clean 6-line function. The shared part (resolve + validate) is two function calls — a new package would be over-engineering.

### 7b. twf symbols: full coverage ✅

Added `WorkerDef`, `NamespaceDef`, `NexusServiceDef` to `extractSymbols()`. Workers list registered workflows/activities/services. Namespaces list workers and endpoints. NexusServices list operations with sync/async kind. Both text and JSON output updated.

### 7c. Nil safety in cmd callers ✅

Added nil check in `parse.go` before `json.MarshalIndent`. Outputs `null` JSON and returns exit code 1 on nil file. `check.go` and `symbols.go` already had nil checks.

### 7d. Dedup error printing ✅

Extracted `printErrors(errs []string)` in `files.go`. All three command handlers (`check.go`, `parse.go`, `symbols.go`) now use it.

### 7e. Flag handling ✅

Removed `--json` from global `Options:` section in usage text. It only applies to `twf symbols`.

### Files changed
- `cmd/twf/symbols.go` — full definition coverage + printErrors
- `cmd/twf/files.go` — printErrors helper
- `cmd/twf/parse.go` — nil safety + printErrors
- `cmd/twf/check.go` — printErrors
- `cmd/twf/main.go` — usage text fix

### Breaking changes
- `twf symbols` output gains `worker`, `namespace`, and `nexusService` kinds with sub-symbols

---

## Group 8: LSP Server Quality

**Goal:** Reduce duplication, fix correctness bugs, remove dead code.

### 8a. Shared AST query layer

Extract `findNodeAtLine` into a single implementation used by hover, symbols, references, definition, and rename handlers. Consider a `Query` struct or package that indexes definitions by line.

### 8b. Decompose signatureFor

Split the 182-line switch into per-node-type functions.

### 8c. Semantic token type safety

Replace magic constants (`semKeyword=0`, `modDeclaration=1<<0`) with typed enums. Ensure legend order changes produce compile errors, not silent misclassification.

### 8d. DocumentStore race condition

`Update()` runs analysis outside the write lock. Move analysis inside the lock, or use a copy-on-write pattern where the new Document is fully analyzed before being stored.

### 8e. Remove dead inlayHintHandler

Delete the stub that returns `(nil, nil)`.

### Parallelism

8a is the largest change — **dedicated agent**.
8b and 8c are independent — **one agent each, or combined**.
8d and 8e are small — **one agent for both**.

### Files touched
- `internal/server/hover.go`, `references.go`, `symbols.go`, `definition.go`, `rename.go`
- `internal/server/semantic_tokens.go`
- `internal/server/document.go`
- `internal/server/inlay_hints.go` (delete)

### Breaking changes
- None (LSP protocol unchanged)

---

## Group 9: twf deps Subcommand

**Goal:** Purpose-built dependency graph output for the Graph View. Pre-computed nodes, edges, containment, and coarsened projections.

**Depends on:** Groups 1–5 (clean AST, complete walker, resolved refs in JSON, sourceFile tracking)

### 9a. Dependency extraction

Walk all workflow bodies, handler bodies, and nexus operation bodies to extract dependency edges. Use the improved walker (Group 2) to cover AsyncTarget nodes in await/promise/awaitOne contexts.

### 9b. Containment hierarchy

Build parent→children relationships from namespace→worker→workflow/activity registration.

### 9c. Graph coarsening

Project workflow-level edges to worker-level and namespace-level. Remove self-loops. Track weight and derivedFrom.

### 9d. Output structure

```json
{
  "nodes": [...],
  "edges": [...],
  "containment": {...},
  "coarsened": { "workerEdges": [...], "namespaceEdges": [...] },
  "unresolved": [...],
  "summary": {...}
}
```

### 9e. Text output

Human-readable default (no `--json` flag) showing namespaces, edges, cross-worker dependencies, and unresolved references.

### Parallelism

9a–9c are sequential (each builds on prior).
9d and 9e are independent output formatters — **two agents in parallel** after 9a–9c.

### Files touched
- New `cmd/twf/deps.go`
- New `parser/deps/` package (or inline in cmd)
- `cmd/twf/main.go` (register command)

### Breaking changes
- New subcommand (additive, not breaking)

---

## Group 10: Minor Cleanup

**Goal:** Small fixes that don't fit elsewhere.

### 10a. Token table brittleness

`tokenTable` is an array indexed by iota. Reordering or inserting a const silently corrupts lookups. Consider a map or add a compile-time size assertion.

### 10b. Lexer: dedup dedent emission

`emitEOF` has two identical dedent-emission loops (lines 193-196 and 202-205). Extract a helper.

### 10c. LookupIdent case sensitivity

`LookupIdent` doesn't lowercase input before map lookup, but keyword map keys are lowercase. Document whether this is intentional or fix it.

### Parallelism

All three are independent — **three agents in parallel** (or one agent if overhead isn't worth it).

### Files touched
- `parser/token/token.go`
- `parser/lexer/lexer.go`
