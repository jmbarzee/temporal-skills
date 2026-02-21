import React from 'react'
import type {
  Statement,
  ActivityCall,
  WorkflowCall,
  AwaitStmt,
  AwaitAllBlock,
  AwaitOneBlock,
  AwaitOneCase,
  SwitchBlock,
  IfStmt,
  ForStmt,
  ReturnStmt,
  CloseStmt,

  RawStmt,
  PromiseStmt,
  SetStmt,
  UnsetStmt,
} from '../../types/ast'
import { DefinitionContext, HandlerContext } from '../WorkflowCanvas'
import { WorkflowContent } from './WorkflowContent'
import { SingleGearIcon, InterlockingGearsIcon } from '../icons/GearIcons'
import { useToggle } from './useToggle'
import './blocks.css'

interface StatementBlockProps {
  statement: Statement
}

export function StatementBlock({ statement }: StatementBlockProps) {
  switch (statement.type) {
    case 'activityCall':
      return <ActivityCallBlock stmt={statement} />
    case 'workflowCall':
      return <WorkflowCallBlock stmt={statement} />
    case 'await':
      return <AwaitStmtBlock stmt={statement} />
    case 'awaitAll':
      return <AwaitAllBlockComponent stmt={statement} />
    case 'awaitOne':
      return <AwaitOneBlockComponent stmt={statement} />
    case 'switch':
      return <SwitchBlockComponent stmt={statement} />
    case 'if':
      return <IfBlock stmt={statement} />
    case 'for':
      return <ForBlock stmt={statement} />
    case 'return':
      return <ReturnBlock stmt={statement} />
    case 'close':
      return <CloseBlock stmt={statement} />
    case 'raw':
      return <RawBlock stmt={statement} />
    case 'break':
      return <SimpleBlock keyword="break" className="block-break" />
    case 'continue':
      return <SimpleBlock keyword="continue" className="block-continue" />
    case 'promise':
      return <PromiseBlock stmt={statement} />
    case 'set':
      return <SetBlock stmt={statement} />
    case 'unset':
      return <UnsetBlock stmt={statement} />
    case 'comment':
      return null // Skip comments in visualization
    default:
      return null
  }
}

// Activity Call - expandable to show activity definition body directly
function ActivityCallBlock({ stmt }: { stmt: ActivityCall }) {
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
        <span className="block-icon"><SingleGearIcon /></span>
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
function WorkflowCallBlock({ stmt }: { stmt: WorkflowCall }) {
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
        <span className="block-icon"><InterlockingGearsIcon /></span>
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

// Single await statement - await timer/signal/update/activity/workflow
function AwaitStmtBlock({ stmt }: { stmt: AwaitStmt }) {
  const context = React.useContext(DefinitionContext)
  const handlers = React.useContext(HandlerContext)

  const { icon, keyword, signature, blockClass, expandableDef, isUnresolved } = getAwaitStmtDisplay(stmt, context, handlers)
  const [expanded, toggle] = useToggle(false, !!expandableDef)

  return (
    <div className={`block ${blockClass} ${expanded ? 'expanded' : 'collapsed'} ${isUnresolved ? 'block-unresolved' : ''}`}>
      <div className="block-header" onClick={toggle}>
        {expandableDef ? (
          <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        ) : (
          <span className="block-toggle-placeholder" />
        )}
        <span className="block-icon">{icon}</span>
        <span className="block-keyword">{keyword}</span>
        <span className="block-signature">{signature}</span>
        {isUnresolved && <span className="block-unresolved-badge">?</span>}
      </div>

      {expanded && expandableDef && (
        <div className="block-body">
          {(expandableDef.body || []).length > 0 ? (
            (expandableDef.body || []).map((s) => (
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

// Shared await target display - both getAwaitStmtDisplay and getAwaitOneCaseDisplay delegate here
function getAwaitTargetDisplay(
  target: { kind: string; timer?: string; signal?: string; signalParams?: string; update?: string; updateParams?: string; activity?: string; activityArgs?: string; activityResult?: string; workflow?: string; workflowMode?: string; workflowArgs?: string; workflowResult?: string; ident?: string; identResult?: string },
  context: { activities: Map<string, any>; workflows: Map<string, any> },
  handlers: { signals: Map<string, any>; updates: Map<string, any> },
): { icon: string; keyword: string; signature: string; expandableDef?: { body?: Statement[] }; isUnresolved: boolean } {
  switch (target.kind) {
    case 'timer':
      return { icon: '⏱', keyword: 'timer', signature: `(${target.timer || ''})`, isUnresolved: false }
    case 'signal': {
      const sig = target.signal || ''
      const params = target.signalParams ? ` → ${target.signalParams}` : ''
      const handler = handlers.signals.get(sig)
      return { icon: '↪', keyword: 'signal', signature: `${sig}${params}`, expandableDef: handler, isUnresolved: !handler }
    }
    case 'update': {
      const sig = target.update || ''
      const params = target.updateParams ? ` → ${target.updateParams}` : ''
      const handler = handlers.updates.get(sig)
      return { icon: '⇄', keyword: 'update', signature: `${sig}${params}`, expandableDef: handler, isUnresolved: !handler }
    }
    case 'activity': {
      const sig = `${target.activity || ''}(${target.activityArgs || ''})`
      const result = target.activityResult ? ` → ${target.activityResult}` : ''
      const def = context.activities.get(target.activity || '')
      return { icon: '⚙', keyword: 'activity', signature: `${sig}${result}`, expandableDef: def, isUnresolved: !def }
    }
    case 'workflow': {
      const modePrefix = target.workflowMode === 'detach' ? 'detach ' : ''
      const sig = `${target.workflow || ''}(${target.workflowArgs || ''})`
      const result = target.workflowResult ? ` → ${target.workflowResult}` : ''
      const def = context.workflows.get(target.workflow || '')
      return { icon: '⚙⚙', keyword: `${modePrefix}workflow`, signature: `${sig}${result}`, expandableDef: def, isUnresolved: !def }
    }
    case 'ident': {
      const name = target.ident || ''
      const result = target.identResult ? ` → ${target.identResult}` : ''
      return { icon: '◉', keyword: '', signature: `${name}${result}`, isUnresolved: false }
    }
    default:
      return { icon: '?', keyword: '', signature: '', isUnresolved: false }
  }
}

// Get display info for single await statements
function getAwaitStmtDisplay(
  stmt: AwaitStmt,
  context: { activities: Map<string, any>; workflows: Map<string, any> },
  handlers: { signals: Map<string, any>; updates: Map<string, any> },
): { icon: string; keyword: string; signature: string; blockClass: string; expandableDef?: { body?: Statement[] }; isUnresolved: boolean } {
  const target = getAwaitTargetDisplay(stmt, context, handlers)
  return {
    ...target,
    // Activity/workflow use SVG icons at block level, not text icons
    icon: (stmt.kind === 'activity' || stmt.kind === 'workflow') ? '' : target.icon,
    keyword: target.keyword ? `await ${target.keyword}` : 'await',
    blockClass: `block-await-stmt block-await-stmt-${stmt.kind}`,
  }
}

// Await All - expandable to show body (waits for all operations to complete)
function AwaitAllBlockComponent({ stmt }: { stmt: AwaitAllBlock }) {
  const [expanded, toggle] = useToggle(true)

  return (
    <div className={`block block-await-all ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={toggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon">⫴</span>
        <span className="block-keyword">await all</span>
        <span className="block-signature">{(stmt.body || []).length} branch(es)</span>
      </div>
      
      {expanded && (
        <div className="block-body">
          {(stmt.body || []).map((s) => (
            <StatementBlock key={`${s.line}:${s.column}`} statement={s} />
          ))}
        </div>
      )}
    </div>
  )
}

// Await One - expandable, shows cases where first to complete wins
function AwaitOneBlockComponent({ stmt }: { stmt: AwaitOneBlock }) {
  const [expanded, toggle] = useToggle(true)
  const caseWord = stmt.cases.length === 1 ? 'case' : 'cases'

  return (
    <div className={`block block-await-one ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={toggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon-placeholder" />
        <span className="block-keyword">await one</span>
        <span className="block-signature">first of {stmt.cases.length} {caseWord}</span>
      </div>
      
      {expanded && (
        <div className="block-body">
          {stmt.cases.map((c) => (
            <AwaitOneCaseBlock key={`${c.line}:${c.column}`} awaitCase={c} />
          ))}
        </div>
      )}
    </div>
  )
}

// Render await one cases with unified tag design
function AwaitOneCaseBlock({ awaitCase }: { awaitCase: AwaitOneCase }) {
  const context = React.useContext(DefinitionContext)
  const handlers = React.useContext(HandlerContext)
  const hasBody = awaitCase.body && awaitCase.body.length > 0
  const isExpandable = hasBody || !!awaitCase.awaitAll
  const [expanded, toggle] = useToggle(false, isExpandable)

  // Determine display based on case kind
  const { contentClass, icon, keyword, signature, isUnresolved } = getAwaitOneCaseDisplay(awaitCase, context, handlers)

  return (
    <div className={`tagged-composite ${expanded ? 'expanded' : ''} ${isUnresolved ? 'tagged-unresolved' : ''}`}>
      <div className="tagged-tag">
        <span className="tagged-tag-label">option</span>
      </div>
      <div className={`tagged-content ${contentClass} ${isExpandable ? 'expandable' : ''}`} onClick={toggle}>
        {isExpandable && <span className="block-toggle">{expanded ? '▼' : '▶'}</span>}
        {!isExpandable && <span className="block-toggle-placeholder" />}
        <span className="tagged-icon">{icon}</span>
        <span className="tagged-kind">{keyword}</span>
        <span className="tagged-name">{signature}</span>
        {isUnresolved && <span className="block-unresolved-badge">?</span>}
      </div>
      {expanded && (
        <div className="tagged-body">
          {/* For await_all cases, show the nested await all block */}
          {awaitCase.awaitAll && (
            <AwaitAllBlockComponent stmt={awaitCase.awaitAll} />
          )}
          {/* Then show the body */}
          {hasBody && awaitCase.body.map((s) => (
            <StatementBlock key={`${s.line}:${s.column}`} statement={s} />
          ))}
        </div>
      )}
    </div>
  )
}

// Get display info for await one cases
function getAwaitOneCaseDisplay(
  c: AwaitOneCase,
  context: { activities: Map<string, any>; workflows: Map<string, any> },
  handlers: { signals: Map<string, any>; updates: Map<string, any> },
): { contentClass: string; icon: string; keyword: string; signature: string; isUnresolved: boolean } {
  // await_all is case-only, handle separately
  if (c.kind === 'await_all') {
    return { contentClass: 'tagged-await-all', icon: '⫴', keyword: 'await all', signature: `${c.awaitAll?.body?.length || 0} branch(es)`, isUnresolved: false }
  }
  const target = getAwaitTargetDisplay(c, context, handlers)
  return {
    icon: target.icon,
    keyword: target.keyword,
    signature: target.signature,
    isUnresolved: target.isUnresolved,
    contentClass: `tagged-${c.kind}`,
  }
}

// Switch - expandable
function SwitchBlockComponent({ stmt }: { stmt: SwitchBlock }) {
  const [expanded, toggle] = useToggle(true)

  return (
    <div className={`block block-switch ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={toggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon-placeholder" />
        <span className="block-keyword">switch</span>
        <span className="block-signature">{stmt.expr}</span>
      </div>
      
      {expanded && (
        <div className="block-body">
          {stmt.cases.map((c) => (
            <SwitchCaseBlock key={`${c.line}:${c.column}`} switchCase={c} />
          ))}
          {stmt.default && stmt.default.length > 0 && (
            <div className="block block-switch-default">
              <div className="block-header">
                <span className="block-toggle-placeholder" />
                <span className="block-icon-placeholder" />
                <span className="block-keyword">default</span>
              </div>
              <div className="block-body">
                {stmt.default.map((s) => (
                  <StatementBlock key={`${s.line}:${s.column}`} statement={s} />
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function SwitchCaseBlock({ switchCase }: { switchCase: SwitchBlock['cases'][0] }) {
  const [expanded, toggle] = useToggle(true)

  return (
    <div className={`block block-switch-case ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={toggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon-placeholder" />
        <span className="block-keyword">case</span>
        <span className="block-signature">{switchCase.value}</span>
      </div>
      
      {expanded && switchCase.body && switchCase.body.length > 0 && (
        <div className="block-body">
          {switchCase.body.map((s) => (
            <StatementBlock key={`${s.line}:${s.column}`} statement={s} />
          ))}
        </div>
      )}
    </div>
  )
}

// If - expandable
function IfBlock({ stmt }: { stmt: IfStmt }) {
  const [expanded, toggle] = useToggle(true)
  const hasElse = stmt.elseBody && stmt.elseBody.length > 0

  return (
    <div className={`block block-if ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={toggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon-placeholder" />
        <span className="block-keyword">if</span>
        <span className="block-signature">{stmt.condition}</span>
      </div>
      
      {expanded && (
        <div className="block-body">
          <div className="block-branch">
            {(stmt.body || []).map((s) => (
              <StatementBlock key={`${s.line}:${s.column}`} statement={s} />
            ))}
          </div>
          {hasElse && (
            <div className="block-branch">
              <div className="branch-label">else:</div>
              {(stmt.elseBody || []).map((s) => (
                <StatementBlock key={`${s.line}:${s.column}`} statement={s} />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// For - expandable
function ForBlock({ stmt }: { stmt: ForStmt }) {
  const [expanded, toggle] = useToggle(true)

  let label = ''
  if (stmt.variant === 'iteration') {
    label = `${stmt.variable} in ${stmt.iterable}`
  } else if (stmt.variant === 'conditional') {
    label = stmt.condition || ''
  } else {
    label = '∞'
  }

  return (
    <div className={`block block-for ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={toggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon">↻</span>
        <span className="block-keyword">for</span>
        <span className="block-signature">{label}</span>
      </div>
      
      {expanded && (
        <div className="block-body">
          {(stmt.body || []).map((s) => (
            <StatementBlock key={`${s.line}:${s.column}`} statement={s} />
          ))}
        </div>
      )}
    </div>
  )
}

// Return
function ReturnBlock({ stmt }: { stmt: ReturnStmt }) {
  return (
    <div className="block block-return collapsed">
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">↩</span>
        <span className="block-keyword">return</span>
        {stmt.value && <span className="block-signature">{stmt.value}</span>}
      </div>
    </div>
  )
}

// Close - workflow termination
function CloseBlock({ stmt }: { stmt: CloseStmt }) {
  // Determine the icon and class based on reason
  const isContinueAsNew = stmt.reason === 'continue_as_new'
  const isFailed = stmt.reason === 'fail'
  const icon = isContinueAsNew ? '⟳' : isFailed ? '✕' : '✓'
  const statusClass = isContinueAsNew ? 'close-continue-as-new' : isFailed ? 'close-failed' : ''

  return (
    <div className={`block block-close ${statusClass} collapsed`}>
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">{icon}</span>
        <span className="block-keyword">close</span>
        <span className="block-signature">
          <span className="close-reason">{stmt.reason}</span>
          {stmt.args && <span>({stmt.args})</span>}
        </span>
      </div>
    </div>
  )
}

// Raw statement (code)
function RawBlock({ stmt }: { stmt: RawStmt }) {
  return (
    <div className="block block-raw collapsed">
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">≡</span>
        <span className="block-code">{stmt.text}</span>
      </div>
    </div>
  )
}

// Simple block (break, continue)
function SimpleBlock({ keyword, className }: { keyword: string; className: string }) {
  return (
    <div className={`block ${className} collapsed`}>
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">•</span>
        <span className="block-keyword">{keyword}</span>
      </div>
    </div>
  )
}

// Promise statement - non-blocking async declaration
function PromiseBlock({ stmt }: { stmt: PromiseStmt }) {
  // Determine the async target description
  let target = ''
  if (stmt.activity) {
    target = `activity ${stmt.activity}(${stmt.activityArgs || ''})`
  } else if (stmt.workflow) {
    const ns = stmt.workflowNamespace ? `nexus "${stmt.workflowNamespace}" ` : ''
    target = `${ns}workflow ${stmt.workflow}(${stmt.workflowArgs || ''})`
  } else if (stmt.timer) {
    target = `timer(${stmt.timer})`
  } else if (stmt.signal) {
    const params = stmt.signalParams ? `(${stmt.signalParams})` : ''
    target = `signal ${stmt.signal}${params}`
  } else if (stmt.update) {
    const params = stmt.updateParams ? `(${stmt.updateParams})` : ''
    target = `update ${stmt.update}${params}`
  }

  return (
    <div className="block block-promise collapsed">
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">◇</span>
        <span className="block-keyword">promise</span>
        <span className="block-signature">{stmt.name} ← {target}</span>
      </div>
    </div>
  )
}

// Set condition to true
function SetBlock({ stmt }: { stmt: SetStmt }) {
  return (
    <div className="block block-set collapsed">
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">◉</span>
        <span className="block-keyword">set</span>
        <span className="block-signature">{stmt.name}</span>
      </div>
    </div>
  )
}

// Unset condition (set to false)
function UnsetBlock({ stmt }: { stmt: UnsetStmt }) {
  return (
    <div className="block block-unset collapsed">
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">○</span>
        <span className="block-keyword">unset</span>
        <span className="block-signature">{stmt.name}</span>
      </div>
    </div>
  )
}

// Helper functions
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
