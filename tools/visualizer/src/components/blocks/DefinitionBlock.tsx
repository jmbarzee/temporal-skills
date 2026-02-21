import React from 'react'
import type { Definition, WorkflowDef, ActivityDef, WorkerDef, WorkerRef, NamespaceDef, NamespaceWorker, NamespaceEndpoint, NexusServiceDef, NexusOperation, SignalDecl, QueryDecl, UpdateDecl } from '../../types/ast'
import { StatementBlock } from './StatementBlock'
import { WorkflowContent, InlineWorkflowBlock, SyncBodyBlock } from './WorkflowContent'
import { THEME, ThemeIcon, WORKER_REF_THEME } from '../../theme/temporal-theme'
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
    case 'namespaceDef':
      return <NamespaceDefBlock def={definition} />
    case 'nexusServiceDef':
      return <NexusServiceDefBlock def={definition} />
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
          <span className="block-icon"><ThemeIcon kind="workflow" /></span>
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
        <span className="block-icon"><ThemeIcon kind="activity" /></span>
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
        <span className="block-icon">{THEME.worker.icon}</span>
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
      : undefined
  const linkedService = refType === 'service'
    ? context.nexusServices.get(ref_.name)
    : undefined
  const isDefined = !!(linkedDef || linkedService)
  const [expanded, toggle] = useToggle(false, isDefined)

  const icon = WORKER_REF_THEME[refType].icon

  return (
    <div className={`worker-ref worker-ref-${refType} ${expanded ? 'expanded' : 'collapsed'} ${!isDefined ? 'worker-ref-unresolved' : ''}`}>
      <div className="worker-ref-header" onClick={toggle}>
        {isDefined ? (
          <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        ) : (
          <span className="block-toggle-placeholder" />
        )}
        <span className={`block-icon ${refType === 'service' ? 'block-icon-nexus-service' : ''}`}>{icon}</span>
        <span className="worker-ref-name">{ref_.name}</span>
        {!isDefined && <span className="block-unresolved-badge">?</span>}
      </div>

      {expanded && isDefined && (
        <div className="block-body">
          {linkedDef?.type === 'workflowDef' ? (
            <WorkflowContent def={linkedDef} />
          ) : linkedDef ? (
            (linkedDef.body || []).map((stmt) => (
              <StatementBlock key={`${stmt.line}:${stmt.column}`} statement={stmt} />
            ))
          ) : linkedService ? (
            (linkedService.operations || []).map((op) => (
              <NexusOperationBlock key={`${op.line}:${op.column}`} operation={op} />
            ))
          ) : null}
        </div>
      )}
    </div>
  )
}

function NamespaceDefBlock({ def }: { def: NamespaceDef }) {
  const [expanded, toggle] = useToggle()

  const totalEntries = (def.workers?.length || 0) + (def.endpoints?.length || 0)

  return (
    <div className={`block block-namespace-def ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={toggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon block-icon-namespace">{THEME.namespace.icon}</span>
        <span className="block-keyword">namespace</span>
        <span className="block-signature">{def.name} ({totalEntries} entries)</span>
      </div>

      {expanded && (
        <div className="block-body">
          {def.workers?.length > 0 && (
            <div className="namespace-entry-section">
              <div className="namespace-entry-label">workers</div>
              {def.workers.map((w) => (
                <NamespaceWorkerEntry key={`${w.line}:${w.column}`} entry={w} />
              ))}
            </div>
          )}
          {def.endpoints?.length > 0 && (
            <div className="namespace-entry-section">
              <div className="namespace-entry-label">nexus endpoints</div>
              {def.endpoints.map((ep) => (
                <NamespaceEndpointEntry key={`${ep.line}:${ep.column}`} entry={ep} />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function NamespaceWorkerEntry({ entry }: { entry: NamespaceWorker }) {
  const context = React.useContext(DefinitionContext)
  const workerDef = context.workers.get(entry.workerName)
  const isDefined = !!workerDef
  const [expanded, toggle] = useToggle(false, isDefined)

  return (
    <div className={`namespace-entry namespace-entry-worker ${expanded ? 'expanded' : 'collapsed'} ${!isDefined ? 'namespace-entry-unresolved' : ''}`}>
      <div className="namespace-entry-header" onClick={toggle}>
        {isDefined ? (
          <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        ) : (
          <span className="block-toggle-placeholder" />
        )}
        <span className="block-icon">{THEME.worker.icon}</span>
        <span className="namespace-entry-name">{entry.workerName}</span>
        {!isDefined && <span className="block-unresolved-badge">?</span>}
      </div>

      {expanded && isDefined && workerDef && (
        <div className="block-body">
          {workerDef.workflows?.length > 0 && (
            <WorkerRefSection label="workflows" refs={workerDef.workflows} refType="workflow" />
          )}
          {workerDef.activities?.length > 0 && (
            <WorkerRefSection label="activities" refs={workerDef.activities} refType="activity" />
          )}
          {workerDef.services?.length > 0 && (
            <WorkerRefSection label="nexus services" refs={workerDef.services} refType="service" />
          )}
        </div>
      )}
    </div>
  )
}

function NamespaceEndpointEntry({ entry }: { entry: NamespaceEndpoint }) {
  return (
    <div className="namespace-entry namespace-entry-endpoint collapsed">
      <div className="namespace-entry-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon block-icon-nexus-endpoint">{THEME.nexusService.icon}</span>
        <span className="namespace-entry-name">{entry.endpointName}</span>
      </div>
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

function NexusServiceDefBlock({ def }: { def: NexusServiceDef }) {
  const [expanded, toggle] = useToggle()
  const opCount = def.operations?.length || 0

  return (
    <div className={`block block-nexus-service-def ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={toggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon block-icon-nexus-service">{THEME.nexusService.icon}</span>
        <span className="block-keyword">service</span>
        <span className="block-signature">{def.name} ({opCount} operation{opCount !== 1 ? 's' : ''})</span>
      </div>

      {expanded && (
        <div className="block-body">
          {(def.operations || []).map((op) => (
            <NexusOperationBlock key={`${op.line}:${op.column}`} operation={op} />
          ))}
        </div>
      )}
    </div>
  )
}

export function NexusOperationBlock({ operation }: { operation: NexusOperation }) {
  const context = React.useContext(DefinitionContext)

  // For async operations, look up the linked workflow to get params/return and body
  const linkedWorkflow = operation.opType === 'async' && operation.workflowName
    ? context.workflows.get(operation.workflowName)
    : undefined

  // Determine expandability
  const isExpandable = operation.opType === 'async'
    ? !!linkedWorkflow
    : !!(operation.body && operation.body.length > 0)
  const isUnresolved = operation.opType === 'async' && operation.workflowName && !linkedWorkflow

  const [expanded, toggle] = useToggle(false, isExpandable)

  // Build signature
  let signature: React.ReactNode
  if (operation.opType === 'async' && linkedWorkflow) {
    signature = (
      <>
        {operation.name}
        <span className="nexus-operation-grayed-sig">({linkedWorkflow.params}){linkedWorkflow.returnType ? ` → ${linkedWorkflow.returnType}` : ''}</span>
      </>
    )
  } else if (operation.opType === 'sync') {
    const params = operation.params || ''
    const ret = operation.returnType ? ` → ${operation.returnType}` : ''
    signature = `${operation.name}(${params})${ret}`
  } else {
    // Async but workflow not found
    signature = operation.name
  }

  return (
    <div className={`block block-nexus-operation nexus-operation-${operation.opType} ${expanded ? 'expanded' : 'collapsed'} ${isUnresolved ? 'block-unresolved' : ''}`}>
      <div className="block-header" onClick={toggle}>
        {isExpandable ? (
          <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        ) : (
          <span className="block-toggle-placeholder" />
        )}
        <span className="block-icon block-icon-nexus-operation">{THEME.nexusOperation.icon}</span>
        <span className="block-keyword">{operation.opType}</span>
        <span className="block-signature">{signature}</span>
        {isUnresolved && <span className="block-unresolved-badge">?</span>}
      </div>

      {expanded && isExpandable && (
        <div className="block-body">
          {operation.opType === 'async' && linkedWorkflow ? (
            <InlineWorkflowBlock def={linkedWorkflow} />
          ) : operation.body ? (
            <SyncBodyBlock body={operation.body} />
          ) : null}
        </div>
      )}
    </div>
  )
}

function formatActivitySignature(def: ActivityDef): string {
  let sig = `${def.name}(${def.params})`
  if (def.returnType) {
    sig += ` → ${def.returnType}`
  }
  return sig
}
