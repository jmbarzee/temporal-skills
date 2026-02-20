import type { WorkflowDef, HandlerDecl } from '../../types/ast'
import { StatementBlock } from './StatementBlock'
import { useToggle } from './useToggle'
import './blocks.css'

const handlerConfig = {
  signalDecl: { icon: '↪', keyword: 'signal', cssClass: 'declaration-signal' },
  queryDecl:  { icon: '↩', keyword: 'query',  cssClass: 'declaration-query' },
  updateDecl: { icon: '⇄', keyword: 'update', cssClass: 'declaration-update' },
} as const

function HandlerDeclBlock({ decl }: { decl: HandlerDecl }) {
  const hasBody = decl.body && decl.body.length > 0
  const [expanded, toggle] = useToggle(false, hasBody)
  const { icon, keyword, cssClass } = handlerConfig[decl.type]

  let signature = `${decl.name}(${decl.params})`
  if ('returnType' in decl && decl.returnType) signature += ` → ${decl.returnType}`

  return (
    <div className={`declaration ${cssClass} ${expanded ? 'expanded' : ''}`}>
      <div className="declaration-header" onClick={toggle}>
        {hasBody && <span className="block-toggle">{expanded ? '▼' : '▶'}</span>}
        {!hasBody && <span className="block-toggle-placeholder" />}
        <span className="declaration-icon">{icon}</span>
        <span className="declaration-keyword">{keyword}</span>
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

export function WorkflowContent({ def }: { def: WorkflowDef }) {
  const [stateExpanded, toggleState] = useToggle()
  const [signalsExpanded, toggleSignals] = useToggle()
  const [queriesExpanded, toggleQueries] = useToggle()
  const [updatesExpanded, toggleUpdates] = useToggle()

  const hasState = def.state && ((def.state.conditions && def.state.conditions.length > 0) || (def.state.rawStmts && def.state.rawStmts.length > 0))
  const hasSignals = def.signals && def.signals.length > 0
  const hasQueries = def.queries && def.queries.length > 0
  const hasUpdates = def.updates && def.updates.length > 0

  const stateItemCount = (def.state?.conditions?.length || 0) + (def.state?.rawStmts?.length || 0)

  return (
    <>
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
                <HandlerDeclBlock key={`${s.line}:${s.column}`} decl={s} />
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
                <HandlerDeclBlock key={`${q.line}:${q.column}`} decl={q} />
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
                <HandlerDeclBlock key={`${u.line}:${u.column}`} decl={u} />
              ))}
            </div>
          )}
        </div>
      )}

      {/* Body statements */}
      <div>
        {(def.body || []).map((stmt) => (
          <StatementBlock key={`${stmt.line}:${stmt.column}`} statement={stmt} />
        ))}
      </div>
    </>
  )
}
