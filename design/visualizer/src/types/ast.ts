// TypeScript types mirroring the Go AST JSON output

export interface Position {
  line: number
  column: number
}

// Per-file parse error
export interface FileError {
  file: string
  error: string
  stderr?: string
}

// Top-level file
export interface TWFFile {
  definitions: Definition[]
  // Added for focused-file visualization
  focusedFile?: string
  // Per-file parse errors and warnings
  errors?: FileError[]
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
  // Source file path (added by extension)
  sourceFile?: string
}

export interface ActivityDef extends Position {
  type: 'activityDef'
  name: string
  params: string
  returnType?: string
  options?: string
  body: Statement[]
  // Source file path (added by extension)
  sourceFile?: string
}

// Declaration types (with handler bodies)
export interface SignalDecl extends Position {
  type: 'signalDecl'
  name: string
  params: string
  body?: Statement[]
}

export interface QueryDecl extends Position {
  type: 'queryDecl'
  name: string
  params: string
  returnType?: string
  body?: Statement[]
}

export interface UpdateDecl extends Position {
  type: 'updateDecl'
  name: string
  params: string
  returnType?: string
  body?: Statement[]
}

// Statement types
export type Statement =
  | ActivityCall
  | WorkflowCall
  | AwaitStmt
  | AwaitAllBlock
  | AwaitOneBlock
  | SwitchBlock
  | IfStmt
  | ForStmt
  | ReturnStmt
  | CloseStmt
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

// Single await statement: await timer/signal/update/activity/workflow
export type AwaitStmtKind = 'timer' | 'signal' | 'update' | 'activity' | 'workflow'

export interface AwaitStmt extends Position {
  type: 'await'
  kind: AwaitStmtKind
  timer?: string
  signal?: string
  signalParams?: string
  update?: string
  updateParams?: string
  activity?: string
  activityArgs?: string
  activityResult?: string
  workflow?: string
  workflowMode?: string
  workflowNamespace?: string
  workflowArgs?: string
  workflowResult?: string
}

// await all: waits for all operations to complete
export interface AwaitAllBlock extends Position {
  type: 'awaitAll'
  body: Statement[]
}

// await one case: signal, update, timer, activity, workflow, or nested await all
export type AwaitOneCaseKind = 'signal' | 'update' | 'timer' | 'activity' | 'workflow' | 'await_all'

export interface AwaitOneCase {
  kind: AwaitOneCaseKind
  // Signal case
  signal?: string
  signalParams?: string
  // Update case
  update?: string
  updateParams?: string
  // Timer case
  timer?: string
  // Activity case
  activity?: string
  activityArgs?: string
  activityResult?: string
  // Workflow case
  workflow?: string
  workflowMode?: string
  workflowNamespace?: string
  workflowArgs?: string
  workflowResult?: string
  // Await all case (nested)
  awaitAll?: AwaitAllBlock
  // Body executed when this case wins (optional - can be empty)
  body: Statement[]
}

// await one: waits for first case to complete
export interface AwaitOneBlock extends Position {
  type: 'awaitOne'
  cases: AwaitOneCase[]
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

export interface CloseStmt extends Position {
  type: 'close'
  reason?: string // 'completed', 'failed', or empty (default completed)
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
