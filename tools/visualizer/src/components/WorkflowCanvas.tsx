import React from 'react'
import type { TWFFile, Definition, WorkflowDef, ActivityDef, WorkerDef, NamespaceDef, NexusServiceDef, SignalDecl, QueryDecl, UpdateDecl, FileError } from '../types/ast'
import { DefinitionBlock } from './blocks/DefinitionBlock'
import { SearchIcon } from './icons/GearIcons'

interface WorkflowCanvasProps {
  ast: TWFFile
  onOpenFile?: (file: string) => void
}

export interface DefinitionContext {
  workflows: Map<string, WorkflowDef>
  activities: Map<string, ActivityDef>
  workers: Map<string, WorkerDef>
  nexusServices: Map<string, NexusServiceDef>
  namespaces: Map<string, NamespaceDef>
}

// Context for looking up signal/query/update handlers in the current workflow
export interface HandlerContext {
  signals: Map<string, SignalDecl>
  queries: Map<string, QueryDecl>
  updates: Map<string, UpdateDecl>
}

export const DefinitionContext = React.createContext<DefinitionContext>({
  workflows: new Map(),
  activities: new Map(),
  workers: new Map(),
  nexusServices: new Map(),
  namespaces: new Map(),
})

export const HandlerContext = React.createContext<HandlerContext>({
  signals: new Map(),
  queries: new Map(),
  updates: new Map(),
})

// Definition type toggle configuration
interface DefTypeConfig {
  type: string
  label: string
  icon: string
  defaultOn: boolean
}

const DEF_TYPE_CONFIGS: DefTypeConfig[] = [
  { type: 'workflowDef', label: 'Workflows', icon: '⚙⚙', defaultOn: true },
  { type: 'activityDef', label: 'Activities', icon: '⚙', defaultOn: false },
  { type: 'nexusServiceDef', label: 'Nexus Services', icon: '★', defaultOn: false },
  { type: 'workerDef', label: 'Workers', icon: '□', defaultOn: false },
  { type: 'namespaceDef', label: 'Namespaces', icon: '⧉', defaultOn: false },
]

const DEFAULT_VISIBLE_TYPES = new Set(
  DEF_TYPE_CONFIGS.filter(c => c.defaultOn).map(c => c.type)
)

export function WorkflowCanvas({ ast, onOpenFile }: WorkflowCanvasProps) {
  // --- Header state ---
  const [selectedFiles, setSelectedFiles] = React.useState<Set<string>>(new Set())
  const [searchActive, setSearchActive] = React.useState(false)
  const [searchQuery, setSearchQuery] = React.useState('')
  const [visibleTypes, setVisibleTypes] = React.useState<Set<string>>(DEFAULT_VISIBLE_TYPES)
  const searchInputRef = React.useRef<HTMLInputElement>(null)

  // Build lookup maps for definitions (all files for expansion support)
  const context = React.useMemo<DefinitionContext>(() => {
    const workflows = new Map<string, WorkflowDef>()
    const activities = new Map<string, ActivityDef>()
    const workers = new Map<string, WorkerDef>()
    const nexusServices = new Map<string, NexusServiceDef>()
    const namespaces = new Map<string, NamespaceDef>()

    for (const def of ast.definitions) {
      if (def.type === 'workflowDef') {
        workflows.set(def.name, def)
      } else if (def.type === 'activityDef') {
        activities.set(def.name, def)
      } else if (def.type === 'workerDef') {
        workers.set(def.name, def)
      } else if (def.type === 'nexusServiceDef') {
        nexusServices.set(def.name, def)
      } else if (def.type === 'namespaceDef') {
        namespaces.set(def.name, def)
      }
    }

    return { workflows, activities, workers, nexusServices, namespaces }
  }, [ast])

  // Extract all unique source files from definitions
  const allFiles = React.useMemo(() => {
    const files = new Set<string>()
    for (const def of ast.definitions) {
      if (def.sourceFile) {
        files.add(def.sourceFile)
      }
    }
    return Array.from(files).sort()
  }, [ast])

  // Initialize selected files: focused file selected by default
  React.useEffect(() => {
    if (ast.focusedFile) {
      setSelectedFiles(new Set([ast.focusedFile]))
    } else {
      setSelectedFiles(new Set())
    }
  }, [ast.focusedFile])

  // Toggle a file in the selection
  const toggleFile = (file: string) => {
    setSelectedFiles(prev => {
      const next = new Set(prev)
      if (next.has(file)) {
        next.delete(file)
      } else {
        next.add(file)
      }
      return next
    })
  }

  // When file filter narrows to exactly one file, open it in the editor
  React.useEffect(() => {
    if (selectedFiles.size === 1 && onOpenFile) {
      onOpenFile(selectedFiles.values().next().value!)
    }
  }, [selectedFiles, onOpenFile])

  // Toggle search bar
  const toggleSearch = () => {
    if (searchActive) {
      setSearchActive(false)
      setSearchQuery('')
    } else {
      setSearchActive(true)
      // Focus the input after it renders
      setTimeout(() => searchInputRef.current?.focus(), 50)
    }
  }

  // Toggle a definition type in visibility
  const toggleType = (type: string) => {
    setVisibleTypes(prev => {
      const next = new Set(prev)
      if (next.has(type)) {
        next.delete(type)
      } else {
        next.add(type)
      }
      return next
    })
  }

  // Filter definitions for display
  const visibleDefinitions = React.useMemo(() => {
    const lowerQuery = searchQuery.toLowerCase()

    return ast.definitions.filter((def): def is Definition => {
      // Type filter
      if (!visibleTypes.has(def.type)) return false

      // File filter: if any files are selected, only show from those files
      // If none selected (all toggled off), show all
      if (selectedFiles.size > 0 && def.sourceFile) {
        if (!selectedFiles.has(def.sourceFile)) return false
      }

      // Search filter
      if (lowerQuery) {
        if (!def.name.toLowerCase().includes(lowerQuery)) return false
      }

      return true
    })
  }, [ast.definitions, selectedFiles, visibleTypes, searchQuery])

  // Partition errors into "shown files" vs "hidden files" based on file filter
  const errors = ast.errors || []
  const { shownFileErrors, hiddenFileErrors } = React.useMemo(() => {
    if (selectedFiles.size === 0) {
      // No file filter active — all errors are "shown"
      return { shownFileErrors: errors, hiddenFileErrors: [] as FileError[] }
    }
    const shown: FileError[] = []
    const hidden: FileError[] = []
    for (const e of errors) {
      if (selectedFiles.has(e.file)) {
        shown.push(e)
      } else {
        hidden.push(e)
      }
    }
    return { shownFileErrors: shown, hiddenFileErrors: hidden }
  }, [errors, selectedFiles])

  const hasFiles = allFiles.length > 0
  const hasErrors = errors.length > 0
  const noFilesSelected = selectedFiles.size === 0

  return (
    <DefinitionContext.Provider value={context}>
      <div className="workflow-canvas">
        {/* === Filter Header === */}
        <div className="canvas-header">
          {/* Files Section — only show if there are files */}
          {hasFiles && (
            <>
              <div className="header-files-section">
                <div className="header-files-row">
                  {allFiles.map(file => {
                    const fileName = file.split('/').pop() || file
                    const isSelected = selectedFiles.has(file)
                    const chipClass = noFilesSelected
                      ? 'header-file-tag all-included'
                      : `header-file-tag ${isSelected ? 'selected' : ''}`
                    return (
                      <button
                        key={file}
                        className={chipClass}
                        onClick={() => toggleFile(file)}
                        title={file}
                      >
                        <span className="header-file-icon">📄</span>
                        <span className="header-file-name">{fileName}</span>
                      </button>
                    )
                  })}
                </div>
              </div>
              <div className="header-divider" />
            </>
          )}

          {/* Definition Type Toggles */}
          <div className="header-types-section">
            <div className="header-types-row">
              {DEF_TYPE_CONFIGS.map(cfg => {
                const isActive = visibleTypes.has(cfg.type)
                return (
                  <button
                    key={cfg.type}
                    className={`header-type-tag ${isActive ? 'active' : ''} header-type-${cfg.type}`}
                    onClick={() => toggleType(cfg.type)}
                    title={isActive ? `Hide ${cfg.label.toLowerCase()}` : `Show ${cfg.label.toLowerCase()}`}
                  >
                    <span className="header-type-icon">{cfg.icon}</span>
                    <span className="header-type-label">{cfg.label}</span>
                  </button>
                )
              })}
            </div>
          </div>
          <div className="header-divider" />

          {/* Controls Section — search */}
          <div className="header-controls-section">
            <div className={`header-search ${searchActive ? 'active' : ''}`}>
              <button
                className="header-search-toggle"
                onClick={toggleSearch}
                title="Search definitions"
              >
                <SearchIcon size={14} />
              </button>
              {searchActive && (
                <input
                  ref={searchInputRef}
                  className="header-search-input"
                  type="text"
                  placeholder="Filter by name..."
                  value={searchQuery}
                  onChange={e => setSearchQuery(e.target.value)}
                  onKeyDown={e => {
                    if (e.key === 'Escape') toggleSearch()
                  }}
                />
              )}
            </div>
          </div>
        </div>

        {/* === Errors Header === */}
        {hasErrors && (
          <ErrorsHeader
            shownFileErrors={shownFileErrors}
            hiddenFileErrors={hiddenFileErrors}
          />
        )}

        {/* Definitions */}
        {visibleDefinitions.length === 0 ? (
          <div className="no-workflows">
            <p>
              {searchQuery
                ? 'No matching definitions found'
                : 'No definitions found for the selected types and files'}
            </p>
          </div>
        ) : (
          visibleDefinitions.map((def) => (
            <DefinitionBlock key={`${def.sourceFile || ''}-${def.type}-${def.name}`} definition={def} />
          ))
        )}
      </div>
    </DefinitionContext.Provider>
  )
}

/** Collapsible errors header — shows compilation errors grouped by shown/hidden files */
function ErrorsHeader({ shownFileErrors, hiddenFileErrors }: {
  shownFileErrors: FileError[]
  hiddenFileErrors: FileError[]
}) {
  const [expanded, setExpanded] = React.useState(false)
  const totalErrors = shownFileErrors.length + hiddenFileErrors.length

  // Build summary text
  const summaryParts: string[] = []
  if (shownFileErrors.length > 0) {
    summaryParts.push(`${shownFileErrors.length} in shown files`)
  }
  if (hiddenFileErrors.length > 0) {
    summaryParts.push(`${hiddenFileErrors.length} in hidden files`)
  }
  const summary = summaryParts.length > 1
    ? ` (${summaryParts.join(', ')})`
    : ''

  return (
    <div className="errors-header">
      <div className="errors-header-bar" onClick={() => setExpanded(!expanded)}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="errors-header-icon">⚠</span>
        <span className="errors-header-title">
          {totalErrors} {totalErrors === 1 ? 'error' : 'errors'}{summary}
        </span>
      </div>

      {expanded && (
        <div className="errors-header-body">
          {shownFileErrors.length > 0 && (
            <ErrorGroup
              label="Shown files"
              errors={shownFileErrors}
              variant="shown"
            />
          )}
          {hiddenFileErrors.length > 0 && (
            <ErrorGroup
              label="Hidden files"
              errors={hiddenFileErrors}
              variant="hidden"
            />
          )}
        </div>
      )}
    </div>
  )
}

/** A group of errors under a sub-label */
function ErrorGroup({ label, errors, variant }: {
  label: string
  errors: FileError[]
  variant: 'shown' | 'hidden'
}) {
  return (
    <div className={`error-group error-group-${variant}`}>
      <div className="error-group-label">
        {label} ({errors.length})
      </div>
      {errors.map((err, i) => (
        <div key={i} className="error-group-item">
          <div className="error-group-file">{err.file.split('/').pop()}</div>
          <pre className="error-group-message">{err.stderr || err.error}</pre>
        </div>
      ))}
    </div>
  )
}
