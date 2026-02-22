# Skills Review

Review the AI skill definitions in `skills/` for accuracy, completeness, and alignment with the current DSL and parser state.

Skills drift silently — they don't fail tests when the DSL changes. This review catches that drift.

## Context

Before starting, read:
- `CHANGELOG.md` — what has changed in the DSL and parser recently
- `tools/lsp/LANGUAGE_SPEC.md` — the current formal grammar spec (ground truth for the DSL)
- `AST_REVISIONS.md` — changes in flight that will land soon
- `skills/design/SKILL.md` and `skills/author-go/SKILL.md` — the skill entry points

## Review Rubric

### 1. DSL Accuracy (Design Skill)
- Does every construct documented in `skills/design/reference/` still parse correctly?
- Are there constructs in `LANGUAGE_SPEC.md` that are absent or poorly documented in the skill?
- Do the `.twf` examples in `skills/design/topics/` pass `twf check`?

### 2. Mapping Accuracy (Author-Go Skill)
- Does the TWF → Go SDK mapping in `skills/author-go/reference/` reflect current DSL semantics?
- Are there DSL constructs that have no corresponding author-go documentation?
- Are the Go SDK examples syntactically and semantically correct for the current Temporal SDK version?

### 3. Coverage Gaps
- What DSL features are new (per CHANGELOG) that lack skill documentation?
- What skill documentation covers deprecated or removed features?

### 4. Clarity & Quality
- Are examples succinct and non-contrived?
- Do examples use the full expressiveness of the DSL, or fall back on simpler constructs?
- Is the skill documentation consistent in voice and structure across topics?

## Workflow

### Phase 1: Validate Examples
Run `twf check` against all `.twf` files in `skills/`:
```
twf check skills/design/topics/*.twf
```
Capture all errors and warnings.

### Phase 2: Audit Documentation
Use sub-agents in parallel — one per skill topic area:
- `design/reference/` — each reference file vs `LANGUAGE_SPEC.md`
- `design/topics/` — each example file for accuracy and quality
- `author-go/reference/` — each mapping file for DSL and SDK accuracy

### Phase 3: Catalog
[Standard finding format. Theme suggestions: "missing coverage", "stale construct", "check failure", "mapping gap", "example quality"]

### Phase 4: Group & Prioritize
**STOP. Present plan and wait for approval. To execute, invoke `/project:address-review`.**

## Constraints
- **Use `twf check` extensively.** Don't trust example correctness by reading alone.
- **Don't modify the parser or DSL.** If the skill documents a feature that doesn't exist, flag it — don't add the feature.
- **Prefer improving existing examples** over creating new ones. More docs ≠ better docs.
