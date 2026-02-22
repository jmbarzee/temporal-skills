# Tree View

The tree view is the primary view for the TWF Visualizer. It renders every definition from the AST as a collapsible, color-coded block in a vertical list. Nesting is achieved through progressive disclosure — clicking a block header expands it to reveal its children.


## User Goals

The tree view answers questions about **individual definitions and their contents**:

1. **"What do these workflows do?"** — Read the step-by-step logic of any workflow, activity, or handler through recursive expand/collapse.
2. **"What are the inputs and outputs?"** — See signatures (params → return type) on every definition and call.
3. **"What handlers do these workflows expose?"** — See signal, query, and update declarations grouped at the top of each workflow body.
4. **"What does this call expand to?"** — Inline expansion shows the full body of any referenced workflow, activity, or nexus operation without navigating away.
5. **"What definitions exist in this file or package?"** — Filter and browse definitions by type, source file, and name.


## Existing codebase context

### Architecture
- **React 18** + **TypeScript** + **Vite** (no additional UI libraries)
- Entry points: `src/main.tsx` (standalone dev) and `src/webview.tsx` (VS Code webview)
- Both entry points load AST data and pass it as `TWFFile` to `<WorkflowCanvas>`
- Shared type definitions in `src/types/ast.ts`
- Theme configuration (icons, labels, CSS variable prefixes) centralized in `src/theme/temporal-theme.tsx`
- Styles split across `src/styles/index.css` (global layout, header, theme variables) and `src/components/blocks/blocks.css` (block-level variables and styles)
- CSS variables provide full light/dark theme support; dark theme activates via `.vscode-dark` or `[data-theme="dark"]`

### File structure
```
src/
  App.tsx                          — Standalone app shell (file upload, query param loading)
  main.tsx                         — Standalone entry point
  webview.tsx                      — VS Code webview entry point
  types/
    ast.ts                         — TypeScript types mirroring the Go AST JSON
  theme/
    temporal-theme.tsx             — Central icon/label/CSS-prefix map for all primitives
  styles/
    index.css                      — Global styles, layout, header, theme variables
  components/
    WorkflowCanvas.tsx             — Main tree view component + DefinitionContext + header/filters
    icons/
      GearIcons.tsx                — SVG icons (search, single gear, interlocking gears)
    blocks/
      blocks.css                   — All block-level CSS variables and styles
      useToggle.ts                 — Shared expand/collapse hook
      DefinitionBlock.tsx          — Top-level definition router + namespace/worker/activity/nexus blocks
      StatementBlock.tsx           — Statement router (dispatches to leaf/call/control-flow blocks)
      WorkflowContent.tsx          — Workflow body renderer (state, handlers, body) + inline workflow/sync body blocks
      CallBlocks.tsx               — Activity call, workflow call, nexus call blocks
      AwaitBlocks.tsx              — Await statement, await all, await one blocks
      ControlFlowBlocks.tsx        — Switch, if, for blocks
      LeafBlocks.tsx               — Return, close, raw, promise, set, unset, break, continue blocks
```


## Data flow

1. AST JSON (`TWFFile`) arrives via file upload, URL query param, or VS Code `postMessage`
2. `WorkflowCanvas` receives the AST as a prop
3. `WorkflowCanvas` builds a `DefinitionContext` — lookup maps keyed by name for workflows, activities, workers, nexus services, and namespaces
4. `DefinitionContext` is provided via React context so any nested block can resolve references (e.g., a workflow call block can look up the target workflow's definition to render its body inline)
5. Each `WorkflowDef` additionally builds a `HandlerContext` with maps for its signals, queries, and updates, so await blocks can resolve handler references


## Header and filtering

The canvas header is a card at the top of the view containing three sections separated by dividers:

### File filter
- Shown only when definitions have `sourceFile` metadata (multi-file ASTs)
- Renders a horizontal row of **filter chips**, one per source file
- Three states: **selected** (active filter), **unselected** (excluded), **all-included** (no files explicitly selected, so everything is shown)
- Clicking a chip toggles it; multiple files can be selected simultaneously
- When exactly one file is selected, the VS Code webview sends an `openFile` message to focus that file in the editor

### Definition type toggles
- A row of toggle buttons, one per definition type: Namespaces, Workers, Nexus Services, Workflows, Activities
- Each button shows the type's icon and label
- Active toggles use the type's accent color; inactive toggles are muted
- Default state: Workers and Workflows are on; Namespaces, Nexus Services, and Activities are off

### Search
- A collapsible search bar toggled by a magnifying glass button
- Filters the visible definitions by name (case-insensitive substring match)
- Escape key closes the search bar and clears the query

### Errors header
- Shown only when the AST contains parse errors
- A collapsible bar below the main header with error count
- Errors are grouped by "shown files" (matching the file filter) and "hidden files"
- Each error displays the file name and the error/stderr message


## Block rendering

### Block anatomy
Every block follows the same layout pattern:

```
┌──────────────────────────────────────────────┐
│ ▶  ⚙⚙  workflow  MyWorkflow(params) → Type  │  ← header (toggle, icon, keyword, signature)
└──────────────────────────────────────────────┘
```

When expanded:

```
┌──────────────────────────────────────────────┐
│ ▼  ⚙⚙  workflow  MyWorkflow(params) → Type  │  ← header
│    ┌─────────────────────────────────────┐   │
│    │  (nested child blocks)              │   │  ← body (indented left margin)
│    └─────────────────────────────────────┘   │
└──────────────────────────────────────────────┘
```

Header elements (left to right):
1. **Toggle indicator** — `▶` (collapsed) or `▼` (expanded). Placeholder space if not expandable.
2. **Icon** — type-specific icon (text emoji or SVG). Workflows use interlocking gears SVG, activities use single gear SVG.
3. **Keyword** — bold text identifying the block type (e.g., `workflow`, `activity`, `await`, `if`)
4. **Signature** — the name, params, and return type. Truncated with ellipsis if too long.

Body:
- Indented via left margin (20px) + left padding (12px)
- Contains child blocks rendered recursively

### Expand/collapse behavior
- `useToggle` hook manages open/closed state per block
- Blocks that reference an unresolvable definition are not expandable (`canToggle: false`)
- Expanded blocks gain a subtle box shadow
- Top-level definition blocks (workflow, worker, namespace) start collapsed
- Control flow blocks (if, for, switch, await all, await one) start expanded
- Default initial state varies by block type


## Definition types

### Namespace (`namespaceDef`)
- Dark grey color palette
- Header shows name and entry count
- Body contains two sections:
  - **Workers** — each renders as a worker-colored sub-entry that, when expanded, shows the full worker body (workflow/activity/service ref lists)
  - **Nexus endpoints** — each renders as a nexus-colored sub-entry (not expandable)

### Worker (`workerDef`)
- Medium grey color palette
- Header shows name and total reference count across all categories
- Body contains up to three labeled sections:
  - **Workflows** — each ref is a purple sub-entry that expands to show the workflow's full content
  - **Activities** — each ref is a blue sub-entry that expands to show the activity's body
  - **Nexus services** — each ref is a pink sub-entry that expands to show the service's operations

### Workflow (`workflowDef`)
- Purple color palette, 2px solid border (heavier than call-level blocks)
- Body uses `WorkflowContent` which renders (in order):
  1. **State** — collapsible group showing conditions and raw state declarations
  2. **Signals** — collapsible group of signal handler declarations (dashed border)
  3. **Queries** — collapsible group of query handler declarations (dashed border)
  4. **Updates** — collapsible group of update handler declarations (dashed border)
  5. **Body statements** — the workflow's statement list rendered recursively

### Activity (`activityDef`)
- Blue color palette, 2px solid border
- Body contains the activity's statement list

### Nexus Service (`nexusServiceDef`)
- Deep pink color palette, 2px solid border
- Body lists nexus operations, each rendered as a sub-block:
  - **Async operations** — expandable if a backing workflow is found; expands to show the workflow inline
  - **Sync operations** — expandable if a body is present; expands to show the handler body


## Statement types

Statements are rendered by `StatementBlock` which routes to specialized block components:

### Call blocks
| Statement | Keyword | Color | Expandable |
|-----------|---------|-------|------------|
| Activity call | `activity` | Blue | Yes — shows activity definition body |
| Workflow call | `workflow` or `detach workflow` | Light purple | Yes — shows workflow content |
| Nexus call | `nexus` or `detach nexus` | Pink | Yes — shows backing workflow or sync handler body |

- Detached calls (fire-and-forget) use a **dashed border** to distinguish from synchronous calls
- Unresolved references (definition not in scope) use a dashed border + reduced opacity + a circled `?` badge

### Await blocks
| Statement | Keyword | Color |
|-----------|---------|-------|
| Await timer | `await timer` | Yellow |
| Await signal | `await signal` | Red/pink |
| Await update | `await update` | Orange |
| Await activity | `await activity` | Blue |
| Await workflow | `await workflow` | Light purple |
| Await nexus | `await nexus` | Pink |
| Await ident | `await` | Teal |
| Await all | `await all` | Grey |
| Await one | `await one` | Grey |

- **Await all** — expands to show its concurrent branch statements
- **Await one** — expands to show its cases, each rendered as a **tagged composite** (a grey "option" tag on the left, the case content block on the right)
- Tagged composites are expandable when the case has a body

### Control flow blocks
| Statement | Keyword | Color |
|-----------|---------|-------|
| If | `if` | Grey/slate |
| For | `for` | Grey/slate |
| Switch | `switch` | Grey/slate |

- `if` bodies show a `then` branch and optionally an `else:` branch
- `for` shows the loop variant in its signature: iteration (`x in items`), conditional (`condition`), or infinite (`∞`)
- `switch` expands to show case blocks and an optional default

### Leaf blocks (non-expandable)
| Statement | Keyword | Color |
|-----------|---------|-------|
| Return | `return` | Green |
| Close (complete) | `close complete` | Green |
| Close (fail) | `close fail` | Red/pink |
| Close (continue as new) | `close continue_as_new` | Orange |
| Promise | `promise` | Cyan/teal |
| Set | `set` | Subtle grey |
| Unset | `unset` | Subtle grey |
| Raw | (code text) | Light grey |
| Break | `break` | Subtle grey |
| Continue | `continue` | Subtle grey |


## Cross-reference resolution

The tree view supports **inline expansion** of referenced definitions. When a call block (activity, workflow, nexus) is expanded, it doesn't navigate elsewhere — it renders the target definition's content directly as nested children. This works through the `DefinitionContext`:

- A workflow call block looks up the target `WorkflowDef` by name and renders `<WorkflowContent>` inline
- A nexus call block resolves the service → operation chain, then either renders the backing workflow inline (async) or the handler body (sync)
- Worker ref items within a worker definition expand to show the referenced workflow, activity, or nexus service content
- Namespace worker entries expand to show the full worker body

Unresolved references (name not found in context) are marked visually but do not prevent rendering.


## Contextual navigation buttons

Every block in the tree view supports **contextual navigation** — small action buttons that appear on hover, positioned at the top-right of the block header, half-overlapping the upper border. These provide focus-shifting actions: navigating to callers, parent containers, or the graph view.

### Available actions by block type

| Block type | Available buttons |
|------------|-------------------|
| Workflow definition | Show callers, Show worker, Show in Graph |
| Activity definition | Show callers, Show worker, Show in Graph |
| NexusService definition | Show callers, Show worker, Show in Graph |
| Worker definition | Show namespace, Show in Graph |
| Namespace definition | Show in Graph |
| Call block (activity/workflow/nexus call) | Show definition, Show in Graph |
| Handler declaration (signal/query/update) | Show callers (workflows that send to this handler) |

Buttons only appear when the action has at least one valid target. If a definition has no callers, "Show callers" does not appear.

### Behavior

- **Single target:** Clicking the button scrolls the tree view to the target, expanding its ancestry if needed, and flashes the target. Same animation sequence as "Show in [View]" (see [NAVIGATION.md](./NAVIGATION.md)).
- **Multiple targets:** Clicking the button opens a small popover listing the targets. The user selects one, then the view navigates to it.
- **Show in Graph:** Follows the cross-view "Show in [View]" sequence from NAVIGATION.md.

### Data requirements

The visualizer builds a **reverse reference index** client-side from the AST's forward references. For each definition, the index maps its name to the set of call sites (workflow + statement location) that reference it. This is computed from the same data already used by `DefinitionContext` — no parser changes needed.


## Visual design

### Color system
Each definition and statement type has a dedicated color palette defined as CSS variables with three values:
- `--block-{type}-bg` — gradient background
- `--block-{type}-border` — border color
- `--block-{type}-text` — text/icon color

Top-level definitions use a 2px border; call-level and statement-level blocks use 1px.

Handler declarations (signals, queries, updates) use a **dashed** border style to visually distinguish them from executable statements.

### Icon system
Icons are defined in the central theme map (`temporal-theme.tsx`). Most are Unicode text characters. Workflows and activities use custom SVG icons (interlocking gears and single gear respectively) for clearer rendering at small sizes.

### Theme support
- Light theme is the default
- Dark theme activates via `.vscode-dark` class (VS Code webview) or `[data-theme="dark"]` attribute
- Every color palette has a matching dark variant defined in the CSS
- Hover brightness shifts direction between themes (`0.95` in light, `1.1` in dark)


## Live Reload Behavior

When the AST updates (file save → parser re-run → new `TWFFile` delivered to the visualizer), the tree view preserves user state where possible.

### Identity Matching

Definitions are matched across AST versions **by name**. A definition with the same name in the new AST is considered the same definition. Renames are treated as a removal of the old name plus an addition of the new name.

### State Preserved Across Reloads

| State | Behavior |
|-------|----------|
| Expand/collapse | Preserved for definitions that still exist (matched by name). New definitions appear collapsed. |
| Scroll position | Preserved. If the scrolled-to definition was removed, scroll to the nearest surviving sibling. |
| Filter selections | Preserved (file filter, type toggles, search query). If a filtered file no longer exists, remove it from the selection. |
| Contextual nav buttons | Recomputed from new AST (reverse index rebuilt). |

### Additions and Removals

- **New definitions** appear in their natural position (sorted by type and order in AST), collapsed, with no special animation.
- **Removed definitions** disappear immediately. If the removed definition was expanded, its children simply vanish with it.

### Transition Indicator

A brief, non-blocking indicator (e.g., a subtle flash on the header bar, or a small "updated" badge that fades) signals that the AST has been refreshed. This should not interrupt the user's current interaction.


## Keyboard Navigation

The tree view supports keyboard navigation following the same model as VS Code's tree widget.

### Key Bindings

| Key | Action |
|-----|--------|
| **↑ / ↓** | Move focus to previous / next visible block (siblings and across nesting levels) |
| **→** | Expand focused block (if collapsed). If already expanded, move focus to first child. |
| **←** | Collapse focused block (if expanded). If already collapsed, move focus to parent. |
| **Enter** | Toggle expand/collapse on focused block |
| **Home / End** | Move focus to first / last visible block |
| **/** or **Ctrl+F** | Open search bar and focus the search input |
| **Escape** | Close search bar (if open), clear selection, or close any open popover |
| **Tab** | Move focus between header controls (file filter, type toggles, search) and the block list |

### Focus Indicator

The currently focused block has a visible focus ring (distinct from hover and selection styles). Focus follows keyboard navigation and is independent of mouse hover.

### Accessibility

ARIA roles should follow the WAI-ARIA Treeview pattern (`role="tree"`, `role="treeitem"`, `aria-expanded`, `aria-level`). Specific ARIA attributes are an implementation concern — the key requirement is that screen readers can announce block type, name, expanded/collapsed state, and nesting depth.


## Cross-View Navigation

The tree view participates in the visualizer's cross-view navigation system. See [NAVIGATION.md](./NAVIGATION.md) for the full spec covering view switching, "Show in Graph" actions, and shared filter vocabulary.
