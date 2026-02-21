import React from 'react'
import type {
  Statement,
  AwaitStmt,
  AwaitAllBlock,
  AwaitOneBlock,
  AwaitOneCase,
} from '../../types/ast'
import { DefinitionContext, HandlerContext } from '../WorkflowCanvas'
import { useToggle } from './useToggle'
import { StatementBlock } from './StatementBlock'

// Shared await target display - both getAwaitStmtDisplay and getAwaitOneCaseDisplay delegate here
function getAwaitTargetDisplay(
  target: { kind: string; timer?: string; signal?: string; signalParams?: string; update?: string; updateParams?: string; activity?: string; activityArgs?: string; activityResult?: string; workflow?: string; workflowMode?: string; workflowArgs?: string; workflowResult?: string; nexus?: string; nexusService?: string; nexusOperation?: string; nexusArgs?: string; nexusResult?: string; nexusDetach?: boolean; ident?: string; identResult?: string },
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
    case 'nexus': {
      const detachPrefix = target.nexusDetach ? 'detach ' : ''
      const sig = `${target.nexus || ''} ${target.nexusService || ''}.${target.nexusOperation || ''}(${target.nexusArgs || ''})`
      const result = target.nexusResult ? ` → ${target.nexusResult}` : ''
      return { icon: '⬡', keyword: `${detachPrefix}nexus`, signature: `${sig}${result}`, isUnresolved: false }
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
    // Activity/workflow/nexus use SVG icons at block level, not text icons
    icon: (stmt.kind === 'activity' || stmt.kind === 'workflow' || stmt.kind === 'nexus') ? '' : target.icon,
    keyword: target.keyword ? `await ${target.keyword}` : 'await',
    blockClass: `block-await-stmt block-await-stmt-${stmt.kind}`,
  }
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

// Single await statement - await timer/signal/update/activity/workflow
export function AwaitStmtBlock({ stmt }: { stmt: AwaitStmt }) {
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

// Await All - expandable to show body (waits for all operations to complete)
export function AwaitAllBlockComponent({ stmt }: { stmt: AwaitAllBlock }) {
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
export function AwaitOneBlockComponent({ stmt }: { stmt: AwaitOneBlock }) {
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
