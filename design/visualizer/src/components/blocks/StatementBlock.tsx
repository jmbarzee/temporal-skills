import React from 'react'
import type {
  Statement,
  ActivityCall,
  WorkflowCall,
  TimerStmt,
  AwaitStmt,
  ParallelBlock,
  SelectBlock,
  SwitchBlock,
  IfStmt,
  ForStmt,
  ReturnStmt,
  ContinueAsNewStmt,
  RawStmt,
} from '../../types/ast'
import { DefinitionContextProvider } from '../WorkflowCanvas'
import { SingleGearIcon, InterlockingGearsIcon } from '../icons/GearIcons'
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
    case 'timer':
      return <TimerBlock stmt={statement} />
    case 'await':
      return <AwaitBlock stmt={statement} />
    case 'parallel':
      return <ParallelBlockComponent stmt={statement} />
    case 'select':
      return <SelectBlockComponent stmt={statement} />
    case 'switch':
      return <SwitchBlockComponent stmt={statement} />
    case 'if':
      return <IfBlock stmt={statement} />
    case 'for':
      return <ForBlock stmt={statement} />
    case 'return':
      return <ReturnBlock stmt={statement} />
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

  const signature = formatActivityCallSignature(stmt)

  return (
    <div className={`block block-activity ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={() => activityDef && setExpanded(!expanded)}>
        {activityDef && <span className="block-toggle">{expanded ? '▼' : '▶'}</span>}
        {!activityDef && <span className="block-toggle-placeholder" />}
        <span className="block-icon"><SingleGearIcon /></span>
        <span className="block-keyword">activity</span>
        <span className="block-signature">{signature}</span>
      </div>
      
      {expanded && activityDef && (
        <div className="block-body">
          {(activityDef.body || []).map((s, i) => (
            <StatementBlock key={i} statement={s} />
          ))}
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

  const modePrefix = stmt.mode === 'spawn' ? 'spawn ' : stmt.mode === 'detach' ? 'detach ' : ''
  const signature = formatWorkflowCallSignature(stmt)

  const hasSignals = workflowDef?.signals && workflowDef.signals.length > 0
  const hasQueries = workflowDef?.queries && workflowDef.queries.length > 0
  const hasUpdates = workflowDef?.updates && workflowDef.updates.length > 0

  return (
    <div className={`block block-workflow-call block-mode-${stmt.mode} ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={() => workflowDef && setExpanded(!expanded)}>
        {workflowDef && <span className="block-toggle">{expanded ? '▼' : '▶'}</span>}
        {!workflowDef && <span className="block-toggle-placeholder" />}
        <span className="block-icon"><InterlockingGearsIcon /></span>
        <span className="block-keyword">{modePrefix}workflow</span>
        <span className="block-signature">{signature}</span>
      </div>
      
      {expanded && workflowDef && (
        <div className="block-body">
          {/* Signals - data flowing IN to workflow */}
          {hasSignals && (
            <div className="block-declarations-group">
              <div className="declarations-header" onClick={() => setSignalsExpanded(!signalsExpanded)}>
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
              <div className="declarations-header" onClick={() => setQueriesExpanded(!queriesExpanded)}>
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
              <div className="declarations-header" onClick={() => setUpdatesExpanded(!updatesExpanded)}>
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
        </div>
      )}
    </div>
  )
}

// Timer
function TimerBlock({ stmt }: { stmt: TimerStmt }) {
  return (
    <div className="block block-timer collapsed">
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">⏱</span>
        <span className="block-keyword">timer</span>
        <span className="block-signature">{stmt.duration}</span>
      </div>
    </div>
  )
}

// Await
function AwaitBlock({ stmt }: { stmt: AwaitStmt }) {
  const targets = stmt.targets.map(t => `${t.kind} ${t.name}`).join(', ')
  
  return (
    <div className="block block-await collapsed">
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">⏸</span>
        <span className="block-keyword">await</span>
        <span className="block-signature">{targets}</span>
      </div>
    </div>
  )
}

// Parallel - expandable to show body
function ParallelBlockComponent({ stmt }: { stmt: ParallelBlock }) {
  const [expanded, setExpanded] = React.useState(true)

  return (
    <div className={`block block-parallel ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={() => setExpanded(!expanded)}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon">⫴</span>
        <span className="block-keyword">parallel</span>
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

// Select - expandable, shows options where first to complete wins
function SelectBlockComponent({ stmt }: { stmt: SelectBlock }) {
  const [expanded, setExpanded] = React.useState(true)
  const optionWord = stmt.cases.length === 1 ? 'option' : 'options'

  return (
    <div className={`block block-select ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={() => setExpanded(!expanded)}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon-placeholder" />
        <span className="block-keyword">select</span>
        <span className="block-signature">soonest of {stmt.cases.length} {optionWord}</span>
      </div>
      
      {expanded && (
        <div className="block-body">
          {stmt.cases.map((c, i) => (
            <SelectOptionBlock key={i} option={c} />
          ))}
        </div>
      )}
    </div>
  )
}

// Render select options using the same styling as standalone primitives
function SelectOptionBlock({ option }: { option: SelectBlock['cases'][0] }) {
  const [expanded, setExpanded] = React.useState(false)
  const hasBody = option.body && option.body.length > 0

  // Determine block class and content based on kind
  const { blockClass, icon, keyword, signature } = getSelectOptionDisplay(option)

  return (
    <div className={`block ${blockClass} ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={() => hasBody && setExpanded(!expanded)}>
        {hasBody && <span className="block-toggle">{expanded ? '▼' : '▶'}</span>}
        {!hasBody && <span className="block-toggle-placeholder" />}
        <span className="block-icon">{icon}</span>
        <span className="block-keyword">{keyword}</span>
        <span className="block-signature">{signature}</span>
      </div>
      
      {expanded && hasBody && (
        <div className="block-body">
          {option.body.map((s, i) => (
            <StatementBlock key={i} statement={s} />
          ))}
        </div>
      )}
    </div>
  )
}

// Get display info matching standalone primitive styling
function getSelectOptionDisplay(c: SelectBlock['cases'][0]): { blockClass: string; icon: string | JSX.Element; keyword: string; signature: string } {
  switch (c.kind) {
    case 'signal':
      return { blockClass: 'block-signal', icon: '↪', keyword: 'signal', signature: c.signalName || '' }
    case 'update':
      return { blockClass: 'block-update', icon: '⇄', keyword: 'update', signature: c.updateName || '' }
    case 'activity':
      return { blockClass: 'block-activity', icon: <SingleGearIcon />, keyword: 'activity', signature: `${c.activityName}(${c.activityArgs || ''})` }
    case 'workflow':
      return { blockClass: 'block-workflow-call', icon: <InterlockingGearsIcon />, keyword: 'workflow', signature: `${c.workflowName}(${c.workflowArgs || ''})` }
    case 'timer':
      return { blockClass: 'block-timer', icon: '⏱', keyword: 'timer', signature: c.timerDuration || '' }
    default:
      return { blockClass: 'block-raw', icon: '?', keyword: 'unknown', signature: '' }
  }
}

// Switch - expandable
function SwitchBlockComponent({ stmt }: { stmt: SwitchBlock }) {
  const [expanded, setExpanded] = React.useState(true)

  return (
    <div className={`block block-switch ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={() => setExpanded(!expanded)}>
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

  return (
    <div className={`block block-switch-case ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={() => setExpanded(!expanded)}>
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
  const hasElse = stmt.elseBody && stmt.elseBody.length > 0

  return (
    <div className={`block block-if ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={() => setExpanded(!expanded)}>
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
      <div className="block-header" onClick={() => setExpanded(!expanded)}>
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

