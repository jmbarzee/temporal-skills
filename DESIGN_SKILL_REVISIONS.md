# Design Skill Revisions

A guide for revising the `skills/design/` skill in cohesive, reviewable groups. Each group is a checkpoint — complete it, verify with `twf check` on all topic `.twf` files, then move to the next.

Within each group, **parallelizable** work is marked. Sub-agents should not cross group boundaries.

**Review origin:** `/project:review-skill skills/design` — 37 findings across 7 themes.

**Guiding principle:** Prefer removal over addition. A shorter, denser skill is almost always better.

---

## Group 1: Eliminate LANGUAGE.md Duplication ✅ COMPLETED

**Goal:** Remove the ~955-line near-duplicate of the canonical spec. Fix the single source of grammar truth.

**Why first:** Two concrete drift bugs exist today (missing priority sub-keys, missing error types). Every other group references grammar — this must be settled first.

**Result:** `reference/LANGUAGE.md` replaced with 5-line redirect to canonical spec. `SKILL.md` link updated. `common-errors.md` expanded with 6 missing error types, 9-row warnings table, and enhanced activity context explanation.

### 1a. Replace `reference/LANGUAGE.md` with a redirect

The file currently duplicates `tools/lsp/LANGUAGE_SPEC.md` with active drift. Replace the entire file with a short redirect:

```markdown
# TWF Language Reference

The canonical grammar specification lives at [`tools/lsp/LANGUAGE_SPEC.md`](../../../tools/lsp/LANGUAGE_SPEC.md).

Always consult the canonical spec for syntax questions. This redirect exists so that skill-internal links resolve.
```

### 1b. Update all skill-internal links

Every file that links to `reference/LANGUAGE.md` must be updated to point to `tools/lsp/LANGUAGE_SPEC.md` (or keep the redirect and accept one hop). Files to check:
- `SKILL.md` (line 112)
- `reference/notation-reference.md`
- `reference/notation-examples.md`
- `reference/common-errors.md`
- Any topic `.md` files

### 1c. Sync `reference/common-errors.md` with canonical spec

Add the missing error types from `LANGUAGE_SPEC.md`:
- Explicit `task_queue` routing errors
- Implicit `task_queue` routing errors
- Empty worker/namespace/workflow/activity body warnings
- Detach-nexus-with-result-binding error

### Parallelism

1a and 1b are one atomic change. 1c is independent — **parallel with 1a+1b**.

### Files touched
- `reference/LANGUAGE.md` (rewrite to redirect)
- `SKILL.md`, various reference/topic `.md` files (link updates)
- `reference/common-errors.md` (add missing error types)

---

## Group 2: Tighten SKILL.md Top Level ✅ COMPLETED

**Goal:** Remove reference content from the top-level entry point. Every line in SKILL.md should create directional momentum, not pre-load syntax.

**Why second:** This affects every invocation of the skill. Reducing top-level noise amplifies all downstream improvements.

**Result:** SKILL.md reduced from 219 to 157 lines (-28%). TWF Syntax section (65 lines of rules, examples, basic structure) replaced with 4-line pointer. Content landed in `notation-examples.md` (Basic Structure + Activity Body Detail sections). `editor-setup.md` deleted; visualizer mention folded into Completion section.

### 2a. Extract TWF Syntax section to reference files

Lines 110–174 of SKILL.md contain:
- Activity body detail levels (3 examples) → move to `reference/notation-examples.md`
- Rules table (6 rows) → move to `reference/notation-reference.md` or `reference/common-errors.md`
- Basic Structure example (26 lines) → move to `reference/notation-examples.md`

Replace with a brief pointer:

```markdown
## TWF Syntax

Full grammar: [`LANGUAGE_SPEC.md`](../../../tools/lsp/LANGUAGE_SPEC.md). Quick reference: [`notation-reference.md`](./reference/notation-reference.md). Examples: [`notation-examples.md`](./reference/notation-examples.md).

All `.twf` must pass `twf check` before presenting to user.
```

### 2b. Remove or absorb `reference/editor-setup.md`

At 13 lines with no judgment content, this file doesn't justify its context-loading cost. Fold the single actionable line ("suggest the visualizer for complex control flow") into the SKILL.md Completion section as a one-liner. Delete the file and remove it from the Reference Index.

### Parallelism

2a and 2b are independent — **two agents in parallel**.

### Files touched
- `SKILL.md` (extract syntax, update reference index)
- `reference/notation-examples.md` (receive extracted content)
- `reference/notation-reference.md` (receive rules table)
- `reference/editor-setup.md` (delete)

---

## Group 3: Enrich Reference File Judgment ✅ COMPLETED

**Goal:** Transform reference files from pure lookup tables into judgment-teaching documents. Add *when to use*, *when not to use*, and decision criteria.

**Why third:** These are the highest-risk files for wrong AI output. An AI that consults primitives-reference and gets no selection guidance will guess.

**Depends on:** Group 1 (grammar links settled), Group 2 (extracted content landed in reference files)

**Result:** All 6 reference files enriched. primitives-reference gained Selection paragraphs + misuse warnings. core-principles renamed to "Determinism & Idempotency" (Option B) with 3 new non-determinism traps, idempotency selection criteria, and Temporal docs link. workflow-boundaries expanded from 23→40 lines with Nexus section and Common Mistakes. notation-reference gained decision hints on 12 rows + heartbeat row. notation-examples gained Async Patterns section (promise, detach, condition, switch, heartbeat, options — all passing `twf check`). design-checklist gained resolution pointers and 2 missing checks.

### 3a. `reference/primitives-reference.md` — add selection guidance and misuse warnings

Currently a pure lookup table. For each primitive group, add brief decision criteria:

- **External Communication**: When signal vs query vs update? (write-only / read-only / read-write is already hinted — make it the explicit decision rule, add "signal has no return value", "query must not modify state", "update validates before committing")
- **Async Coordination**: When promise vs sequential? When `await all` vs `await one`? When `detach`? ("detach = fire-and-forget, you cannot observe the result")
- **Time**: When `timer` vs activity-level timeout?
- **State**: When `condition` vs local variable?

Keep it terse — one sentence per decision point, not paragraphs.

### 3b. `reference/core-principles.md` — expand scope or rename

The file covers only determinism and idempotency but the title claims "core principles." Two options (pick one):

**Option A (expand):** Add brief sections on the other four completion criteria: workflow decomposition (link to workflow-boundaries.md), failure handling strategy, history management (link to long-running.md topics), and activity design. Keep each to 5–10 lines.

**Option B (rename + scope):** Rename to `determinism-idempotency.md`, update the Reference Index, and let other principles be covered by their dedicated reference/topic files.

Also:
- Add missing non-determinism traps: map iteration order, goroutines/threads, mutable global state
- Add selection criteria to the idempotency patterns table (when to pick each)
- Add a Temporal docs reference for the replay model

### 3c. `reference/workflow-boundaries.md` — substantial expansion

At 23 lines, this is too thin for one of the skill's primary goals. Either:
- Expand with the richer decision table from `topics/child-workflows.md` (don't duplicate — cross-reference or consolidate)
- Add the nexus boundary decision: when a child workflow should become a nexus call
- Add the "wrapper workflow" anti-pattern (single-activity child workflow)
- Fix "bounded completion time" → "short, predictable completion time"

### 3d. `reference/notation-reference.md` — add decision hints

Add brief annotations to the "Meaning" column:
- `promise p <- activity Foo()` → "Use when you need the result later, not immediately"
- `detach workflow Bar()` → "Fire-and-forget; you cannot observe the result"
- `await one:` → "First match wins; use for races, timeouts"
- etc.

Also add the missing `heartbeat` row.

### 3e. `reference/notation-examples.md` — add missing construct examples

Add examples for: `promise`, `detach`, `state:`/`condition`/`set`/`unset`, `switch`/`case`, `heartbeat`, call-level `options:`. These are the constructs where an AI is most likely to invent incorrect syntax.

Also add brief "why" annotations to the control flow example (when `await all` vs sequential, when `for` vs parallel).

### 3f. `reference/design-checklist.md` — add resolution pointers

Each checklist item should link to the reference or topic file that helps when the check fails. For example:
- "Each failure mode identified" → see `reference/anti-patterns.md`
- "Recovery strategy defined" → see `topics/patterns.md` (saga pattern)
- "Loops have deterministic bounds" → see `reference/core-principles.md`

Also add missing checks: non-deterministic data structure iteration, version-specific branching.

Remove deployment topology items that are redundant with `twf check` validation, or clarify they are for design review ("should this topology exist?"), not syntax validation.

### Parallelism

3a, 3b, 3c, 3d, 3e, 3f are independent files — **up to six agents in parallel**.

### Files touched
- `reference/primitives-reference.md`
- `reference/core-principles.md`
- `reference/workflow-boundaries.md`
- `reference/notation-reference.md`
- `reference/notation-examples.md`
- `reference/design-checklist.md`

---

## Group 4: Anti-Patterns Overhaul

**Goal:** Make the anti-patterns file the definitive catalog of Temporal design mistakes. Surface anti-patterns that are currently buried in topic files.

**Depends on:** Group 3 (primitives-reference misuse warnings settled — avoid duplication)

### 4a. Expand `reference/anti-patterns.md`

Currently 3 entries, all duplicating core-principles. Restructure into categories and add the missing high-frequency anti-patterns:

**Structural anti-patterns:**
- Unbounded history growth (no continue-as-new) — extract from `topics/long-running.md`
- Over-decomposition / wrapper workflows (one activity per child workflow)
- Monolithic workflow (everything in one workflow, no child decomposition)
- Large payloads in workflow state

**Primitive misuse:**
- Signal for RPC-style calls (no return value)
- Query that modifies state
- Blocking in query handlers
- Update without validation
- Detach when you need the result

**Activity anti-patterns:**
- Orchestration in activities (already present — keep)
- Non-idempotent activities (already present — keep, enrich)
- Non-determinism in workflows (already present — keep, add map iteration)

For each: show the wrong approach, explain *why* it's wrong (not just that it is), and show the correct alternative. Use consistent notation (TWF where possible, `pseudo` where TWF can't express it).

### 4b. Refocus `reference/common-errors.md` title

The file is a `twf check` troubleshooting guide, not a design error guide. Either:
- Rename to `twf-check-errors.md` (honest title)
- Or add a brief "Design Errors" section that cross-references `anti-patterns.md`

Also add the "why" to the activity context restriction error: explain that activities cannot contain temporal primitives because they run outside the replay-safe workflow context.

### Parallelism

4a and 4b are independent — **two agents in parallel**.

### Files touched
- `reference/anti-patterns.md` (major expansion)
- `reference/common-errors.md` (rename or augment)
- `SKILL.md` Reference Index (update if renamed)

---

## Group 5: Example Validity Fixes

**Goal:** Every `.twf` example passes `twf check`. Every `.md` code block is correctly tagged. `.twf` and `.md` companions tell the same story.

**Depends on:** Groups 1–4 (grammar and reference content settled)

### 5a. Fix `topics/versioning.twf` + `topics/versioning.md` alignment

**Critical.** The `.twf` uses boolean flags; the `.md` uses `patched()` (not a TWF keyword). Multiple `.md` code blocks tagged ` ```twf ` will fail `twf check`.

Fix:
- Re-tag all `patched()` code blocks in `.md` as ` ```pseudo `
- Add an explicit bridge paragraph explaining: the `.twf` file uses flag-based gating as the DSL-level representation; `patched()` is the SDK implementation detail that maps to these flags
- Ensure the `.twf` examples demonstrate the same versioning scenarios the `.md` discusses

### 5b. Fix `topics/task-queues.twf` validation errors

The only topic file that fails `twf check`. Three validation errors from intentional unregistered definitions. Fix by either:
- Registering the definitions on a worker with a comment explaining the intent
- Splitting the file so the unregistered section is a separate `.twf` that is expected to fail (with a note)

### 5c. Enrich `topics/patterns.twf` saga example

The saga `BookingWorkflow` is happy-path-only. `CompensateBooking` is defined but never connected. Add failure handling: show how/when `CompensateBooking` is invoked (e.g., via `close fail` in an error branch, or an explicit compensation trigger).

### 5d. Align `topics/patterns.md` state machine with `.twf`

The `.md` version of `DocumentApproval` diverges structurally from the `.twf` (different handler placements, missing timer expiration, different await forms). Update the `.md` to match the `.twf`, or vice versa — they must show the same design.

Also fix the "First Successful" variant: it uses `await all` but claims to be an early-exit pattern. Either rename it or use `await one` for a true race.

### 5e. Strengthen `topics/testing.twf` examples

The four workflows are too simple to exercise testing patterns. Replace with (or add) workflows that surface real testing challenges: a workflow with continue-as-new, parallel branches, or condition-based update handlers. The `.md` companion is excellent — the `.twf` should serve it.

### 5f. Decide on `topics/skill-basics.twf`

This is an unfocused syntax sampler with no companion `.md`, overlapping with SKILL.md and every other topic. Options:
- **Remove it.** The specialized topic files cover each construct better.
- **Rename to `syntax-sampler.twf` and add a brief `.md`** explaining its role as a quick reference for all constructs in one file.

### 5g. Fix minor `.md` structural issues

- `topics/long-running.md:104-139`: fix misplaced query/update indentation in entity example
- `topics/activities-advanced.md:182-223`: consistently tag local activity blocks as `pseudo`, not mixed TWF/pseudo

### Parallelism

5a through 5g are independent files — **up to seven agents in parallel**.

### Files touched
- `topics/versioning.twf`, `topics/versioning.md`
- `topics/task-queues.twf`
- `topics/patterns.twf`, `topics/patterns.md`
- `topics/testing.twf`
- `topics/skill-basics.twf` (rename or delete)
- `topics/long-running.md`
- `topics/activities-advanced.md`

### Validation

After all changes: `go run ./tools/lsp/cmd/twf check skills/design/topics/*.twf` — all files must pass.

---

## Group 6: Topic Scope Trimming

**Goal:** Remove content that drifts into SDK/ops territory. Topics should teach TWF design patterns, not platform configuration.

### 6a. Trim `topics/timers-scheduling.md` Schedules section

Lines 123–291 are YAML platform configuration that explicitly says "not TWF notation." Options:
- **Remove entirely** — schedules are infrastructure, not workflow design
- **Collapse to a 5-line note** acknowledging that cron workflows exist as a Temporal platform concept, with a pointer to Temporal docs

### 6b. Trim `topics/long-running.md` secondary sections

Lines 178–394 contain three minor history management variants (redundant with each other), signal handling across continue-as-new (SDK territory), search attributes (operational), and monitoring (operational). Collapse to the core lessons:
- Continue-as-new pattern (already well-covered)
- Entity workflow pattern (already well-covered)
- Remove or radically shorten everything after

### Parallelism

6a and 6b are independent — **two agents in parallel**.

### Files touched
- `topics/timers-scheduling.md`
- `topics/long-running.md`

---

## Group 7: Minor Cleanup

**Goal:** Small fixes that don't justify their own group.

### 7a. `reference/primitives-reference.md` vocabulary

- `unset` description: "clear" → "set to false"
- `worker` / `namespace` descriptions: clarify that namespace defines deployment topology

### 7b. `reference/notation-reference.md` missing heartbeat

Add `heartbeat` row to the reference table.

*Note: if 3d already handles this, skip.*

### 7c. `topics/child-workflows.twf` boilerplate reduction

Collapse the 16 trivial activity definitions (~50 lines) into fewer, more descriptive activities. Since activity bodies are free-form pseudocode, consolidate where the activity name and signature are self-explanatory.

### 7d. Cross-file name collision awareness

Topic `.twf` files reuse names (ValidateOrder, ProcessPayment, SendEmail, etc.) preventing combined `twf check`. Consider prefixing with topic domain or using distinct domain scenarios per file. This is low priority — each file passes individually — but improves tooling experience.

### Parallelism

7a–7d are independent — **four agents in parallel** (or one agent for all).

### Files touched
- `reference/primitives-reference.md`
- `reference/notation-reference.md`
- `topics/child-workflows.twf`
- Multiple topic `.twf` files (if name prefixing is adopted)

---

## Execution Order

| Group | Goal | Dependencies |
|-------|------|-------------|
| 1 | Eliminate LANGUAGE.md duplication | None |
| 2 | Tighten SKILL.md top level | None (can parallel with 1) |
| 3 | Enrich reference file judgment | Groups 1, 2 |
| 4 | Anti-patterns overhaul | Group 3 |
| 5 | Example validity fixes | Groups 1–4 |
| 6 | Topic scope trimming | None (can parallel with 3–5) |
| 7 | Minor cleanup | None (can parallel with anything) |

---

## Validation Checklist

After all groups:
- [ ] `go run ./tools/lsp/cmd/twf check skills/design/topics/*.twf` — all files pass
- [ ] No broken internal links (grep for `](./` and verify targets exist)
- [ ] `SKILL.md` Reference Index matches actual files
- [ ] No duplicate content between SKILL.md and reference files
- [ ] `reference/LANGUAGE.md` is a redirect, not a duplicate
