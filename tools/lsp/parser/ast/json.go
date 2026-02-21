package ast

import "encoding/json"

// JSON serialization with type discriminators for interface types.

// resolvedRefJSON is a lightweight JSON reference to a resolved AST definition.
// Instead of embedding the entire target node, we emit its name and source position.
type resolvedRefJSON struct {
	Name   string `json:"name"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// FileJSON is the JSON-serializable representation of a File.
type FileJSON struct {
	Definitions []json.RawMessage `json:"definitions"`
}

// MarshalJSON implements json.Marshaler for File.
func (f *File) MarshalJSON() ([]byte, error) {
	fj := FileJSON{
		Definitions: make([]json.RawMessage, 0, len(f.Definitions)),
	}
	for _, def := range f.Definitions {
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
		return json.Marshal(def)
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
	Name       string             `json:"name"`
	Params     string             `json:"params"`
	ReturnType string             `json:"returnType,omitempty"`
	State      *StateBlockJSON    `json:"state,omitempty"`
	Signals    []*SignalDeclJSON  `json:"signals,omitempty"`
	Queries    []*QueryDeclJSON   `json:"queries,omitempty"`
	Updates    []*UpdateDeclJSON  `json:"updates,omitempty"`
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
		Name:       w.Name,
		Params:     w.Params,
		ReturnType: w.ReturnType,
		Signals:    make([]*SignalDeclJSON, 0, len(w.Signals)),
		Queries:    make([]*QueryDeclJSON, 0, len(w.Queries)),
		Updates:    make([]*UpdateDeclJSON, 0, len(w.Updates)),
		Body:       make([]json.RawMessage, 0, len(w.Body)),
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
	for _, s := range w.Signals {
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
		wj.Signals = append(wj.Signals, sj)
	}
	for _, q := range w.Queries {
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
		wj.Queries = append(wj.Queries, qj)
	}
	for _, u := range w.Updates {
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
		wj.Updates = append(wj.Updates, uj)
	}
	var err error
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
	Type      string                  `json:"type"`
	Line      int                     `json:"line"`
	Column    int                     `json:"column"`
	Name      string                  `json:"name"`
	Workers   []NamespaceWorkerJSON   `json:"workers,omitempty"`
	Endpoints []NamespaceEndpointJSON `json:"endpoints,omitempty"`
}

func (n *NamespaceDef) MarshalJSON() ([]byte, error) {
	nj := NamespaceDefJSON{
		Type:   "namespaceDef",
		Line:   n.Line,
		Column: n.Column,
		Name:   n.Name,
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
		return json.Marshal(activityCallJSON{
			Type:    "activityCall",
			Line:    s.Line,
			Column:  s.Column,
			Name:    s.Activity.Name,
			Args:    s.Args,
			Result:  s.Result,
			Options: marshalOptionsBlock(s.Options),
		})
	case *WorkflowCall:
		return json.Marshal(workflowCallJSON{
			Type:    "workflowCall",
			Line:    s.Line,
			Column:  s.Column,
			Mode:    workflowCallModeString(s.Mode),
			Name:    s.Workflow.Name,
			Args:    s.Args,
			Result:  s.Result,
			Options: marshalOptionsBlock(s.Options),
		})
	case *AwaitStmt:
		aj := awaitStmtJSON{
			Type:                  "await",
			Line:                  s.Line,
			Column:                s.Column,
			asyncTargetFieldsJSON: marshalAsyncTargetFields(s.Target),
		}
		return json.Marshal(aj)
	case *AwaitAllBlock:
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
	case *AwaitOneBlock:
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
				cj.Kind = "await_all"
				data, err := marshalStatement(c.AwaitAll)
				if err != nil {
					return nil, err
				}
				cj.AwaitAll = data
			} else if c.Target != nil {
				cj.asyncTargetFieldsJSON = marshalAsyncTargetFields(c.Target)
			}
			cases = append(cases, cj)
		}
		return json.Marshal(awaitOneBlockJSON{
			Type:   "awaitOne",
			Line:   s.Line,
			Column: s.Column,
			Cases:  cases,
		})
	case *SwitchBlock:
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
	case *IfStmt:
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
	case *ForStmt:
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
	case *ReturnStmt:
		return json.Marshal(returnStmtJSON{
			Type:   "return",
			Line:   s.Line,
			Column: s.Column,
			Value:  s.Value,
		})
	case *CloseStmt:
		return json.Marshal(closeStmtJSON{
			Type:   "close",
			Line:   s.Line,
			Column: s.Column,
			Reason: s.Reason,
			Args:   s.Args,
		})
	case *BreakStmt:
		return json.Marshal(breakStmtJSON{
			Type:   "break",
			Line:   s.Line,
			Column: s.Column,
		})
	case *ContinueStmt:
		return json.Marshal(continueStmtJSON{
			Type:   "continue",
			Line:   s.Line,
			Column: s.Column,
		})
	case *RawStmt:
		return json.Marshal(rawStmtJSON{
			Type:   "raw",
			Line:   s.Line,
			Column: s.Column,
			Text:   s.Text,
		})
	case *Comment:
		return json.Marshal(commentJSON{
			Type:   "comment",
			Line:   s.Line,
			Column: s.Column,
			Text:   s.Text,
		})
	case *PromiseStmt:
		pj := promiseStmtJSON{
			Type:                  "promise",
			Line:                  s.Line,
			Column:                s.Column,
			Name:                  s.Name,
			asyncTargetFieldsJSON: marshalAsyncTargetFields(s.Target),
		}
		return json.Marshal(pj)
	case *NexusCall:
		nj := nexusCallJSON{
			Type:      "nexusCall",
			Line:      s.Line,
			Column:    s.Column,
			Detach:    s.Detach,
			Endpoint:  s.Endpoint,
			Service:   s.Service,
			Operation: s.Operation,
			Args:      s.Args,
			Result:    s.Result,
			Options:   marshalOptionsBlock(s.Options),
		}
		nj.ResolvedEndpoint, nj.ResolvedService, nj.ResolvedOperation, nj.ResolvedEndpointNamespace =
			marshalNexusResolved(s.ResolvedEndpoint, s.ResolvedEndpointNamespace, s.ResolvedService, s.ResolvedOperation)
		return json.Marshal(nj)
	case *SetStmt:
		return json.Marshal(setStmtJSON{
			Type:   "set",
			Line:   s.Line,
			Column: s.Column,
			Name:   s.Condition.Name,
		})
	case *UnsetStmt:
		return json.Marshal(unsetStmtJSON{
			Type:   "unset",
			Line:   s.Line,
			Column: s.Column,
			Name:   s.Condition.Name,
		})
	default:
		return json.Marshal(stmt)
	}
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
	Type    string            `json:"type"`
	Line    int               `json:"line"`
	Column  int               `json:"column"`
	Name    string            `json:"name"`
	Args    string            `json:"args"`
	Result  string            `json:"result,omitempty"`
	Options *OptionsBlockJSON `json:"options,omitempty"`
}

type workflowCallJSON struct {
	Type    string            `json:"type"`
	Line    int               `json:"line"`
	Column  int               `json:"column"`
	Mode    string            `json:"mode"`
	Name    string            `json:"name"`
	Args    string            `json:"args"`
	Result  string            `json:"result,omitempty"`
	Options *OptionsBlockJSON `json:"options,omitempty"`
}

// marshalNexusResolved converts nexus resolution fields to JSON reference form.
func marshalNexusResolved(ep *NamespaceEndpoint, epNS string, svc *NexusServiceDef, op *NexusOperation) (
	epRef, svcRef, opRef *resolvedRefJSON, ns string,
) {
	if ep != nil {
		epRef = &resolvedRefJSON{Name: ep.EndpointName, Line: ep.Line, Column: ep.Column}
	}
	ns = epNS
	if svc != nil {
		svcRef = &resolvedRefJSON{Name: svc.Name, Line: svc.Line, Column: svc.Column}
	}
	if op != nil {
		opRef = &resolvedRefJSON{Name: op.Name, Line: op.Line, Column: op.Column}
	}
	return
}

// asyncTargetFieldsJSON holds the flat JSON fields for an async target.
// Embedded in awaitStmtJSON, awaitOneCaseJSON, and promiseStmtJSON
// to maintain backward-compatible flat JSON format.
type asyncTargetFieldsJSON struct {
	Kind           string `json:"kind"`
	Timer          string `json:"timer,omitempty"`
	Signal         string `json:"signal,omitempty"`
	SignalParams   string `json:"signalParams,omitempty"`
	Update         string `json:"update,omitempty"`
	UpdateParams   string `json:"updateParams,omitempty"`
	Activity       string `json:"activity,omitempty"`
	ActivityArgs   string `json:"activityArgs,omitempty"`
	ActivityResult string `json:"activityResult,omitempty"`
	Workflow       string `json:"workflow,omitempty"`
	WorkflowMode   string `json:"workflowMode,omitempty"`
	WorkflowArgs   string `json:"workflowArgs,omitempty"`
	WorkflowResult string `json:"workflowResult,omitempty"`
	Nexus          string `json:"nexus,omitempty"`
	NexusService   string `json:"nexusService,omitempty"`
	NexusOperation string `json:"nexusOperation,omitempty"`
	NexusArgs      string `json:"nexusArgs,omitempty"`
	NexusResult    string `json:"nexusResult,omitempty"`
	NexusDetach    bool   `json:"nexusDetach,omitempty"`
	// Nexus resolution links
	NexusResolvedEndpoint          *resolvedRefJSON `json:"nexusResolvedEndpoint,omitempty"`
	NexusResolvedEndpointNamespace string           `json:"nexusResolvedEndpointNamespace,omitempty"`
	NexusResolvedService           *resolvedRefJSON `json:"nexusResolvedService,omitempty"`
	NexusResolvedOperation         *resolvedRefJSON `json:"nexusResolvedOperation,omitempty"`
	Ident                          string           `json:"ident,omitempty"`
	IdentResult                    string           `json:"identResult,omitempty"`
}

func marshalAsyncTargetFields(target AsyncTarget) asyncTargetFieldsJSON {
	f := asyncTargetFieldsJSON{Kind: AsyncTargetKind(target)}
	switch t := target.(type) {
	case *TimerTarget:
		f.Timer = t.Duration
	case *SignalTarget:
		f.Signal = t.Signal.Name
		f.SignalParams = t.Params
	case *UpdateTarget:
		f.Update = t.Update.Name
		f.UpdateParams = t.Params
	case *ActivityTarget:
		f.Activity = t.Activity.Name
		f.ActivityArgs = t.Args
		f.ActivityResult = t.Result
	case *WorkflowTarget:
		f.Workflow = t.Workflow.Name
		f.WorkflowMode = workflowCallModeString(t.Mode)
		f.WorkflowArgs = t.Args
		f.WorkflowResult = t.Result
	case *NexusTarget:
		f.Nexus = t.Endpoint
		f.NexusService = t.Service
		f.NexusOperation = t.Operation
		f.NexusArgs = t.Args
		f.NexusResult = t.Result
		f.NexusDetach = t.Detach
		f.NexusResolvedEndpoint, f.NexusResolvedService, f.NexusResolvedOperation, f.NexusResolvedEndpointNamespace =
			marshalNexusResolved(t.ResolvedEndpoint, t.ResolvedEndpointNamespace, t.ResolvedService, t.ResolvedOperation)
	case *IdentTarget:
		f.Ident = t.Name
		f.IdentResult = t.Result
	}
	return f
}

type awaitStmtJSON struct {
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	asyncTargetFieldsJSON
}

type awaitAllBlockJSON struct {
	Type   string            `json:"type"`
	Line   int               `json:"line"`
	Column int               `json:"column"`
	Body   []json.RawMessage `json:"body"`
}

type awaitOneCaseJSON struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	asyncTargetFieldsJSON
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
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Name   string `json:"name"`
	asyncTargetFieldsJSON
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
	Name       string                `json:"name"`
	Operations []*NexusOperationJSON `json:"operations,omitempty"`
}

func (n *NexusServiceDef) MarshalJSON() ([]byte, error) {
	nj := NexusServiceDefJSON{
		Type:   "nexusServiceDef",
		Line:   n.Line,
		Column: n.Column,
		Name:   n.Name,
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
