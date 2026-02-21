import React from 'react'
import type { ActivityCall, WorkflowCall, NexusCall } from '../../types/ast'
import { DefinitionContext } from '../WorkflowCanvas'
import { WorkflowContent, InlineWorkflowBlock, SyncBodyBlock } from './WorkflowContent'
import { THEME, ThemeIcon } from '../../theme/temporal-theme'
import { useToggle } from './useToggle'
import { StatementBlock } from './StatementBlock'

// Activity Call - expandable to show activity definition body directly
export function ActivityCallBlock({ stmt }: { stmt: ActivityCall }) {
  const context = React.useContext(DefinitionContext)
  const activityDef = context.activities.get(stmt.name)
  const isDefined = !!activityDef
  const [expanded, toggle] = useToggle(false, isDefined)

  const signature = formatActivityCallSignature(stmt)

  return (
    <div className={`block block-activity ${expanded ? 'expanded' : 'collapsed'} ${!isDefined ? 'block-unresolved' : ''}`}>
      <div className="block-header" onClick={toggle}>
        {isDefined ? (
          <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        ) : (
          <span className="block-toggle-placeholder" />
        )}
        <span className="block-icon"><ThemeIcon kind="activity" /></span>
        <span className="block-keyword">activity</span>
        <span className="block-signature">{signature}</span>
        {!isDefined && <span className="block-unresolved-badge">?</span>}
      </div>

      {expanded && isDefined && (
        <div className="block-body">
          {(activityDef.body || []).length > 0 ? (
            (activityDef.body || []).map((s) => (
              <StatementBlock key={`${s.line}:${s.column}`} statement={s} />
            ))
          ) : (
            <div className="block-empty-body">No implementation defined</div>
          )}
        </div>
      )}
    </div>
  )
}

// Workflow Call - expandable to show workflow definition body directly
export function WorkflowCallBlock({ stmt }: { stmt: WorkflowCall }) {
  const context = React.useContext(DefinitionContext)
  const workflowDef = context.workflows.get(stmt.name)
  const isDefined = !!workflowDef
  const [expanded, toggle] = useToggle(false, isDefined)

  const modePrefix = stmt.mode === 'detach' ? 'detach ' : ''
  const signature = formatWorkflowCallSignature(stmt)

  return (
    <div className={`block block-workflow-call block-mode-${stmt.mode} ${expanded ? 'expanded' : 'collapsed'} ${!isDefined ? 'block-unresolved' : ''}`}>
      <div className="block-header" onClick={toggle}>
        {isDefined ? (
          <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        ) : (
          <span className="block-toggle-placeholder" />
        )}
        <span className="block-icon"><ThemeIcon kind="workflow" /></span>
        <span className="block-keyword">{modePrefix}workflow</span>
        <span className="block-signature">{signature}</span>
        {!isDefined && <span className="block-unresolved-badge">?</span>}
      </div>

      {expanded && isDefined && (
        <div className="block-body">
          <WorkflowContent def={workflowDef} />
        </div>
      )}
    </div>
  )
}

// Nexus Call - calls a nexus service operation (expandable via service context lookup)
export function NexusCallBlock({ stmt }: { stmt: NexusCall }) {
  const context = React.useContext(DefinitionContext)

  // Look up the service and operation from context
  const serviceDef = context.nexusServices.get(stmt.service)
  const operation = serviceDef?.operations?.find(op => op.name === stmt.operation)
  const isDefined = !!operation

  // For expansion: async shows linked workflow, sync shows body
  const linkedWorkflow = operation?.opType === 'async' && operation.workflowName
    ? context.workflows.get(operation.workflowName)
    : undefined
  const isExpandable = operation?.opType === 'async' ? !!linkedWorkflow : !!(operation?.body && operation.body.length > 0)

  const [expanded, toggle] = useToggle(false, isExpandable)

  const modePrefix = stmt.detach ? 'detach ' : ''
  const signature = `${stmt.endpoint} ${stmt.service}.${stmt.operation}(${stmt.args})`
  const result = stmt.result ? ` → ${stmt.result}` : ''

  return (
    <div className={`block block-nexus-call ${stmt.detach ? 'block-mode-detach' : ''} ${expanded ? 'expanded' : 'collapsed'} ${!isDefined && stmt.service ? 'block-unresolved' : ''}`}>
      <div className="block-header" onClick={toggle}>
        {isExpandable ? (
          <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        ) : (
          <span className="block-toggle-placeholder" />
        )}
        <span className="block-icon block-icon-nexus-call">{THEME.nexusCall.icon}</span>
        <span className="block-keyword">{modePrefix}nexus</span>
        <span className="block-signature">{signature}{result}</span>
        {!isDefined && stmt.service && <span className="block-unresolved-badge">?</span>}
      </div>

      {expanded && isExpandable && (
        <div className="block-body">
          {operation?.opType === 'async' && linkedWorkflow ? (
            <InlineWorkflowBlock def={linkedWorkflow} />
          ) : operation?.body ? (
            <SyncBodyBlock body={operation.body} />
          ) : (
            <div className="block-empty-body">No implementation defined</div>
          )}
        </div>
      )}
    </div>
  )
}

function formatActivityCallSignature(stmt: ActivityCall): string {
  let sig = `${stmt.name}(${stmt.args})`
  if (stmt.result) {
    sig += ` → ${stmt.result}`
  }
  return sig
}

function formatWorkflowCallSignature(stmt: WorkflowCall): string {
  let sig = `${stmt.name}(${stmt.args})`
  if (stmt.result) {
    sig += ` → ${stmt.result}`
  }
  return sig
}
