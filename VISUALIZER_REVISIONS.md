# Visualizer Spec Revisions

Product vision review of TREE_VIEW.md and GRAPH_VIEW.md. Focused on spec completeness, user question coverage, and product coherence — not implementation status.

## Summary

**What's working well:**
- Graph View spec is user-goal-driven. Four explicit questions ground every design decision.
- Tree View spec is thorough about visual rendering: every block type, color palette, icon, and expand/collapse behavior is specified.
- The graph data model (Namespace → Worker → Workflow containment hierarchy + graph coarsening) is clean.
- Force simulation, semantic zoom, and animated level transitions are well designed.
- Tree view's inline cross-reference expansion (click a call → see the target's body in-place) is the right interaction model.

**What's missing:**
- Tree View has no user goals section — it describes *what it renders* but never *why*. Without goals, the feature set can't be evaluated.
- The graph model excludes activity and nexus call dependencies, which are major Temporal dependency vectors.
- No spec exists for how the two views compose into a single product (navigation, linked state, shared mental model).
- Reverse references ("who calls this?") are unaddressed in both views.
- The "blast radius" user goal is declared but underspecified — only immediate neighbors are shown.

**Blocked on parser data:**
- Reverse reference index: computable from forward references in the AST, no parser change needed. Visualizer can build it client-side.
- Activity-to-worker mapping: already in AST (workers list their registered activities).
- Nexus-to-namespace mapping: already in AST (namespaces list endpoints, services have operations with backing workflows).
- Signal/query/update flow graph: NOT in current AST. Would require parser to emit "workflow X sends signal Y to workflow Z" — currently the DSL only declares handlers, not signal targets.

---

## Group 1: Tree View User Goals ✅

**Priority:** Tier 1 — foundational. Every other tree view assessment depends on this.

**Status:** Complete. Added "User Goals" section to TREE_VIEW.md with 5 explicit user questions.

### Problem

The Graph View spec opens with a "User Goals" section that names four questions. The Tree View spec has none. It's a detailed rendering manual with no stated purpose. This makes it impossible to evaluate whether the feature set is right, whether features are missing, or what should be cut.

### User questions this addresses

These are the implicit questions the tree view already serves — they need to be made explicit:

1. **"What does this workflow do?"** — Read the step-by-step logic: calls, awaits, control flow, state changes.
2. **"What are the inputs and outputs?"** — See signatures (params → return type) on every definition and call.
3. **"What handlers does this workflow expose?"** — See signal, query, and update declarations grouped at the top of each workflow.
4. **"What does this call expand to?"** — Inline expansion shows the full body of any referenced workflow, activity, or nexus operation.
5. **"What definitions exist in this file?"** — Filter and browse by type, file, and name.

### Target experience

No behavior change. This is a spec revision: add a "User Goals" section to TREE_VIEW.md that frames the view's purpose. Every existing feature should trace back to one of these goals.

### Data requirements

None.

### Spec additions needed

- Add "User Goals" section to TREE_VIEW.md (parallel to Graph View's).
- Map each existing feature to the goal it serves. This may reveal features that serve no goal or goals that no feature serves.

---

## Group 2: Activity and Nexus Dependencies in Graph Model ✅

**Priority:** Tier 1 — without this, the graph can't answer "what depends on what?" for the most common Temporal dependency type.

**Status:** Complete. Level 3 expanded to include Workflows, Activities, and NexusServices. Dependency edges broadened to include cross-worker activity calls and nexus calls (traced to backing workflow). Nexus edges carry metadata shown on hover. Per-type Level 3 visibility noted as future work.

### Problem

The graph data model defines only **Workflow → Workflow** fundamental dependency edges. But in real Temporal systems, the most common cross-boundary dependencies are:

- **Workflow → Activity (on a different worker):** Workflow A calls Activity B, which is registered on Worker C. This creates a real Worker A's-worker → Worker C dependency, but the graph model doesn't capture it because Activity B isn't a Workflow node.
- **Workflow → Nexus operation (in a different namespace):** Workflow A calls NexusService.Operation, whose backing workflow runs in Namespace B. This creates a Namespace → Namespace dependency, but only if the graph traces through the nexus indirection.

Without these, the derived Worker → Worker and Namespace → Namespace edges are incomplete. A system where all cross-boundary communication happens through activities (extremely common in Temporal) would show zero inter-worker dependency edges.

### User questions this addresses

- "What depends on what?" — directly
- "What is the blast radius of a change?" — directly (missing edges = missing blast radius)
- "What does this system look like?" — edges ARE the system topology

### Target experience

The graph shows dependency edges for all cross-boundary calls, not just workflow-to-workflow. When a user hovers Worker A and sees its edges, they see dependencies created by activity calls and nexus calls, not just workflow calls.

### Data requirements

Activity-to-worker and nexus-to-namespace mappings are already in the AST:
- Workers list their registered activities and workflows.
- Namespaces list their nexus endpoints.
- Nexus operations reference backing workflows.

The graph construction step (step 4: "Resolve dependency edges from call/await references") needs to be broadened to include activity calls and nexus calls, not just workflow calls.

### Spec additions needed

- Expand "Fundamental Edges" to include:
  - **Workflow → Activity** ("calls") — a workflow calls an activity
  - **Activity → Worker** ("member of") — an activity is registered on a worker
- Expand "Derived Edges" derivation to project activity-call dependencies up through the Worker→Worker and Namespace→Namespace levels.
- Add equivalent treatment for nexus call dependencies.
- Decide: should Activity and NexusService appear as node types in the graph at certain zoom levels, or should they remain invisible (their dependencies projected onto Workflow/Worker/Namespace edges only)?

---

## Group 3: Cross-View Navigation ✅

**Priority:** Tier 1 — two disconnected views is not a product.

**Status:** Complete. Created NAVIGATION.md with tab switching model, "Show in [View]" contextual action (5-step animation sequence: switch → adjust filters → animate filter bar → animate view → flash target), independent filter state per view, shared filter vocabulary, visual consistency requirements. View-level filter sync left as future design question.

### Problem

The tree view and graph view are specced as independent views with no relationship. There's no spec for:
- How the user switches between views
- Whether selection state carries across views (select a workflow in graph → see it in tree)
- Whether the views can be shown simultaneously
- How colors, icons, and visual language align between views

Without cross-view navigation, the graph can answer "what does this system look like?" but can't lead the user to "what does this workflow do?" — that requires jumping to the tree view with context preserved.

### User questions this addresses

- All graph view questions eventually lead to tree view detail: overview → drill-down is the fundamental navigation pattern.
- "How is this namespace composed?" starts in graph, but understanding a specific workflow requires the tree.

### Target experience

The user sees a system overview in the graph. They click a workflow node. The tree view opens (or scrolls to) that workflow, expanded. They understand the detail, then return to the graph with their position preserved.

### Data requirements

Both views already consume the same `TWFFile` AST. Shared selection state is a UI concern.

### Spec additions needed

- New section: "View Navigation" covering:
  - View switcher UI (tabs? toggle? split?)
  - Click-through from graph node to tree view definition (and vice versa: tree definition → highlight in graph)
  - Shared selection/highlight state between views
  - Whether both views can be visible simultaneously (side-by-side)
- Ensure color palette and icon mapping are consistent between views (tree view's purple for workflows should match graph view's workflow node color).

---

## Group 4: Reverse References ("Who Calls This?") ✅

**Priority:** Tier 1 — required for "blast radius" and dependency understanding.

**Status:** Complete. Reframed from static data to contextual navigation buttons: hover-action buttons on blocks that navigate to callers, parent containers, or graph view. Actions vary by block type. Multiple targets show a popover for selection. Graph hover/selection explicitly includes both incoming and outgoing edges. Reverse index built client-side from forward references.

### Problem

Both views show forward references: "this workflow calls X." Neither shows reverse references: "X is called by these workflows." This is the difference between:

- Forward: "MyWorkflow calls ProcessPayment" (tree view inline expansion)
- Reverse: "ProcessPayment is called by MyWorkflow, OrderFlow, and RefundHandler"

The graph's hover interaction shows immediate edges, which includes reverse dependencies at the graph level. But the tree view has no equivalent — expanding a workflow shows what it calls, never what calls it.

The "blast radius" user goal fundamentally requires reverse references. You can't assess the impact of changing ProcessPayment without knowing its callers.

### User questions this addresses

- "What is the blast radius of a change?" — directly
- "What depends on what?" — the reverse direction

### Target experience

In the tree view, a workflow or activity definition shows its callers — the set of workflows that reference it. This could be:
- A "Referenced by" section in the definition header or body
- A badge showing caller count, expandable to show the list
- Callers are clickable to navigate to the calling workflow

In the graph view, selecting a node highlights not just immediate edges but can expand to show transitive dependents (all nodes that directly or transitively depend on the selected node).

### Data requirements

Computable client-side from the AST. The visualizer already builds `DefinitionContext` maps for forward resolution. A reverse index (definition name → list of call sites) can be built from the same data.

### Spec additions needed

- Tree View: add "Reverse References" section describing how callers are surfaced per definition.
- Graph View: extend "Selection" interaction to explicitly describe transitive dependency highlighting.

---

## Group 5: Blast Radius Spec ✅

**Priority:** Tier 2 — depends on Group 2 (complete edges) and Group 4 (reverse references). Elevates an existing user goal from "declared" to "designed."

**Status:** Complete. Reframed from selection-based to hover-based: default hover shows transitive downstream dependencies, modifier+hover shows transitive upstream (blast radius). Selection locks the highlight. Added hotkey discoverability spec. Note: Shift modifier conflicts with future multi-select spec — needs reconciliation later.

### Problem

"What is the blast radius of a change?" is listed as Graph View user goal #3, but the spec only provides:
- Hover: highlight node + immediate edges, dim everything else.
- Selection: stays highlighted, "optionally reveals info panel."

Neither interaction answers the blast radius question. Blast radius requires showing **transitive dependents** — not just the nodes directly connected, but everything that depends on them recursively. The current spec stops at one hop.

### User questions this addresses

- "What is the blast radius of a change?" — directly
- "What depends on what?" — the transitive version

### Target experience

The user selects a workflow node. The graph highlights that node and all nodes that transitively depend on it (direct callers, their callers, etc.) at whatever abstraction level is currently selected. Non-dependent nodes dim. The user sees the full "impact cone."

Optional depth control: a slider or stepper that expands the highlight one hop at a time, so the user can see 1-hop, 2-hop, 3-hop impact incrementally.

### Data requirements

Transitive dependency graph is computable from the edge set. Requires Group 2 (complete edges including activity/nexus) to be accurate.

### Spec additions needed

- Expand Graph View "Selection" section with explicit transitive dependency highlight behavior.
- Decide on interaction model: auto-transitive on select, or progressive expansion.
- Specify visual treatment for "impact depth" (e.g., opacity gradient by distance).

---

## Group 6: Graph Search and Filtering ✅

**Priority:** Tier 2 — the tree view has rich filtering; the graph has none.

**Status:** Complete. Added search and filtering spec to GRAPH_VIEW.md: source file filter, name search, hidden-match badges on filter controls, search result selection. Added reactive composition design principle. Future note: hidden-match badge pattern may be useful for tree view too.

### Problem

The tree view spec includes file filter, definition type toggles, and name search. The graph view has only the semantic zoom level selector. At scale (many namespaces, many workers), finding a specific node in the graph requires visual scanning.

### User questions this addresses

- "What does this system look like?" — filtered to the part I care about
- "What depends on what?" — focused on a specific area

### Target experience

The graph view has:
- **Name search**: type a name, matching nodes highlight (non-matches dim). Select a search result to center the viewport on it.
- **File filter**: same as tree view — filter visible nodes by source file.

The semantic zoom level selector already acts as a "type filter" (show only namespaces, or only workflows, etc.), so explicit type toggles may not be needed.

### Data requirements

Node names and source files are already in the AST.

### Spec additions needed

- Add "Search and Filtering" section to GRAPH_VIEW.md.
- Specify how search interacts with semantic zoom (search for a workflow name when only namespace-level is visible: auto-switch level? show nothing?).

---

## Group 7: Signal, Query, and Update Flows in Graph

**Priority:** Tier 3 — adds a new relationship dimension.

### Problem

The graph models "calls" relationships (workflow calls workflow, workflow calls activity). But Temporal workflows also communicate through:
- **Signals**: async messages sent to a running workflow
- **Queries**: synchronous reads of workflow state
- **Updates**: synchronous mutations of workflow state

These are fundamentally different from calls — they're messages, not invocations — and they represent real dependencies. A workflow that sends signals to another workflow depends on it, but this dependency is invisible in the current graph model.

### User questions this addresses

- "What depends on what?" — the messaging dimension
- "How do these workflows communicate?" — new question, not currently served

### Target experience

Signal/query/update flows appear as a second edge type in the graph (visually distinct from call edges — e.g., different line style or color). The user can toggle these on/off to reduce visual complexity.

### Data requirements

**Blocked on parser data.** The current DSL declares signal/query/update *handlers* on the receiving workflow but does not capture *senders*. The AST says "WorkflowA handles signal X" but not "WorkflowB sends signal X to WorkflowA." Without send-side data, flow edges can't be derived.

This would require either:
- DSL syntax for signal/query/update sends (e.g., `signal MyWorkflow.SignalName(...)`)
- Parser emitting send-side references when they appear as statements

### Spec additions needed

- Decide if message flows are in scope for the graph.
- If yes, define the data contract (what the parser needs to emit).
- Add edge types for signal/query/update flows with visual encoding.

---

## Group 8: State Persistence and Live Reload

**Priority:** Tier 3 — developer experience during active editing.

### Problem

Neither spec addresses what happens during the typical development loop: edit a `.twf` file → save → visualizer updates. Questions unanswered:

- **Tree view**: Do expand/collapse states persist across AST reloads? If a user has WorkflowA expanded three levels deep and saves their file, do they lose that context?
- **Graph view**: Does the layout position persist, or does the force simulation restart from scratch? Restart means the user loses spatial memory.
- **Both views**: Is there a loading/transition state while the new AST is parsed and delivered?

### User questions this addresses

- All questions — this is about the quality of the experience while iterating, not a specific question.

### Target experience

When the AST updates:
- Tree view preserves expand/collapse state for definitions that still exist. New definitions appear collapsed. Removed definitions disappear.
- Graph view preserves node positions for nodes that still exist. New nodes are seeded near their parent and the simulation reheats locally. Removed nodes fade out.
- Brief transition indicator (not a full loading screen) signals the update.

### Data requirements

Definitions need stable identity across AST versions. Currently, definitions are identified by name, which is sufficient unless names change.

### Spec additions needed

- Add "Live Reload Behavior" section to both specs.
- Define identity matching strategy (by name? by position?).
- Specify what "stable" means for each piece of view state.

---

## Group 9: Accessibility and Keyboard Navigation

**Priority:** Tier 3 — correctness and inclusivity.

### Problem

Neither spec addresses keyboard navigation or screen reader behavior:
- Tree view: no spec for arrow key navigation through blocks, Enter to expand/collapse, Tab to move between sections.
- Graph view: no spec for keyboard-driven node selection, focus cycling, zoom/pan without mouse.
- Neither view: no ARIA roles, live region announcements, or focus management.

### User questions this addresses

All questions — this is about who can use the product, not what it shows.

### Target experience

Tree view supports arrow key navigation (up/down to move between sibling blocks, right to expand, left to collapse — similar to VS Code's tree widget). Graph view supports Tab to cycle through nodes, Enter to select, arrow keys to pan.

### Data requirements

None.

### Spec additions needed

- Add "Keyboard Navigation" section to both specs.
- Define focus order, key bindings, and ARIA semantics.
