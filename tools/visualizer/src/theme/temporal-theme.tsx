import React from 'react'
import { SingleGearIcon, InterlockingGearsIcon } from '../components/icons/GearIcons'

// --- Core types ---

export interface PrimitiveTheme {
  icon: string
  label: string
  cssVarPrefix: string
  SvgIcon?: React.ComponentType<{ size?: number }>
}

export type PrimitiveKind =
  | 'workflow' | 'activity' | 'worker' | 'namespace'
  | 'nexusService' | 'nexusOperation' | 'nexusCall'
  | 'signal' | 'query' | 'update'
  | 'timer' | 'conditionSet' | 'conditionUnset'
  | 'promise' | 'return'
  | 'closeComplete' | 'closeFail' | 'closeContinueAsNew'
  | 'forLoop' | 'awaitAll' | 'raw' | 'breakContinue' | 'error'

// --- Central theme map ---

export const THEME: Record<PrimitiveKind, PrimitiveTheme> = {
  workflow:           { icon: '⚙⚙', label: 'Workflow',              cssVarPrefix: 'workflow',         SvgIcon: InterlockingGearsIcon },
  activity:           { icon: '⚙',   label: 'Activity',              cssVarPrefix: 'activity',         SvgIcon: SingleGearIcon },
  worker:             { icon: '□',   label: 'Worker',                cssVarPrefix: 'worker' },
  namespace:          { icon: '⧉',   label: 'Namespace',             cssVarPrefix: 'namespace' },
  nexusService:       { icon: '★',   label: 'Nexus Service',         cssVarPrefix: 'nexus-service' },
  nexusOperation:     { icon: '☆',   label: 'Nexus Operation',       cssVarPrefix: 'nexus-operation' },
  nexusCall:          { icon: '☆',   label: 'Nexus Call',            cssVarPrefix: 'nexus' },
  signal:             { icon: '↪',   label: 'Signal',                cssVarPrefix: 'signal' },
  query:              { icon: '↩',   label: 'Query',                 cssVarPrefix: 'query' },
  update:             { icon: '⇄',   label: 'Update',                cssVarPrefix: 'update' },
  timer:              { icon: '⏱',   label: 'Timer',                 cssVarPrefix: 'timer' },
  conditionSet:       { icon: '◉',   label: 'Set Condition',         cssVarPrefix: 'subtle' },
  conditionUnset:     { icon: '○',   label: 'Unset Condition',       cssVarPrefix: 'subtle' },
  promise:            { icon: '◇',   label: 'Promise',               cssVarPrefix: 'promise' },
  return:             { icon: '↩',   label: 'Return',                cssVarPrefix: 'return' },
  closeComplete:      { icon: '✓',   label: 'Close (Complete)',      cssVarPrefix: 'return' },
  closeFail:          { icon: '✕',   label: 'Close (Fail)',          cssVarPrefix: 'signal' },
  closeContinueAsNew: { icon: '⟳',   label: 'Close (Continue As New)', cssVarPrefix: 'continue-new' },
  forLoop:            { icon: '↻',   label: 'For Loop',              cssVarPrefix: 'control' },
  awaitAll:           { icon: '⫴',   label: 'Await All',             cssVarPrefix: 'control' },
  raw:                { icon: '≡',   label: 'Raw Code',              cssVarPrefix: 'raw' },
  breakContinue:      { icon: '•',   label: 'Break/Continue',        cssVarPrefix: 'subtle' },
  error:              { icon: '⚠',   label: 'Error',                 cssVarPrefix: 'signal' },
}

// --- Derived lookup tables ---

export interface DefTypeConfig {
  type: string
  icon: string
  label: string
  defaultOn: boolean
}

export const DEF_TYPE_CONFIGS: DefTypeConfig[] = [
  { type: 'namespaceDef',    icon: THEME.namespace.icon,    label: 'Namespaces',     defaultOn: false },
  { type: 'workerDef',       icon: THEME.worker.icon,       label: 'Workers',        defaultOn: true },
  { type: 'nexusServiceDef', icon: THEME.nexusService.icon, label: 'Nexus Services', defaultOn: false },
  { type: 'workflowDef',     icon: THEME.workflow.icon,     label: 'Workflows',      defaultOn: true },
  { type: 'activityDef',     icon: THEME.activity.icon,     label: 'Activities',     defaultOn: false },
]

export const DEF_TYPE_ORDER = new Map(DEF_TYPE_CONFIGS.map((cfg, i) => [cfg.type, i]))

export const HANDLER_CONFIG = {
  signalDecl: { icon: THEME.signal.icon, keyword: 'signal', cssClass: 'declaration-signal' },
  queryDecl:  { icon: THEME.query.icon,  keyword: 'query',  cssClass: 'declaration-query' },
  updateDecl: { icon: THEME.update.icon, keyword: 'update', cssClass: 'declaration-update' },
} as const

export const CLOSE_REASON_THEME: Record<string, PrimitiveTheme> = {
  complete:        THEME.closeComplete,
  fail:            THEME.closeFail,
  continue_as_new: THEME.closeContinueAsNew,
}

export const AWAIT_TARGET_THEME: Record<string, PrimitiveTheme> = {
  timer:    THEME.timer,
  signal:   THEME.signal,
  update:   THEME.update,
  activity: THEME.activity,
  workflow: THEME.workflow,
  nexus:    THEME.nexusCall,
  ident:    THEME.conditionSet,
}

export const WORKER_REF_THEME: Record<string, PrimitiveTheme> = {
  workflow: THEME.workflow,
  activity: THEME.activity,
  service:  THEME.nexusService,
}

// --- Helper component ---

export function ThemeIcon({ kind, size }: { kind: PrimitiveKind; size?: number }) {
  const entry = THEME[kind]
  if (entry.SvgIcon) return <entry.SvgIcon size={size} />
  return <>{entry.icon}</>
}
