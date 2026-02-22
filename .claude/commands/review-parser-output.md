# Parser Output Review

Review the JSON output of `twf parse` and `twf symbols` from the perspective of downstream TypeScript consumers: the **Tree View**, the **Graph View**, and the VS Code extension.

This review operates at the **boundary between Go and TypeScript**. The parser's JSON is the contract — it must be complete, consistent, and directly consumable without client-side workarounds.

## Context

Read these files before starting:
- `AST_REVISIONS.md` — current revision plan, what's been done
- `tools/visualizer/spec/TREE_VIEW.md` — Tree View requirements and current behavior
- `tools/visualizer/spec/GRAPH_VIEW.md` — Graph View requirements and planned behavior

## Review Checklist

Evaluate the JSON output against each of these concerns:

### 1. Resolution Completeness
- Does every `Ref[T]` in the AST emit its `resolved` field in JSON when the resolver successfully resolves it?
- Are there reference types where the resolver does work but the JSON serialization drops the result?
- Can the visualizer build lookup maps purely from the JSON, or must it re-resolve references client-side?

### 2. Field Scoping
- Does each JSON node emit only fields relevant to its type/kind?
- Are there fields that bleed across variants (e.g., `workflowMode` on a timer target)?
- For discriminated unions (AsyncTarget, Definition), is the `kind`/`type` discriminator reliable and exhaustive?

### 3. Nullability & Presence
- Are array fields always present (empty array) or sometimes absent (`omitempty`)?
- Does the JSON shape match what TypeScript interfaces declare? Specifically: are fields marked required in TS always present in JSON?
- Are optional fields consistently `null` vs absent?

### 4. Multi-file Support
- When multiple files are parsed together, does each definition carry `sourceFile`?
- Does cross-file resolution work (workflow in file A references activity in file B)?
- Can the visualizer attribute definitions to files without client-side workarounds?

### 5. Consumer Fitness
- Can the **Tree View** render directly from the JSON without building intermediate lookup maps?
- Can the **Graph View** extract dependency edges without walking statement bodies?
- Is there information the visualizer needs that requires a separate subcommand (e.g., `twf deps`)?
- Is there information in the JSON that no consumer uses (noise)?

### 6. Consistency
- Do similar constructs serialize the same way? (e.g., all call types, all declaration types)
- Are naming conventions consistent across the JSON schema? (`camelCase` everywhere? `name` vs `Name`?)
- Do error/warning representations align between `twf check` and `twf parse`?

## Workflow

### Phase 1: Generate Test Output

Run `twf parse` and `twf symbols` against representative `.twf` files. Use files that exercise:
- Multiple definition types (workflow, activity, worker, namespace, nexus service)
- Cross-definition references (workflow calls activity, worker registers workflow)
- Nexus calls (endpoint → service → operation chain)
- Async targets (await, promise, await-one with mixed target types)
- Multi-file input (if test fixtures support it)

Capture the raw JSON output for analysis.

### Phase 2: Analyze Against Checklist

For each checklist item, provide:
- **Status**: `pass` | `fail` | `partial`
- **Evidence**: specific JSON paths or field values demonstrating the status
- **Impact**: which consumer is affected and how
- **Recommendation**: what should change (if anything)

### Phase 3: Cross-reference with AST_REVISIONS.md

- Which checklist failures are already addressed by planned revisions?
- Which are **new** findings not yet tracked?
- Are there planned revisions that would break something that currently works for consumers?

### Phase 4: Report

Present findings grouped by severity:
1. **Blockers** — JSON output actively breaks or misleads consumers
2. **Gaps** — missing data that forces client-side workarounds
3. **Noise** — unnecessary data that inflates payloads or confuses consumers
4. **Style** — inconsistencies that don't break anything but hurt DX

**STOP here. Present findings and wait for approval. To execute changes, invoke `/project:address-review`.**

## Constraints

- **Focus on the output, not the implementation.** This review is about what the JSON *contains*, not how the Go code produces it. Implementation concerns belong in `/review-parser-internals`.
- **Test with real output.** Don't infer JSON structure from Go code alone — run the CLI and inspect actual output.
- **No backwards compatibility.** This is pre-v1. If the JSON shape should change, say so. Document breaking changes for the TS team.
- **Stay at the boundary.** Don't review TypeScript code. Don't review Go internals beyond what's needed to understand the output.
