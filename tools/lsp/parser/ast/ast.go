package ast

// Node is the base interface for all AST nodes.
type Node interface {
	NodeLine() int
	NodeColumn() int
}

// Definition is a top-level definition (workflow, activity, or worker).
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

// WorkerRef is a reference to a workflow or activity name inside a worker block.
type WorkerRef struct {
	Pos
	Name string
}

type WorkerDef struct {
	Pos
	Name       string
	Namespace  string
	TaskQueue  string
	Workflows  []WorkerRef
	Activities []WorkerRef
}

func (*WorkerDef) defNode() {}

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
	Name     string
	Args     string
	Result   string // optional
	Options  *OptionsBlock
	Resolved *ActivityDef
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
	Mode      WorkflowCallMode
	Namespace string // optional, from nexus "ns"
	Name      string
	Args      string
	Result    string // optional
	Options   *OptionsBlock
	Resolved  *WorkflowDef
}

func (*WorkflowCall) stmtNode() {}

// AwaitStmt represents a single await statement.
type AwaitStmt struct {
	Pos
	// Timer await
	Timer string // duration, e.g. "5m"

	// Signal await
	Signal       string // signal name
	SignalParams string // optional parameter binding, e.g. "(approver, timestamp)"
	SignalResolved *SignalDecl

	// Update await
	Update       string // update name
	UpdateParams string // optional parameter binding
	UpdateResolved *UpdateDecl

	// Activity await
	Activity string // activity name
	ActivityArgs string
	ActivityResult string // optional result binding
	ActivityResolved *ActivityDef

	// Workflow await
	Workflow string // workflow name
	WorkflowMode WorkflowCallMode
	WorkflowNamespace string // optional nexus namespace
	WorkflowArgs string
	WorkflowResult string // optional result binding
	WorkflowResolved *WorkflowDef

	// Ident await (promise or condition reference)
	Ident       string // promise or condition name
	IdentResult string // optional result binding (promises only)
}

// AwaitKind returns the kind of await statement.
func (a *AwaitStmt) AwaitKind() string {
	switch {
	case a.Timer != "":
		return "timer"
	case a.Signal != "":
		return "signal"
	case a.Update != "":
		return "update"
	case a.Activity != "":
		return "activity"
	case a.Workflow != "":
		return "workflow"
	case a.Ident != "":
		return "ident"
	default:
		return "unknown"
	}
}

func (*AwaitStmt) stmtNode() {}

// AwaitAllBlock represents an "await all:" block that waits for all operations to complete.
type AwaitAllBlock struct {
	Pos
	Body []Statement
}

func (*AwaitAllBlock) stmtNode() {}

// AwaitOneCase represents a single case in an "await one:" block.
// Can be signal, update, timer, activity, workflow, or nested await all.
type AwaitOneCase struct {
	Pos

	// Signal case
	Signal string // signal name
	SignalParams string // optional parameter binding, e.g. "(approver, timestamp)"
	SignalResolved *SignalDecl

	// Update case
	Update string // update name
	UpdateParams string // optional parameter binding
	UpdateResolved *UpdateDecl

	// Timer case
	Timer string // duration

	// Activity case
	Activity string // activity name
	ActivityArgs string
	ActivityResult string // optional result binding
	ActivityResolved *ActivityDef

	// Workflow case
	Workflow string // workflow name
	WorkflowMode WorkflowCallMode // spawn/detach/child
	WorkflowNamespace string // optional nexus namespace
	WorkflowArgs string
	WorkflowResult string // optional result binding
	WorkflowResolved *WorkflowDef

	// Await all case (nested)
	AwaitAll *AwaitAllBlock

	// Ident case (promise or condition reference)
	Ident       string // promise or condition name
	IdentResult string // optional result binding (promises only)

	Body []Statement // optional body (can be empty)
}

// CaseKind returns the kind of this await one case.
func (c *AwaitOneCase) CaseKind() string {
	switch {
	case c.Signal != "":
		return "signal"
	case c.Update != "":
		return "update"
	case c.Timer != "":
		return "timer"
	case c.Activity != "":
		return "activity"
	case c.Workflow != "":
		return "workflow"
	case c.AwaitAll != nil:
		return "await_all"
	case c.Ident != "":
		return "ident"
	default:
		return "unknown"
	}
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

type CloseStmt struct {
	Pos
	Reason string // "complete", "fail", or "continue_as_new"
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
	Name string

	// The async target (exactly one set, mirrors AwaitStmt fields)
	Timer string

	Signal       string
	SignalParams string

	Update       string
	UpdateParams string

	Activity       string
	ActivityArgs   string

	Workflow          string
	WorkflowNamespace string
	WorkflowArgs      string
}

func (*PromiseStmt) stmtNode() {}

// SetStmt represents: set conditionName
type SetStmt struct {
	Pos
	Name string
}

func (*SetStmt) stmtNode() {}

// UnsetStmt represents: unset conditionName
type UnsetStmt struct {
	Pos
	Name string
}

func (*UnsetStmt) stmtNode() {}

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
