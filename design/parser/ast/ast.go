package ast

// Node is the base interface for all AST nodes.
type Node interface {
	NodeLine() int
	NodeColumn() int
}

// Definition is a top-level definition (workflow or activity).
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
	Options    string // opaque, optional
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
	Options    string
	Body       []Statement
}

func (*ActivityDef) defNode() {}

// ---------------------------------------------------------------------------
// Workflow-level declarations (embedded in WorkflowDef)
// ---------------------------------------------------------------------------

type SignalDecl struct {
	Pos
	Name   string
	Params string
}

func (*SignalDecl) stmtNode() {}

type QueryDecl struct {
	Pos
	Name       string
	Params     string
	ReturnType string
}

func (*QueryDecl) stmtNode() {}

type UpdateDecl struct {
	Pos
	Name       string
	Params     string
	ReturnType string
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
	Options  string // optional
	Resolved *ActivityDef
}

func (*ActivityCall) stmtNode() {}

// WorkflowCallMode describes how a workflow call is executed.
type WorkflowCallMode int

const (
	CallChild  WorkflowCallMode = iota // bare workflow call (child)
	CallSpawn                          // spawn workflow
	CallDetach                         // detach workflow (fire-and-forget)
)

type WorkflowCall struct {
	Pos
	Mode      WorkflowCallMode
	Namespace string // optional, from nexus "ns"
	Name      string
	Args      string
	Result    string // optional
	Options   string // optional
	Resolved  *WorkflowDef
}

func (*WorkflowCall) stmtNode() {}

type TimerStmt struct {
	Pos
	Duration string // opaque
}

func (*TimerStmt) stmtNode() {}

// AwaitTarget represents a single target in an await statement.
type AwaitTarget struct {
	Pos
	Kind     string // "signal" or "update"
	Name     string
	Args     string // optional
	Resolved Node   // *SignalDecl or *UpdateDecl after resolution
}

type AwaitStmt struct {
	Pos
	Targets []*AwaitTarget
}

func (*AwaitStmt) stmtNode() {}

type ParallelBlock struct {
	Pos
	Body []Statement
}

func (*ParallelBlock) stmtNode() {}

// SelectCase represents a single case in a select block.
type SelectCase struct {
	Pos
	// Exactly one of these groups is set:
	// Workflow case
	WorkflowMode      WorkflowCallMode // only CallChild or CallSpawn
	WorkflowNamespace string
	WorkflowName      string
	WorkflowArgs      string
	WorkflowResult    string

	// Activity case
	ActivityName string
	ActivityArgs string
	ActivityResult string

	// Signal case
	SignalName string
	SignalArgs string

	// Update case
	UpdateName string
	UpdateArgs string

	// Timer case
	TimerDuration string

	Body []Statement
}

// CaseKind returns the kind of this select case.
func (sc *SelectCase) CaseKind() string {
	switch {
	case sc.WorkflowName != "":
		return "workflow"
	case sc.ActivityName != "":
		return "activity"
	case sc.SignalName != "":
		return "signal"
	case sc.UpdateName != "":
		return "update"
	case sc.TimerDuration != "":
		return "timer"
	default:
		return "unknown"
	}
}

func (*SelectCase) stmtNode() {}

type SelectBlock struct {
	Pos
	Cases []*SelectCase
}

func (*SelectBlock) stmtNode() {}

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

type ContinueAsNewStmt struct {
	Pos
	Args string
}

func (*ContinueAsNewStmt) stmtNode() {}

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
