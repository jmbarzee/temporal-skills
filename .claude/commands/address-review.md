# Address Review

Execute the groups produced by a review command. This command owns the inner loop: plan → approve → execute → validate → document → repeat.

Invoke this after any review command has produced a grouped finding plan and you are ready to begin addressing it.

## Input

The grouped plan from the review command should be present in conversation context. If it is not, ask the user to paste it or point to the file where it was written.

Each group in the plan should have:
- A theme name
- A list of findings with locations
- An estimated scope (which files are touched)

## Workflow

Repeat the following loop for each group, in order. Do not skip ahead.

### For Each Group:

**Step 1: Plan**

Before touching any code, write a concrete execution plan for this group:
- List every file that will change and why
- Identify which changes are independent (can be done in parallel by sub-agents)
- Identify which changes are sequential (must be ordered)
- Flag any finding in this group where the right approach is ambiguous

Present the plan. **Wait for approval before proceeding.**

**Step 2: Execute**

Carry out the plan. Where changes are independent, spawn parallel sub-agents — one per file or logical unit. Each sub-agent receives:
- The specific finding(s) it is addressing
- The file(s) it owns
- A constraint to make no changes outside its assigned scope

**Step 3: Validate**

Run the appropriate checks for the layer that was changed:
- Go code: `go build ./...` and `go test ./...` from `tools/lsp/`
- Skills/examples: `twf check` against affected `.twf` files
- TypeScript: `npm run build` from `tools/visualizer/`

If validation fails, fix before moving on. Do not paper over failures.

**Step 4: Summarize**

Present a brief summary:
- What changed (files + nature of change)
- Validation result
- Any new findings surfaced during execution that weren't in the original plan

**Step 5: Document**

Update `AST_REVISIONS.md`:
- Mark this group as completed
- Add any new findings surfaced during execution as new tracked items

**Step 6: Continue?**

Ask: proceed to the next group, or stop here?

If stopping: note which groups remain and confirm they are still recorded in `AST_REVISIONS.md` for the next session.

## Constraints

- **One group at a time.** Do not start group N+1 until group N is validated and documented.
- **Surgical changes only.** If execution reveals a larger problem, add it as a new finding — don't expand the current group's scope.
- **Sub-agents execute, don't decide.** Ambiguity gets escalated to the user, not resolved silently by a sub-agent.
- **Validation is not optional.** A group is not done until the build passes.
