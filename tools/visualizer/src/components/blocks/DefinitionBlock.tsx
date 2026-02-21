import React from 'react'
import type { Definition, WorkflowDef, ActivityDef, WorkerDef, WorkerRef, SignalDecl, QueryDecl, UpdateDecl } from '../../types/ast'
import { StatementBlock } from './StatementBlock'
import { WorkflowContent } from './WorkflowContent'
import { SingleGearIcon, InterlockingGearsIcon } from '../icons/GearIcons'
import { useToggle } from './useToggle'
import { DefinitionContext, HandlerContext } from '../WorkflowCanvas'
import './blocks.css'

interface DefinitionBlockProps {
  definition: Definition
}

export function DefinitionBlock({ definition }: DefinitionBlockProps) {
  switch (definition.type) {
    case 'workflowDef':
      return <WorkflowDefBlock def={definition} />
    case 'activityDef':
      return <ActivityDefBlock def={definition} />
    case 'workerDef':
      return <WorkerDefBlock def={definition} />
    default:
      return null
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

function WorkerDefBlock({ def }: { def: WorkerDef }) {
  const [expanded, toggle] = useToggle()

  const totalRefs = (def.workflows?.length || 0) + (def.activities?.length || 0) + (def.services?.length || 0)

  return (
    <div className={`block block-worker-def ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={toggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon">⧉</span>
        <span className="block-keyword">worker</span>
        <span className="block-signature">{def.name} ({totalRefs} types)</span>
      </div>

      {expanded && (
        <div className="block-body">
          {def.workflows?.length > 0 && (
            <WorkerRefSection label="workflows" refs={def.workflows} refType="workflow" />
          )}
          {def.activities?.length > 0 && (
            <WorkerRefSection label="activities" refs={def.activities} refType="activity" />
          )}
          {def.services?.length > 0 && (
            <WorkerRefSection label="nexus services" refs={def.services} refType="service" />
          )}
        </div>
      )}
    </div>
  )
}

function WorkerRefSection({ label, refs, refType }: { label: string; refs: WorkerRef[]; refType: 'workflow' | 'activity' | 'service' }) {
  return (
    <div className="worker-ref-section">
      <div className="worker-ref-label">{label}</div>
      {refs.map((ref) => (
        <WorkerRefItem key={`${ref.line}:${ref.column}`} ref_={ref} refType={refType} />
      ))}
    </div>
  )
}

function WorkerRefItem({ ref_, refType }: { ref_: WorkerRef; refType: 'workflow' | 'activity' | 'service' }) {
  const context = React.useContext(DefinitionContext)

  // Look up linked definition for expansion
  const linkedDef = refType === 'workflow'
    ? context.workflows.get(ref_.name)
    : refType === 'activity'
      ? context.activities.get(ref_.name)
      : undefined // nexus services not yet in context
  const isDefined = !!linkedDef
  const [expanded, toggle] = useToggle(false, isDefined)

  const icon = refType === 'workflow' ? '⚙⚙' : refType === 'activity' ? '⚙' : '⬡'

  return (
    <div className={`worker-ref worker-ref-${refType} ${expanded ? 'expanded' : 'collapsed'} ${!isDefined ? 'worker-ref-unresolved' : ''}`}>
      <div className="worker-ref-header" onClick={toggle}>
        {isDefined ? (
          <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        ) : (
          <span className="block-toggle-placeholder" />
        )}
        <span className="block-icon">{icon}</span>
        <span className="worker-ref-name">{ref_.name}</span>
        {!isDefined && <span className="block-unresolved-badge">?</span>}
      </div>

      {expanded && isDefined && linkedDef && (
        <div className="block-body">
          {linkedDef.type === 'workflowDef' ? (
            <WorkflowContent def={linkedDef} />
          ) : (
            (linkedDef.body || []).map((stmt) => (
              <StatementBlock key={`${stmt.line}:${stmt.column}`} statement={stmt} />
            ))
          )}
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
