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
  ContinueAsNewStmt,
  RawStmt,
} from '../../types/ast'
import { DefinitionContextProvider, HandlerContextProvider } from '../WorkflowCanvas'
import { SingleGearIcon, InterlockingGearsIcon } from '../icons/GearIcons'
import { useRefocus } from './useRefocus'
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
    case 'continueAsNew':
      return <ContinueAsNewBlock stmt={statement} />
    case 'raw':
      return <RawBlock stmt={statement} />
    case 'break':
      return <SimpleBlock keyword="break" className="block-break" />
    case 'continue':
      return <SimpleBlock keyword="continue" className="block-continue" />
    case 'comment':
      return null // Skip comments in visualization
    default:
      return null
  }
}

// Activity Call - expandable to show activity definition body directly
function ActivityCallBlock({ stmt }: { stmt: ActivityCall }) {
  const [expanded, setExpanded] = React.useState(false)
  const context = React.useContext(DefinitionContextProvider)
  const activityDef = context.activities.get(stmt.name)
  const refocus = useRefocus()

  const signature = formatActivityCallSignature(stmt)

  const handleToggle = () => {
    setExpanded(!expanded)
    refocus()
  }

  return (
    <div className={`block block-activity ${expanded ? 'expanded' : 'collapsed'} ${!activityDef ? 'block-undefined' : ''}`}>
      <div className="block-header" onClick={handleToggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon"><SingleGearIcon /></span>
        <span className="block-keyword">activity</span>
        <span className="block-signature">{signature}</span>
        {!activityDef && <span className="block-undefined-badge">?</span>}
      </div>
      
      {expanded && (
        <div className="block-body">
          {activityDef ? (
            (activityDef.body || []).length > 0 ? (
              (activityDef.body || []).map((s, i) => (
                <StatementBlock key={i} statement={s} />
              ))
            ) : (
              <div className="block-empty-body">No implementation defined</div>
            )
          ) : (
            <MissingDefinition kind="activity" name={stmt.name} />
          )}
        </div>
      )}
    </div>
  )
}

// Workflow Call - expandable to show workflow definition body directly
function WorkflowCallBlock({ stmt }: { stmt: WorkflowCall }) {
  const [expanded, setExpanded] = React.useState(false)
  const [signalsExpanded, setSignalsExpanded] = React.useState(false)
  const [queriesExpanded, setQueriesExpanded] = React.useState(false)
  const [updatesExpanded, setUpdatesExpanded] = React.useState(false)
  const context = React.useContext(DefinitionContextProvider)
  const workflowDef = context.workflows.get(stmt.name)
  const refocus = useRefocus()

  const modePrefix = stmt.mode === 'spawn' ? 'spawn ' : stmt.mode === 'detach' ? 'detach ' : ''
  const signature = formatWorkflowCallSignature(stmt)

  const hasSignals = workflowDef?.signals && workflowDef.signals.length > 0
  const hasQueries = workflowDef?.queries && workflowDef.queries.length > 0
  const hasUpdates = workflowDef?.updates && workflowDef.updates.length > 0

  const handleToggle = () => { setExpanded(!expanded); refocus() }
  const toggleSignals = () => { setSignalsExpanded(!signalsExpanded); refocus() }
  const toggleQueries = () => { setQueriesExpanded(!queriesExpanded); refocus() }
  const toggleUpdates = () => { setUpdatesExpanded(!updatesExpanded); refocus() }

  return (
    <div className={`block block-workflow-call block-mode-${stmt.mode} ${expanded ? 'expanded' : 'collapsed'} ${!workflowDef ? 'block-undefined' : ''}`}>
      <div className="block-header" onClick={handleToggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon"><InterlockingGearsIcon /></span>
        <span className="block-keyword">{modePrefix}workflow</span>
        <span className="block-signature">{signature}</span>
        {!workflowDef && <span className="block-undefined-badge">?</span>}
      </div>
      
      {expanded && (
        <div className="block-body">
          {workflowDef ? (
            <>
              {/* Signals - data flowing IN to workflow */}
              {hasSignals && (
                <div className="block-declarations-group">
                  <div className="declarations-header" onClick={toggleSignals}>
                    <span className="block-toggle">{signalsExpanded ? '▼' : '▶'}</span>
                    <span className="declarations-icon declaration-signal">↪</span>
                    <span className="declarations-label">signals</span>
                    <span className="declarations-count">({workflowDef.signals!.length})</span>
                  </div>
                  {signalsExpanded && (
                    <div className="block-declarations">
                      {workflowDef.signals!.map((s, i) => (
                        <div key={i} className="declaration declaration-signal">
                          <span className="declaration-icon">↪</span>
                          <span className="declaration-keyword">signal</span>
                          <span className="declaration-name">{s.name}</span>
                          <span className="declaration-params">({s.params})</span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}
              {/* Queries - data flowing OUT of workflow */}
              {hasQueries && (
                <div className="block-declarations-group">
                  <div className="declarations-header" onClick={toggleQueries}>
                    <span className="block-toggle">{queriesExpanded ? '▼' : '▶'}</span>
                    <span className="declarations-icon declaration-query">↩</span>
                    <span className="declarations-label">queries</span>
                    <span className="declarations-count">({workflowDef.queries!.length})</span>
                  </div>
                  {queriesExpanded && (
                    <div className="block-declarations">
                      {workflowDef.queries!.map((q, i) => (
                        <div key={i} className="declaration declaration-query">
                          <span className="declaration-icon">↩</span>
                          <span className="declaration-keyword">query</span>
                          <span className="declaration-name">{q.name}</span>
                          <span className="declaration-params">({q.params})</span>
                          {q.returnType && <span className="declaration-return">→ {q.returnType}</span>}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}
              {/* Updates - data flowing BOTH ways */}
              {hasUpdates && (
                <div className="block-declarations-group">
                  <div className="declarations-header" onClick={toggleUpdates}>
                    <span className="block-toggle">{updatesExpanded ? '▼' : '▶'}</span>
                    <span className="declarations-icon declaration-update">⇄</span>
                    <span className="declarations-label">updates</span>
                    <span className="declarations-count">({workflowDef.updates!.length})</span>
                  </div>
                  {updatesExpanded && (
                    <div className="block-declarations">
                      {workflowDef.updates!.map((u, i) => (
                        <div key={i} className="declaration declaration-update">
                          <span className="declaration-icon">⇄</span>
                          <span className="declaration-keyword">update</span>
                          <span className="declaration-name">{u.name}</span>
                          <span className="declaration-params">({u.params})</span>
                          {u.returnType && <span className="declaration-return">→ {u.returnType}</span>}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}
              
              {/* Body statements */}
              <div className="block-statements">
                {(workflowDef.body || []).map((s, i) => (
                  <StatementBlock key={i} statement={s} />
                ))}
              </div>
            </>
          ) : (
            <MissingDefinition kind="workflow" name={stmt.name} />
          )}
        </div>
      )}
    </div>
  )
}

// Single await statement - await timer/signal/update/activity/workflow
function AwaitStmtBlock({ stmt }: { stmt: AwaitStmt }) {
  const [expanded, setExpanded] = React.useState(false)
  const context = React.useContext(DefinitionContextProvider)
  const handlers = React.useContext(HandlerContextProvider)
  const refocus = useRefocus()

  const { icon, keyword, signature, blockClass, expandableDef } = getAwaitStmtDisplay(stmt, context, handlers)

  const handleToggle = () => {
    if (expandableDef) { setExpanded(!expanded) }
    refocus()
  }

  return (
    <div className={`block ${blockClass} ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={handleToggle}>
        {expandableDef ? (
          <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        ) : (
          <span className="block-toggle-placeholder" />
        )}
        <span className="block-icon">{icon}</span>
        <span className="block-keyword">{keyword}</span>
        <span className="block-signature">{signature}</span>
      </div>

      {expanded && expandableDef && (
        <div className="block-body">
          {(expandableDef.body || []).length > 0 ? (
            (expandableDef.body || []).map((s, i) => (
              <StatementBlock key={i} statement={s} />
            ))
          ) : (
            <div className="block-empty-body">No implementation defined</div>
          )}
        </div>
      )}
    </div>
  )
}

// Get display info for single await statements
function getAwaitStmtDisplay(
  stmt: AwaitStmt,
  context: { activities: Map<string, any>; workflows: Map<string, any> },
  handlers: { signals: Map<string, any>; updates: Map<string, any> },
): { icon: string; keyword: string; signature: string; blockClass: string; expandableDef?: { body?: Statement[] } } {
  switch (stmt.kind) {
    case 'timer':
      return { icon: '⏱', keyword: 'await timer', signature: `(${stmt.timer || ''})`, blockClass: 'block-await-stmt block-await-stmt-timer' }
    case 'signal': {
      const sig = stmt.signal || ''
      const params = stmt.signalParams ? ` → ${stmt.signalParams}` : ''
      const handler = handlers.signals.get(sig)
      return { icon: '↪', keyword: 'await signal', signature: `${sig}${params}`, blockClass: 'block-await-stmt block-await-stmt-signal', expandableDef: handler }
    }
    case 'update': {
      const sig = stmt.update || ''
      const params = stmt.updateParams ? ` → ${stmt.updateParams}` : ''
      const handler = handlers.updates.get(sig)
      return { icon: '⇄', keyword: 'await update', signature: `${sig}${params}`, blockClass: 'block-await-stmt block-await-stmt-update', expandableDef: handler }
    }
    case 'activity': {
      const sig = `${stmt.activity || ''}(${stmt.activityArgs || ''})`
      const result = stmt.activityResult ? ` → ${stmt.activityResult}` : ''
      const def = context.activities.get(stmt.activity || '')
      return { icon: '', keyword: 'await activity', signature: `${sig}${result}`, blockClass: 'block-await-stmt block-await-stmt-activity', expandableDef: def }
    }
    case 'workflow': {
      const modePrefix = stmt.workflowMode === 'spawn' ? 'spawn ' : stmt.workflowMode === 'detach' ? 'detach ' : ''
      const sig = `${stmt.workflow || ''}(${stmt.workflowArgs || ''})`
      const result = stmt.workflowResult ? ` → ${stmt.workflowResult}` : ''
      const def = context.workflows.get(stmt.workflow || '')
      return { icon: '', keyword: `await ${modePrefix}workflow`, signature: `${sig}${result}`, blockClass: 'block-await-stmt block-await-stmt-workflow', expandableDef: def }
    }
    default:
      return { icon: '?', keyword: 'await', signature: '', blockClass: 'block-await-stmt' }
  }
}

// Await All - expandable to show body (waits for all operations to complete)
function AwaitAllBlockComponent({ stmt }: { stmt: AwaitAllBlock }) {
  const [expanded, setExpanded] = React.useState(true)
  const refocus = useRefocus()

  const handleToggle = () => { setExpanded(!expanded); refocus() }

  return (
    <div className={`block block-await-all ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={handleToggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon">⫴</span>
        <span className="block-keyword">await all</span>
        <span className="block-signature">{(stmt.body || []).length} branch(es)</span>
      </div>
      
      {expanded && (
        <div className="block-body">
          {(stmt.body || []).map((s, i) => (
            <StatementBlock key={i} statement={s} />
          ))}
        </div>
      )}
    </div>
  )
}

// Await One - expandable, shows cases where first to complete wins
function AwaitOneBlockComponent({ stmt }: { stmt: AwaitOneBlock }) {
  const [expanded, setExpanded] = React.useState(true)
  const refocus = useRefocus()
  const caseWord = stmt.cases.length === 1 ? 'case' : 'cases'

  const handleToggle = () => { setExpanded(!expanded); refocus() }

  return (
    <div className={`block block-await-one ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={handleToggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon-placeholder" />
        <span className="block-keyword">await one</span>
        <span className="block-signature">first of {stmt.cases.length} {caseWord}</span>
      </div>
      
      {expanded && (
        <div className="block-body">
          {stmt.cases.map((c, i) => (
            <AwaitOneCaseBlock key={i} awaitCase={c} />
          ))}
        </div>
      )}
    </div>
  )
}

// Render await one cases with unified tag design
function AwaitOneCaseBlock({ awaitCase }: { awaitCase: AwaitOneCase }) {
  const [expanded, setExpanded] = React.useState(false)
  const refocus = useRefocus()
  const hasBody = awaitCase.body && awaitCase.body.length > 0
  const isExpandable = hasBody || awaitCase.awaitAll

  // Determine display based on case kind
  const { contentClass, icon, keyword, signature } = getAwaitOneCaseDisplay(awaitCase)

  const handleToggle = () => { 
    if (isExpandable) { setExpanded(!expanded) }
    refocus()
  }

  return (
    <div className={`tagged-composite ${expanded ? 'expanded' : ''}`}>
      <div className="tagged-tag">
        <span className="tagged-tag-label">option</span>
      </div>
      <div className={`tagged-content ${contentClass} ${isExpandable ? 'expandable' : ''}`} onClick={handleToggle}>
        {isExpandable && <span className="block-toggle">{expanded ? '▼' : '▶'}</span>}
        {!isExpandable && <span className="block-toggle-placeholder" />}
        <span className="tagged-icon">{icon}</span>
        <span className="tagged-kind">{keyword}</span>
        <span className="tagged-name">{signature}</span>
      </div>
      {expanded && (
        <div className="tagged-body">
          {/* For await_all cases, show the nested await all block */}
          {awaitCase.awaitAll && (
            <AwaitAllBlockComponent stmt={awaitCase.awaitAll} />
          )}
          {/* Then show the body */}
          {hasBody && awaitCase.body.map((s, i) => (
            <StatementBlock key={i} statement={s} />
          ))}
        </div>
      )}
    </div>
  )
}

// Get display info for await one cases
function getAwaitOneCaseDisplay(c: AwaitOneCase): { contentClass: string; icon: string; keyword: string; signature: string } {
  switch (c.kind) {
    case 'signal': {
      const params = c.signalParams ? ` → ${c.signalParams}` : ''
      return { contentClass: 'tagged-signal', icon: '↪', keyword: 'signal', signature: `${c.signal || ''}${params}` }
    }
    case 'update': {
      const params = c.updateParams ? ` → ${c.updateParams}` : ''
      return { contentClass: 'tagged-update', icon: '⇄', keyword: 'update', signature: `${c.update || ''}${params}` }
    }
    case 'timer':
      return { contentClass: 'tagged-timer', icon: '⏱', keyword: 'timer', signature: `(${c.timer || ''})` }
    case 'activity': {
      const sig = `${c.activity || ''}(${c.activityArgs || ''})`
      const result = c.activityResult ? ` → ${c.activityResult}` : ''
      return { contentClass: 'tagged-activity', icon: '⚙', keyword: 'activity', signature: `${sig}${result}` }
    }
    case 'workflow': {
      const modePrefix = c.workflowMode === 'spawn' ? 'spawn ' : c.workflowMode === 'detach' ? 'detach ' : ''
      const sig = `${c.workflow || ''}(${c.workflowArgs || ''})`
      const result = c.workflowResult ? ` → ${c.workflowResult}` : ''
      return { contentClass: 'tagged-workflow', icon: '⚙⚙', keyword: `${modePrefix}workflow`, signature: `${sig}${result}` }
    }
    case 'await_all':
      return { contentClass: 'tagged-await-all', icon: '⫴', keyword: 'await all', signature: `${c.awaitAll?.body?.length || 0} branch(es)` }
    default:
      return { contentClass: 'tagged-raw', icon: '?', keyword: 'unknown', signature: '' }
  }
}

// Switch - expandable
function SwitchBlockComponent({ stmt }: { stmt: SwitchBlock }) {
  const [expanded, setExpanded] = React.useState(true)
  const refocus = useRefocus()

  const handleToggle = () => { setExpanded(!expanded); refocus() }

  return (
    <div className={`block block-switch ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={handleToggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon-placeholder" />
        <span className="block-keyword">switch</span>
        <span className="block-signature">{stmt.expr}</span>
      </div>
      
      {expanded && (
        <div className="block-body">
          {stmt.cases.map((c, i) => (
            <SwitchCaseBlock key={i} switchCase={c} />
          ))}
          {stmt.default && stmt.default.length > 0 && (
            <div className="block block-switch-default">
              <div className="block-header">
                <span className="block-toggle-placeholder" />
                <span className="block-icon-placeholder" />
                <span className="block-keyword">default</span>
              </div>
              <div className="block-body">
                {stmt.default.map((s, i) => (
                  <StatementBlock key={i} statement={s} />
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
  const [expanded, setExpanded] = React.useState(true)
  const refocus = useRefocus()

  const handleToggle = () => { setExpanded(!expanded); refocus() }

  return (
    <div className={`block block-switch-case ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={handleToggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon-placeholder" />
        <span className="block-keyword">case</span>
        <span className="block-signature">{switchCase.value}</span>
      </div>
      
      {expanded && switchCase.body && switchCase.body.length > 0 && (
        <div className="block-body">
          {switchCase.body.map((s, i) => (
            <StatementBlock key={i} statement={s} />
          ))}
        </div>
      )}
    </div>
  )
}

// If - expandable
function IfBlock({ stmt }: { stmt: IfStmt }) {
  const [expanded, setExpanded] = React.useState(true)
  const refocus = useRefocus()
  const hasElse = stmt.elseBody && stmt.elseBody.length > 0

  const handleToggle = () => { setExpanded(!expanded); refocus() }

  return (
    <div className={`block block-if ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={handleToggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon-placeholder" />
        <span className="block-keyword">if</span>
        <span className="block-signature">{stmt.condition}</span>
      </div>
      
      {expanded && (
        <div className="block-body">
          <div className="block-branch block-then">
            {(stmt.body || []).map((s, i) => (
              <StatementBlock key={i} statement={s} />
            ))}
          </div>
          {hasElse && (
            <div className="block-branch block-else">
              <div className="branch-label">else:</div>
              {(stmt.elseBody || []).map((s, i) => (
                <StatementBlock key={i} statement={s} />
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
  const [expanded, setExpanded] = React.useState(true)
  const refocus = useRefocus()
  
  let label = ''
  if (stmt.variant === 'iteration') {
    label = `${stmt.variable} in ${stmt.iterable}`
  } else if (stmt.variant === 'conditional') {
    label = stmt.condition || ''
  } else {
    label = '∞'
  }

  const handleToggle = () => { setExpanded(!expanded); refocus() }

  return (
    <div className={`block block-for ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={handleToggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon">↻</span>
        <span className="block-keyword">for</span>
        <span className="block-signature">{label}</span>
      </div>
      
      {expanded && (
        <div className="block-body">
          {(stmt.body || []).map((s, i) => (
            <StatementBlock key={i} statement={s} />
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
  const isFailed = stmt.reason === 'failed'
  const icon = isFailed ? '✕' : '✓'
  const statusClass = isFailed ? 'close-failed' : 'close-completed'
  
  // Build the label
  let label = 'close'
  if (stmt.reason) {
    label += ` ${stmt.reason}`
  }
  if (stmt.value) {
    label += ` "${stmt.value}"`
  }
  
  return (
    <div className={`block block-close ${statusClass} collapsed`}>
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">{icon}</span>
        <span className="block-keyword">close</span>
        {(stmt.reason || stmt.value) && (
          <span className="block-signature">
            {stmt.reason && <span className="close-reason">{stmt.reason}</span>}
            {stmt.value && <span className="close-value">"{stmt.value}"</span>}
          </span>
        )}
      </div>
    </div>
  )
}

// Continue as new
function ContinueAsNewBlock({ stmt }: { stmt: ContinueAsNewStmt }) {
  return (
    <div className="block block-continue-as-new collapsed">
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">⟳</span>
        <span className="block-keyword">continue_as_new</span>
        <span className="block-signature">{stmt.args}</span>
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

// Missing definition indicator
function MissingDefinition({ kind, name }: { kind: string; name: string }) {
  return (
    <div className="missing-definition">
      <span className="missing-definition-icon">⚠</span>
      <span className="missing-definition-text">
        No definition found for {kind} <strong>{name}</strong>
      </span>
      <span className="missing-definition-hint">
        Define it in a .twf file in this folder to see its contents
      </span>
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
