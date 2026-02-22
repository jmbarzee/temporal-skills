# Graph View

A second view for the Visualizer. The existing tree view shows definitions in isolation. This graph view shows **how definitions relate to each other** — which namespaces depend on which, how workers compose, and where workflows call across boundaries.

---

## User Goals

The graph view answers questions that the tree view cannot:

1. **"What does this system look like?"** — A spatial overview of all namespaces, workers, and workflows, and the edges between them.
2. **"What depends on what?"** — Trace a workflow's cross-worker and cross-namespace calls visually.
3. **"What is the blast radius of a change?"** — Select a node and see its transitive dependents at any level of abstraction.
4. **"How is this namespace composed?"** — Zoom into a namespace to see its workers, then into a worker to see its workflows.

---

## Graph Data Model

### Node Types (3 levels of a hierarchy)

| Level | Node Type    | Derived From                            |
|-------|--------------|-----------------------------------------|
| 1     | **Namespace** | Namespace definition                   |
| 2     | **Worker**    | Worker instantiation within a namespace |
| 3     | **Workflow**  | Workflow implementation within a worker |

These form a strict containment hierarchy: every Workflow belongs to exactly one Worker, and every Worker belongs to exactly one Namespace.

### Fundamental Edges

These edges come directly from the data:

1. **Workflow → Workflow** ("depends on") — A workflow calls or awaits another workflow.
2. **Workflow → Worker** ("member of") — A workflow is registered on a specific worker.
3. **Worker → Namespace** ("member of") — A worker is instantiated within a specific namespace.

### Derived Edges (Graph Coarsening)

Higher-level dependency edges are **derived** by projecting lower-level edges upward through the containment hierarchy:

1. **Worker → Worker** ("depends on") — Exists when any workflow in Worker A depends on any workflow in Worker B. Discard self-loops (both ends resolve to the same worker).
2. **Namespace → Namespace** ("depends on") — Exists when any worker in Namespace A depends on any worker in Namespace B. Discard self-loops (both ends resolve to the same namespace).

### Graph Construction Order

1. Build Namespace nodes from namespace definitions.
2. Build Worker nodes from worker instantiations; attach each to its parent Namespace.
3. Build Workflow nodes from workflow implementations; attach each to its parent Worker.
4. Resolve Workflow → Workflow dependency edges from call/await references.
5. Project Workflow-level dependencies up to Worker-level; discard self-loops.
6. Project Worker-level dependencies up to Namespace-level; discard self-loops.

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

### Per-Type Strength Parameters (8 total)

Three **charge strengths** (one per node type):
- Namespace node repulsion
- Worker node repulsion
- Workflow node repulsion

Five **link strengths** (one per edge type):
- Namespace ↔ Namespace (dependency)
- Namespace ↔ Worker (containment)
- Worker ↔ Worker (dependency)
- Worker ↔ Workflow (containment)
- Workflow ↔ Workflow (dependency)

### Simulation Lifecycle

1. **Initialize** — Place nodes at initial positions (see *Level Transitions* below).
2. **Tick** — On each animation frame, compute forces, update velocities, apply velocity damping (friction), update positions.
3. **Cool** — Over time, reduce the simulation's *alpha* (energy). As alpha approaches zero, the layout stabilizes and ticking can pause.
4. **Reheat** — When the graph structure changes (nodes added/removed, level transition), reset alpha to restart the simulation.

The simulation should use **requestAnimationFrame** for rendering, decoupled from the physics tick rate if needed for performance.

---

## Semantic Zoom: Level Selection

The three node levels (Namespace, Worker, Workflow) represent a **semantic zoom** — not a geometric magnification, but a change in the *level of abstraction* being displayed. The user selects which levels are visible.

### Level Selector Control

A **range selector** (not a dropdown, not independent toggles) that enforces a single contiguous span of levels. The user can select:

- Level 1 only (Namespaces)
- Levels 1–2 (Namespaces + Workers)
- Level 2 only (Workers)
- Levels 2–3 (Workers + Workflows)
- Level 3 only (Workflows)
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
| 2–3             | Workers, Workflows          | Worker → Worker, Worker ↔ Workflow, Workflow → Workflow |
| 3               | Workflows                   | Workflow → Workflow                                |
| 1–3             | All                         | All                                                |

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

| Node Type   | Shape Suggestion      | Size      |
|-------------|-----------------------|-----------|
| Namespace   | Rounded rectangle     | Large     |
| Worker      | Rectangle             | Medium    |
| Workflow    | Circle or pill        | Small     |

All nodes display their name as a label. Labels should remain legible at typical zoom levels — consider truncation with a tooltip for long names.

### Edge Appearance

| Edge Type                        | Line Style   | Direction Indicator |
|----------------------------------|--------------|---------------------|
| Dependency (→ same level)        | Solid        | Arrowhead           |
| Containment (↔ adjacent levels)  | Dashed       | None (undirected)   |

Edge opacity and thickness can be secondary signals — thicker or more opaque for higher-traffic connections if multiplicity data is available in the future.

### Color

Use the existing visualizer color palette as a starting point. Assign a distinct hue per node type. Edges inherit the color of their source node, or use a neutral color to reduce visual noise.

---

## Viewport Controls

The graph lives on an infinite 2D canvas. Standard viewport interactions:

| Interaction       | Action                      |
|-------------------|-----------------------------|
| **Scroll / pinch** | Geometric zoom in/out      |
| **Click-drag on background** | Pan the viewport  |
| **Click-drag on a node**     | Drag the node; pin its position while dragging, release to unpin |
| **Double-click a node**      | Center and zoom to fit the node and its immediate neighbors |

Dragging a node should **reheat** the simulation locally so nearby nodes can adjust.

### Fit-to-View

A button (or automatic behavior on level change) that adjusts the viewport to frame all currently visible nodes with padding.

---

## Control Panel

A collapsible sidebar or bottom drawer containing the tuning controls. This serves both as a power-user tool and as a transparency mechanism — users can see *why* the layout looks the way it does.

### Contents

1. **Level selector** — The contiguous range selector described above.
2. **Force strength sliders** — The 8 sliders (3 charge, 5 link). Grouped visually:
   - *Node repulsion* group (Namespace, Worker, Workflow)
   - *Edge attraction* group (organized by edge type)
3. **Simulation controls** — Play/pause the simulation. Optionally a "shake" or reheat button to escape local minima.
4. **Presets** (optional, future) — Named slider configurations (e.g., "Tight clusters", "Spread out", "Namespace focus") that animate the sliders to known-good values.

Sliders should show their current numeric value and respond to direct input (click the number to type a value). All sliders should be **live** — dragging a slider immediately affects the running simulation.

---

## Interaction States

### Hover

Hovering a node should:
- Highlight the node and all its immediate edges.
- Dim all other nodes and edges (reduce opacity to ~20–30%).
- Show a **tooltip** with the node's full name and type.

### Selection

Clicking a node selects it. A selected node:
- Stays highlighted even after the cursor moves away.
- Optionally reveals an info panel showing the node's properties (name, type, parent, connected nodes).
- Click the background or press Escape to deselect.

### Multi-Select (future consideration)

Shift-click or lasso to select multiple nodes. Useful for "what connects these two namespaces?" queries.

---

## Performance Considerations

- **Node count** — For typical TWF projects, expect tens to low hundreds of nodes. A naive O(n²) charge calculation is acceptable at this scale. If needed later, apply a **Barnes-Hut approximation** (quadtree-based) to reduce to O(n log n).
- **Rendering** — Canvas-based rendering (2D context) will outperform SVG for larger node counts, but SVG is easier to style and integrate with React. Start with whichever matches the team's velocity; the simulation logic is renderer-agnostic.
- **Offscreen culling** — Only render nodes and edges within (or near) the visible viewport. The simulation still runs for all nodes.

---

## Summary of Patterns Used

| Pattern                        | Purpose                                                     |
|--------------------------------|-------------------------------------------------------------|
| **Force-directed layout**      | Automatic, aesthetically pleasing spatial arrangement        |
| **Semantic zoom**              | Navigate abstraction levels, not just magnification         |
| **Graph coarsening**           | Derive higher-level relationships from lower-level data      |
| **Containment hierarchy**      | Strict parent-child nesting gives structure to the graph     |
| **Animated interpolation**     | Smooth transitions preserve spatial context during changes   |
| **Focus + context (dimming)**  | Highlight what matters; push everything else to the background |
| **Direct manipulation**        | Drag nodes, drag the viewport, scrub sliders — all immediate |
| **Linked views (sliders ↔ simulation)** | Controls reflect simulation state; changes flow both ways  |
