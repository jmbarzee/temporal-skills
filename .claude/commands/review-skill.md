# Skill Review

Review an AI skill for quality, focus, and effectiveness. Evaluate whether the skill makes an AI maximally capable at its stated goal.

## Skill Selection

If a skill path is not provided in context, list the contents of `skills/` and present the options. Wait for selection before proceeding.

## Context

Read in this order before starting:
1. `{skill}/README.md` — the declared goal, scope, and structure. This is the calibration point for the entire review.
2. `{skill}/SKILL.md` — the top-level entry point the AI reads first.

If `README.md` does not exist, stop and report it. Do not attempt the review.

## Review Rubric

A skill is a context budget. Every word is an expenditure. These lenses evaluate whether the budget is spent well.

### 1. Top-Level Focus
- Does `SKILL.md` open with a razor-sharp statement of the goal from `README.md`?
- Is the top level free of content that belongs in reference files?
- Would an AI reading only `SKILL.md` have strong directional momentum toward the right outcome — or diffuse awareness of many concerns?
- What is the ratio of goal-setting to reference content at the top level?

### 2. Information Density
- Is vocabulary precise? Are there vague or hedging phrases that could be tightened?
- Does each section earn its place? Could anything be removed without losing guidance?
- Are there verbose explanations where a single precise term or example would suffice?
- Does the language signal to the AI that every word matters — or does noise train it to expect noise?

### 3. Gradual Context Expansion
- Is the skill structured so the AI can navigate to depth *on demand*, rather than being pre-loaded?
- Does `SKILL.md` make it clear *when* and *why* to consult each reference file?
- Can the AI self-direct through the skill's structure based on the task at hand, or must it read everything upfront?
- Do reference files stay focused on their topic, or do they duplicate content from `SKILL.md`?

### 4. Judgment Guidance
- Does the skill teach *when* to use each construct, not just *how*?
- Are the non-obvious decision points documented? (e.g., "use X when Y, not when Z")
- Are there common judgment calls the skill leaves implicit that should be made explicit?

### 5. Anti-Pattern Coverage
- Are failure modes documented? What does wrong look like?
- Are there tempting-but-incorrect approaches that an AI would plausibly generate?
- Does the skill guard against confident-but-wrong output in high-risk areas?

### 6. Example Quality
- Do examples exercise real complexity, not just happy paths?
- Are examples succinct — demonstrating the concept without noise?
- Do examples use the full expressiveness of the DSL/SDK, or fall back to simpler constructs?
- Run `twf check` against all `.twf` examples: do they pass?

### 7. Grounding
- Are authoritative references (LANGUAGE_SPEC.md, Temporal docs MCP) used instead of relying on trained knowledge?
- Are there areas where the skill trusts the AI's memory for something that should be looked up?

### 8. Scope & Handoff
- Are scope boundaries explicit? What does this skill explicitly not do?
- Is the handoff to adjacent skills (e.g., design → author-go) clearly documented?
- Does the skill declare what it produces and what it expects to consume?

## Workflow

**Follow this phased approach strictly. Do not skip or combine phases.**

### Phase 1: Explore

Use sub-agents in parallel:
- **Top-level agent**: Read `README.md` and `SKILL.md`. Evaluate against lenses 1–3 (focus, density, expansion). Return structured findings.
- **Reference agents**: One agent per file in `reference/`. Each evaluates its file against lenses 2, 4, 5, and 7 (density, judgment, anti-patterns, grounding).
- **Example agent**: Read all files in `topics/` or equivalent. Run `twf check` on all `.twf` files. Evaluate against lens 6 (example quality). Return findings + check results.

### Phase 2: Catalog

Each finding must include:
- **Location**: `file:section` or `file:line`
- **Lens**: which rubric section (1–8)
- **Severity**: `critical` | `moderate` | `minor`
- **Theme**: a short grouping label (e.g., "top-level dilution", "missing judgment guidance", "stale example", "scope creep")
- **Finding**: 1–2 sentences on the issue and why it costs the AI

### Phase 3: Group & Prioritize

Group by theme. Order by:
1. Top-level focus and density issues first — these affect every use of the skill
2. Missing judgment and anti-patterns — highest risk for wrong output
3. Example and grounding issues
4. Minor density and structure improvements

**STOP. Present the grouped plan and wait for approval. To execute, invoke `/project:address-review`.**

## Constraints
- **Evaluate against the skill's own stated goal.** `README.md` defines intent. Judge the skill against that, not against what you think it should do.
- **No `twf check` substitutes for judgment.** Passing check means syntactically valid, not well-designed.
- **Prefer removal over addition.** A shorter, denser skill is almost always better. Resist the urge to add content — surface what should be cut.
- **Don't rewrite the skill during review.** Catalog and group first. Changes happen in `address-review`.
