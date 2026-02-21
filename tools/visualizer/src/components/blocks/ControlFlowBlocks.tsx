import type { SwitchBlock, IfStmt, ForStmt } from '../../types/ast'
import { useToggle } from './useToggle'
import { StatementBlock } from './StatementBlock'
import { THEME } from '../../theme/temporal-theme'

// Switch - expandable
export function SwitchBlockComponent({ stmt }: { stmt: SwitchBlock }) {
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
export function IfBlock({ stmt }: { stmt: IfStmt }) {
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
export function ForBlock({ stmt }: { stmt: ForStmt }) {
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
        <span className="block-icon">{THEME.forLoop.icon}</span>
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
