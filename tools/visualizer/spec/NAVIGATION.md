# View Navigation

The visualizer has two views — Tree View and Graph View — that show the same data from different perspectives. This document specifies how they compose into a single product: view switching, shared state, and cross-view navigation.


## View Model

The views use a **tab switching** model. One view is active at a time. Each view maintains its own state (filter selections, scroll/zoom position, expand/collapse state) independently.


## Tab Bar

A tab bar at the top of the visualizer with two tabs: **Tree** and **Graph**. The active tab is visually highlighted. Clicking a tab switches the active view.

**State on switch:** Each view preserves its own state across tab switches. Switching from Tree to Graph and back returns the Tree to exactly where it was — scroll position, expanded blocks, filter selections, search query. Same for Graph: zoom level, node positions, semantic zoom selection, simulation state.


## "Show in [View]" Action

A contextual action available on any identifiable item (a definition block in the tree, a node in the graph) that opens the other view focused on that item.

### Trigger Points

- **Tree View:** Action available on definition block headers (workflow, activity, worker, namespace, nexus service). Accessible via right-click context menu or a small icon in the block header.
- **Graph View:** Action available on nodes. Accessible via right-click context menu or double-click (which currently centers — this may need reconciliation).

### Behavior Sequence

When the user invokes "Show in [target view]" on an item:

1. **Switch tab** — The target view becomes active (instant, no animation).
2. **Adjust filters** — Make the minimum filter changes needed to ensure the item is visible in the target view:
   - If the item's type is toggled off, toggle it on.
   - If the item's source file is filtered out, add it to the file selection.
   - If the semantic zoom level (graph) excludes the item's level, adjust the range to include it.
   - Do not clear other active filters — only expand, never narrow.
3. **Animate filter bar** — The filter controls animate to reflect the changes (chips activate, toggles flip, level selector bubble slides). This gives the user a visual explanation of what changed.
4. **Animate view to target:**
   - **Tree View:** Smooth scroll to the target definition. Expand the target block's ancestry (any collapsed parents that contain it). Expand the target block itself.
   - **Graph View:** Pan and zoom to center the target node. Lock viewport focus on the target node until the simulation stabilizes (the node stops moving). The user can break the lock by manually panning.
5. **Flash target** — After the view has settled, briefly highlight the target item (a pulse or glow effect) to draw the eye.

### Filter-as-Source-of-Truth

Filters are always authoritative. "Show in [view]" never bypasses filters — it adjusts them. The user can always see exactly what's filtered and manually change it. The animation of the filter bar in step 3 makes the adjustment visible and reversible.


## Shared Filter Vocabulary

Both views support filtering by:
- **Source file** — which `.twf` files contribute definitions
- **Name search** — find definitions by name

Each view adds its own filter dimensions:
- **Tree View:** Definition type toggles (Namespace, Worker, Workflow, Activity, NexusService)
- **Graph View:** Semantic zoom level selector (which hierarchy levels are visible)

Filter state is **independent per view** by default. Each view's filters can diverge without affecting the other.

**Future:** A mechanism for the user to align filter state across views (e.g., "apply these filters to both views"). The right interaction pattern for this will depend on observed usage. Candidates include a sync toggle, a filter bar action, or a tab modifier. To be designed after cross-view navigation is in use.


## Visual Consistency

Both views must use the same color and icon vocabulary for shared concepts:

| Concept       | Tree View Color | Graph View Color |
|---------------|-----------------|------------------|
| Workflow      | Purple          | Purple           |
| Activity      | Blue            | Blue             |
| NexusService  | Pink            | Pink             |
| Worker        | Grey            | Grey             |
| Namespace     | Dark grey       | Dark grey        |

The existing tree view color palette (defined in CSS variables and `temporal-theme.tsx`) is the authoritative source. The graph view adopts it.
