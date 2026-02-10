import React from 'react'
import type { TWFFile, WorkflowDef, ActivityDef } from '../types/ast'
import { DefinitionBlock } from './blocks/DefinitionBlock'

interface WorkflowCanvasProps {
  ast: TWFFile
}

export interface DefinitionContext {
  workflows: Map<string, WorkflowDef>
  activities: Map<string, ActivityDef>
}

export const DefinitionContextProvider = React.createContext<DefinitionContext>({
  workflows: new Map(),
  activities: new Map(),
})

export function WorkflowCanvas({ ast }: WorkflowCanvasProps) {
  // Build lookup maps for definitions
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

  // Filter to show only workflows at top level (activities are shown when expanded from calls)
  const workflows = ast.definitions.filter(
    (def): def is WorkflowDef => def.type === 'workflowDef'
  )

  return (
    <DefinitionContextProvider.Provider value={context}>
      <div className="workflow-canvas">
        {workflows.map((workflow) => (
          <DefinitionBlock key={workflow.name} definition={workflow} />
        ))}
      </div>
    </DefinitionContextProvider.Provider>
  )
}
