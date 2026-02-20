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
  state?: StateBlock
  signals: SignalDecl[]
  queries: QueryDecl[]
  updates: UpdateDecl[]
  body: Statement[]
  // Source file path (added by extension)
  sourceFile?: string
}

// State block declared at the top of a workflow definition
export interface StateBlock {
  conditions?: ConditionDecl[]
  rawStmts?: RawStmt[]
}

export interface ConditionDecl extends Position {
  name: string
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

// Handler declaration union (signal, query, update)
export type HandlerDecl = SignalDecl | QueryDecl | UpdateDecl

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
  | BreakStmt
  | ContinueStmt
  | RawStmt
  | Comment
  | PromiseStmt
  | SetStmt
  | UnsetStmt

export interface ActivityCall extends Position {
  type: 'activityCall'
  name: string
  args: string
  result?: string
  options?: string
}

export type WorkflowCallMode = 'child' | 'detach'

export interface WorkflowCall extends Position {
  type: 'workflowCall'
  mode: WorkflowCallMode
  namespace?: string
  name: string
  args: string
  result?: string
  options?: string
}

// Single await statement: await timer/signal/update/activity/workflow/ident
export type AwaitStmtKind = 'timer' | 'signal' | 'update' | 'activity' | 'workflow' | 'ident'

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
  // Ident await (promise or condition reference)
  ident?: string
  identResult?: string
}

// await all: waits for all operations to complete
export interface AwaitAllBlock extends Position {
  type: 'awaitAll'
  body: Statement[]
}

// await one case: signal, update, timer, activity, workflow, nested await all, or ident
export type AwaitOneCaseKind = 'signal' | 'update' | 'timer' | 'activity' | 'workflow' | 'await_all' | 'ident'

export interface AwaitOneCase extends Position {
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
  // Ident case (promise or condition reference)
  ident?: string
  identResult?: string
  // Body executed when this case wins (optional - can be empty)
  body: Statement[]
}

// await one: waits for first case to complete
export interface AwaitOneBlock extends Position {
  type: 'awaitOne'
  cases: AwaitOneCase[]
}

export interface SwitchCase extends Position {
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
  reason: string // 'complete', 'fail', or 'continue_as_new'
  args?: string
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

// Promise statement: promise name <- async_target
export interface PromiseStmt extends Position {
  type: 'promise'
  name: string
  // Async target (exactly one set)
  timer?: string
  signal?: string
  signalParams?: string
  update?: string
  updateParams?: string
  activity?: string
  activityArgs?: string
  workflow?: string
  workflowNamespace?: string
  workflowArgs?: string
}

// Set a condition to true
export interface SetStmt extends Position {
  type: 'set'
  name: string
}

// Set a condition to false
export interface UnsetStmt extends Position {
  type: 'unset'
  name: string
}

// Type guards
export function isWorkflowDef(def: Definition): def is WorkflowDef {
  return def.type === 'workflowDef'
}

export function isActivityDef(def: Definition): def is ActivityDef {
  return def.type === 'activityDef'
}
