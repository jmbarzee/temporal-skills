import React from 'react'
import type { Definition, WorkflowDef, ActivityDef } from '../../types/ast'
import { StatementBlock } from './StatementBlock'
import { SingleGearIcon, InterlockingGearsIcon } from '../icons/GearIcons'
import './blocks.css'

interface DefinitionBlockProps {
  definition: Definition
}

export function DefinitionBlock({ definition }: DefinitionBlockProps) {
  const [expanded, setExpanded] = React.useState(false)

  if (definition.type === 'workflowDef') {
    return <WorkflowDefBlock def={definition} expanded={expanded} onToggle={() => setExpanded(!expanded)} />
  } else {
    return <ActivityDefBlock def={definition} expanded={expanded} onToggle={() => setExpanded(!expanded)} />
  }
}

interface WorkflowDefBlockProps {
  def: WorkflowDef
  expanded: boolean
  onToggle: () => void
}

function WorkflowDefBlock({ def, expanded, onToggle }: WorkflowDefBlockProps) {
  const signature = formatWorkflowSignature(def)
  const [signalsExpanded, setSignalsExpanded] = React.useState(false)
  const [queriesExpanded, setQueriesExpanded] = React.useState(false)
  const [updatesExpanded, setUpdatesExpanded] = React.useState(false)

  const hasSignals = def.signals && def.signals.length > 0
  const hasQueries = def.queries && def.queries.length > 0
  const hasUpdates = def.updates && def.updates.length > 0

  return (
    <div className={`block block-workflow ${expanded ? 'expanded' : 'collapsed'}`}>
      <div className="block-header" onClick={onToggle}>
        <span className="block-toggle">{expanded ? '▼' : '▶'}</span>
        <span className="block-icon"><InterlockingGearsIcon /></span>
        <span className="block-keyword">workflow</span>
        <span className="block-signature">{signature}</span>
      </div>
      
      {expanded && (
        <div className="block-body">
          {/* Signals - data flowing IN to workflow */}
          {hasSignals && (
            <div className="block-declarations-group">
              <div className="declarations-header" onClick={() => setSignalsExpanded(!signalsExpanded)}>
                <span className="block-toggle">{signalsExpanded ? '▼' : '▶'}</span>
                <span className="declarations-icon declaration-signal">↪</span>
                <span className="declarations-label">signals</span>
                <span className="declarations-count">({def.signals!.length})</span>
              </div>
              {signalsExpanded && (
                <div className="block-declarations">
                  {def.signals!.map((s, i) => (
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
                <span className="declarations-count">({def.queries!.length})</span>
              </div>
              {queriesExpanded && (
                <div className="block-declarations">
                  {def.queries!.map((q, i) => (
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
                <span className="declarations-count">({def.updates!.length})</span>
              </div>
              {updatesExpanded && (
                <div className="block-declarations">
                  {def.updates!.map((u, i) => (
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
            {(def.body || []).map((stmt, i) => (
              <StatementBlock key={i} statement={stmt} />
            ))}
          </div>
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
          {(def.body || []).map((stmt, i) => (
            <StatementBlock key={i} statement={stmt} />
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
