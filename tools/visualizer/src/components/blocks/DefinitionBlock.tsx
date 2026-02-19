import React from 'react'
import type { Definition, WorkflowDef, ActivityDef, SignalDecl, QueryDecl, UpdateDecl } from '../../types/ast'
import { StatementBlock } from './StatementBlock'
import { SingleGearIcon, InterlockingGearsIcon } from '../icons/GearIcons'
import { useRefocus } from './useRefocus'
import { HandlerContext } from '../WorkflowCanvas'
import './blocks.css'

interface DefinitionBlockProps {
  definition: Definition
}

export function DefinitionBlock({ definition }: DefinitionBlockProps) {
  const [expanded, setExpanded] = React.useState(false)
  const refocus = useRefocus()

  const handleToggle = () => {
    setExpanded(!expanded)
    refocus()
  }

  if (definition.type === 'workflowDef') {
    return <WorkflowDefBlock def={definition} expanded={expanded} onToggle={handleToggle} />
  } else {
    return <ActivityDefBlock def={definition} expanded={expanded} onToggle={handleToggle} />
  }
}

interface WorkflowDefBlockProps {
  def: WorkflowDef
  expanded: boolean
  onToggle: () => void
}

function WorkflowDefBlock({ def, expanded, onToggle }: WorkflowDefBlockProps) {
  const signature = formatWorkflowSignature(def)
  const [stateExpanded, setStateExpanded] = React.useState(false)
  const [signalsExpanded, setSignalsExpanded] = React.useState(false)
  const [queriesExpanded, setQueriesExpanded] = React.useState(false)
  const [updatesExpanded, setUpdatesExpanded] = React.useState(false)
  const refocus = useRefocus()

  const hasState = def.state && ((def.state.conditions && def.state.conditions.length > 0) || (def.state.rawStmts && def.state.rawStmts.length > 0))
  const hasSignals = def.signals && def.signals.length > 0
  const hasQueries = def.queries && def.queries.length > 0
  const hasUpdates = def.updates && def.updates.length > 0

  const stateItemCount = (def.state?.conditions?.length || 0) + (def.state?.rawStmts?.length || 0)
  const toggleState = () => { setStateExpanded(!stateExpanded); refocus() }
  const toggleSignals = () => { setSignalsExpanded(!signalsExpanded); refocus() }
  const toggleQueries = () => { setQueriesExpanded(!queriesExpanded); refocus() }
  const toggleUpdates = () => { setUpdatesExpanded(!updatesExpanded); refocus() }

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
        <div className="block-header" onClick={onToggle}>
          <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
          <span className="block-icon"><InterlockingGearsIcon /></span>
          <span className="block-keyword">workflow</span>
          <span className="block-signature">{signature}</span>
        </div>
        
        {expanded && (
          <div className="block-body">
          {/* State block - conditions and raw state declarations */}
          {hasState && (
            <div className="block-declarations-group">
              <div className="declarations-header" onClick={toggleState}>
                <span className="block-toggle">{stateExpanded ? '▼' : '▶'}</span>
                <span className="declarations-icon declaration-condition">◉</span>
                <span className="declarations-label">state</span>
                <span className="declarations-count">({stateItemCount})</span>
              </div>
              {stateExpanded && (
                <div className="block-declarations">
                  {(def.state!.conditions || []).map((c) => (
                    <div key={`${c.line}:${c.column}`} className="declaration declaration-condition">
                      <div className="declaration-header">
                        <span className="block-toggle-placeholder" />
                        <span className="declaration-icon">◉</span>
                        <span className="declaration-keyword">condition</span>
                        <span className="declaration-name">{c.name}</span>
                      </div>
                    </div>
                  ))}
                  {(def.state!.rawStmts || []).map((r) => (
                    <div key={`${r.line}:${r.column}`} className="declaration declaration-raw-state">
                      <div className="declaration-header">
                        <span className="block-toggle-placeholder" />
                        <span className="declaration-icon">≡</span>
                        <span className="declaration-name">{r.text}</span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
          {/* Signals - data flowing IN to workflow */}
          {hasSignals && (
            <div className="block-declarations-group">
              <div className="declarations-header" onClick={toggleSignals}>
                <span className="block-toggle">{signalsExpanded ? '▼' : '▶'}</span>
                <span className="declarations-icon declaration-signal">↪</span>
                <span className="declarations-label">signals</span>
                <span className="declarations-count">({def.signals!.length})</span>
              </div>
              {signalsExpanded && (
                <div className="block-declarations">
                  {def.signals!.map((s) => (
                    <SignalDeclBlock key={`${s.line}:${s.column}`} decl={s} />
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
                <span className="declarations-count">({def.queries!.length})</span>
              </div>
              {queriesExpanded && (
                <div className="block-declarations">
                  {def.queries!.map((q) => (
                    <QueryDeclBlock key={`${q.line}:${q.column}`} decl={q} />
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
                <span className="declarations-count">({def.updates!.length})</span>
              </div>
              {updatesExpanded && (
                <div className="block-declarations">
                  {def.updates!.map((u) => (
                    <UpdateDeclBlock key={`${u.line}:${u.column}`} decl={u} />
                  ))}
                </div>
              )}
            </div>
          )}
          
          {/* Body statements */}
          <div className="block-statements">
            {(def.body || []).map((stmt) => (
              <StatementBlock key={`${stmt.line}:${stmt.column}`} statement={stmt} />
            ))}
          </div>
        </div>
      )}
      </div>
    </HandlerContext.Provider>
  )
}

// Signal declaration with expandable handler body
function SignalDeclBlock({ decl }: { decl: SignalDecl }) {
  const [expanded, setExpanded] = React.useState(false)
  const refocus = useRefocus()
  const hasBody = decl.body && decl.body.length > 0

  const handleToggle = () => { 
    if (hasBody) { setExpanded(!expanded) }
    refocus()
  }

  const signature = `${decl.name}(${decl.params})`

  return (
    <div className={`declaration declaration-signal ${expanded ? 'expanded' : ''} ${hasBody ? 'has-body' : ''}`}>
      <div className="declaration-header" onClick={handleToggle}>
        {hasBody && <span className="block-toggle">{expanded ? '▼' : '▶'}</span>}
        {!hasBody && <span className="block-toggle-placeholder" />}
        <span className="declaration-icon">↪</span>
        <span className="declaration-keyword">signal</span>
        <span className="declaration-name">{signature}</span>
      </div>
      {expanded && hasBody && (
        <div className="declaration-body">
          {decl.body!.map((stmt) => (
            <StatementBlock key={`${stmt.line}:${stmt.column}`} statement={stmt} />
          ))}
        </div>
      )}
    </div>
  )
}

// Query declaration with expandable handler body
function QueryDeclBlock({ decl }: { decl: QueryDecl }) {
  const [expanded, setExpanded] = React.useState(false)
  const refocus = useRefocus()
  const hasBody = decl.body && decl.body.length > 0

  const handleToggle = () => { 
    if (hasBody) { setExpanded(!expanded) }
    refocus()
  }

  let signature = `${decl.name}(${decl.params})`
  if (decl.returnType) signature += ` → ${decl.returnType}`

  return (
    <div className={`declaration declaration-query ${expanded ? 'expanded' : ''} ${hasBody ? 'has-body' : ''}`}>
      <div className="declaration-header" onClick={handleToggle}>
        {hasBody && <span className="block-toggle">{expanded ? '▼' : '▶'}</span>}
        {!hasBody && <span className="block-toggle-placeholder" />}
        <span className="declaration-icon">↩</span>
        <span className="declaration-keyword">query</span>
        <span className="declaration-name">{signature}</span>
      </div>
      {expanded && hasBody && (
        <div className="declaration-body">
          {decl.body!.map((stmt) => (
            <StatementBlock key={`${stmt.line}:${stmt.column}`} statement={stmt} />
          ))}
        </div>
      )}
    </div>
  )
}

// Update declaration with expandable handler body
function UpdateDeclBlock({ decl }: { decl: UpdateDecl }) {
  const [expanded, setExpanded] = React.useState(false)
  const refocus = useRefocus()
  const hasBody = decl.body && decl.body.length > 0

  const handleToggle = () => { 
    if (hasBody) { setExpanded(!expanded) }
    refocus()
  }

  let signature = `${decl.name}(${decl.params})`
  if (decl.returnType) signature += ` → ${decl.returnType}`

  return (
    <div className={`declaration declaration-update ${expanded ? 'expanded' : ''} ${hasBody ? 'has-body' : ''}`}>
      <div className="declaration-header" onClick={handleToggle}>
        {hasBody && <span className="block-toggle">{expanded ? '▼' : '▶'}</span>}
        {!hasBody && <span className="block-toggle-placeholder" />}
        <span className="declaration-icon">⇄</span>
        <span className="declaration-keyword">update</span>
        <span className="declaration-name">{signature}</span>
      </div>
      {expanded && hasBody && (
        <div className="declaration-body">
          {decl.body!.map((stmt) => (
            <StatementBlock key={`${stmt.line}:${stmt.column}`} statement={stmt} />
          ))}
        </div>
      )}
    </div>
  )
}

interface ActivityDefBlockProps {
  def: ActivityDef
  expanded: boolean
  onToggle: () => void
}

function ActivityDefBlock({ def, expanded, onToggle }: ActivityDefBlockProps) {
  const signature = formatActivitySignature(def)

  return (
    <div className={`block block-activity-def ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={onToggle}>
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
