# Graph View

A second view for the Visualizer. The existing tree view shows definitions in isolation. This graph view shows **how definitions relate to each other** — which namespaces depend on which, how workers compose, and where workflows call across boundaries.

---

## User Goals

This view serves goals 6–9 (system architecture questions) from [PRODUCT.md](./PRODUCT.md) § User Goals.

---

## Graph Data Model

### Node Types (3 levels of a hierarchy)

| Level | Node Type    | Derived From                            |
|-------|--------------|-----------------------------------------|
| 1     | **Namespace** | Namespace definition                   |
| 2     | **Worker**    | Worker instantiation within a namespace |
| 3     | **Workflow**  | Workflow registered on a worker         |
| 3     | **Activity**  | Activity registered on a worker         |
| 3     | **NexusService** | Nexus service registered on a worker |

Level 3 contains all definition types that are registered on a Worker. These form a strict containment hierarchy: every Level 3 node belongs to exactly one Worker, and every Worker belongs to exactly one Namespace.

**Future:** Per-type visibility toggle at Level 3, allowing users to show/hide Workflows, Activities, and NexusServices independently.

### Fundamental Edges

**Containment edges** come from registration:

1. **Level 3 node → Worker** ("member of") — A workflow, activity, or nexus service is registered on a specific worker.
2. **Worker → Namespace** ("member of") — A worker is instantiated within a specific namespace.

**Dependency edges** come from cross-boundary calls. Only calls that cross a Worker boundary are included — same-worker calls are implicit in the containment hierarchy.

1. **Workflow → Workflow** ("calls") — A workflow calls or awaits another workflow on a different worker.
2. **Workflow → Activity** ("calls") — A workflow calls an activity on a different worker.
3. **Workflow → Workflow via nexus** ("calls via nexus") — A workflow calls a nexus operation whose backing workflow is on a different worker (or in a different namespace). The edge connects caller to backing workflow directly; the nexus service and operation are metadata on the edge, not intermediary nodes.

Nexus edges are visually distinct from direct call edges. Hovering a nexus edge reveals the endpoint, service, and operation. Edges sharing a nexus service or endpoint can be highlighted together to show shared scope.

### Derived Edges (Graph Coarsening)

Higher-level dependency edges are **derived** by projecting Level 3 dependency edges upward through the containment hierarchy:

1. **Worker → Worker** ("depends on") — Exists when any Level 3 node in Worker A depends on any Level 3 node in Worker B. Discard self-loops.
2. **Namespace → Namespace** ("depends on") — Exists when any Worker in Namespace A depends on any Worker in Namespace B. Discard self-loops.

### Graph Construction Order

1. Build Namespace nodes from namespace definitions.
2. Build Worker nodes from worker instantiations; attach each to its parent Namespace.
3. Build Workflow, Activity, and NexusService nodes from registrations on each worker; attach each to its parent Worker.
4. Resolve cross-worker dependency edges:
   a. Workflow → Workflow edges from cross-worker workflow calls and awaits.
   b. Workflow → Activity edges from cross-worker activity calls.
   c. Workflow → Workflow (via nexus) edges by tracing nexus calls through to their backing workflows.
5. Project Level 3 dependencies up to Worker-level; discard self-loops.
6. Project Worker-level dependencies up to Namespace-level; discard self-loops.

### Orphan Definitions

The containment hierarchy assumes every Level 3 node belongs to a Worker and every Worker belongs to a Namespace. But the DSL allows definitions without assignments — a workflow not registered on any worker, a worker not placed in any namespace.

**Treatment:** Orphan definitions appear in the graph as uncontained nodes, visually distinct from contained ones (e.g., positioned outside any grouping, with a subtle "unassigned" indicator such as a dashed outline or muted badge). They participate in dependency edges normally.

**Semantic zoom behavior:**
- Orphan workers (no namespace) are visible when Level 2 is selected.
- Orphan Level 3 nodes (no worker) are visible when Level 3 is selected.
- When their level is not in the selected range, orphans are hidden like any other node at that level.

**Data:** Orphan status is derivable from the AST — a definition is orphan if no worker or namespace references it. No parser changes needed.

---

## Layout: Force-Directed Simulation

The graph is rendered using a **force-directed layout** (also called a force simulation or spring-electrical model). Nodes are positioned by a continuous physics simulation where forces push and pull until the system reaches equilibrium.

### Force Types

| Force             | Applies To       | Behavior                                                                 |
|-------------------|------------------|--------------------------------------------------------------------------|
| **Charge (repulsion)** | Every node pair | Nodes repel each other, preventing overlap. Follows an inverse-square falloff (like electrostatic charge). |
| **Link (attraction)**  | Connected node pairs | Edges act as springs pulling connected nodes toward a target distance.    |
| **Center**        | All nodes        | A weak drift toward the viewport center to keep the graph from wandering. |

Each force has a **strength** parameter that controls its magnitude. These strengths are the primary tuning knobs for the layout.

### Per-Level Strength Parameters (8 total)

Three **charge strengths** (one per level):
- Level 1 (Namespace) node repulsion
- Level 2 (Worker) node repulsion
- Level 3 (Workflow/Activity/NexusService) node repulsion

Five **link strengths** (one per edge type):
- Namespace ↔ Namespace (dependency)
- Namespace ↔ Worker (containment)
- Worker ↔ Worker (dependency)
- Worker ↔ Level 3 (containment)
- Level 3 ↔ Level 3 (dependency)

### Simulation Lifecycle

1. **Initialize** — Place nodes at initial positions (see *Level Transitions* below).
2. **Tick** — On each animation frame, compute forces, update velocities, apply velocity damping (friction), update positions.
3. **Cool** — Over time, reduce the simulation's *alpha* (energy). As alpha approaches zero, the layout stabilizes and ticking can pause.
4. **Reheat** — When the graph structure changes (nodes added/removed, level transition), reset alpha to restart the simulation.

The simulation should use **requestAnimationFrame** for rendering, decoupled from the physics tick rate if needed for performance.

---

## Semantic Zoom: Level Selection

The three node levels (Namespace, Worker, Level 3 definitions) represent a **semantic zoom** — not a geometric magnification, but a change in the *level of abstraction* being displayed. The user selects which levels are visible.

### Level Selector Control

A **range selector** (not a dropdown, not independent toggles) that enforces a single contiguous span of levels. The user can select:

- Level 1 only (Namespaces)
- Levels 1–2 (Namespaces + Workers)
- Level 2 only (Workers)
- Levels 2–3 (Workers + Definitions)
- Level 3 only (Definitions)
- Levels 1–3 (all)

Display this as three horizontally arranged segments. A **bubble** or **highlight region** covers the selected span. Interaction model:

- **Click** a single level to select it alone.
- **Click-and-drag** across levels to select a contiguous range.
- The bubble resizes and slides to always cover a contiguous selection.

This is conceptually a **dual-thumb range slider** mapped to three discrete stops, but styled as a segmented capsule rather than a traditional slider track.

### Which Nodes and Edges Are Visible

When a level is selected, its nodes and its **intra-level** dependency edges are shown. When two adjacent levels are both selected, the **containment edges** between them are also shown. Specifically:

| Selected Levels | Visible Nodes               | Visible Edges                                     |
|-----------------|-----------------------------|----------------------------------------------------|
| 1               | Namespaces                  | Namespace → Namespace                              |
| 1–2             | Namespaces, Workers         | Namespace → Namespace, Namespace ↔ Worker, Worker → Worker |
| 2               | Workers                     | Worker → Worker                                    |
| 2–3             | Workers, Level 3 nodes      | Worker → Worker, Worker ↔ Level 3, Level 3 → Level 3 |
| 3               | Level 3 nodes               | Level 3 → Level 3                                  |
| 1–3             | All                         | All                                                |

Level 3 includes all node types registered on workers (Workflows, Activities, NexusServices). All types are shown by default. **Future:** per-type visibility toggle to show/hide specific Level 3 node types.

---

## Level Transitions (Animated)

Switching between levels should feel spatial and continuous, not like a hard cut. The design below supports eventual seamless animation even if the first implementation is simpler.

### Revealing a Lower Level (e.g., showing Workers beneath Namespaces)

1. **Seed positions** — Place incoming lower-level nodes at the position of their parent node (the one already on screen). This keeps the spatial context intact.
2. **Set initial forces** —
   - Containment link strength: **maximum** (children cling to parent).
   - Intra-level dependency link strength: **zero** (children ignore peer edges initially).
   - Child charge strength: **zero** (children don't repel yet).
3. **Animate in** — Over a transition duration (~400–600 ms, eased):
   - Fade child nodes and edges from fully transparent to fully opaque.
   - Ramp child charge strength up to its target value (children begin to spread out).
   - Ramp intra-level link strength up to its target value (dependency edges start pulling).
   - Ramp containment link strength down toward its resting target (parent grip loosens).
4. **Result** — Children fan out from their parent, finding their own equilibrium, with dependency edges guiding the final arrangement.

### Hiding a Lower Level (e.g., collapsing Workers back into Namespaces)

Reverse the process: fade out, collapse charge and link strengths, then remove child nodes when they've converged back to their parent position.

### Force Strength Interpolation

All strength transitions should be **interpolated over time** (not snapped). Use an easing curve (e.g., ease-in-out cubic) so the simulation smoothly adjusts. The sliders in the control panel should visibly animate in sync with the force changes, providing a direct visual mapping between the controls and the layout behavior.

---

## Visual Encoding

### Node Appearance

Each node type should be visually distinct using redundant encoding (don't rely on color alone):

| Node Type      | Shape Suggestion      | Size      |
|----------------|-----------------------|-----------|
| Namespace      | Rounded rectangle     | Large     |
| Worker         | Rectangle             | Medium    |
| Workflow       | Circle or pill        | Small     |
| Activity       | Circle or pill        | Small     |
| NexusService   | Circle or pill        | Small     |

Level 3 node types share the same size tier but are distinguished by color and icon (matching the tree view's existing color system: purple for workflows, blue for activities, pink for nexus services).

All nodes display their name as a label. Labels should remain legible at typical zoom levels — consider truncation with a tooltip for long names.

### Edge Appearance

| Edge Type                        | Line Style   | Direction Indicator |
|----------------------------------|--------------|---------------------|
| Direct dependency (→ same level) | Solid        | Arrowhead           |
| Nexus dependency (→ via nexus)   | Solid, distinct color | Arrowhead    |
| Containment (↔ adjacent levels)  | Dashed       | None (undirected)   |

Edge opacity and thickness can be secondary signals — thicker or more opaque for higher-traffic connections if multiplicity data is available in the future.

Nexus edges carry metadata (endpoint, service, operation) shown on hover. Hovering a nexus edge can highlight all edges sharing the same nexus scope (endpoint, service, or operation) to reveal shared routing.

### Color

See [PRODUCT.md](./PRODUCT.md) § Visual Identity for the shared color palette. Edges inherit the color of their source node, or use a neutral color to reduce visual noise.

---

## Viewport Controls

The graph lives on an infinite 2D canvas. Standard viewport interactions:

| Interaction       | Action                      |
|-------------------|-----------------------------|
| **Scroll / pinch** | Geometric zoom in/out      |
| **Click-drag on background** | Pan the viewport  |
| **Click-drag on a node**     | Drag the node; pin its position while dragging, release to unpin |
| **Double-click a node**      | Center and zoom to fit the node and its immediate neighbors (graph-native; not used for cross-view navigation) |

Dragging a node should **reheat** the simulation locally so nearby nodes can adjust.

### Fit-to-View

A button (or automatic behavior on level change) that adjusts the viewport to frame all currently visible nodes with padding.

---

## Control Panel

A collapsible sidebar or bottom drawer containing the tuning controls. This serves both as a power-user tool and as a transparency mechanism — users can see *why* the layout looks the way it does.

### Contents

1. **Level selector** — The contiguous range selector described above.
2. **Force strength sliders** — The 8 sliders (3 charge, 5 link). Grouped visually:
   - *Node repulsion* group (Level 1: Namespace, Level 2: Worker, Level 3: Definitions)
   - *Edge attraction* group (organized by edge type)
3. **Simulation controls** — Play/pause the simulation. Optionally a "shake" or reheat button to escape local minima.
4. **Presets** (optional, future) — Named slider configurations (e.g., "Tight clusters", "Spread out", "Namespace focus") that animate the sliders to known-good values.

Sliders should show their current numeric value and respond to direct input (click the number to type a value). All sliders should be **live** — dragging a slider immediately affects the running simulation.

---

## Errors Header

The graph view surfaces parse errors using the shared error handling pattern. See [VIEW_FRAMEWORK.md](./VIEW_FRAMEWORK.md) § Error Handling. The error bar appears between the graph header controls (filter chips, level selector) and the graph canvas. The canvas still renders whatever valid nodes and edges exist in the partial AST.

---

## Search and Filtering

The graph view shares a filtering vocabulary with the tree view (see [VIEW_FRAMEWORK.md](./VIEW_FRAMEWORK.md) § Shared Filter Vocabulary) but adds graph-specific dimensions.

### Filter Controls

| Filter | Behavior | Shared with Tree View |
|--------|----------|----------------------|
| **Source file** | Filter visible nodes by which `.twf` file defines them. Same chip-based UI as tree view. | Yes (same vocabulary, independent state) |
| **Name search** | Text input that matches node names (case-insensitive substring). Matching visible nodes highlight; non-matching visible nodes dim. | Yes (same vocabulary, independent state) |
| **Semantic zoom** | Level selector (described above). Controls which hierarchy levels are visible. | No (graph-specific) |
| **Level 3 type toggle** | Show/hide Workflows, Activities, NexusServices independently within Level 3. | Future |

### Search and Hidden Matches

Search matches against **all** nodes, not just visible ones. The results are split:

- **Visible matches** — nodes that match the search AND are shown at the current semantic zoom level and filter state. These are highlighted in the graph.
- **Hidden matches** — nodes that match the search but are excluded by filters (wrong zoom level, filtered file, toggled-off type). These are NOT shown in the graph.

Hidden match counts appear as **badge overlays** on the filter controls that are hiding them:
- If 3 matching workflows are hidden because semantic zoom is at Namespace-only, the level selector shows a badge: "3".
- If 2 matching nodes are hidden because their source file is filtered out, the corresponding file chip shows a badge: "2".

This lets the user discover that matches exist, understand *why* they're hidden, and decide whether to adjust filters to reveal them. Search informs but never overrides filters.

### Selecting a Search Result

Clicking a visible search match centers the viewport on it and selects it (triggering the dependency highlight from the Interaction States spec).

---

## Interaction States

### Hover: Dependency Highlighting

Hovering a node highlights its **transitive dependency chain**, not just immediate neighbors. The direction of traversal is controlled by a modifier key:

**Default hover (downstream):** Highlight the hovered node and all nodes it **transitively depends on** — follow outgoing dependency edges through the full call chain. This answers: "what does this node need?"

**Modifier+hover (upstream):** Hold a modifier key (e.g., Shift) while hovering to reverse direction. Highlight the hovered node and all nodes that **transitively depend on it** — follow incoming dependency edges through the full caller chain. This answers: "what breaks if I change this?" (blast radius).

In both modes:
- The hovered node and all highlighted nodes are shown at full opacity.
- All edges along the traversal path are highlighted.
- All other nodes and edges dim (reduce opacity to ~20–30%).
- Show a **tooltip** with the hovered node's full name and type.

The transitive chain follows edges at the **currently visible** abstraction level. If only Namespace-level is shown, the chain follows Namespace → Namespace edges. If all levels are shown, it follows the finest-grained edges available.

### Selection

Clicking a node selects it. A selected node:
- Stays highlighted even after the cursor moves away.
- Retains the dependency highlight from hover (downstream by default, upstream if modifier was held during click).
- Optionally reveals an info panel showing the node's properties (name, type, parent, connected nodes, callers, callees).
- Click the background or press Escape to deselect.

### Multi-Select (future consideration)

Lasso or modifier+click to select multiple nodes. Useful for "what connects these two namespaces?" queries. See [VIEW_FRAMEWORK.md](./VIEW_FRAMEWORK.md) § Keyboard Modifier Vocabulary for modifier key assignments — multi-select should use Ctrl/Meta+click (not Shift, which is reserved for dependency direction).

### Hotkey Discoverability

The graph view uses modifier keys for interaction variants (e.g., Shift+hover for upstream dependencies). These need to be discoverable:
- Tooltip hint on first hover (e.g., "Hold Shift to show dependents").
- A keyboard shortcut reference accessible from the control panel or a `?` button.
- Modifier state reflected in the UI — when Shift is held, a subtle indicator appears (e.g., the cursor changes, or a small label like "upstream" appears near the hovered node).

---

## Future: Message Flow Edges

The current graph models **call** relationships (workflow calls workflow, workflow calls activity). Temporal workflows also communicate through **messages** — signals, queries, and updates — which represent a different kind of dependency.

### Vision

When the DSL supports typed signal/query/update send statements (see `POSSIBLE_DSL_FEATURES.md`), the graph can derive **message flow edges** alongside call edges:

- **Workflow → Workflow** ("signals/queries/updates") — WorkflowA sends a signal to WorkflowB.

These are visually distinct from call edges (different line style or color) and toggleable — the user can show/hide message flow edges independently to manage visual complexity.

Message flow edges participate in the same systems as call edges:
- Graph coarsening projects them up to Worker → Worker and Namespace → Namespace.
- Transitive hover highlights follow them.
- Search and filtering apply to them.

### Data Contract (not yet available)

Requires the parser to emit typed send statements in the AST with:
- Source workflow (the sender)
- Target workflow (the receiver)
- Handler name (which signal/query/update)
- Message type (signal vs query vs update)

The DSL does not currently support send-side syntax. This feature is blocked on DSL and parser work.

### Design Anticipation

The edge data model, visual encoding tables, and interaction specs in this document are designed to accommodate message flow edges without structural changes. When the data becomes available:
- Add a new edge type to the Fundamental Edges table.
- Add a row to the Edge Appearance table (distinct line style for message flows).
- Add a toggle to the filter controls for message flow visibility.

---

## Performance Considerations

- **Node count** — For typical TWF projects, expect tens to low hundreds of nodes. A naive O(n²) charge calculation is acceptable at this scale. If needed later, apply a **Barnes-Hut approximation** (quadtree-based) to reduce to O(n log n).
- **Rendering** — Canvas-based rendering (2D context) will outperform SVG for larger node counts, but SVG is easier to style and integrate with React. Start with whichever matches the team's velocity; the simulation logic is renderer-agnostic.
- **Offscreen culling** — Only render nodes and edges within (or near) the visible viewport. The simulation still runs for all nodes.

---

## Live Reload

See [VIEW_FRAMEWORK.md](./VIEW_FRAMEWORK.md) § Live Reload for the shared reload behavior (identity matching, state preservation, transition indicator). Graph-view-specific reload details are documented there.

---

## Keyboard Navigation

The graph view supports keyboard navigation for node selection, viewport control, and interaction states.

### Key Bindings

| Key | Action |
|-----|--------|
| **Tab** | Cycle focus to the next node (order: by containment hierarchy, then alphabetical within peers) |
| **Shift+Tab** | Cycle focus to the previous node |
| **Enter** | Select the focused node (same as click — triggers dependency highlight) |
| **Escape** | Deselect all. Close any open panel or popover. |
| **Arrow keys** | Pan the viewport |
| **+** / **-** | Zoom in / out |
| **F** | Fit-to-view (frame all visible nodes) |
| **/** or **Ctrl+F** | Open search bar and focus the search input |
| **Space** | Toggle simulation play/pause |
| **?** | Toggle keyboard shortcut reference panel |

### Focus Indicator

The currently focused node has a visible focus ring (distinct from hover highlight and selection highlight). When a node is focused via keyboard, the tooltip appears as it would on mouse hover.

### Modifier Keys

See [VIEW_FRAMEWORK.md](./VIEW_FRAMEWORK.md) § Keyboard Modifier Vocabulary for modifier key semantics. The Shift modifier for upstream dependency highlighting (see § Interaction States) also works with keyboard focus — holding Shift while a node is focused reverses the transitive highlight direction.

### Accessibility

See [VIEW_FRAMEWORK.md](./VIEW_FRAMEWORK.md) § Accessibility for the shared accessibility approach (ARIA roles, focus indicators). Graph nodes are focusable elements with labels announcing node type and name.


## Cross-View Navigation

The graph view participates in the visualizer's cross-view navigation system. See [VIEW_FRAMEWORK.md](./VIEW_FRAMEWORK.md) for view switching, "Show in Tree" actions, shared filter vocabulary, and other shared behaviors.
