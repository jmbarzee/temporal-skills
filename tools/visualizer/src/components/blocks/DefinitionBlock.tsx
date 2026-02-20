import React from 'react'
import type { Definition, WorkflowDef, ActivityDef, SignalDecl, QueryDecl, UpdateDecl } from '../../types/ast'
import { StatementBlock } from './StatementBlock'
import { WorkflowContent } from './WorkflowContent'
import { SingleGearIcon, InterlockingGearsIcon } from '../icons/GearIcons'
import { useToggle } from './useToggle'
import { HandlerContext } from '../WorkflowCanvas'
import './blocks.css'

interface DefinitionBlockProps {
  definition: Definition
}

export function DefinitionBlock({ definition }: DefinitionBlockProps) {
  if (definition.type === 'workflowDef') {
    return <WorkflowDefBlock def={definition} />
  } else {
    return <ActivityDefBlock def={definition} />
  }
}

function WorkflowDefBlock({ def }: { def: WorkflowDef }) {
  const signature = formatWorkflowSignature(def)
  const [expanded, toggle] = useToggle()

  // Build handler context for this workflow
  const handlerContext = React.useMemo<HandlerContext>(() => {
    const signals = new Map<string, SignalDecl>()
    const queries = new Map<string, QueryDecl>()
    const updates = new Map<string, UpdateDecl>()

    for (const s of def.signals || []) signals.set(s.name, s)
    for (const q of def.queries || []) queries.set(q.name, q)
    for (const u of def.updates || []) updates.set(u.name, u)

    return { signals, queries, updates }
  }, [def])

  return (
    <HandlerContext.Provider value={handlerContext}>
      <div className={`block block-workflow ${expanded ? 'expanded' : 'collapsed'}`}>
        <div className="block-header" onClick={toggle}>
          <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
          <span className="block-icon"><InterlockingGearsIcon /></span>
          <span className="block-keyword">workflow</span>
          <span className="block-signature">{signature}</span>
        </div>

        {expanded && (
          <div className="block-body">
            <WorkflowContent def={def} />
          </div>
        )}
      </div>
    </HandlerContext.Provider>
  )
}

function ActivityDefBlock({ def }: { def: ActivityDef }) {
  const [expanded, toggle] = useToggle()
  const signature = formatActivitySignature(def)

  return (
    <div className={`block block-activity-def ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={toggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon"><SingleGearIcon /></span>
        <span className="block-keyword">activity</span>
        <span className="block-signature">{signature}</span>
      </div>

      {expanded && (
        <div className="block-body">
          {(def.body || []).map((stmt) => (
            <StatementBlock key={`${stmt.line}:${stmt.column}`} statement={stmt} />
          ))}
        </div>
      )}
    </div>
  )
}

function formatWorkflowSignature(def: WorkflowDef): string {
  let sig = `${def.name}(${def.params})`
  if (def.returnType) {
    sig += ` → ${def.returnType}`
  }
  return sig
}

function formatActivitySignature(def: ActivityDef): string {
  let sig = `${def.name}(${def.params})`
  if (def.returnType) {
    sig += ` → ${def.returnType}`
  }
  return sig
}
