# Code Review: `tools/lsp/`

Findings grouped by theme, ordered by priority. Each group is self-contained — pick up any one independently.

---

## Group 1: Dead Code & Incomplete Stubs

**Severity: critical + moderate | Risk: low | Scope: small**

Remove unused code that adds confusion and maintenance burden.

| Location | Severity | Finding |
|---|---|---|
| `parser/lexer/lexer.go:339-348` | moderate | `LexError` type implements `error` but is never instantiated or returned. No function in the lexer API returns an error. Dead type. |
| `internal/server/inlay_hints.go` (entire file) | critical | `collectStatementHints()` never appends any hints — it only recurses into nested statements. The handler is a stub that generates nothing. Remove or complete. |
| `internal/server/code_actions.go:181-290` | critical | Four functions defined but never called: `findCallInStatements`, `findAwaitOneBlocks`, `findSignalsThatModify`, `containsAssignment`. Dead experiments. |

**Action:** Delete dead code. If `inlay_hints.go` has planned use, add a `// TODO` and gut the handler to return empty.

---

## Group 2: Lexer Robustness

**Severity: critical + moderate | Risk: moderate | Scope: `parser/lexer/lexer.go`**

Harden the lexer against edge cases and improve internal structure.

| Location | Severity | Finding |
|---|---|---|
| `lexer.go:186` | critical | `emitEOF()` mutates `l.input` (appends `'\n'`) as an idempotency guard. If called twice (e.g., in a test harness), input is silently corrupted. Replace with a boolean flag (`eofEmitted`). |
| `lexer.go:171-177` | moderate | `emitDedentsTo()` does not validate that the target indent level exists in the stack or that `target >= 0`. No error recovery for malformed indentation. |
| `lexer.go:221-237, 239-255` | minor | `scanArgs()` and `scanString()` replicate the same "scan until delimiter, tracking newlines" pattern. Could share a `scanDelimited(open, close byte)` helper. |

**Action:** Add `eofEmitted` flag. Add stack-membership validation in `emitDedentsTo()`. Optionally extract shared scan helper.

---

## Group 3: AST Node Bloat — Async Target Abstraction

**Severity: critical | Risk: high | Scope: foundational — cascades into Groups 4, 5, 8**

`AwaitStmt`, `AwaitOneCase`, and `PromiseStmt` each carry 12-18 optional fields representing mutually exclusive async targets (timer, signal, activity, workflow, nexus, ident). This creates massive duplication in every layer that touches these nodes.

| Location | Severity | Finding |
|---|---|---|
| `parser/ast/ast.go:173-240` | critical | `AwaitStmt` has 15+ optional fields. Only one "target group" is populated at a time. Should be a discriminated union — e.g., an `AsyncTarget` interface with concrete types `TimerTarget`, `SignalTarget`, `ActivityTarget`, `WorkflowTarget`, `NexusTarget`, `IdentTarget`. |
| `parser/ast/ast.go:253-304` | critical | `AwaitOneCase` mirrors `AwaitStmt` with 18 optional fields. Same structural problem. |
| `parser/ast/ast.go:443-472` | moderate | `PromiseStmt` replicates the same async target fields. Third copy. |
| `parser/ast/ast.go + json.go` | critical | All three types share identical field patterns for each target kind. A shared `AsyncTarget` type would collapse the duplication. |

**Action:** Define an `AsyncTarget` interface (or concrete struct with a `Kind` discriminator) that encapsulates target-specific fields. Replace the flat optional fields in `AwaitStmt`, `AwaitOneCase`, and `PromiseStmt` with a single `Target AsyncTarget` field. This is the highest-impact structural change — it simplifies parsing, resolution, serialization, and LSP handlers downstream.

**Tradeoff:** Interface-based union is idiomatic Go but requires type switches at every consumer. A single struct with a `Kind` enum and optional field groups is simpler but less type-safe. Discuss before implementing.

---

## Group 4: Await/Promise Parsing & Resolution Duplication

**Severity: critical | Scope: `parser/parser/statements.go`, `parser/resolver/resolver.go` | Depends on: Group 3**

The parser and resolver both duplicate target-dispatch logic across await variants.

| Location | Severity | Finding |
|---|---|---|
| `parser/statements.go:119-340` | critical | `parseSingleAwait()` is 200+ lines with deeply nested switches covering 10+ target types. |
| `parser/statements.go:427-737` | critical | `parseAwaitOneCase()` duplicates nearly all parsing logic from `parseSingleAwait()` — same target dispatch, same field population. Refactoring to a shared `parseAsyncTarget()` would save ~300 lines. |
| `resolver/resolver.go:580-651` | critical | AwaitStmt resolution dispatches over signal/update/activity/workflow/nexus/ident with field-specific logic. |
| `resolver/resolver.go:961-1042` | critical | `resolveAwaitOneCase()` duplicates 25+ lines from the AwaitStmt case — identical resolution logic for each target kind. |

**Action:** Extract `parseAsyncTarget()` in the parser and `resolveAsyncTarget()` in the resolver. Both consume the shared `AsyncTarget` type from Group 3.

---

## Group 5: JSON Serialization Explosion

**Severity: critical + moderate | Scope: `parser/ast/json.go` | Depends on: Group 3**

`json.go` is 1043 lines (2x `ast.go`) with pervasive duplication in marshaling logic.

| Location | Severity | Finding |
|---|---|---|
| `ast/json.go` (entire) | critical | 1043 lines with 50+ JSON struct types. File is unwieldy — needs splitting or a generation approach. |
| `ast/json.go:391-729` | critical | `marshalStatement` is a 340-line type switch with 20+ cases. Each case duplicates position/type field marshaling. Extract common marshaling into a helper; use per-type wrappers. |
| `ast/json.go:119-208` | moderate | WorkflowDef, ActivityDef, etc. all repeat: create JSON struct, loop body, call marshalStatement, append. |
| `ast/json.go:440-452, 515-526, 674-685, 700-711` | moderate | Nexus resolution boilerplate (`if NexusResolvedEndpoint`, `if NexusResolvedService`, `if NexusResolvedOperation`) repeated in 4 locations. Extract `marshalNexusResolvedRefs()`. |
| `ast/json.go:267-300` | moderate | WorkerDef marshal has three nearly-identical loops over Workflows, Activities, Services. Extract `marshalWorkerRefs()`. |

**Action:** After Group 3 lands (shared `AsyncTarget`), the type switch cases collapse significantly. Then split the file and extract repeated helpers (`marshalNexusResolvedRefs`, `marshalWorkerRefs`, `marshalStatementList`).

---

## Group 6: Resolver Separation of Concerns

**Severity: critical | Scope: `parser/resolver/resolver.go`**

The resolver conflates symbol resolution with deployment/routing validation.

| Location | Severity | Finding |
|---|---|---|
| `resolver.go:29-237` | critical | Resolver performs 4 passes including task queue routing checks (`checkCallRouting` ~line 808). Routing is an execution/deployment concern, not symbol resolution — it belongs in a downstream validation layer. |
| `resolver.go:240-462` | moderate | `resolveWorkersAndNamespaces()` is 222 lines handling 4 distinct concerns: empty-body checks, worker validation, namespace validation, and task queue coherence. Should be split into focused functions. |
| `resolver.go:411-459` | critical | Task queue coherence validation has 3 nested loops with inline struct definition and high cyclomatic complexity. |
| `resolver.go:214-229` | moderate | Nexus service validation scattered across 3 locations (Pass 2b, `resolveNexusRef`, and within statement resolution). |

**Action:** Extract routing/deployment validation into a separate pass or package (e.g., `validator`). Keep the resolver focused on name resolution and type checking. Split `resolveWorkersAndNamespaces()` into per-concern functions.

**Tradeoff:** Moving validation downstream requires deciding where it lives. Options: (a) new `validator` package in the parser pipeline, (b) keep in resolver but clearly separated as a distinct pass with its own entry point. Discuss before implementing.

---

## Group 7: Error Type Design

**Severity: moderate | Scope: `parser/resolver/`, `internal/server/code_actions.go`, `cmd/twf/main.go`**

Error types are stringly-typed, making downstream consumers fragile.

| Location | Severity | Finding |
|---|---|---|
| `server/code_actions.go:35-120` | moderate | Parses resolver error messages via string splitting (`strings.Contains(err.Msg, "undefined activity:")`). Breaks silently if message format changes. |
| `cmd/main.go:104` | minor | Parser returns `[]*ParseError` and resolver returns `[]*ResolveError` — separate types with no shared interface. Main treats them uniformly. |
| `server/document.go` (Get method) | moderate | `DocumentStore.Get()` returns `nil` for missing documents instead of idiomatic `(*Document, bool)`. Forces nil-checks in 14+ handler functions. |

**Action:** Add structured fields to `ResolveError` (e.g., `Kind ErrorKind`, `Name string`) so consumers can switch on error kind instead of parsing messages. Consider a shared `Diagnostic` interface for parser and resolver errors. Change `DocumentStore.Get()` to return `(*Document, bool)`.

---

## Group 8: AST Traversal Duplication in Server (Visitor Pattern)

**Severity: moderate-critical | Scope: `parser/ast/`, `internal/server/*.go` | Depends on: Group 3**

Every LSP handler independently walks the AST with copy-pasted type switches.

| Location | Severity | Finding |
|---|---|---|
| `parser/ast/ast.go:1-7` | moderate | `Node` interface provides position but no `Accept(Visitor)` method. All consumers must write their own type switches. |
| `server/definition.go:35-104` | critical | `resolvedTarget()` has 8 type cases with nested pointer derefs over `AwaitStmt` fields. Fragile — silently breaks if AST adds new fields. |
| `server/hover.go:40-143` | moderate | `findNodeAtLine()` has 8 sequential type switches with repeated subchecks (Signals, Queries, Updates). |
| `server/references.go`, `server/hover.go` | moderate | Both traverse identical AST structures in nearly identical ways. Candidate for a shared walker. |
| `server/code_actions.go`, `server/rename.go` | moderate | More traversal duplication. Rename reuses references logic but duplicates parts. |

**Action:** Add a Visitor interface or a generic `Walk(node, func)` helper to the `ast` package. Replace per-handler traversal with visitor callbacks. If Group 3 lands first (discriminated union for async targets), the visitors become much simpler.

---

## Group 9: DRY Violations in Resolver

**Severity: moderate | Scope: `parser/resolver/resolver.go`**

Repetitive patterns in the resolver that could be unified without abstraction overhead.

| Location | Severity | Finding |
|---|---|---|
| `resolver.go:37-86` | moderate | Pass 1 duplicate-check-then-store pattern repeated 5x (workflows, activities, workers, namespaces, nexusServices) with only map/name/type differences. Extract a generic `collectDefinitions[T]()` helper or a loop over a descriptor slice. |
| `resolver.go:267-304` | moderate | Worker type set validation loops tripled — resolves workflow/activity/service refs with identical if-ok-else-error pattern. Only names and error messages differ. |
| `resolver.go:355-366, 427-439` | moderate | `instantiatedWorkers` coverage tracking: identical set-construction logic built in two separate places. |

**Action:** Extract `collectAndDedup()` for Pass 1. Unify worker-ref resolution into a loop over a `[]struct{kind, map, refs}` descriptor. Deduplicate coverage set construction.

---

## Group 10: Definition Parser Boilerplate

**Severity: moderate | Scope: `parser/parser/definitions.go`, `parser/parser/workers.go`, `parser/parser/namespaces.go`**

Definition parsers follow identical structural patterns with copy-pasted scaffolding.

| Location | Severity | Finding |
|---|---|---|
| `definitions.go:11-325` | moderate | 5 definition parsers (workflow, activity, signal, query, update) follow identical INDENT → parse body → DEDENT pattern. A generic `parseIndentedBlock()` helper would reduce boilerplate. |
| `definitions.go:172-214` | moderate | Signal/query/update declarations save and restore `inWorkflow`/`inActivity` flags manually — pattern repeats 4x. A `withContext(ctx)` helper would encapsulate save/restore. |
| `parser.go:26-36` | minor | Dual boolean flags (`inWorkflow`, `inActivity`) are mutually exclusive — should be a single enum (`parserContext`) to prevent invalid state combinations. |
| `workers.go:45-95` | moderate | Worker reference parsing (workflow, activity, nexus service) repeated 3x with trivial differences. |
| `namespaces.go:46-89` | moderate | Worker and endpoint instantiation in namespaces follow similar patterns. |

**Action:** Replace `inWorkflow`/`inActivity` with a `context` enum. Extract `parseIndentedBlock()` and `withContext()` helpers. Unify worker-ref parsing into a loop or helper.

---

## Group 11: Missing Test Coverage

**Severity: critical (breadth) | Scope: `ast/`, `internal/server/`, `cmd/twf/`, gaps in `lexer/`, `resolver/`**

Large areas of the codebase have no test coverage. This is an ongoing concern that should be addressed incrementally alongside other groups.

| Location | Severity | Finding |
|---|---|---|
| `parser/ast/` | minor | No `*_test.go` files. `NodeLine()`/`NodeColumn()`, `AwaitKind()`, `CaseKind()` untested. |
| `internal/server/` | critical | Zero test files for the entire LSP server package. Handlers involve complex AST traversal, null checks, and protocol edge cases. |
| `cmd/twf/` | moderate | No `main_test.go`. CLI commands (`check`, `parse`, `symbols`, `lsp`) untested. |
| `parser/lexer/lexer_test.go` | moderate | Missing edge cases: unclosed strings, unclosed args, invalid duration suffixes (e.g., `1x`), mixed tabs/spaces, column number accuracy. |
| `parser/resolver/resolver_test.go` | minor | No multi-namespace overlapping task queue tests. Limited coverage for complex deployment topologies. |

**Action:** Add tests incrementally as each group is implemented. Prioritize server handler tests (highest risk) and lexer edge cases (most likely to regress).

---

## Group 12: CLI Structure

**Severity: moderate | Scope: `cmd/twf/main.go`**

The CLI is a single 336-line file mixing dispatch, I/O, parsing, and formatting.

| Location | Severity | Finding |
|---|---|---|
| `main.go:1-336` | moderate | Single monolithic file mixes CLI dispatch, file I/O, parsing orchestration, and output formatting. |
| `main.go:75-129` | moderate | `parseFiles()` combines arg parsing, file reading, parser invocation, and error collection — too many responsibilities for one function. |
| `main.go:79-85, 210-216` | moderate | Flag parsing is inlined and duplicated across commands (`--lenient`, `--json`). No `flag.FlagSet` or shared parsing. |
| `main.go:238-335` | moderate | `printSymbolsText()` and `printSymbolsJSON()` both iterate definitions with duplicated type assertions and field extraction. Extract shared symbol-extraction logic. |

**Action:** Split into subcommand files or a `commands/` subpackage. Use `flag.FlagSet` per command. Extract `extractSymbols()` to share between text and JSON output.

---

## Group 13: Token Map Redundancy & Option Schema Duplication

**Severity: moderate-minor | Scope: `parser/token/token.go`, `parser/parser/options.go`**

Small maintenance-burden issues that accumulate over time.

| Location | Severity | Finding |
|---|---|---|
| `token/token.go:88-199` | moderate | `tokenNames` (Type → display string) and `keywords` (string → Type) maps must be kept in sync manually. Adding a keyword requires updating both. Consider generating one from the other, or using a single `[]struct{Type, Name}` table. |
| `parser/options.go:27-90` | moderate | Option schemas are duplicated across parsing contexts. A shared registry keyed by context would reduce maintenance. |
| `parser/options.go:317-341` | moderate | `joinStrings()` manually builds comma-separated output instead of using `strings.Join()`. |

**Action:** Unify token maps into a single source-of-truth table. Consolidate option schemas into a registry. Replace `joinStrings()` with `strings.Join()`.
