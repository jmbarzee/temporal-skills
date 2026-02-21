# Parser & LSP Code Review

## Group 1: AST Reference Model — String/Pointer Duality

**Severity: critical (foundational)** | Files: `ast/ast.go`, `resolver/resolver.go`

The AST has **three inconsistent patterns** for references:

| # | Finding | Location |
|---|---------|----------|
| 1a | **Repeated Name+Resolved pairs.** `ActivityCall`, `WorkflowCall`, `SignalTarget`, `UpdateTarget`, `ActivityTarget`, `WorkflowTarget` all have `Name string` + `Resolved *ConcreteType`. This is a repeated structural pattern with no shared abstraction. | `ast.go:142-244` |
| 1b | **Duplicated Nexus resolution fields.** `NexusCall` and `NexusTarget` each carry 4 identical Resolved fields (`ResolvedEndpoint`, `ResolvedEndpointNamespace`, `ResolvedService`, `ResolvedOperation`). The resolver assigns them with identical code in two places. | `ast.go:246-257, 464-477`, `resolver.go:280-285, 488-493` |
| 1c | **Unresolved references.** `SetStmt.Name`, `UnsetStmt.Name`, `IdentTarget.Name`, and `NexusOperation.WorkflowName` reference definitions by string but have **no Resolved field**. The resolver validates they exist but doesn't store the pointer, so downstream consumers (LSP hover, go-to-def) must re-resolve. | `ast.go:416-429, 261-264, 444-448` |
| 1d | **WorkerRef uses interface for Resolved.** `WorkerRef.Resolved` is `Definition` (interface) while all other Resolved fields use concrete pointers. Consumers must type-assert. | `ast.go:64-68` |
| 1e | **`ResolvedEndpointNamespace string`** on NexusCall/NexusTarget is a string that names a NamespaceDef, but it's not a pointer to the NamespaceDef — it's the only Resolved field that remains a string. | `ast.go:254, 475` |

**Possible approaches** (for discussion):

- **A: Generic `Ref[T]` type** — `type Ref[T any] struct { Name string; Resolved T }` eliminates boilerplate. `ActivityCall.Activity Ref[*ActivityDef]` replaces `Name string` + `Resolved *ActivityDef`. The resolver sets `.Resolved` uniformly.
- **B: Extract `NexusResolution` struct** — shared between NexusCall and NexusTarget, assigned once.
- **C: Add missing Resolved fields** — `SetStmt`, `UnsetStmt`, `IdentTarget`, `NexusOperation` gain Resolved pointers for consistency.

---

## Group 2: Resolver Boilerplate & Duplication

**Severity: moderate** | Files: `resolver/resolver.go`

| # | Finding | Location |
|---|---------|----------|
| 2a | **Six near-identical resolve blocks** in `resolveStatement` and `resolveAsyncTarget`: lookup in map, set Resolved or append error. Same 8-line pattern repeated for ActivityCall, WorkflowCall, SignalTarget, UpdateTarget, ActivityTarget, WorkflowTarget. | `resolver.go:254-278, 440-487` |
| 2b | **Identical nexus assignment** in two places (resolveStatement for NexusCall, resolveAsyncTarget for NexusTarget). | `resolver.go:280-285, 488-493` |
| 2c | **`resolveWorkerRefs` generic adds complexity without benefit.** The `T ast.Definition` constraint isn't leveraged beyond map lookup. | `resolver.go:537-552` |

If Group 1 introduces `Ref[T]`, the resolver can use a single generic helper `resolveRef[T](ref *Ref[T], defs map[string]T, ...)` — collapsing all six blocks into one.

---

## Group 3: LSP Server — Duplicated Tree Traversal

**Severity: moderate** | Files: `internal/server/hover.go`, `symbols.go`, `references.go`, `code_actions.go`, `definition.go`, `folding.go`, `rename.go`

| # | Finding | Location |
|---|---------|----------|
| 3a | **`findNodeAtLine` reimplemented 3+ times** across hover.go, symbols.go, and references.go. Each handler walks the definition list and matches by line number independently. | `hover.go:40-156`, `symbols.go`, `references.go` |
| 3b | **`signatureFor` is 182 lines** with 40+ case branches. Should split by node category. | `hover.go:158-341` |
| 3c | **Repetitive definition-type matching** (WorkflowDef/ActivityDef/NexusServiceDef/WorkerDef/NamespaceDef) in `collectReferences` (103 lines), `resolvedTarget`, rename handlers. | `references.go:109-214`, `definition.go:35-102`, `rename.go:8-69` |
| 3d | **`lastLineInStmts` defined in symbols.go but called from folding.go.** Cross-file private helper in wrong home. | `folding.go:120-130`, `symbols.go:264-272` |
| 3e | **Dead code: `inlayHintHandler`** always returns nil. | `inlay_hints.go:1-13` |

---

## Group 4: AST Walk Visitor Gaps

**Severity: moderate** | Files: `ast/walk.go`, `ast/walk_test.go`

| # | Finding | Location |
|---|---------|----------|
| 4a | **`WalkStatements` never visits AsyncTarget nodes.** AwaitStmt, PromiseStmt, AwaitOneCase all contain AsyncTarget but the walker doesn't recurse into them. Any analysis pass using Walk misses these references. | `walk.go:1-61` |
| 4b | **SwitchCase vs AwaitOneCase inconsistency**: AwaitOneCase implements `stmtNode()` but SwitchCase does not. Both are containers for statements. | `ast.go:286-308` |

---

## Group 5: Parser — Nexus/Target Parsing Duplication

**Severity: moderate** | Files: `parser/statements.go`, `parser/nexus.go`

| # | Finding | Location |
|---|---------|----------|
| 5a | **`parseDetachableNexusTarget` and `parseNexusCallTarget` share ~90% code** (endpoint, service, operation, args parsing). | `statements.go:280-323` |
| 5b | **`parseWorkflowOrNexusTarget` and `parseDetachableNexusTarget` both handle nexus dispatch**, creating two entry points for nexus parsing with inconsistent error position handling. | `statements.go:134-162, 234-276` |
| 5c | **`collectRawUntil` column-spacing math assumes same line** — fails silently for tokens on different lines. | `helpers.go:58-81` |

---

## Group 6: Token Table Brittleness

**Severity: minor** | Files: `token/token.go`

| # | Finding | Location |
|---|---------|----------|
| 6a | **`tokenTable` array indexed by iota** — insertion or reordering silently corrupts lookups with no compiler error. | `token.go:100-152` |
| 6b | **No programmatic way to query "all keywords"** — only comments group them. | `token.go:46-90` |

---

## Group 7: CMD Nil Safety & Duplication

**Severity: moderate** | Files: `cmd/twf/files.go`, `parse.go`, `check.go`, `symbols.go`

| # | Finding | Location |
|---|---------|----------|
| 7a | **Nil pointer dereference risk**: `parseFiles()` can return nil file on read error, but callers (`parse.go:34`, `check.go`, `symbols.go`) don't check. | `files.go:45-54`, `parse.go:34` |
| 7b | **Identical error-printing loop** repeated in 3 command handlers. | `check.go:28-30`, `parse.go:29-31`, `symbols.go:94-96` |

---

## Group 8: DocumentStore Concurrency

**Severity: critical (but narrow)** | Files: `internal/server/document.go`

| # | Finding | Location |
|---|---------|----------|
| 8a | **`DocumentStore.Update()` runs analysis outside the lock** — concurrent reads can see partially-analyzed state (parsed but not yet resolved). | `document.go:39-89` |

---

## Recommended Execution Order

1. **Group 1** (AST reference model) — foundational; shapes all downstream work
2. **Group 2** (Resolver boilerplate) — directly follows from Group 1 changes
3. **Group 4** (Walk visitor gaps) — small, self-contained, improves correctness
4. **Group 1c addendum** (Add missing Resolved fields) — may combine with Group 2
5. **Group 5** (Parser nexus duplication) — moderate risk, self-contained
6. **Group 3** (LSP server duplication) — large scope, can be incremental
7. **Groups 6, 7, 8** — independent, lower priority
