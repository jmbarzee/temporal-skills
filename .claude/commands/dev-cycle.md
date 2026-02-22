# Development Cycle

Orchestrate a full development cycle across the repository — from review through execution, documentation, and downstream propagation.

Use this command when you want to run the complete loop, not just a single review. It drives the process described in the project's development loop, coordinating the specific review commands and propagation steps in sequence with appropriate approval gates.

## The Loop

```
[Identify] → [Group] → [Iterate Groups] → [Document] → [Propagate] → [Repeat or Close]
```

## Workflow

### Phase 1: Scope

Before launching any reviews, ask the user:
1. What is the **starting point**? Options:
   - "Fresh cycle" — full review from scratch
   - "Resume" — groups are already identified, provide the plan file or describe current state
   - "Post-change" — changes were made, now need documentation and propagation only
2. Which **layers** should be reviewed this cycle?
   - Parser internals (`/review-parser-internals`)
   - Parser output / JSON contract (`/review-parser-output`)
   - Visualizer TypeScript (`/review-visualizer`)
   - Skills alignment (`/review-skills`)
   - All of the above

Present the proposed review scope and **wait for confirmation** before starting.

### Phase 2: Discovery (Parallel)

Launch the confirmed review commands as parallel sub-agents. Each sub-agent runs its full Phase 1 + Phase 2 (Explore + Catalog) and returns a structured finding list. Do NOT have sub-agents proceed to execution — catalog only.

Collect all findings from all sub-agents.

### Phase 3: Unified Catalog & Grouping

Merge all findings into a single cross-layer catalog. Deduplicate (the same root cause may surface in multiple layers). Apply themes across layers — a theme like "resolver error model" may span parser internals, LSP server, and visualizer.

Group by theme with cross-layer impact noted. Order by:
1. Foundation-first (parser before visualizer, design skill before author-go)
2. Critical severity before moderate
3. High parallelism groups before sequential ones

Present the unified grouped plan.

**STOP. Wait for user approval and group selection before any execution.**

### Phase 4: Execute Groups

Work through approved groups, one at a time:
1. Announce which group is starting
2. If the group spans multiple independent files/layers, spawn parallel sub-agents for execution
3. Validate: run `go build ./...`, `go test ./...`, `twf check`, `npm run build` as appropriate
4. Present a brief diff summary
5. Ask: "Continue to next group, or stop here?"

### Phase 5: Document

After completing all approved groups:
1. Update `AST_REVISIONS.md` — mark completed items, add new breaking changes found
2. Create a brief session summary: what was changed, what was skipped, what remains
3. List any skipped groups with rationale

**STOP. Present the session summary for review.**

### Phase 6: Assess Next Cycle

Based on what changed:
1. Run `/propagate-changes` to assess downstream impact
2. Recommend which review commands should be run in the next cycle
3. Identify any groups that were skipped and should carry forward

Present the next-cycle recommendation.

**STOP. User decides whether to continue or close the cycle.**

## Approval Gates (Summary)

| Gate | What you see | Decision |
|------|-------------|----------|
| Phase 1 end | Proposed review scope | Confirm or narrow scope |
| Phase 3 end | Unified grouped plan | Approve groups, reorder, drop |
| Phase 4 (each group) | Diff summary | Continue or pause |
| Phase 5 end | Session summary | Accept documentation |
| Phase 6 end | Next-cycle recommendations | Continue cycle or close |

## Constraints
- **Never skip an approval gate.** The value of this command is structured human oversight, not automation.
- **Prefer narrow scope over broad.** It's better to finish 3 groups well than start 8 and finish none.
- **One thing changes at a time.** Don't let a group balloon. If a fix surfaces new work, add it to the next cycle, don't expand the current group.
- **State is persistent.** After each gate, write current state to `AST_REVISIONS.md` so the cycle can be resumed in a new session if needed.
