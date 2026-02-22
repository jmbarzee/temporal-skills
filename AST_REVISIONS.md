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

## Group 2: Walker Completeness

**Goal:** `WalkStatements` provides complete coverage of all reference-carrying nodes. No consumer needs to combine Walk + manual AsyncTarget extraction.

### 2a. Walk visits AsyncTarget nodes

Currently `WalkStatements` stops at the Statement level. `AwaitStmt`, `PromiseStmt`, and `AwaitOneCase` contain `AsyncTarget` but the walker doesn't recurse into them. Consumers must use `AsyncTargetOf()` separately.

Options:
- **(A)** Add an `AsyncTargetVisitor` callback to `WalkStatements` (second function parameter)
- **(B)** Make `WalkStatements` call `fn` on AsyncTargets by introducing a wrapper type that satisfies `Statement`
- **(C)** Add a separate `WalkAll` that visits both statements and targets

Recommendation: **(A)** — optional second callback keeps existing callers unchanged. Signature becomes:
```go
func WalkStatements(stmts []Statement, fn func(Statement) bool, opts ...WalkOption)
```
Where `WalkOption` can include `WithAsyncTargets(func(AsyncTarget) bool)`.

### 2b. SwitchCase vs AwaitOneCase consistency

`AwaitOneCase` implements `stmtNode()` (is a Statement). `SwitchCase` does not. Both are containers for child statements. Either both should be Statements or neither — decide and align.

### Parallelism

2a and 2b are independent — **two agents in parallel**.

### Files touched
- `parser/ast/walk.go`
- `parser/ast/walk_test.go`
- `parser/ast/ast.go` (if SwitchCase changes)

### Breaking changes
- Walk API signature change (backward compatible if using variadic options)
- SwitchCase may gain/lose Statement interface

---

## Group 3: Resolver Simplification

**Goal:** Collapse repeated resolution patterns now that all references use `Ref[T]`.

**Depends on:** Group 1

### 3a. Nexus resolution uses resolveRef chains

After Group 1, `resolveNexusRef` can become three sequential `resolveRef` calls instead of a custom 70-line function with inline map lookups. The cascading lookup (endpoint → service → operation) still exists but each step is a standard `resolveRef`.

### 3b. Collapse resolveStatement switch

Many cases in `resolveStatement` (13+ branches) follow the same pattern: extract Ref field, call resolveRef. Consider a `Resolvable` interface:
```go
type Resolvable interface {
    Refs() []*UntypedRef  // returns all references needing resolution
}
```

If type erasure is too complex, at minimum extract the body-recursion cases (AwaitAllBlock, SwitchBlock, IfStmt, ForStmt) into a shared helper, leaving only the resolve-specific logic in each case.

### 3c. Collapse resolveAsyncTarget switch

Same pattern as 3b — 8 cases, most calling resolveRef. After Group 1 makes NexusTarget use Ref[T], the nexus case becomes three resolveRef calls like any other.

### 3d. Unify resolveWorkerRefs

`resolveWorkerRefs` is nearly identical to a loop of `resolveRef` calls. Inline it or make it a thin wrapper.

### 3e. Document ErrorKind constants

The 23 `ErrorKind` constants lack documentation. Add a one-line comment to each explaining when it's used and whether it's a warning or error.

### Parallelism

3a, 3d, and 3e are independent — **three agents in parallel**.
3b and 3c are related (both simplify switches) — **one agent for both**.

### Files touched
- `parser/resolver/resolver.go`
- `parser/resolver/resolver_test.go`

### Breaking changes
- None (internal refactor). Resolver output unchanged.

---

## Group 4: JSON Serialization Redesign

**Goal:** JSON output is a clean, complete representation of the resolved AST. No dropped data, no field bleeding, no historical cruft.

**Depends on:** Groups 1, 2, 3

### 4a. Emit resolved refs everywhere

Currently `json.go` drops `Ref[T].Resolved` for ActivityCall, WorkflowCall, and WorkerDef refs. Only NexusCall emits resolved data. Fix: emit `resolvedRefJSON` (`{name, line, column}`) on every `Ref[T]` that has a non-nil Resolved.

This is the single highest-impact change for downstream consumers. The visualizer currently re-resolves every reference in TypeScript because the parser throws away its own work.

### 4b. AsyncTarget → discriminated union JSON

Replace the flat 22-field `asyncTargetFieldsJSON` with a nested structure:
```json
{
  "kind": "activity",
  "activity": {
    "name": "ValidateOrder",
    "args": "order",
    "result": "valid",
    "resolved": { "name": "ValidateOrder", "line": 45, "column": 1 }
  }
}
```

Each `kind` emits only its own fields. No more `workflowMode: "child"` on timer cases.

### 4c. Always emit empty arrays

Remove `omitempty` from `signals`, `queries`, `updates` on WorkflowDefJSON. Always emit `"signals": []` instead of omitting the key. Eliminates a class of runtime bugs in TypeScript consumers.

### 4d. Break apart marshalStatement

The 214-line switch in `marshalStatement` should become per-type marshal functions (`marshalActivityCall`, `marshalWorkflowCall`, etc.). Each function returns `json.RawMessage`.

### 4e. Extract marshalDeclList helper

WorkflowDef.MarshalJSON has three nearly-identical loops for signals/queries/updates. Extract a generic `marshalDeclList[T]()` helper.

### 4f. Add summary metadata

Add a top-level `summary` object to parse output:
```json
{
  "summary": {
    "namespaces": 1,
    "workers": 3,
    "workflows": 6,
    "activities": 8,
    "nexusServices": 2,
    "errors": 0
  },
  "definitions": [...]
}
```

### 4g. marshalDefinition exhaustiveness

Replace the silent `default` case in `marshalDefinition` with a panic or error. New Definition types must be explicitly handled.

### Parallelism

4a and 4c are small, surgical changes — **one agent for both**.
4b is the largest change (new JSON shape for all async targets) — **dedicated agent**.
4d and 4e are DRY refactors — **one agent for both**.
4f and 4g are independent additions — **one agent for both**.

### Files touched
- `parser/ast/json.go` (primary)
- `parser/ast/ast.go` (if File struct needs summary field)
- `cmd/twf/parse.go` (if top-level JSON wrapper changes)

### Breaking changes (document all for TS team)
- Every Ref[T] now emits `resolved` field when resolved
- AsyncTarget JSON is nested by kind, not flat
- `signals`, `queries`, `updates` always present (empty array, not absent)
- Top-level JSON gains `summary` object
- NexusTarget/NexusCall JSON fields restructured (from Group 1)
- CloseStmt reason is enum string, not freeform string

---

## Group 5: Multi-file sourceFile Tracking

**Goal:** When multiple `.twf` files are parsed together, each definition carries its `sourceFile`. Cross-file resolution works.

### 5a. Parser tracks file boundaries

When `parseFiles()` concatenates multiple files, inject markers or track byte offsets so each definition can be attributed to its source file. Options:
- Track a `[]fileBoundary` (filename + start line) and assign during parsing
- Parse files into separate ASTs, then merge definitions with sourceFile stamped

### 5b. AST carries sourceFile

Add `SourceFile string` to each Definition type (or to the embedded `Pos` struct if all nodes should carry it).

### 5c. JSON emits sourceFile

Include `"sourceFile": "orders.twf"` on each definition in JSON output.

### Parallelism

Sequential — 5a must come before 5b, 5b before 5c.

### Files touched
- `parser/ast/ast.go` (new field)
- `parser/ast/json.go` (emit field)
- `parser/parser/parser.go` (track boundaries)
- `cmd/twf/files.go` (pass file info to parser)

### Breaking changes
- New `sourceFile` field on all definitions in JSON output

---

## Group 6: Parser DRY

**Goal:** Reduce duplication in the parser package. Smaller files, shared helpers, no boolean-flag dispatch.

### 6a. Unify call parsers

`parseActivityCall`, `parseWorkflowCall`, `parseNexusCall` share identical structure: consume keyword → expect name → handle arrow → consume NEWLINE → parseOptionalOptionsLine. Extract a shared `parseCall` helper parameterized by keyword and AST constructor.

### 6b. Refactor parseAsyncTarget

Replace `allowArrows bool, allowDetach bool` parameters with separate functions: `parsePromiseTarget()` and `parseAwaitTarget()`. Each knows its own allowed syntax without flag-checking.

### 6c. Merge nexus target entry points

`parseDetachableNexusTarget` and `parseNexusCallTarget` share ~90% of their code. Extract shared nexus parsing logic.

### 6d. Split statements.go

At 903 lines, `statements.go` handles too many concerns. Split by category:
- `statements_calls.go` (activity/workflow/nexus calls)
- `statements_async.go` (await, promise, await-one, await-all)
- `statements_control.go` (if, for, switch)
- `statements_misc.go` (return, close, set, unset, raw, comment, break, continue)

### 6e. Remove goto in parseWorkflowDef

Replace `goto parseBody` in `definitions.go` with a labeled break or extracted helper function.

### 6f. Fix collectRawUntil column math

`helpers.go:collectRawUntil` column-spacing assumes tokens are on the same line. Fails silently for multi-line raw content.

### Parallelism

6a, 6b, 6c are related (all touch statement parsing) — **one agent**.
6d is a file-splitting refactor — **dedicated agent** (can run in parallel with the above since it's structural, not logic changes).
6e and 6f are small fixes — **one agent for both**.

### Files touched
- `parser/parser/statements.go` (primary)
- `parser/parser/definitions.go` (goto removal)
- `parser/parser/nexus.go` (nexus merge)
- `parser/parser/helpers.go` (collectRawUntil fix)

### Breaking changes
- None (internal refactor). Parser output unchanged.

---

## Group 7: CLI Improvements

**Goal:** Clean CLI layer with proper pipeline encapsulation, full symbol coverage, and safe error handling.

### 7a. Pipeline encapsulation

Create a top-level function (e.g., `parser.ParseAndResolve()` or a `Pipeline` struct) that runs parse → resolve → validate in one call. CLI commands become thin wrappers.

### 7b. twf symbols: full coverage

Add workers, namespaces, and nexus services to `twf symbols` output. Workers should list their registered workflow/activity/service names.

### 7c. Nil safety in cmd callers

`parseFiles()` can return nil file on read error. Callers (`parse.go`, `check.go`, `symbols.go`) must check before using.

### 7d. Dedup error printing

Extract shared error-printing loop used identically in three command handlers.

### 7e. Flag handling

Either accept `--json` as a no-op on `twf parse` (it always outputs JSON) or remove it from help text. Don't silently swallow flag parse errors.

### Parallelism

7a is the structural change — **dedicated agent**.
7b through 7e are independent fixes — **one agent for all four**.

### Files touched
- `cmd/twf/files.go`, `check.go`, `parse.go`, `symbols.go`, `main.go`
- Possibly a new `parser/pipeline.go` for the encapsulated function

### Breaking changes
- `twf symbols` output gains new definition types

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
