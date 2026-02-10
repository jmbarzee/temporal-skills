import React from 'react'
import type { TWFFile, WorkflowDef, ActivityDef, SignalDecl, QueryDecl, UpdateDecl } from '../types/ast'
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

  return (
    <DefinitionContextProvider.Provider value={context}>
      <div className="workflow-canvas">
        {ast.focusedFile && (
          <div className="focused-file-indicator">
            <span className="file-icon">📄</span>
            <span className="file-name">{focusedFileName}</span>
          </div>
        )}
        {workflows.length === 0 ? (
          <div className="no-workflows">
            <p>No workflows defined in this file</p>
          </div>
        ) : (
          workflows.map((workflow) => (
            <DefinitionBlock key={workflow.name} definition={workflow} />
          ))
        )}
      </div>
    </DefinitionContextProvider.Provider>
  )
}
