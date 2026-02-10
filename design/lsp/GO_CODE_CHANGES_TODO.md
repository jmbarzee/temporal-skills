# Go Code Changes for Await Keywords

## Summary
This tracks the remaining Go code changes needed to implement the `await all` / `await one` keyword changes.

## Completed ✅
1. ✅ Updated token definitions (`token/token.go`)
   - Added `ALL`, `ONE` tokens
   - Removed `OR`, `PARALLEL`, `SELECT` tokens
   - Updated `tokenNames` map
   - Updated `keywords` map

2. ✅ Updated LANGUAGE.md spec
3. ✅ Updated all example files

## Remaining Tasks

### 1. Parser (`parser/parser/`)

**statements.go:**
- ✅ Remove `parseAwaitStmt` function
- ✅ Remove `OR` token handling from await parsing
- ✅ Update `parseParallelBlock` → `parseAwaitAllBlock`
- ✅ Update `parseSelectBlock` → `parseAwaitOneBlock`
- ✅ Update statement parser dispatch map
- ✅ Update `parseHintStmt` to support query hints

**parser.go:**
- ✅ Update `workflowStmtParsers` map:
  - Remove `token.PARALLEL` entry
  - Remove `token.SELECT` entry
  - Add `token.AWAIT` entry pointing to `parseAwaitBlock` (dispatches to all/one)
- ✅ Update `temporalKeywords` map (removed PARALLEL/SELECT, added ALL/ONE)

**New function needed:**
```go
func parseAwaitBlock(p *Parser) (ast.Statement, error) {
    p.advance() // consume AWAIT

    switch p.current.Type {
    case token.ALL:
        return parseAwaitAllBlock(p)
    case token.ONE:
        return parseAwaitOneBlock(p)
    default:
        return nil, p.errorf("expected 'all' or 'one' after 'await', got %s", p.current.Type)
    }
}
```

### 2. AST (`parser/ast/`)

**ast.go:**
- ✅ Remove `AwaitStmt` type
- ✅ Remove `AwaitTarget` type
- ✅ Rename `ParallelBlock` → `AwaitAllBlock`
- ✅ Rename `SelectBlock` → `AwaitOneBlock`
- ✅ Update `SelectCase` → `AwaitOneCase` and simplify (timer + nested await all only)
- ✅ Update `HintStmt` comment to include query

**json.go:**
- ✅ Update JSON marshaling for renamed types
- ✅ Remove `AwaitStmt` JSON handling
- ✅ Update type strings: `"parallel"` → `"awaitAll"`, `"select"` → `"awaitOne"`
- ✅ Simplify `AwaitOneCase` JSON (no workflow/activity fields)

### 3. Resolver (`parser/resolver/`)

**resolver.go:**
- ✅ Update `resolveStatement` switch cases:
  - Remove `*ast.AwaitStmt` case
  - Rename `*ast.ParallelBlock` → `*ast.AwaitAllBlock`
  - Rename `*ast.SelectBlock` → `*ast.AwaitOneBlock`
- ✅ Remove `resolveAwaitTarget` function
- ✅ Simplify and rename `resolveSelectCase` → `resolveAwaitOneCase`
- ✅ Add queries map and support query hints

### 4. LSP Server (`internal/server/`)

**symbols.go:**
- ✅ Update symbol extraction for renamed block types

**semantic_tokens.go:**
- ✅ Remove `token.OR` handling
- ✅ Remove `token.PARALLEL` handling
- ✅ Remove `token.SELECT` handling
- ✅ Add `token.ALL` handling
- ✅ Add `token.ONE` handling

**Other LSP files:**
- ✅ definition.go - Remove AwaitTarget case
- ✅ folding.go - Update block types
- ✅ hover.go - Remove AwaitStmt/AwaitTarget, update block types
- ✅ references.go - Remove AwaitStmt/AwaitTarget/SelectCase, update block types

### 5. Tests

**lexer_test.go:**
- ✅ Update test cases to use `all`, `one`, `hint` instead of `or`, `parallel`, `select`

**parser_test.go:**
- ✅ Update test cases for renamed block types
- ✅ Remove `await signal X` test cases (TestAwaitSingle, TestAwaitMultiTarget, TestAwaitWithArgs)
- ✅ Update tests for `await all:` and `await one:` (TestAwaitAllBlock, TestAwaitOneBlock)
- ✅ Update hint tests to support query
- ✅ Remove/update tests with obsolete syntax (TestSelectWithActivityCase)

**resolver_test.go:**
- ✅ Update test cases for renamed types
- ✅ Replace `await signal/update` with `hint signal/update` in tests
- ✅ Update `parallel:` → `await all:` in TestNestedResolution
- ✅ Remove obsolete tests (TestSelectCaseResolution, TestSelectCaseUndefinedWorkflow)

### 6. Command-line Tools

**cmd/parse/main.go:**
- Should work automatically once AST types are updated

**cmd/twf-lsp/main.go:**
- Should work automatically

## Implementation Order

1. **AST changes first** - Rename types so parser can reference them
2. **Parser changes** - Update parsing logic
3. **Resolver changes** - Update traversal logic
4. **LSP changes** - Update server features
5. **Tests** - Update all test cases
6. **Verify** - Test with example files

## Notes

- Keep backward compat testing in mind - old example files should fail gracefully
- Consider adding migration warnings if old keywords are detected
- Update any error messages that reference old keywords
