# TWF CLI

Command-line interface for working with Temporal Workflow Format (.twf) files.

## Installation

```bash
go install github.com/jmbarzee/temporal-skills/design/lsp/cmd/twf@latest
```

Or from the repo root:
```bash
cd design/lsp
go install ./cmd/twf
```

## Commands

### `twf check`

Parse and validate TWF files, reporting any errors.

```bash
twf check workflow.twf
twf check *.twf
twf check --lenient workflow.twf  # Continue even with resolve errors
```

**Output:**
- Parse errors (syntax)
- Resolve errors (undefined references, type mismatches)
- Success message with counts

**Exit codes:**
- `0` - Success, no errors
- `1` - Errors found

**Example:**
```bash
$ twf check examples/skill-basics.twf
✓ OK: 4 workflow(s), 10 activity(s)
```

---

### `twf parse`

Output the Abstract Syntax Tree (AST) as JSON.

```bash
twf parse workflow.twf
twf parse --lenient workflow.twf  # Parse even with resolve errors
```

**Output:** Complete AST in JSON format, suitable for:
- Code generation
- Analysis tools
- AI assistants
- Custom tooling

**Example:**
```bash
$ twf parse workflow.twf | jq '.definitions[0].name'
"OrderWorkflow"
```

---

### `twf symbols`

List all workflows and activities in the file.

```bash
twf symbols workflow.twf
twf symbols --json workflow.twf  # JSON output
```

**Text output:**
```
workflow ProcessOrder(order: Order) -> (Result)
activity ValidateOrder(order: Order) -> (ValidateResult)
activity ProcessPayment(order: Order) -> (Payment)
```

**JSON output:**
```json
[
  {
    "kind": "workflow",
    "name": "ProcessOrder",
    "params": "order: Order",
    "returnType": "Result",
    "signals": ["PaymentReceived"],
    "queries": ["GetStatus"],
    "updates": ["UpdateAddress"]
  }
]
```

---

## Use Cases

### CI/CD Validation

```bash
# Validate all TWF files in CI
twf check $(find . -name "*.twf")
```

### AI Integration

```bash
# AI assistants can analyze workflow structure
twf symbols --json workflow.twf | ai-tool analyze

# Parse complete AST for code generation
twf parse workflow.twf | ai-tool generate-go
```

### Development Workflow

```bash
# Quick validation during development
twf check workflow.twf

# List all definitions to understand structure
twf symbols workflow.twf

# Generate documentation from AST
twf parse workflow.twf | doc-generator
```

### Build Scripts

```bash
# Validate before building
if ! twf check workflows/*.twf; then
    echo "TWF validation failed"
    exit 1
fi

# Generate code from TWF
for file in workflows/*.twf; do
    twf parse "$file" | code-generator > "generated/$(basename "$file" .twf).go"
done
```

---

## Options

- `--json` - Output in JSON format (for `symbols` command)
- `--lenient` - Continue even with resolve errors (useful for partial/incomplete code)

---

## Architecture

The CLI uses the same parser and resolver as the LSP server, ensuring consistent behavior between:
- Editor integration (LSP)
- Command-line tools (CLI)
- Build scripts
- CI/CD pipelines

**Shared components:**
- `parser/lexer` - Tokenization
- `parser/parser` - AST construction
- `parser/resolver` - Symbol resolution and validation

---

## Exit Codes

- `0` - Success
- `1` - Error (parse error, resolve error, file not found, etc.)

---

## See Also

- [Language Specification](../../LANGUAGE.md) - TWF syntax reference
- [LSP Server](../twf-lsp/) - Editor integration
- [Code Actions](../../CODE_ACTIONS.md) - Quick fixes and refactorings
