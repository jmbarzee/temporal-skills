# TWF Tooling Overview

This document provides an overview of the TWF (Temporal Workflow Format) tooling ecosystem.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Shared Components                        │
│  parser/lexer  │  parser/parser  │  parser/resolver         │
│  parser/ast    │  parser/token                              │
└─────────────────────────────────────────────────────────────┘
           ↓                    ↓                    ↓
    ┌──────────┐         ┌──────────┐         ┌──────────┐
    │   CLI    │         │   LSP    │         │  Custom  │
    │   twf    │         │ Server   │         │  Tools   │
    └──────────┘         └──────────┘         └──────────┘
         ↓                     ↓                     ↓
   ┌──────────┐         ┌──────────┐         ┌──────────┐
   │  CI/CD   │         │ Editors  │         │    AI    │
   │  Build   │         │  VS Code │         │  Agents  │
   └──────────┘         │  Cursor  │         └──────────┘
                        └──────────┘
```

## Tools

### 1. CLI Tool (`cmd/twf`)

**Purpose:** Command-line interface for parsing, validating, and analyzing TWF files

**Commands:**
- `twf check` - Validate syntax and semantics
- `twf parse` - Output AST as JSON
- `twf symbols` - List workflows/activities

**Use cases:**
- CI/CD validation
- Build scripts
- Code generation
- AI integration
- Quick validation during development

**See:** [cmd/twf/README.md](cmd/twf/README.md)

---

### 2. LSP Server (`cmd/twf-lsp`)

**Purpose:** Language Server Protocol implementation for editor integration

**Features:**
- Real-time diagnostics (parse/resolve errors)
- Go to definition
- Find references
- Hover documentation
- Autocomplete
- Rename symbols
- **Code actions** (quick fixes & refactorings)
- Semantic highlighting
- Signature help
- Document symbols
- Folding ranges

**Supported editors:**
- VS Code (via extension)
- Cursor (via extension)
- Any LSP-compatible editor

**See:** [CODE_ACTIONS.md](CODE_ACTIONS.md)

---

## Standard Interface: LSP

Both tools use the **Language Server Protocol** as the standard interface for:
- Linting and diagnostics
- Symbol information
- Code navigation
- Refactoring

This means:
- ✅ **Consistent behavior** across CLI and editors
- ✅ **AI assistants** can use the same tooling
- ✅ **Single source of truth** for validation
- ✅ **Easy integration** with any LSP-compatible tool

### Using LSP from AI Tools

AI assistants (like Claude, Copilot, etc.) can connect to the LSP server to:

```bash
# Start LSP server
twf-lsp --stdio

# Send LSP requests (JSON-RPC)
{
  "method": "textDocument/diagnostic",
  "params": { "textDocument": { "uri": "file:///path/to/workflow.twf" } }
}

# Or use the CLI for simpler interface
twf check workflow.twf
```

---

## Shared Components

All tools use the same underlying packages:

### `parser/lexer`
- Tokenization
- Indentation tracking
- Comment handling

### `parser/parser`
- AST construction
- Error recovery
- Multi-file support

### `parser/resolver`
- Symbol resolution
- Type checking
- Reference validation

### `parser/ast`
- Abstract Syntax Tree definitions
- JSON serialization
- Visitor pattern support

---

## Development Workflow

### For Developers

```bash
# Validate during development
twf check workflow.twf

# See structure
twf symbols workflow.twf

# Use editor integration
# (LSP provides real-time feedback)
```

### For CI/CD

```bash
# Validate all TWF files
twf check $(find . -name "*.twf")

# Generate code
twf parse workflow.twf | codegen
```

### For AI Assistants

```bash
# Analyze structure
twf symbols --json workflow.twf

# Get full AST
twf parse workflow.twf

# Or connect to LSP for rich features
# (diagnostics, code actions, etc.)
```

---

## Installation

```bash
cd design/lsp

# Install CLI
go install ./cmd/twf

# Install LSP server
go install ./cmd/twf-lsp

# Or build both
go build ./...
```

---

## Future Enhancements

### CLI
- `twf format` - Auto-format TWF files
- `twf test` - Run workflow tests
- `twf codegen` - Generate Go/Java/Python from TWF
- `twf visualize` - Generate diagrams

### LSP
- Inlay hints (inline type information)
- Code lens (run/test buttons)
- Call hierarchy
- Document formatting
- Workspace symbols

### Shared
- Performance optimizations
- Incremental parsing
- Better error recovery
- Enhanced type inference

---

## See Also

- [LANGUAGE.md](LANGUAGE.md) - TWF syntax specification
- [CODE_ACTIONS.md](CODE_ACTIONS.md) - Available quick fixes
- [cmd/twf/README.md](cmd/twf/README.md) - CLI documentation
