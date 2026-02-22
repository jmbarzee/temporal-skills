# Visualizer Spec Revisions — Round 3 (Reorganization + Content)

Combines the spec reorganization plan with Round 2's content findings into a single execution sequence. Each group touches a primary file, and the ordering ensures earlier groups establish structure that later groups reference.

## Summary

**Reorganization goal:** Move from 3 files (TREE_VIEW, GRAPH_VIEW, NAVIGATION) to 4 files with clearer ownership:

| File | Owns |
|------|------|
| **PRODUCT.md** (new) | Product vision, combined user goal hierarchy, UX principles, visual identity system |
| **VIEW_FRAMEWORK.md** (renamed from VIEW_FRAMEWORK.md) | All shared view behaviors: view composition, cross-view nav, shared filtering, live reload, empty states, error handling, accessibility, keyboard modifier vocabulary |
| **TREE_VIEW.md** (slimmed) | Tree-view-specific rendering, block anatomy, statement types, expand/collapse, contextual nav buttons |
| **GRAPH_VIEW.md** (slimmed) | Graph-view-specific data model, layout, semantic zoom, interaction states, force simulation |

**Round 2 content findings** are folded into the reorganization where they naturally fit. Each is tagged with its Round 2 origin.

---

## Group 1: Create PRODUCT.md

**Primary file:** `tools/visualizer/spec/PRODUCT.md` (new)

This document is the top-level entry point for understanding the visualizer as a product. It answers: what is this, who is it for, and what principles guide design decisions?

### Contents to write

**1a. Product vision statement.**
What the visualizer is, what it isn't, how the two views compose. The tree view answers questions about individual definitions and their contents. The graph view answers questions about how definitions relate across boundaries. Together they give a developer both the detail and the big picture.

**1b. Combined user goal hierarchy.**
Extract from both view specs and reframe as product-level goals. Each view spec currently has its own User Goals section — the product doc should present the unified set, noting which view(s) serve each goal. View specs will then cross-reference this section rather than maintaining their own goals.

Current goals to consolidate:
- Tree: "What do these workflows do?", "What are the inputs and outputs?", "What handlers do these workflows expose?", "What does this call expand to?", "What definitions exist in this file or package?"
- Graph: "What does this system look like?", "What depends on what?", "What is the blast radius of a change?", "How is this namespace composed?"

**1c. Core UX principles.**
These are currently scattered as one-off remarks in the specs. Consolidate them as named principles:
- **Progressive disclosure** — collapse by default, expand on demand (tree blocks, graph semantic zoom)
- **Filter-as-source-of-truth** — filters are always authoritative; navigation adjusts filters, never bypasses them
- **Reactive composition** — features (search, filters, hover, simulation) share state but don't manipulate each other; compose through shared data, not inter-feature wiring
- **Focus + context** — highlight what matters, dim everything else (hover dependency chains, search matching)
- **Direct manipulation** — drag nodes, scrub sliders, click to expand; all immediate feedback

**1d. Visual identity system.**
Extract from TREE_VIEW.md § Visual Design and VIEW_FRAMEWORK.md § Visual Consistency. Consolidate into one authoritative source:
- Color palette by definition type (the existing CSS variable naming + color assignments)
- Icon system (theme map, SVG vs Unicode)
- Theming (light/dark, activation mechanism, hover brightness direction)
- Border conventions (2px for top-level definitions, 1px for statements, dashed for handlers and detached calls)

### Decisions

- Each view spec retains a brief cross-reference listing which product goals it serves, linking to PRODUCT.md for the full hierarchy.
- The file structure / architecture overview stays in TREE_VIEW.md as implementation context — PRODUCT.md is product vision, not codebase docs.

---

## Group 2: Expand VIEW_FRAMEWORK.md — Shared View Behaviors

**Primary file:** `tools/visualizer/spec/VIEW_FRAMEWORK.md`

VIEW_FRAMEWORK.md currently covers cross-view navigation (tabs, "Show in [View]", filter vocabulary, visual consistency). Expand it to own all shared behaviors that apply identically to both views.

### Content to extract and consolidate

**2a. Live Reload Behavior.**
Currently near-identical sections in both TREE_VIEW.md and GRAPH_VIEW.md. Extract into one shared section covering:
- Identity matching (by name; renames = removal + addition)
- State preservation table (expand/collapse, scroll/viewport, filters, search, selection)
- Additions and removals behavior
- Transition indicator

Each view spec can retain a brief note with view-specific reload details (e.g., graph: "new nodes seeded at parent position, simulation reheats locally"; tree: "new definitions appear collapsed").

**2b. Error Handling.** *(absorbs Round 2 Group 2)*
Tree view has a detailed errors header spec. Graph view has none. Extract the shared error pattern:
- Collapsible error bar between header controls and view content
- Error count, grouped by shown/hidden files
- Each error shows file name and message
- Errors are informational, not blocking — the view still renders valid content

Both view specs cross-reference this section.

**2c. Empty States and Initial Defaults.** *(absorbs Round 2 Group 1)*
New shared section covering:
- **Default view:** Tree View loads first (familiar interaction model, works immediately with any AST)
- **Default tree filters:** Workers + Workflows ON; Namespaces, Nexus Services, Activities OFF (already specced)
- **Default graph zoom:** Levels 1–2 (Namespaces + Workers) on first switch to graph
- **Empty states:** Three cases, same across both views:
  - No AST loaded: "Open a .twf file or connect to the extension to get started."
  - AST loaded, no definitions match filters: "No definitions match the current filters." with hint to adjust
  - AST loaded, only parse errors: Show error bar with no content below

**2d. Accessibility Approach.**
Currently brief ARIA notes in both view specs. Consolidate the shared philosophy:
- ARIA roles follow WAI-ARIA tree (tree view) and interactive graphics (graph view) patterns
- Key requirement: screen readers announce identity, state, and relationships
- Specific ARIA attributes are implementation concerns, not specced
- Focus indicators are visible and distinct from hover/selection

Each view spec retains its view-specific key bindings table.

**2e. Keyboard Modifier Vocabulary.** *(absorbs Round 2 Group 6)*
New shared section defining modifier key semantics product-wide:
- **Shift** — direction reversal (upstream vs downstream dependency highlighting)
- **Shift+Tab** — reverse focus cycling (standard browser/OS convention)
- **Future: Multi-select** — note that Shift+click is reserved for upstream selection semantics, so multi-select should use a different modifier (Ctrl+click or Meta+click)

This prevents each view spec from independently defining modifier meanings that could conflict.

### Existing sections that stay

- View Model, Tab Bar, "Show in [View]" Action, Shared Filter Vocabulary — these stay and get refined as needed.

### Decisions

- Renamed from NAVIGATION.md to VIEW_FRAMEWORK.md — the expanded scope (live reload, errors, empty states, accessibility, modifier vocabulary) warrants a name that reflects "the shared structure both views inherit from."
- Visual Consistency table: move the authoritative color/icon definitions to PRODUCT.md § Visual Identity. VIEW_FRAMEWORK.md can retain a brief cross-reference.

---

## Group 3: Slim TREE_VIEW.md + Accuracy Fixes

**Primary file:** `tools/visualizer/spec/TREE_VIEW.md`

Remove extracted content, add cross-references to shared docs, and apply Round 2 accuracy findings.

### Content to remove (with cross-references)

- **User Goals** — replace with a one-line cross-reference to PRODUCT.md § User Goal Hierarchy, listing which goals this view serves
- **Visual Design § Color system** — replace with cross-reference to PRODUCT.md § Visual Identity. Keep view-specific rendering details (border weights, gradient patterns) if they go beyond the shared palette.
- **Visual Design § Theme support** — move to PRODUCT.md. Keep only a note that the tree view follows the shared theme.
- **Live Reload Behavior** — replace with cross-reference to VIEW_FRAMEWORK.md § Live Reload. Keep any tree-specific reload notes (e.g., "new definitions appear collapsed").
- **Keyboard Navigation § Accessibility** — replace with cross-reference to VIEW_FRAMEWORK.md § Accessibility. Keep the key bindings table (it's view-specific).
- **Cross-View Navigation** — this is already just a cross-reference. No change needed.

### Content to keep (tree-view-specific)

- Existing codebase context + file structure (implementation reference)
- Data flow (DefinitionContext, HandlerContext)
- Header and filtering (tree-specific filter UI — file chips, type toggles, search)
- Block rendering (anatomy, expand/collapse, all block types)
- Definition types (namespace, worker, workflow, activity, nexus service rendering)
- Statement types (call blocks, await blocks, control flow, leaf blocks)
- Cross-reference resolution (inline expansion)
- Contextual navigation buttons

### Round 2 accuracy fixes to apply

**4a. Handler "Show callers" annotation.** *(Round 2 Group 4a)*
In the contextual navigation table, update the handler declaration row:
> Handler declaration (signal/query/update) | Show callers *(future — requires send-side DSL syntax)*

**4b. Comment rendering.** *(Round 2 Group 4b)*
Add a `Comment` row to the leaf blocks table:
| Comment | `comment` | Light grey | No |

Or add a note that comments are routed through `StatementBlock` as a leaf block type.

**5a. Truncation tooltip.** *(Round 2 Group 5a)*
Add to block anatomy section: "Truncated signatures show full text on hover via tooltip."

**5b. Bulk expand/collapse.** *(Round 2 Group 5b)*
Add as a future consideration or spec a keyboard shortcut: Ctrl+Shift+Arrow Right (expand all at level), Ctrl+Shift+Arrow Left (collapse all at level).

---

## Group 4: Slim GRAPH_VIEW.md + Content Additions

**Primary file:** `tools/visualizer/spec/GRAPH_VIEW.md`

Remove extracted content, add cross-references, and apply Round 2 content additions.

### Content to remove (with cross-references)

- **User Goals** — replace with cross-reference to PRODUCT.md, listing which goals this view serves
- **Visual Encoding § Color** — replace with cross-reference to PRODUCT.md § Visual Identity. Keep node shape/size table and edge appearance table (graph-specific).
- **Live Reload Behavior** — replace with cross-reference to VIEW_FRAMEWORK.md § Live Reload. Keep graph-specific reload notes (new nodes seeded at parent position, simulation reheats locally, node fade-out on removal).
- **Summary of Patterns Used** — move to PRODUCT.md § UX Principles (most are product-level patterns, not graph-specific). If any are truly graph-only (e.g., force-directed layout), keep them in the graph spec.
- **Keyboard Navigation § Accessibility** — replace with cross-reference to VIEW_FRAMEWORK.md § Accessibility. Keep the key bindings table.
- **Keyboard Navigation § Modifier Keys with Keyboard** — replace with cross-reference to VIEW_FRAMEWORK.md § Keyboard Modifier Vocabulary.
- **Cross-View Navigation** — already just a cross-reference. No change needed.
- **Search and Filtering § Design Principle: Reactive Composition** — move to PRODUCT.md § UX Principles. This is a product-level principle, not graph-specific.

### Content to keep (graph-view-specific)

- Graph Data Model (node types, edges, coarsening, construction order)
- Layout: Force-Directed Simulation (forces, parameters, lifecycle)
- Semantic Zoom: Level Selection (range selector, visibility rules)
- Level Transitions (animation spec)
- Visual Encoding (node shapes/sizes, edge line styles — minus shared color)
- Viewport Controls
- Control Panel (sliders, simulation controls, presets)
- Search and Filtering (filter controls, hidden match badges, search result selection)
- Interaction States (hover, selection, multi-select future)
- Hotkey Discoverability
- Future: Message Flow Edges
- Performance Considerations

### Round 2 content additions

**Orphan Definitions.** *(Round 2 Group 3)*
Add new section under Graph Data Model:
- Orphan definitions (no worker/namespace assignment) appear as uncontained nodes
- Visually distinct: positioned outside groupings, subtle "unassigned" indicator
- Participate in dependency edges normally
- Semantic zoom behavior: orphan workers visible at Level 2, orphan Level 3 nodes visible at Level 3, hidden when their level is not selected

**Double-click resolution.** *(Round 2 Group 4c)*
In Viewport Controls: double-click remains "center and zoom to fit node and neighbors" (graph-native behavior).
Update VIEW_FRAMEWORK.md § "Show in [View]" trigger points: remove double-click as a cross-view navigation trigger. "Show in Tree" is accessible via right-click context menu and contextual navigation buttons only.

**Error handling cross-reference.** *(Round 2 Group 2)*
Add a brief "Errors Header" section referencing VIEW_FRAMEWORK.md § Error Handling, noting that the error bar appears between the control panel header and the graph canvas.

---

## Execution Order

Groups must be executed in order: Group 1 establishes the product doc that later groups reference. Group 2 establishes shared sections that Groups 3–4 cross-reference. Groups 3 and 4 can be parallelized.

| Group | File | Dependencies |
|-------|------|-------------|
| 1 | PRODUCT.md (create) | None |
| 2 | VIEW_FRAMEWORK.md (expand) | None (can parallel with 1) |
| 3 | TREE_VIEW.md (slim + fix) | Groups 1, 2 (needs to cross-reference both) |
| 4 | GRAPH_VIEW.md (slim + add) | Groups 1, 2 (needs to cross-reference both) |

---

## Round 2 Finding Absorption Map

Shows where each Round 2 finding lands in the reorganized structure:

| Round 2 Group | Finding | Absorbed Into |
|---------------|---------|---------------|
| 1 | Initial experience and empty states | Group 2 (VIEW_FRAMEWORK.md § Empty States and Initial Defaults) |
| 2 | Graph error handling | Group 2 (VIEW_FRAMEWORK.md § Error Handling) + Group 4 (GRAPH_VIEW.md cross-reference) |
| 3 | Orphan definitions in graph | Group 4 (GRAPH_VIEW.md § Orphan Definitions) |
| 4a | Handler "Show callers" annotation | Group 3 (TREE_VIEW.md contextual nav table) |
| 4b | Comment rendering | Group 3 (TREE_VIEW.md leaf blocks table) |
| 4c | Double-click conflict | Group 4 (GRAPH_VIEW.md viewport controls) + Group 2 (VIEW_FRAMEWORK.md trigger points) |
| 5a | Truncation tooltip | Group 3 (TREE_VIEW.md block anatomy) |
| 5b | Bulk expand/collapse | Group 3 (TREE_VIEW.md keyboard nav or future section) |
| 6 | Modifier conflicts | Group 2 (VIEW_FRAMEWORK.md § Keyboard Modifier Vocabulary) |
