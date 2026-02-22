package ast

import (
	"encoding/json"
	"fmt"
)

// JSON serialization with type discriminators for interface types.

// resolvedRefJSON is a lightweight JSON reference to a resolved AST definition.
// Instead of embedding the entire target node, we emit its name and source position.
type resolvedRefJSON struct {
	Name   string `json:"name"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// FileSummary provides a count of each definition type in the file.
type FileSummary struct {
	Namespaces    int `json:"namespaces"`
	Workers       int `json:"workers"`
	Workflows     int `json:"workflows"`
	Activities    int `json:"activities"`
	NexusServices int `json:"nexusServices"`
}

// FileJSON is the JSON-serializable representation of a File.
type FileJSON struct {
	Summary     FileSummary       `json:"summary"`
	Definitions []json.RawMessage `json:"definitions"`
}

// MarshalJSON implements json.Marshaler for File.
func (f *File) MarshalJSON() ([]byte, error) {
	fj := FileJSON{
		Definitions: make([]json.RawMessage, 0, len(f.Definitions)),
	}
	for _, def := range f.Definitions {
		switch def.(type) {
		case *WorkflowDef:
			fj.Summary.Workflows++
		case *ActivityDef:
			fj.Summary.Activities++
		case *WorkerDef:
			fj.Summary.Workers++
		case *NamespaceDef:
			fj.Summary.Namespaces++
		case *NexusServiceDef:
			fj.Summary.NexusServices++
		}
		data, err := marshalDefinition(def)
		if err != nil {
			return nil, err
		}
		fj.Definitions = append(fj.Definitions, data)
	}
	return json.Marshal(fj)
}

// marshalStatements marshals a slice of statements into JSON.
func marshalStatements(stmts []Statement) ([]json.RawMessage, error) {
	if len(stmts) == 0 {
		return nil, nil
	}
	out := make([]json.RawMessage, 0, len(stmts))
	for _, stmt := range stmts {
		data, err := marshalStatement(stmt)
		if err != nil {
			return nil, err
		}
		out = append(out, data)
	}
	return out, nil
}

// marshalDeclList marshals a slice of declarations using the given per-element function.
func marshalDeclList[D any, J any](decls []D, fn func(D) (*J, error)) ([]*J, error) {
	out := make([]*J, 0, len(decls))
	for _, d := range decls {
		j, err := fn(d)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, nil
}

func marshalSignalDecl(s *SignalDecl) (*SignalDeclJSON, error) {
	sj := &SignalDeclJSON{
		Type:   "signalDecl",
		Line:   s.Line,
		Column: s.Column,
		Name:   s.Name,
		Params: s.Params,
	}
	var err error
	if sj.Body, err = marshalStatements(s.Body); err != nil {
		return nil, err
	}
	return sj, nil
}

func marshalQueryDecl(q *QueryDecl) (*QueryDeclJSON, error) {
	qj := &QueryDeclJSON{
		Type:       "queryDecl",
		Line:       q.Line,
		Column:     q.Column,
		Name:       q.Name,
		Params:     q.Params,
		ReturnType: q.ReturnType,
	}
	var err error
	if qj.Body, err = marshalStatements(q.Body); err != nil {
		return nil, err
	}
	return qj, nil
}

func marshalUpdateDecl(u *UpdateDecl) (*UpdateDeclJSON, error) {
	uj := &UpdateDeclJSON{
		Type:       "updateDecl",
		Line:       u.Line,
		Column:     u.Column,
		Name:       u.Name,
		Params:     u.Params,
		ReturnType: u.ReturnType,
	}
	var err error
	if uj.Body, err = marshalStatements(u.Body); err != nil {
		return nil, err
	}
	return uj, nil
}

func marshalDefinition(def Definition) (json.RawMessage, error) {
	switch d := def.(type) {
	case *WorkflowDef:
		return json.Marshal(d)
	case *ActivityDef:
		return json.Marshal(d)
	case *WorkerDef:
		return json.Marshal(d)
	case *NamespaceDef:
		return json.Marshal(d)
	case *NexusServiceDef:
		return json.Marshal(d)
	default:
		return nil, fmt.Errorf("marshalDefinition: unhandled definition type %T", def)
	}
}

// OptionsBlockJSON is the JSON representation of an options block.
type OptionsBlockJSON struct {
	Entries []OptionEntryJSON `json:"entries"`
}

// OptionEntryJSON is the JSON representation of a single option entry.
type OptionEntryJSON struct {
	Key       string            `json:"key"`
	Value     string            `json:"value,omitempty"`
	ValueType string            `json:"valueType,omitempty"`
	Nested    []OptionEntryJSON `json:"nested,omitempty"`
}

func marshalOptionsBlock(ob *OptionsBlock) *OptionsBlockJSON {
	if ob == nil {
		return nil
	}
	obj := &OptionsBlockJSON{
		Entries: marshalOptionEntries(ob.Entries),
	}
	return obj
}

func marshalOptionEntries(entries []*OptionEntry) []OptionEntryJSON {
	result := make([]OptionEntryJSON, 0, len(entries))
	for _, e := range entries {
		ej := OptionEntryJSON{
			Key:       e.Key,
			Value:     e.Value,
			ValueType: e.ValueType,
		}
		if len(e.Nested) > 0 {
			ej.Nested = marshalOptionEntries(e.Nested)
		}
		result = append(result, ej)
	}
	return result
}

// WorkflowDefJSON is the JSON representation of WorkflowDef.
type WorkflowDefJSON struct {
	Type       string             `json:"type"`
	Line       int                `json:"line"`
	Column     int                `json:"column"`
	SourceFile string             `json:"sourceFile,omitempty"`
	Name       string             `json:"name"`
	Params     string             `json:"params"`
	ReturnType string             `json:"returnType,omitempty"`
	State      *StateBlockJSON    `json:"state,omitempty"`
	Signals    []*SignalDeclJSON  `json:"signals"`
	Queries    []*QueryDeclJSON   `json:"queries"`
	Updates    []*UpdateDeclJSON  `json:"updates"`
	Body       []json.RawMessage  `json:"body"`
}

// StateBlockJSON is the JSON representation of a state: block.
type StateBlockJSON struct {
	Conditions []*ConditionDeclJSON `json:"conditions,omitempty"`
	RawStmts   []rawStmtJSON        `json:"rawStmts,omitempty"`
}

// ConditionDeclJSON is the JSON representation of a condition declaration.
type ConditionDeclJSON struct {
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Name   string `json:"name"`
}

func (w *WorkflowDef) MarshalJSON() ([]byte, error) {
	wj := WorkflowDefJSON{
		Type:       "workflowDef",
		Line:       w.Line,
		Column:     w.Column,
		SourceFile: w.SourceFile,
		Name:       w.Name,
		Params:     w.Params,
		ReturnType: w.ReturnType,
	}
	if w.State != nil {
		sj := &StateBlockJSON{}
		for _, c := range w.State.Conditions {
			sj.Conditions = append(sj.Conditions, &ConditionDeclJSON{Line: c.Line, Column: c.Column, Name: c.Name})
		}
		for _, r := range w.State.RawStmts {
			sj.RawStmts = append(sj.RawStmts, rawStmtJSON{
				Type:   "raw",
				Line:   r.Line,
				Column: r.Column,
				Text:   r.Text,
			})
		}
		wj.State = sj
	}
	var err error
	if wj.Signals, err = marshalDeclList(w.Signals, marshalSignalDecl); err != nil {
		return nil, err
	}
	if wj.Queries, err = marshalDeclList(w.Queries, marshalQueryDecl); err != nil {
		return nil, err
	}
	if wj.Updates, err = marshalDeclList(w.Updates, marshalUpdateDecl); err != nil {
		return nil, err
	}
	if wj.Body, err = marshalStatements(w.Body); err != nil {
		return nil, err
	}
	return json.Marshal(wj)
}

// ActivityDefJSON is the JSON representation of ActivityDef.
type ActivityDefJSON struct {
	Type       string            `json:"type"`
	Line       int               `json:"line"`
	Column     int               `json:"column"`
	SourceFile string            `json:"sourceFile,omitempty"`
	Name       string            `json:"name"`
	Params     string            `json:"params"`
	ReturnType string            `json:"returnType,omitempty"`
	Body       []json.RawMessage `json:"body"`
}

func (a *ActivityDef) MarshalJSON() ([]byte, error) {
	aj := ActivityDefJSON{
		Type:       "activityDef",
		Line:       a.Line,
		Column:     a.Column,
		SourceFile: a.SourceFile,
		Name:       a.Name,
		Params:     a.Params,
		ReturnType: a.ReturnType,
	}
	var err error
	if aj.Body, err = marshalStatements(a.Body); err != nil {
		return nil, err
	}
	return json.Marshal(aj)
}

// WorkerRefJSON is the JSON representation of a Ref used in worker definitions.
type WorkerRefJSON struct {
	Name     string           `json:"name"`
	Line     int              `json:"line"`
	Column   int              `json:"column"`
	Resolved *resolvedRefJSON `json:"resolved,omitempty"`
}

// marshalWorkerRefs converts a slice of Ref[T] to JSON form.
func marshalWorkerRefs[T interface{ comparable; Node }](refs []Ref[T]) []WorkerRefJSON {
	if len(refs) == 0 {
		return nil
	}
	out := make([]WorkerRefJSON, 0, len(refs))
	for _, ref := range refs {
		rj := WorkerRefJSON{
			Name:   ref.Name,
			Line:   ref.Line,
			Column: ref.Column,
		}
		var zero T
		if ref.Resolved != zero {
			rj.Resolved = &resolvedRefJSON{
				Name:   ref.Name,
				Line:   ref.Resolved.NodeLine(),
				Column: ref.Resolved.NodeColumn(),
			}
		}
		out = append(out, rj)
	}
	return out
}

// WorkerDefJSON is the JSON representation of WorkerDef.
type WorkerDefJSON struct {
	Type       string          `json:"type"`
	Line       int             `json:"line"`
	Column     int             `json:"column"`
	SourceFile string          `json:"sourceFile,omitempty"`
	Name       string          `json:"name"`
	Workflows  []WorkerRefJSON `json:"workflows,omitempty"`
	Activities []WorkerRefJSON `json:"activities,omitempty"`
	Services   []WorkerRefJSON `json:"services,omitempty"`
}

func (w *WorkerDef) MarshalJSON() ([]byte, error) {
	wj := WorkerDefJSON{
		Type:       "workerDef",
		Line:       w.Line,
		Column:     w.Column,
		SourceFile: w.SourceFile,
		Name:       w.Name,
		Workflows:  marshalWorkerRefs(w.Workflows),
		Activities: marshalWorkerRefs(w.Activities),
		Services:   marshalWorkerRefs(w.Services),
	}
	return json.Marshal(wj)
}

// NamespaceWorkerJSON is the JSON representation of a worker instantiation in a namespace.
type NamespaceWorkerJSON struct {
	WorkerName     string            `json:"workerName"`
	Line           int               `json:"line"`
	Column         int               `json:"column"`
	Options        *OptionsBlockJSON `json:"options,omitempty"`
	ResolvedWorker *resolvedRefJSON  `json:"resolvedWorker,omitempty"`
}

// NamespaceEndpointJSON is the JSON representation of a nexus endpoint in a namespace.
type NamespaceEndpointJSON struct {
	EndpointName string            `json:"endpointName"`
	Line         int               `json:"line"`
	Column       int               `json:"column"`
	Options      *OptionsBlockJSON `json:"options,omitempty"`
}

// NamespaceDefJSON is the JSON representation of NamespaceDef.
type NamespaceDefJSON struct {
	Type       string                  `json:"type"`
	Line       int                     `json:"line"`
	Column     int                     `json:"column"`
	SourceFile string                  `json:"sourceFile,omitempty"`
	Name       string                  `json:"name"`
	Workers    []NamespaceWorkerJSON   `json:"workers,omitempty"`
	Endpoints  []NamespaceEndpointJSON `json:"endpoints,omitempty"`
}

func (n *NamespaceDef) MarshalJSON() ([]byte, error) {
	nj := NamespaceDefJSON{
		Type:       "namespaceDef",
		Line:       n.Line,
		Column:     n.Column,
		SourceFile: n.SourceFile,
		Name:       n.Name,
	}
	for _, w := range n.Workers {
		wj := NamespaceWorkerJSON{
			WorkerName: w.Worker.Name,
			Line:       w.Line,
			Column:     w.Column,
			Options:    marshalOptionsBlock(w.Options),
		}
		if w.Worker.Resolved != nil {
			wj.ResolvedWorker = &resolvedRefJSON{Name: w.Worker.Resolved.Name, Line: w.Worker.Resolved.Line, Column: w.Worker.Resolved.Column}
		}
		nj.Workers = append(nj.Workers, wj)
	}
	for _, ep := range n.Endpoints {
		nj.Endpoints = append(nj.Endpoints, NamespaceEndpointJSON{
			EndpointName: ep.EndpointName,
			Line:         ep.Line,
			Column:       ep.Column,
			Options:      marshalOptionsBlock(ep.Options),
		})
	}
	return json.Marshal(nj)
}

// Declaration JSON types
type SignalDeclJSON struct {
	Type   string            `json:"type"`
	Line   int               `json:"line"`
	Column int               `json:"column"`
	Name   string            `json:"name"`
	Params string            `json:"params"`
	Body   []json.RawMessage `json:"body,omitempty"`
}

type QueryDeclJSON struct {
	Type       string            `json:"type"`
	Line       int               `json:"line"`
	Column     int               `json:"column"`
	Name       string            `json:"name"`
	Params     string            `json:"params"`
	ReturnType string            `json:"returnType,omitempty"`
	Body       []json.RawMessage `json:"body,omitempty"`
}

type UpdateDeclJSON struct {
	Type       string            `json:"type"`
	Line       int               `json:"line"`
	Column     int               `json:"column"`
	Name       string            `json:"name"`
	Params     string            `json:"params"`
	ReturnType string            `json:"returnType,omitempty"`
	Body       []json.RawMessage `json:"body,omitempty"`
}

// marshalStatement marshals a Statement with type discrimination.
func marshalStatement(stmt Statement) (json.RawMessage, error) {
	switch s := stmt.(type) {
	case *ActivityCall:
		return marshalActivityCall(s)
	case *WorkflowCall:
		return marshalWorkflowCall(s)
	case *NexusCall:
		return marshalNexusCall(s)
	case *AwaitStmt:
		return marshalAwaitStmt(s)
	case *AwaitAllBlock:
		return marshalAwaitAllBlock(s)
	case *AwaitOneBlock:
		return marshalAwaitOneBlock(s)
	case *SwitchBlock:
		return marshalSwitchBlock(s)
	case *IfStmt:
		return marshalIfStmt(s)
	case *ForStmt:
		return marshalForStmt(s)
	case *ReturnStmt:
		return marshalReturnStmt(s)
	case *CloseStmt:
		return marshalCloseStmt(s)
	case *BreakStmt:
		return marshalBreakStmt(s)
	case *ContinueStmt:
		return marshalContinueStmt(s)
	case *RawStmt:
		return marshalRawStmt(s)
	case *Comment:
		return marshalComment(s)
	case *PromiseStmt:
		return marshalPromiseStmt(s)
	case *SetStmt:
		return marshalSetStmt(s)
	case *UnsetStmt:
		return marshalUnsetStmt(s)
	default:
		return nil, fmt.Errorf("marshalStatement: unhandled statement type %T", stmt)
	}
}

func marshalActivityCall(s *ActivityCall) (json.RawMessage, error) {
	aj := activityCallJSON{
		Type:    "activityCall",
		Line:    s.Line,
		Column:  s.Column,
		Name:    s.Activity.Name,
		Args:    s.Args,
		Result:  s.Result,
		Options: marshalOptionsBlock(s.Options),
	}
	if s.Activity.Resolved != nil {
		aj.Resolved = &resolvedRefJSON{
			Name:   s.Activity.Resolved.Name,
			Line:   s.Activity.Resolved.Line,
			Column: s.Activity.Resolved.Column,
		}
	}
	return json.Marshal(aj)
}

func marshalWorkflowCall(s *WorkflowCall) (json.RawMessage, error) {
	wj := workflowCallJSON{
		Type:    "workflowCall",
		Line:    s.Line,
		Column:  s.Column,
		Mode:    workflowCallModeString(s.Mode),
		Name:    s.Workflow.Name,
		Args:    s.Args,
		Result:  s.Result,
		Options: marshalOptionsBlock(s.Options),
	}
	if s.Workflow.Resolved != nil {
		wj.Resolved = &resolvedRefJSON{
			Name:   s.Workflow.Resolved.Name,
			Line:   s.Workflow.Resolved.Line,
			Column: s.Workflow.Resolved.Column,
		}
	}
	return json.Marshal(wj)
}

func marshalNexusCall(s *NexusCall) (json.RawMessage, error) {
	nj := nexusCallJSON{
		Type:      "nexusCall",
		Line:      s.Line,
		Column:    s.Column,
		Detach:    s.Detach,
		Endpoint:  s.Endpoint.Name,
		Service:   s.Service.Name,
		Operation: s.Operation.Name,
		Args:      s.Args,
		Result:    s.Result,
		Options:   marshalOptionsBlock(s.Options),
	}
	if s.Endpoint.Resolved != nil {
		nj.ResolvedEndpoint = &resolvedRefJSON{Name: s.Endpoint.Resolved.EndpointName, Line: s.Endpoint.Resolved.Line, Column: s.Endpoint.Resolved.Column}
		nj.ResolvedEndpointNamespace = s.Endpoint.Resolved.Namespace
	}
	if s.Service.Resolved != nil {
		nj.ResolvedService = &resolvedRefJSON{Name: s.Service.Resolved.Name, Line: s.Service.Resolved.Line, Column: s.Service.Resolved.Column}
	}
	if s.Operation.Resolved != nil {
		nj.ResolvedOperation = &resolvedRefJSON{Name: s.Operation.Resolved.Name, Line: s.Operation.Resolved.Line, Column: s.Operation.Resolved.Column}
	}
	return json.Marshal(nj)
}

func marshalAwaitStmt(s *AwaitStmt) (json.RawMessage, error) {
	return json.Marshal(awaitStmtJSON{
		Type:   "await",
		Line:   s.Line,
		Column: s.Column,
		Target: marshalAsyncTarget(s.Target),
	})
}

func marshalAwaitAllBlock(s *AwaitAllBlock) (json.RawMessage, error) {
	body, err := marshalStatements(s.Body)
	if err != nil {
		return nil, err
	}
	return json.Marshal(awaitAllBlockJSON{
		Type:   "awaitAll",
		Line:   s.Line,
		Column: s.Column,
		Body:   body,
	})
}

func marshalAwaitOneBlock(s *AwaitOneBlock) (json.RawMessage, error) {
	cases := make([]awaitOneCaseJSON, 0, len(s.Cases))
	for _, c := range s.Cases {
		caseBody, err := marshalStatements(c.Body)
		if err != nil {
			return nil, err
		}
		cj := awaitOneCaseJSON{
			Line:   c.Line,
			Column: c.Column,
			Body:   caseBody,
		}
		if c.AwaitAll != nil {
			data, err := marshalStatement(c.AwaitAll)
			if err != nil {
				return nil, err
			}
			cj.AwaitAll = data
		} else if c.Target != nil {
			t := marshalAsyncTarget(c.Target)
			cj.Target = &t
		}
		cases = append(cases, cj)
	}
	return json.Marshal(awaitOneBlockJSON{
		Type:   "awaitOne",
		Line:   s.Line,
		Column: s.Column,
		Cases:  cases,
	})
}

func marshalSwitchBlock(s *SwitchBlock) (json.RawMessage, error) {
	cases := make([]switchCaseJSON, 0, len(s.Cases))
	for _, c := range s.Cases {
		caseBody, err := marshalStatements(c.Body)
		if err != nil {
			return nil, err
		}
		cases = append(cases, switchCaseJSON{
			Line:   c.Line,
			Column: c.Column,
			Value:  c.Value,
			Body:   caseBody,
		})
	}
	defaultBody, err := marshalStatements(s.Default)
	if err != nil {
		return nil, err
	}
	return json.Marshal(switchBlockJSON{
		Type:    "switch",
		Line:    s.Line,
		Column:  s.Column,
		Expr:    s.Expr,
		Cases:   cases,
		Default: defaultBody,
	})
}

func marshalIfStmt(s *IfStmt) (json.RawMessage, error) {
	body, err := marshalStatements(s.Body)
	if err != nil {
		return nil, err
	}
	elseBody, err := marshalStatements(s.ElseBody)
	if err != nil {
		return nil, err
	}
	return json.Marshal(ifStmtJSON{
		Type:      "if",
		Line:      s.Line,
		Column:    s.Column,
		Condition: s.Condition,
		Body:      body,
		ElseBody:  elseBody,
	})
}

func marshalForStmt(s *ForStmt) (json.RawMessage, error) {
	body, err := marshalStatements(s.Body)
	if err != nil {
		return nil, err
	}
	return json.Marshal(forStmtJSON{
		Type:      "for",
		Line:      s.Line,
		Column:    s.Column,
		Variant:   forVariantString(s.Variant),
		Condition: s.Condition,
		Variable:  s.Variable,
		Iterable:  s.Iterable,
		Body:      body,
	})
}

func marshalReturnStmt(s *ReturnStmt) (json.RawMessage, error) {
	return json.Marshal(returnStmtJSON{
		Type:   "return",
		Line:   s.Line,
		Column: s.Column,
		Value:  s.Value,
	})
}

func marshalCloseStmt(s *CloseStmt) (json.RawMessage, error) {
	return json.Marshal(closeStmtJSON{
		Type:   "close",
		Line:   s.Line,
		Column: s.Column,
		Reason: closeReasonString(s.Reason),
		Args:   s.Args,
	})
}

func marshalBreakStmt(s *BreakStmt) (json.RawMessage, error) {
	return json.Marshal(breakStmtJSON{Type: "break", Line: s.Line, Column: s.Column})
}

func marshalContinueStmt(s *ContinueStmt) (json.RawMessage, error) {
	return json.Marshal(continueStmtJSON{Type: "continue", Line: s.Line, Column: s.Column})
}

func marshalRawStmt(s *RawStmt) (json.RawMessage, error) {
	return json.Marshal(rawStmtJSON{Type: "raw", Line: s.Line, Column: s.Column, Text: s.Text})
}

func marshalComment(s *Comment) (json.RawMessage, error) {
	return json.Marshal(commentJSON{Type: "comment", Line: s.Line, Column: s.Column, Text: s.Text})
}

func marshalPromiseStmt(s *PromiseStmt) (json.RawMessage, error) {
	return json.Marshal(promiseStmtJSON{
		Type:   "promise",
		Line:   s.Line,
		Column: s.Column,
		Name:   s.Name,
		Target: marshalAsyncTarget(s.Target),
	})
}

func marshalSetStmt(s *SetStmt) (json.RawMessage, error) {
	return json.Marshal(setStmtJSON{Type: "set", Line: s.Line, Column: s.Column, Name: s.Condition.Name})
}

func marshalUnsetStmt(s *UnsetStmt) (json.RawMessage, error) {
	return json.Marshal(unsetStmtJSON{Type: "unset", Line: s.Line, Column: s.Column, Name: s.Condition.Name})
}

func workflowCallModeString(mode WorkflowCallMode) string {
	switch mode {
	case CallChild:
		return "child"
	case CallDetach:
		return "detach"
	default:
		return "child"
	}
}

func forVariantString(v ForVariant) string {
	switch v {
	case ForInfinite:
		return "infinite"
	case ForConditional:
		return "conditional"
	case ForIteration:
		return "iteration"
	default:
		return "infinite"
	}
}

// Statement JSON types
type activityCallJSON struct {
	Type     string            `json:"type"`
	Line     int               `json:"line"`
	Column   int               `json:"column"`
	Name     string            `json:"name"`
	Args     string            `json:"args"`
	Result   string            `json:"result,omitempty"`
	Options  *OptionsBlockJSON `json:"options,omitempty"`
	Resolved *resolvedRefJSON  `json:"resolved,omitempty"`
}

type workflowCallJSON struct {
	Type     string            `json:"type"`
	Line     int               `json:"line"`
	Column   int               `json:"column"`
	Mode     string            `json:"mode"`
	Name     string            `json:"name"`
	Args     string            `json:"args"`
	Result   string            `json:"result,omitempty"`
	Options  *OptionsBlockJSON `json:"options,omitempty"`
	Resolved *resolvedRefJSON  `json:"resolved,omitempty"`
}

func closeReasonString(r CloseReason) string {
	switch r {
	case CloseComplete:
		return "complete"
	case CloseFailWorkflow:
		return "fail"
	case CloseContinueAsNew:
		return "continue_as_new"
	default:
		return "complete"
	}
}

// asyncTargetJSON is a discriminated union for async target JSON serialization.
// Each kind populates exactly one of the per-kind fields.
type asyncTargetJSON struct {
	Kind     string              `json:"kind"`
	Timer    *timerTargetJSON    `json:"timer,omitempty"`
	Signal   *signalTargetJSON   `json:"signal,omitempty"`
	Update   *updateTargetJSON   `json:"update,omitempty"`
	Activity *activityTargetJSON `json:"activity,omitempty"`
	Workflow *workflowTargetJSON `json:"workflow,omitempty"`
	Nexus    *nexusTargetJSON    `json:"nexus,omitempty"`
	Ident    *identTargetJSON    `json:"ident,omitempty"`
}

type timerTargetJSON struct {
	Duration string `json:"duration"`
}

type signalTargetJSON struct {
	Name   string `json:"name"`
	Params string `json:"params,omitempty"`
}

type updateTargetJSON struct {
	Name   string `json:"name"`
	Params string `json:"params,omitempty"`
}

type activityTargetJSON struct {
	Name     string           `json:"name"`
	Args     string           `json:"args,omitempty"`
	Result   string           `json:"result,omitempty"`
	Resolved *resolvedRefJSON `json:"resolved,omitempty"`
}

type workflowTargetJSON struct {
	Name     string           `json:"name"`
	Mode     string           `json:"mode"`
	Args     string           `json:"args,omitempty"`
	Result   string           `json:"result,omitempty"`
	Resolved *resolvedRefJSON `json:"resolved,omitempty"`
}

type nexusTargetJSON struct {
	Endpoint                      string           `json:"endpoint"`
	Service                       string           `json:"service"`
	Operation                     string           `json:"operation"`
	Args                          string           `json:"args,omitempty"`
	Result                        string           `json:"result,omitempty"`
	Detach                        bool             `json:"detach,omitempty"`
	ResolvedEndpoint              *resolvedRefJSON `json:"resolvedEndpoint,omitempty"`
	ResolvedEndpointNamespace     string           `json:"resolvedEndpointNamespace,omitempty"`
	ResolvedService               *resolvedRefJSON `json:"resolvedService,omitempty"`
	ResolvedOperation             *resolvedRefJSON `json:"resolvedOperation,omitempty"`
}

type identTargetJSON struct {
	Name   string `json:"name"`
	Result string `json:"result,omitempty"`
}

func marshalAsyncTarget(target AsyncTarget) asyncTargetJSON {
	at := asyncTargetJSON{Kind: AsyncTargetKind(target)}
	switch t := target.(type) {
	case *TimerTarget:
		at.Timer = &timerTargetJSON{Duration: t.Duration}
	case *SignalTarget:
		at.Signal = &signalTargetJSON{Name: t.Signal.Name, Params: t.Params}
	case *UpdateTarget:
		at.Update = &updateTargetJSON{Name: t.Update.Name, Params: t.Params}
	case *ActivityTarget:
		aj := &activityTargetJSON{Name: t.Activity.Name, Args: t.Args, Result: t.Result}
		if t.Activity.Resolved != nil {
			aj.Resolved = &resolvedRefJSON{Name: t.Activity.Resolved.Name, Line: t.Activity.Resolved.Line, Column: t.Activity.Resolved.Column}
		}
		at.Activity = aj
	case *WorkflowTarget:
		wj := &workflowTargetJSON{Name: t.Workflow.Name, Mode: workflowCallModeString(t.Mode), Args: t.Args, Result: t.Result}
		if t.Workflow.Resolved != nil {
			wj.Resolved = &resolvedRefJSON{Name: t.Workflow.Resolved.Name, Line: t.Workflow.Resolved.Line, Column: t.Workflow.Resolved.Column}
		}
		at.Workflow = wj
	case *NexusTarget:
		nj := &nexusTargetJSON{
			Endpoint:  t.Endpoint.Name,
			Service:   t.Service.Name,
			Operation: t.Operation.Name,
			Args:      t.Args,
			Result:    t.Result,
			Detach:    t.Detach,
		}
		if t.Endpoint.Resolved != nil {
			nj.ResolvedEndpoint = &resolvedRefJSON{Name: t.Endpoint.Resolved.EndpointName, Line: t.Endpoint.Resolved.Line, Column: t.Endpoint.Resolved.Column}
			nj.ResolvedEndpointNamespace = t.Endpoint.Resolved.Namespace
		}
		if t.Service.Resolved != nil {
			nj.ResolvedService = &resolvedRefJSON{Name: t.Service.Resolved.Name, Line: t.Service.Resolved.Line, Column: t.Service.Resolved.Column}
		}
		if t.Operation.Resolved != nil {
			nj.ResolvedOperation = &resolvedRefJSON{Name: t.Operation.Resolved.Name, Line: t.Operation.Resolved.Line, Column: t.Operation.Resolved.Column}
		}
		at.Nexus = nj
	case *IdentTarget:
		at.Ident = &identTargetJSON{Name: t.Name, Result: t.Result}
	}
	return at
}

type awaitStmtJSON struct {
	Type   string          `json:"type"`
	Line   int             `json:"line"`
	Column int             `json:"column"`
	Target asyncTargetJSON `json:"target"`
}

type awaitAllBlockJSON struct {
	Type   string            `json:"type"`
	Line   int               `json:"line"`
	Column int               `json:"column"`
	Body   []json.RawMessage `json:"body"`
}

type awaitOneCaseJSON struct {
	Line     int               `json:"line"`
	Column   int               `json:"column"`
	Target   *asyncTargetJSON  `json:"target,omitempty"`
	AwaitAll json.RawMessage   `json:"awaitAll,omitempty"`
	Body     []json.RawMessage `json:"body"`
}

type awaitOneBlockJSON struct {
	Type   string             `json:"type"`
	Line   int                `json:"line"`
	Column int                `json:"column"`
	Cases  []awaitOneCaseJSON `json:"cases"`
}

type switchCaseJSON struct {
	Line   int               `json:"line"`
	Column int               `json:"column"`
	Value  string            `json:"value"`
	Body   []json.RawMessage `json:"body"`
}

type switchBlockJSON struct {
	Type    string            `json:"type"`
	Line    int               `json:"line"`
	Column  int               `json:"column"`
	Expr    string            `json:"expr"`
	Cases   []switchCaseJSON  `json:"cases"`
	Default []json.RawMessage `json:"default,omitempty"`
}

type ifStmtJSON struct {
	Type      string            `json:"type"`
	Line      int               `json:"line"`
	Column    int               `json:"column"`
	Condition string            `json:"condition"`
	Body      []json.RawMessage `json:"body"`
	ElseBody  []json.RawMessage `json:"elseBody,omitempty"`
}

type forStmtJSON struct {
	Type      string            `json:"type"`
	Line      int               `json:"line"`
	Column    int               `json:"column"`
	Variant   string            `json:"variant"`
	Condition string            `json:"condition,omitempty"`
	Variable  string            `json:"variable,omitempty"`
	Iterable  string            `json:"iterable,omitempty"`
	Body      []json.RawMessage `json:"body"`
}

type returnStmtJSON struct {
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Value  string `json:"value,omitempty"`
}

type closeStmtJSON struct {
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Reason string `json:"reason"`
	Args   string `json:"args,omitempty"`
}

type breakStmtJSON struct {
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

type continueStmtJSON struct {
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

type rawStmtJSON struct{
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Text   string `json:"text"`
}

type commentJSON struct {
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Text   string `json:"text"`
}

type promiseStmtJSON struct {
	Type   string          `json:"type"`
	Line   int             `json:"line"`
	Column int             `json:"column"`
	Name   string          `json:"name"`
	Target asyncTargetJSON `json:"target"`
}

type setStmtJSON struct {
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Name   string `json:"name"`
}

type unsetStmtJSON struct {
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Name   string `json:"name"`
}

type nexusCallJSON struct {
	Type      string            `json:"type"`
	Line      int               `json:"line"`
	Column    int               `json:"column"`
	Detach    bool              `json:"detach,omitempty"`
	Endpoint  string            `json:"endpoint"`
	Service   string            `json:"service"`
	Operation string            `json:"operation"`
	Args      string            `json:"args"`
	Result    string            `json:"result,omitempty"`
	Options   *OptionsBlockJSON `json:"options,omitempty"`
	// Resolution links
	ResolvedEndpoint          *resolvedRefJSON `json:"resolvedEndpoint,omitempty"`
	ResolvedEndpointNamespace string           `json:"resolvedEndpointNamespace,omitempty"`
	ResolvedService           *resolvedRefJSON `json:"resolvedService,omitempty"`
	ResolvedOperation         *resolvedRefJSON `json:"resolvedOperation,omitempty"`
}

// NexusOperationJSON is the JSON representation of a nexus operation.
type NexusOperationJSON struct {
	OpType       string            `json:"opType"`
	Line         int               `json:"line"`
	Column       int               `json:"column"`
	Name         string            `json:"name"`
	WorkflowName string            `json:"workflowName,omitempty"`
	Params       string            `json:"params,omitempty"`
	ReturnType   string            `json:"returnType,omitempty"`
	Body         []json.RawMessage `json:"body,omitempty"`
}

// NexusServiceDefJSON is the JSON representation of NexusServiceDef.
type NexusServiceDefJSON struct {
	Type       string                `json:"type"`
	Line       int                   `json:"line"`
	Column     int                   `json:"column"`
	SourceFile string                `json:"sourceFile,omitempty"`
	Name       string                `json:"name"`
	Operations []*NexusOperationJSON `json:"operations,omitempty"`
}

func (n *NexusServiceDef) MarshalJSON() ([]byte, error) {
	nj := NexusServiceDefJSON{
		Type:       "nexusServiceDef",
		Line:       n.Line,
		Column:     n.Column,
		SourceFile: n.SourceFile,
		Name:       n.Name,
	}
	for _, op := range n.Operations {
		opj := &NexusOperationJSON{
			Line:         op.Line,
			Column:       op.Column,
			Name:         op.Name,
			WorkflowName: op.Workflow.Name,
			Params:       op.Params,
			ReturnType:   op.ReturnType,
		}
		if op.OpType == NexusOpAsync {
			opj.OpType = "async"
		} else {
			opj.OpType = "sync"
		}
		var err error
		if opj.Body, err = marshalStatements(op.Body); err != nil {
			return nil, err
		}
		nj.Operations = append(nj.Operations, opj)
	}
	return json.Marshal(nj)
}
