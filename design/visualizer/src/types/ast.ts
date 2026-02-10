// TypeScript types mirroring the Go AST JSON output

export interface Position {
  line: number
  column: number
}

// Top-level file
export interface TWFFile {
  definitions: Definition[]
}

// Definition types
export type Definition = WorkflowDef | ActivityDef

export interface WorkflowDef extends Position {
  type: 'workflowDef'
  name: string
  params: string
  returnType?: string
  options?: string
  signals: SignalDecl[]
  queries: QueryDecl[]
  updates: UpdateDecl[]
  body: Statement[]
}

export interface ActivityDef extends Position {
  type: 'activityDef'
  name: string
  params: string
  returnType?: string
  options?: string
  body: Statement[]
}

// Declaration types
export interface SignalDecl extends Position {
  type: 'signalDecl'
  name: string
  params: string
}

export interface QueryDecl extends Position {
  type: 'queryDecl'
  name: string
  params: string
  returnType?: string
}

export interface UpdateDecl extends Position {
  type: 'updateDecl'
  name: string
  params: string
  returnType?: string
}

// Statement types
export type Statement =
  | ActivityCall
  | WorkflowCall
  | TimerStmt
  | AwaitStmt
  | ParallelBlock
  | SelectBlock
  | SwitchBlock
  | IfStmt
  | ForStmt
  | ReturnStmt
  | ContinueAsNewStmt
  | BreakStmt
  | ContinueStmt
  | RawStmt
  | Comment

export interface ActivityCall extends Position {
  type: 'activityCall'
  name: string
  args: string
  result?: string
  options?: string
}

export type WorkflowCallMode = 'child' | 'spawn' | 'detach'

export interface WorkflowCall extends Position {
  type: 'workflowCall'
  mode: WorkflowCallMode
  namespace?: string
  name: string
  args: string
  result?: string
  options?: string
}

export interface TimerStmt extends Position {
  type: 'timer'
  duration: string
}

export interface AwaitTarget {
  kind: 'signal' | 'update'
  name: string
  args?: string
}

export interface AwaitStmt extends Position {
  type: 'await'
  targets: AwaitTarget[]
}

export interface ParallelBlock extends Position {
  type: 'parallel'
  body: Statement[]
}

export type SelectCaseKind = 'workflow' | 'activity' | 'signal' | 'update' | 'timer'

export interface SelectCase {
  kind: SelectCaseKind
  // Workflow case
  workflowMode?: WorkflowCallMode
  workflowNamespace?: string
  workflowName?: string
  workflowArgs?: string
  workflowResult?: string
  // Activity case
  activityName?: string
  activityArgs?: string
  activityResult?: string
  // Signal case
  signalName?: string
  signalArgs?: string
  // Update case
  updateName?: string
  updateArgs?: string
  // Timer case
  timerDuration?: string
  // Body
  body: Statement[]
}

export interface SelectBlock extends Position {
  type: 'select'
  cases: SelectCase[]
}

export interface SwitchCase {
  value: string
  body: Statement[]
}

export interface SwitchBlock extends Position {
  type: 'switch'
  expr: string
  cases: SwitchCase[]
  default?: Statement[]
}

export interface IfStmt extends Position {
  type: 'if'
  condition: string
  body: Statement[]
  elseBody?: Statement[]
}

export type ForVariant = 'infinite' | 'conditional' | 'iteration'

export interface ForStmt extends Position {
  type: 'for'
  variant: ForVariant
  condition?: string
  variable?: string
  iterable?: string
  body: Statement[]
}

export interface ReturnStmt extends Position {
  type: 'return'
  value?: string
}

export interface ContinueAsNewStmt extends Position {
  type: 'continueAsNew'
  args: string
}

export interface BreakStmt extends Position {
  type: 'break'
}

export interface ContinueStmt extends Position {
  type: 'continue'
}

export interface RawStmt extends Position {
  type: 'raw'
  text: string
}

export interface Comment extends Position {
  type: 'comment'
  text: string
}

// Type guards
export function isWorkflowDef(def: Definition): def is WorkflowDef {
  return def.type === 'workflowDef'
}

export function isActivityDef(def: Definition): def is ActivityDef {
  return def.type === 'activityDef'
}
