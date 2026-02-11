import React from 'react'
import type { TWFFile, WorkflowDef, ActivityDef, SignalDecl, QueryDecl, UpdateDecl, FileError } from '../types/ast'
import { DefinitionBlock } from './blocks/DefinitionBlock'

interface WorkflowCanvasProps {
  ast: TWFFile
}

export interface DefinitionContext {
  workflows: Map<string, WorkflowDef>
  activities: Map<string, ActivityDef>
}

// Context for looking up signal/query/update handlers in the current workflow
export interface HandlerContext {
  signals: Map<string, SignalDecl>
  queries: Map<string, QueryDecl>
  updates: Map<string, UpdateDecl>
}

export const DefinitionContextProvider = React.createContext<DefinitionContext>({
  workflows: new Map(),
  activities: new Map(),
})

export const HandlerContextProvider = React.createContext<HandlerContext>({
  signals: new Map(),
  queries: new Map(),
  updates: new Map(),
})

export function WorkflowCanvas({ ast }: WorkflowCanvasProps) {
  // Build lookup maps for definitions (all files for expansion support)
  const context = React.useMemo<DefinitionContext>(() => {
    const workflows = new Map<string, WorkflowDef>()
    const activities = new Map<string, ActivityDef>()

    for (const def of ast.definitions) {
      if (def.type === 'workflowDef') {
        workflows.set(def.name, def)
      } else if (def.type === 'activityDef') {
        activities.set(def.name, def)
      }
    }

    return { workflows, activities }
  }, [ast])

  // Filter to show only workflows from the focused file at top level
  // If no focused file is set, show all workflows (backward compatible)
  const workflows = ast.definitions.filter(
    (def): def is WorkflowDef => {
      if (def.type !== 'workflowDef') return false
      // If no focused file specified, show all
      if (!ast.focusedFile) return true
      // Otherwise only show workflows from the focused file
      return def.sourceFile === ast.focusedFile
    }
  )

  // Extract just the filename for display
  const focusedFileName = ast.focusedFile?.split('/').pop() || 'All Workflows'

  // Filter errors relevant to the focused file (or show all if no focused file)
  const errors = ast.errors || []
  const relevantErrors = ast.focusedFile
    ? errors.filter(e => e.file === ast.focusedFile)
    : errors

  return (
    <DefinitionContextProvider.Provider value={context}>
      <div className="workflow-canvas">
        {ast.focusedFile && (
          <div className="focused-file-indicator">
            <span className="file-icon">📄</span>
            <span className="file-name">{focusedFileName}</span>
          </div>
        )}
        {relevantErrors.length > 0 && (
          <ParseErrors errors={relevantErrors} />
        )}
        {workflows.length === 0 && relevantErrors.length === 0 ? (
          <div className="no-workflows">
            <p>No workflows defined in this file</p>
          </div>
        ) : (
          workflows.map((workflow) => (
            <DefinitionBlock key={workflow.name} definition={workflow} />
          ))
        )}
        {/* Show other file errors collapsed if viewing a specific file */}
        {ast.focusedFile && errors.length > relevantErrors.length && (
          <OtherFileErrors errors={errors.filter(e => e.file !== ast.focusedFile)} />
        )}
      </div>
    </DefinitionContextProvider.Provider>
  )
}

/** Display parse errors prominently */
function ParseErrors({ errors }: { errors: FileError[] }) {
  return (
    <div className="parse-errors">
      <div className="parse-errors-header">
        <span className="parse-errors-icon">⚠</span>
        <span className="parse-errors-title">
          {errors.length === 1 ? 'Parse error' : `${errors.length} parse errors`}
        </span>
      </div>
      {errors.map((err, i) => (
        <div key={i} className="parse-error-item">
          <div className="parse-error-file">{err.file.split('/').pop()}</div>
          <pre className="parse-error-message">{err.stderr || err.error}</pre>
        </div>
      ))}
    </div>
  )
}

/** Collapsed section for errors in other files */
function OtherFileErrors({ errors }: { errors: FileError[] }) {
  const [expanded, setExpanded] = React.useState(false)

  return (
    <div className="other-file-errors">
      <div className="other-file-errors-header" onClick={() => setExpanded(!expanded)}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="other-file-errors-icon">⚠</span>
        <span className="other-file-errors-title">
          {errors.length} error{errors.length !== 1 ? 's' : ''} in other files
        </span>
      </div>
      {expanded && (
        <div className="other-file-errors-body">
          {errors.map((err, i) => (
            <div key={i} className="parse-error-item">
              <div className="parse-error-file">{err.file.split('/').pop()}</div>
              <pre className="parse-error-message">{err.stderr || err.error}</pre>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
