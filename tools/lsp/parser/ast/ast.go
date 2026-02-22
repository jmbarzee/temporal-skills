package ast

// Node is the base interface for all AST nodes.
type Node interface {
	NodeLine() int
	NodeColumn() int
}

// Definition is a top-level definition (workflow, activity, worker, or namespace).
type Definition interface {
	Node
	defNode()
}

// Statement is a statement inside a body.
type Statement interface {
	Node
	stmtNode()
}

// Pos holds source position information.
type Pos struct {
	Line   int
	Column int
}

func (p Pos) NodeLine() int   { return p.Line }
func (p Pos) NodeColumn() int { return p.Column }

// Ref is a named reference to another AST node, resolved after parsing.
type Ref[T any] struct {
	Pos
	Name     string
	Resolved T
}

// File represents a parsed .twf file.
type File struct {
	Definitions []Definition
}

// ---------------------------------------------------------------------------
// Top-level definitions
// ---------------------------------------------------------------------------

type WorkflowDef struct {
	Pos
	Name       string
	Params     string // opaque content inside parens
	ReturnType string // opaque, optional
	State      *StateBlock
	Signals    []*SignalDecl
	Queries    []*QueryDecl
	Updates    []*UpdateDecl
	Body       []Statement
}

func (*WorkflowDef) defNode() {}

type ActivityDef struct {
	Pos
	Name       string
	Params     string
	ReturnType string
	Body       []Statement
}

func (*ActivityDef) defNode() {}

type WorkerDef struct {
	Pos
	Name       string
	Workflows  []Ref[*WorkflowDef]
	Activities []Ref[*ActivityDef]
	Services   []Ref[*NexusServiceDef] // nexus service references
}

func (*WorkerDef) defNode() {}

// NamespaceWorker is a worker instantiation inside a namespace block.
type NamespaceWorker struct {
	Pos
	Worker  Ref[*WorkerDef]
	Options *OptionsBlock
}

// NamespaceEndpoint is a nexus endpoint instantiation inside a namespace block.
type NamespaceEndpoint struct {
	Pos
	EndpointName string
	Namespace    string // set by resolver: name of the owning namespace
	Options      *OptionsBlock
}

// NamespaceDef is a namespace definition that instantiates workers with options.
type NamespaceDef struct {
	Pos
	Name      string
	Workers   []NamespaceWorker
	Endpoints []NamespaceEndpoint
}

func (*NamespaceDef) defNode() {}

// ---------------------------------------------------------------------------
// Workflow-level declarations (embedded in WorkflowDef)
// ---------------------------------------------------------------------------

type SignalDecl struct {
	Pos
	Name   string
	Params string
	Body   []Statement // handler body
}

func (*SignalDecl) stmtNode() {}

type QueryDecl struct {
	Pos
	Name       string
	Params     string
	ReturnType string
	Body       []Statement // handler body (restricted: no temporal primitives)
}

func (*QueryDecl) stmtNode() {}

type UpdateDecl struct {
	Pos
	Name       string
	Params     string
	ReturnType string
	Body       []Statement // handler body
}

func (*UpdateDecl) stmtNode() {}

// ---------------------------------------------------------------------------
// Statements
// ---------------------------------------------------------------------------

type ActivityCall struct {
	Pos
	Activity Ref[*ActivityDef]
	Args     string
	Result   string // optional
	Options  *OptionsBlock
}

func (*ActivityCall) stmtNode() {}

// WorkflowCallMode describes how a workflow call is executed.
type WorkflowCallMode int

const (
	CallChild  WorkflowCallMode = iota // bare workflow call (child)
	CallDetach                         // detach workflow (fire-and-forget)
)

type WorkflowCall struct {
	Pos
	Mode     WorkflowCallMode
	Workflow Ref[*WorkflowDef]
	Args     string
	Result   string // optional
	Options  *OptionsBlock
}

func (*WorkflowCall) stmtNode() {}

// ---------------------------------------------------------------------------
// Async target discriminated union
// ---------------------------------------------------------------------------

// AsyncTarget is the interface for async operation targets used in await,
// await-one-case, and promise statements. Exactly one concrete type is used.
type AsyncTarget interface {
	asyncTarget() // unexported marker
}

// AsyncTargetKind returns a string identifying the kind of async target.
func AsyncTargetKind(t AsyncTarget) string {
	switch t.(type) {
	case *TimerTarget:
		return "timer"
	case *SignalTarget:
		return "signal"
	case *UpdateTarget:
		return "update"
	case *ActivityTarget:
		return "activity"
	case *WorkflowTarget:
		return "workflow"
	case *NexusTarget:
		return "nexus"
	case *IdentTarget:
		return "ident"
	default:
		return "unknown"
	}
}

type TimerTarget struct {
	Duration string
}

func (*TimerTarget) asyncTarget() {}

type SignalTarget struct {
	Signal Ref[*SignalDecl]
	Params string
}

func (*SignalTarget) asyncTarget() {}

type UpdateTarget struct {
	Update Ref[*UpdateDecl]
	Params string
}

func (*UpdateTarget) asyncTarget() {}

type ActivityTarget struct {
	Activity Ref[*ActivityDef]
	Args     string
	Result   string
}

func (*ActivityTarget) asyncTarget() {}

type WorkflowTarget struct {
	Workflow Ref[*WorkflowDef]
	Mode     WorkflowCallMode
	Args     string
	Result   string
}

func (*WorkflowTarget) asyncTarget() {}

type NexusTarget struct {
	Endpoint  Ref[*NamespaceEndpoint]
	Service   Ref[*NexusServiceDef]
	Operation Ref[*NexusOperation]
	Args      string
	Result    string
	Detach    bool
}

func (*NexusTarget) asyncTarget() {}

// IdentResolution holds the resolved target of an ident reference.
// Exactly one field is non-nil after successful resolution.
type IdentResolution struct {
	Promise   *PromiseStmt
	Condition *ConditionDecl
}

type IdentTarget struct {
	Name     string
	Result   string
	Resolved IdentResolution
}

func (*IdentTarget) asyncTarget() {}

// AwaitStmt represents a single await statement.
type AwaitStmt struct {
	Pos
	Target AsyncTarget
}

func (*AwaitStmt) stmtNode() {}

// AwaitAllBlock represents an "await all:" block that waits for all operations to complete.
type AwaitAllBlock struct {
	Pos
	Body []Statement
}

func (*AwaitAllBlock) stmtNode() {}

// AwaitOneCase represents a single case in an "await one:" block.
// Can be signal, update, timer, activity, workflow, nexus, ident, or nested await all.
type AwaitOneCase struct {
	Pos
	Target   AsyncTarget    // nil when AwaitAll is set
	AwaitAll *AwaitAllBlock // nil when Target is set
	Body     []Statement
}

func (*AwaitOneCase) stmtNode() {}

// AwaitOneBlock represents an "await one:" block that waits for the first case to complete.
type AwaitOneBlock struct {
	Pos
	Cases []*AwaitOneCase
}

func (*AwaitOneBlock) stmtNode() {}

// SwitchCase represents a single case in a switch block.
type SwitchCase struct {
	Pos
	Value string // opaque expression after "case"
	Body  []Statement
}

type SwitchBlock struct {
	Pos
	Expr    string // opaque, paren-delimited
	Cases   []*SwitchCase
	Default []Statement // optional else block
}

func (*SwitchBlock) stmtNode() {}

type IfStmt struct {
	Pos
	Condition string // opaque, paren-delimited
	Body      []Statement
	ElseBody  []Statement // optional
}

func (*IfStmt) stmtNode() {}

// ForVariant describes the kind of for loop.
type ForVariant int

const (
	ForInfinite    ForVariant = iota // for:
	ForConditional                   // for (condition):
	ForIteration                     // for (var in collection):
)

type ForStmt struct {
	Pos
	Variant   ForVariant
	Condition string // for conditional loops
	Variable  string // for iteration loops
	Iterable  string // for iteration loops
	Body      []Statement
}

func (*ForStmt) stmtNode() {}

type ReturnStmt struct {
	Pos
	Value string // opaque, optional
}

func (*ReturnStmt) stmtNode() {}

// CloseReason classifies the kind of workflow close.
type CloseReason int

const (
	CloseComplete      CloseReason = iota // close complete
	CloseFailWorkflow                     // close fail
	CloseContinueAsNew                    // close continue_as_new
)

type CloseStmt struct {
	Pos
	Reason CloseReason
	Args   string // opaque, optional (parenthesized args)
}

func (*CloseStmt) stmtNode() {}

type BreakStmt struct {
	Pos
}

func (*BreakStmt) stmtNode() {}

type ContinueStmt struct {
	Pos
}

func (*ContinueStmt) stmtNode() {}

type RawStmt struct {
	Pos
	Text string
}

func (*RawStmt) stmtNode() {}

type Comment struct {
	Pos
	Text string
}

func (*Comment) stmtNode() {}

// ---------------------------------------------------------------------------
// State block and new primitives
// ---------------------------------------------------------------------------

// StateBlock represents a state: block at the top of a workflow definition.
type StateBlock struct {
	Pos
	Conditions []*ConditionDecl
	RawStmts   []*RawStmt
}

// ConditionDecl represents a condition declaration inside a state block.
type ConditionDecl struct {
	Pos
	Name string
}

// PromiseStmt represents a promise declaration: promise name <- async_target
type PromiseStmt struct {
	Pos
	Name   string
	Target AsyncTarget
}

func (*PromiseStmt) stmtNode() {}

// SetStmt represents: set conditionName
type SetStmt struct {
	Pos
	Condition Ref[*ConditionDecl]
}

func (*SetStmt) stmtNode() {}

// UnsetStmt represents: unset conditionName
type UnsetStmt struct {
	Pos
	Condition Ref[*ConditionDecl]
}

func (*UnsetStmt) stmtNode() {}

// ---------------------------------------------------------------------------
// Nexus definitions and calls
// ---------------------------------------------------------------------------

// NexusOperationType distinguishes async vs sync nexus operations.
type NexusOperationType int

const (
	NexusOpAsync NexusOperationType = iota
	NexusOpSync
)

// NexusOperation is an operation inside a nexus service definition.
type NexusOperation struct {
	Pos
	OpType       NexusOperationType
	Name         string
	Workflow      Ref[*WorkflowDef] // async only: backing workflow
	Params       string      // sync only
	ReturnType   string      // sync only
	Body         []Statement // sync only
}

// NexusServiceDef is a top-level nexus service definition.
type NexusServiceDef struct {
	Pos
	Name       string
	Operations []*NexusOperation
}

func (*NexusServiceDef) defNode() {}

// NexusCall is a nexus service operation call inside a workflow body.
type NexusCall struct {
	Pos
	Detach    bool
	Endpoint  Ref[*NamespaceEndpoint]
	Service   Ref[*NexusServiceDef]
	Operation Ref[*NexusOperation]
	Args      string
	Result    string // optional
	Options   *OptionsBlock
}

func (*NexusCall) stmtNode() {}

// ---------------------------------------------------------------------------
// Options blocks
// ---------------------------------------------------------------------------

// OptionsBlock represents a structured options { ... } block.
type OptionsBlock struct {
	Pos
	Entries []*OptionEntry
}

// OptionEntry represents a single key-value pair or nested block inside options.
type OptionEntry struct {
	Pos
	Key       string
	Value     string         // literal for flat entries
	ValueType string         // "string", "duration", "number", "bool", "enum"
	Nested    []*OptionEntry // non-nil for nested blocks (e.g. retry_policy)
}
