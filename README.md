# temporal-skills

AI skills for designing and developing Temporal workflows, distributed as a VS Code / Cursor extension.

## Install

Install **Temporal Workflow (.twf)** from the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=jmbarzee.temporal-twf) or [Open VSX](https://open-vsx.org/extension/jmbarzee/temporal-twf). The extension bundles:

- **AI Skills** — automatically installed to `~/.cursor/skills/` so Cursor's agent can use them immediately
- **`twf` CLI** — parser, validator, and language server for `.twf` files, added to your terminal PATH
- **Syntax highlighting** and **workflow visualization** for `.twf` files

## Skills

- **[design](./skills/design/SKILL.md)** — Design Temporal workflows using `.twf`, a language-agnostic DSL with parser/LSP tooling and visualization
- **[author-go](./skills/author-go/SKILL.md)** — Generate Go code from `.twf` workflow designs using the Temporal Go SDK

### Planned

- **Implementers** — More authorship skills (TypeScript, Python, Java, etc.)
- **Translators** — Analyze existing systems (event-based architectures, SQS/Lambda, etc.) and generate equivalent DSL designs
- **Debuggers & Optimizers** — Assist with debugging, profiling, and optimizing existing Temporal workflows

## Repository Structure

```
packages/    VS Code / Cursor extension
tools/       Go LSP + CLI (twf), React workflow visualizer
skills/      AI skill definitions (SKILL.md + reference docs)
```

## Development

```bash
# Build everything (current platform)
make build

# Run Go tests
make test

# Package a local .vsix
make package

# Package for all platforms (CI)
make package-all

# Publish to marketplaces
VSCE_TOKEN=... OVSX_TOKEN=... make publish
```

### Build Targets

| Target | Description |
|--------|-------------|
| `build-lsp` | Compile the `twf` Go binary |
| `build-visualizer` | Build the React webview |
| `build-skills` | Copy skills into the extension package |
| `build-extension` | Compile extension TypeScript |
| `build` | All of the above |
| `package` | Build + create `.vsix` (local platform) |
| `package-all` | Cross-compile + create `.vsix` per platform |
| `publish-vscode` | Publish to VS Code Marketplace |
| `publish-ovsx` | Publish to Open VSX |
| `publish` | Publish to both |
| `test` | Run Go tests |
| `vet` | Run Go vet |
| `release` | Bump version, tag, push (triggers CI release) |
| `clean` | Remove build artifacts |
