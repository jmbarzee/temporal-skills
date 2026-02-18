# Subagent Adoption Plan

Notes on splitting the author-go skill into an orchestrator + worker subagents.

---

## Why subagents

The author-go skill has two kinds of work:

1. **Interactive** -- context gathering, ambiguity resolution, activity body clarification, review. Requires user dialog.
2. **Mechanical** -- SDK exploration, type generation, workflow body generation, build verification. Self-contained, parallelizable, context-heavy.

Mixing both in one context window means the 20 reference files, SDK source, and build output compete for tokens with the user conversation. Subagents isolate the mechanical work and return only summaries.

### Supporting data

- Google Research (180 agent configs): multi-agent excels on **parallelizable/decomposable** tasks (80.9% improvement), but **degrades 39-70%** on sequential reasoning. Architecture should match task structure.
- Anthropic guidance: start with single agent + skills, add subagents when context isolation or parallelism justifies it. Orchestrator-worker is the sanctioned pattern (no peer-to-peer).
- Anthropic Agent SDK: subagents get isolated context windows, return summaries to orchestrator. Cannot spawn nested subagents.

---

## Architecture

```
author-go/SKILL.md (orchestrator, inline)
  Owns all user dialog
  Runs in main conversation context
  Spawns subagents for mechanical work
  │
  ├── sdk-explorer (Explore agent)
  │   Reads SDK source / docs to identify available types and imports
  │   Returns: dependency manifest (types, imports, patterns available)
  │
  ├── go-codegen (general-purpose agent)
  │   Preloads reference skills for TWF → Go mapping
  │   Generates types, signatures, workflow bodies
  │   Returns: generated .go files
  │
  └── build-verifier (Bash agent)
      Runs go build / go vet
      Diagnoses errors
      Returns: pass/fail summary with actionable fixes
```

### Phase mapping

| Phase | Owner | Why |
|-------|-------|-----|
| Phase 1: Context gathering | Orchestrator + sdk-explorer | Orchestrator asks user about project context. sdk-explorer runs in parallel scanning SDK/dependencies. |
| Phase 2: Planning | Orchestrator | Requires ambiguity resolution with user. Consumes sdk-explorer's manifest. |
| Phase 3 Layer 1: Types + signatures | go-codegen | Mechanical. Reference-heavy. Parallelizable per workflow tree. |
| Phase 3 Layer 2: Workflow bodies | go-codegen | Mechanical TWF → Go SDK mapping. |
| Phase 3 Layer 3: Activity bodies | Orchestrator | Pseudocode in .twf requires user clarification. |
| Phase 3 Layer 4: Build verification | build-verifier | Verbose output stays isolated. Returns summary. |
| Phase 4: Review | Orchestrator | User feedback loop. |

---

## Subagent definitions

### sdk-explorer

**Purpose:** Scan SDK and project dependencies to build a manifest of available types, imports, and patterns. Answers "what already exists?" so the orchestrator can follow the "prefer imports over generation" principle.

**Critical:** Don't just `go doc` individual types in isolation. Trace the actual call chain from activity code to SDK method and verify every type in the signature. SDK types often have union/wrapper variants (e.g., `ToolParam` vs `ToolUnionParam`) that only surface when you check the struct field that accepts them. The manifest must reflect the types the code will actually compile against, not the types that look right by name.

**Sources (suggest both, don't constrain):**
- Local module cache: `$GOPATH/pkg/mod/go.temporal.io/sdk@version/`
- Web: SDK docs, godoc, GitHub source

**Returns:** Structured summary -- available types, import paths, relevant SDK patterns. For each SDK call site, include the exact method signature and the concrete types of every parameter and return value.

```yaml
---
name: sdk-explorer
description: Explore Temporal SDK and project dependencies to identify available types and imports.
tools: Read, Glob, Grep, Bash, WebFetch, WebSearch
model: sonnet
---
```

### go-codegen

**Purpose:** Generate Go code from TWF designs using the reference material. Handles Layers 1-2 (types, signatures, workflow bodies). Can run one instance per independent workflow tree for parallelism.

**Preloads:** The author-go reference files via the `skills` field. This keeps the 20 reference docs out of the main conversation's context.

```yaml
---
name: go-codegen
description: Generate Go Temporal workflow code from TWF designs. Use for types, signatures, and workflow bodies.
tools: Read, Write, Edit, Glob, Grep
model: sonnet
skills:
  - temporal-go-reference
---
```

### build-verifier

**Purpose:** Run `go build` and `go vet`, diagnose errors, return actionable summary. Keeps verbose compiler output out of the main context.

```yaml
---
name: build-verifier
description: Run Go build and vet, diagnose errors, return summary.
tools: Bash, Read, Glob, Grep
model: haiku
---
```

---

## Skill composition

Claude Code supports two composition directions:

| Direction | Mechanism | Use case |
|-----------|-----------|----------|
| Skill spawns subagent | `context: fork` in skill frontmatter | Entire skill runs as isolated subagent |
| Subagent loads skills | `skills` field in agent frontmatter | Subagent gets domain knowledge injected at startup |

For author-go, we use **both**:
- The main skill runs **inline** (no `context: fork`) because it needs user dialog
- The go-codegen subagent **loads reference skills** so it has the TWF → Go mapping knowledge
- The orchestrator (main conversation with skill loaded) spawns subagents via the Task tool

---

## Per-workflow-tree parallelism

After Phase 2 identifies independent root workflows, each root + its children + its activities can be a separate go-codegen subagent invocation. They don't share state and can generate in parallel.

Example: a .twf file with 3 independent root workflows → 3 parallel go-codegen subagents, each producing its own .go files.

---

## Cross-language generalization

This pattern generalizes to other language targets:

```
.claude/agents/
├── sdk-explorer.md          # shared, parameterized by language/SDK
├── go-codegen.md            # Go-specific, loads go reference skills
├── python-codegen.md        # Python-specific, loads python reference skills
└── build-verifier.md        # parameterized (go build vs pytest vs tsc)
```

Each codegen agent loads only its language's reference material. The orchestrator skill per language (author-go, author-python, etc.) handles the dialog. A user request like "implement in Go and Python" spawns both codegen agents in parallel with full context isolation.

---

## Open questions

- **Skill splitting granularity:** Should reference files become a separate skill (`temporal-go-reference`) that the codegen agent preloads, or should they stay as files the agent reads on demand?
- **sdk-explorer scope:** Should it also scan the user's existing project code for conventions, or is that the orchestrator's job during Phase 1?
- **Error recovery:** When build-verifier reports failures, should the orchestrator fix them directly or re-invoke go-codegen with the error context?
- **Activity body generation:** Some activity bodies are unambiguous (simple SDK calls). Could a heuristic route "clear" bodies to a subagent and only surface "ambiguous" ones to the user?
