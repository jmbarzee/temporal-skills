# TWF Visualizer — Product Vision

The TWF Visualizer renders Temporal workflow definitions from the `.twf` DSL as an interactive, explorable interface. It gives developers a visual understanding of their workflow system — both the fine-grained logic of individual definitions and the architectural relationships between them.

Two views compose into a single product:

- **Tree View** — a collapsible, color-coded block list. Answers questions about individual definitions: their logic, inputs/outputs, handlers, and call structure. The familiar interaction model (expand/collapse) makes it the default view on first load.
- **Graph View** — a force-directed dependency graph with semantic zoom. Answers questions about system architecture: which namespaces depend on which, how workers compose, and where workflows call across boundaries. A "power view" the user discovers via the tab bar.

The views share a visual identity, filtering vocabulary, and interaction patterns. Cross-view navigation lets the user follow a definition from one perspective to the other without losing context.


## User Goals

The visualizer answers two categories of developer questions:

### Individual definitions (served by Tree View)

1. **"What do these workflows do?"** — Read the step-by-step logic of any workflow, activity, or handler through recursive expand/collapse.
2. **"What are the inputs and outputs?"** — See signatures (params → return type) on every definition and call.
3. **"What handlers do these workflows expose?"** — See signal, query, and update declarations grouped at the top of each workflow body.
4. **"What does this call expand to?"** — Inline expansion shows the full body of any referenced workflow, activity, or nexus operation without navigating away.
5. **"What definitions exist in this file or package?"** — Filter and browse definitions by type, source file, and name.

### System architecture (served by Graph View)

6. **"What does this system look like?"** — A spatial overview of all namespaces, workers, and the definitions they host, with edges showing cross-boundary dependencies.
7. **"What depends on what?"** — Trace cross-worker and cross-namespace calls visually.
8. **"What is the blast radius of a change?"** — Hover a node and see its transitive dependents at any level of abstraction.
9. **"How is this namespace composed?"** — Zoom into a namespace to see its workers, then into a worker to see its workflows, activities, and services.

### Cross-cutting (served by both views together)

10. **"Where is this used?"** — Contextual navigation buttons (tree) and dependency highlighting (graph) surface reverse references — who calls this definition, what worker hosts it, what namespace contains it.
11. **"Show me the same thing from the other angle."** — "Show in [View]" navigates from a tree block to its graph node or vice versa, adjusting filters as needed to reveal the target.


## UX Principles

### Progressive disclosure

Collapse by default, expand on demand. Tree blocks start collapsed; the user drills into what interests them. The graph's semantic zoom starts at a high level of abstraction; the user zooms into finer detail. Information density is always under user control.

### Filter-as-source-of-truth

Filters are always authoritative. No feature bypasses filters — navigation adjusts them. When "Show in [View]" needs to reveal a hidden item, it makes the minimum filter changes required, animates the filter bar to show what changed, and lets the user see exactly what's filtered and reverse it. The user is never surprised by what's visible or hidden.

### Reactive composition

Features (search, filters, hover highlighting, force simulation) share state but don't directly manipulate each other. Each feature reads from shared data and produces its own output. This prevents cascading complexity as new features are added. The spec describes *what* each feature does; the implementation composes them through shared data, not inter-feature wiring.

### Focus + context

Highlight what matters; push everything else to the background. Hover a graph node and its dependency chain lights up while everything else dims. Search a name and matching nodes highlight while non-matches fade. The user always sees both the focal point and its surroundings.

### Direct manipulation

Drag nodes, scrub sliders, click to expand — all with immediate visual feedback. The graph's force simulation responds live to slider changes. Tree blocks expand and collapse on click. The control panel's sliders visibly animate during level transitions. The user acts on the visualization directly, not through menus or dialogs.


## Visual Identity

### Color palette

Each definition type has a dedicated color used consistently across both views. Colors are defined as CSS variables with three values per type:

- `--block-{type}-bg` — gradient background
- `--block-{type}-border` — border color
- `--block-{type}-text` — text/icon color

| Type | Color | Used In |
|------|-------|---------|
| Workflow | Purple | Tree blocks, graph nodes, call blocks |
| Activity | Blue | Tree blocks, graph nodes, call blocks |
| NexusService | Deep pink | Tree blocks, graph nodes, call blocks |
| Worker | Medium grey | Tree blocks, graph nodes |
| Namespace | Dark grey | Tree blocks, graph nodes |

Every color palette has a matching dark variant.

### Icon system

Icons are defined in a central theme map (`temporal-theme.tsx`). Most are Unicode text characters. Workflows and activities use custom SVG icons (interlocking gears and single gear respectively) for clarity at small sizes. Icons appear in tree block headers and graph node labels.

### Theming

- Light theme is the default.
- Dark theme activates via `.vscode-dark` class (VS Code webview) or `[data-theme="dark"]` attribute.
- Hover brightness shifts direction between themes (`0.95` in light, `1.1` in dark).

### Border conventions

| Context | Border weight | Style |
|---------|--------------|-------|
| Top-level definitions (workflow, activity, nexus service) | 2px | Solid |
| Call-level and statement-level blocks | 1px | Solid |
| Handler declarations (signal, query, update) | 1px | Dashed |
| Detached (fire-and-forget) calls | 1px | Dashed |
| Unresolved references | 1px | Dashed + reduced opacity + `?` badge |

### Node shapes (Graph View)

| Node Type | Shape | Size |
|-----------|-------|------|
| Namespace | Rounded rectangle | Large |
| Worker | Rectangle | Medium |
| Workflow / Activity / NexusService | Circle or pill | Small |

Level 3 node types share the same size tier but are distinguished by color and icon.

### Edge styles (Graph View)

| Edge Type | Line Style | Direction |
|-----------|-----------|-----------|
| Direct dependency | Solid | Arrowhead |
| Nexus dependency | Solid, distinct color | Arrowhead |
| Containment | Dashed | None (undirected) |
