import type {
  ReturnStmt,
  CloseStmt,
  RawStmt,
  PromiseStmt,
  SetStmt,
  UnsetStmt,
} from '../../types/ast'
import { THEME, CLOSE_REASON_THEME } from '../../theme/temporal-theme'

// Return
export function ReturnBlock({ stmt }: { stmt: ReturnStmt }) {
  return (
    <div className="block block-return collapsed">
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">{THEME.return.icon}</span>
        <span className="block-keyword">return</span>
        {stmt.value && <span className="block-signature">{stmt.value}</span>}
      </div>
    </div>
  )
}

// Close - workflow termination
export function CloseBlock({ stmt }: { stmt: CloseStmt }) {
  const icon = (CLOSE_REASON_THEME[stmt.reason] ?? THEME.closeComplete).icon
  const statusClass = stmt.reason === 'continue_as_new' ? 'close-continue-as-new' : stmt.reason === 'fail' ? 'close-failed' : ''

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
export function RawBlock({ stmt }: { stmt: RawStmt }) {
  return (
    <div className="block block-raw collapsed">
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">{THEME.raw.icon}</span>
        <span className="block-code">{stmt.text}</span>
      </div>
    </div>
  )
}

// Simple block (break, continue)
export function SimpleBlock({ keyword, className }: { keyword: string; className: string }) {
  return (
    <div className={`block ${className} collapsed`}>
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">{THEME.breakContinue.icon}</span>
        <span className="block-keyword">{keyword}</span>
      </div>
    </div>
  )
}

// Promise statement - non-blocking async declaration
export function PromiseBlock({ stmt }: { stmt: PromiseStmt }) {
  // Determine the async target description
  let target = ''
  if (stmt.activity) {
    target = `activity ${stmt.activity}(${stmt.activityArgs || ''})`
  } else if (stmt.workflow) {
    target = `workflow ${stmt.workflow}(${stmt.workflowArgs || ''})`
  } else if (stmt.nexus) {
    target = `nexus ${stmt.nexus} ${stmt.nexusService || ''}.${stmt.nexusOperation || ''}(${stmt.nexusArgs || ''})`
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
        <span className="block-icon">{THEME.promise.icon}</span>
        <span className="block-keyword">promise</span>
        <span className="block-signature">{stmt.name} ← {target}</span>
      </div>
    </div>
  )
}

// Set condition to true
export function SetBlock({ stmt }: { stmt: SetStmt }) {
  return (
    <div className="block block-set collapsed">
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">{THEME.conditionSet.icon}</span>
        <span className="block-keyword">set</span>
        <span className="block-signature">{stmt.name}</span>
      </div>
    </div>
  )
}

// Unset condition (set to false)
export function UnsetBlock({ stmt }: { stmt: UnsetStmt }) {
  return (
    <div className="block block-unset collapsed">
      <div className="block-header">
        <span className="block-toggle-placeholder" />
        <span className="block-icon">{THEME.conditionUnset.icon}</span>
        <span className="block-keyword">unset</span>
        <span className="block-signature">{stmt.name}</span>
      </div>
    </div>
  )
}
