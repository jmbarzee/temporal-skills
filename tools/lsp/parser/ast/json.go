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
		for _, stmt := range s.Body {
			data, err := marshalStatement(stmt)
			if err != nil {
				return nil, err
			}
			sj.Body = append(sj.Body, data)
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
		for _, stmt := range q.Body {
			data, err := marshalStatement(stmt)
			if err != nil {
				return nil, err
			}
			qj.Body = append(qj.Body, data)
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
		for _, stmt := range u.Body {
			data, err := marshalStatement(stmt)
			if err != nil {
				return nil, err
			}
			uj.Body = append(uj.Body, data)
		}
		wj.Updates = append(wj.Updates, uj)
	}
	for _, stmt := range w.Body {
		data, err := marshalStatement(stmt)
		if err != nil {
			return nil, err
		}
		wj.Body = append(wj.Body, data)
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
		Body:       make([]json.RawMessage, 0, len(a.Body)),
	}
	for _, stmt := range a.Body {
		data, err := marshalStatement(stmt)
		if err != nil {
			return nil, err
		}
		aj.Body = append(aj.Body, data)
	}
	return json.Marshal(aj)
}

// WorkerRefJSON is the JSON representation of a worker reference.
type WorkerRefJSON struct {
	Name     string           `json:"name"`
	Line     int              `json:"line"`
	Column   int              `json:"column"`
	Resolved *resolvedRefJSON `json:"resolved,omitempty"`
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
		Type:   "workerDef",
		Line:   w.Line,
		Column: w.Column,
		Name:   w.Name,
	}
	for _, ref := range w.Workflows {
		rj := WorkerRefJSON{
			Name:   ref.Name,
			Line:   ref.Line,
			Column: ref.Column,
		}
		if ref.Resolved != nil {
			rj.Resolved = &resolvedRefJSON{Name: ref.Name, Line: ref.Resolved.NodeLine(), Column: ref.Resolved.NodeColumn()}
		}
		wj.Workflows = append(wj.Workflows, rj)
	}
	for _, ref := range w.Activities {
		rj := WorkerRefJSON{
			Name:   ref.Name,
			Line:   ref.Line,
			Column: ref.Column,
		}
		if ref.Resolved != nil {
			rj.Resolved = &resolvedRefJSON{Name: ref.Name, Line: ref.Resolved.NodeLine(), Column: ref.Resolved.NodeColumn()}
		}
		wj.Activities = append(wj.Activities, rj)
	}
	for _, ref := range w.Services {
		rj := WorkerRefJSON{
			Name:   ref.Name,
			Line:   ref.Line,
			Column: ref.Column,
		}
		if ref.Resolved != nil {
			rj.Resolved = &resolvedRefJSON{Name: ref.Name, Line: ref.Resolved.NodeLine(), Column: ref.Resolved.NodeColumn()}
		}
		wj.Services = append(wj.Services, rj)
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
			WorkerName: w.WorkerName,
			Line:       w.Line,
			Column:     w.Column,
			Options:    marshalOptionsBlock(w.Options),
		}
		if w.ResolvedWorker != nil {
			wj.ResolvedWorker = &resolvedRefJSON{Name: w.ResolvedWorker.Name, Line: w.ResolvedWorker.Line, Column: w.ResolvedWorker.Column}
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
			Name:    s.Name,
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
			Name:    s.Name,
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
		body := make([]json.RawMessage, 0, len(s.Body))
		for _, stmt := range s.Body {
			data, err := marshalStatement(stmt)
			if err != nil {
				return nil, err
			}
			body = append(body, data)
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
			caseBody := make([]json.RawMessage, 0, len(c.Body))
			for _, stmt := range c.Body {
				data, err := marshalStatement(stmt)
				if err != nil {
					return nil, err
				}
				caseBody = append(caseBody, data)
			}
			cj := awaitOneCaseJSON{
				Line: c.Line,
				Column: c.Column,
				Body: caseBody,
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
			caseBody := make([]json.RawMessage, 0, len(c.Body))
			for _, stmt := range c.Body {
				data, err := marshalStatement(stmt)
				if err != nil {
					return nil, err
				}
				caseBody = append(caseBody, data)
			}
			cases = append(cases, switchCaseJSON{
				Line:   c.Line,
				Column: c.Column,
				Value:  c.Value,
				Body:   caseBody,
			})
		}
		var defaultBody []json.RawMessage
		for _, stmt := range s.Default {
			data, err := marshalStatement(stmt)
			if err != nil {
				return nil, err
			}
			defaultBody = append(defaultBody, data)
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
		body := make([]json.RawMessage, 0, len(s.Body))
		for _, stmt := range s.Body {
			data, err := marshalStatement(stmt)
			if err != nil {
				return nil, err
			}
			body = append(body, data)
		}
		var elseBody []json.RawMessage
		for _, stmt := range s.ElseBody {
			data, err := marshalStatement(stmt)
			if err != nil {
				return nil, err
			}
			elseBody = append(elseBody, data)
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
		body := make([]json.RawMessage, 0, len(s.Body))
		for _, stmt := range s.Body {
			data, err := marshalStatement(stmt)
			if err != nil {
				return nil, err
			}
			body = append(body, data)
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
		if s.ResolvedEndpoint != nil {
			nj.ResolvedEndpoint = &resolvedRefJSON{Name: s.ResolvedEndpoint.EndpointName, Line: s.ResolvedEndpoint.Line, Column: s.ResolvedEndpoint.Column}
		}
		if s.ResolvedEndpointNamespace != "" {
			nj.ResolvedEndpointNamespace = s.ResolvedEndpointNamespace
		}
		if s.ResolvedService != nil {
			nj.ResolvedService = &resolvedRefJSON{Name: s.ResolvedService.Name, Line: s.ResolvedService.Line, Column: s.ResolvedService.Column}
		}
		if s.ResolvedOperation != nil {
			nj.ResolvedOperation = &resolvedRefJSON{Name: s.ResolvedOperation.Name, Line: s.ResolvedOperation.Line, Column: s.ResolvedOperation.Column}
		}
		return json.Marshal(nj)
	case *SetStmt:
		return json.Marshal(setStmtJSON{
			Type:   "set",
			Line:   s.Line,
			Column: s.Column,
			Name:   s.Name,
		})
	case *UnsetStmt:
		return json.Marshal(unsetStmtJSON{
			Type:   "unset",
			Line:   s.Line,
			Column: s.Column,
			Name:   s.Name,
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
		f.Signal = t.Name
		f.SignalParams = t.Params
	case *UpdateTarget:
		f.Update = t.Name
		f.UpdateParams = t.Params
	case *ActivityTarget:
		f.Activity = t.Name
		f.ActivityArgs = t.Args
		f.ActivityResult = t.Result
	case *WorkflowTarget:
		f.Workflow = t.Name
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
		if t.ResolvedEndpoint != nil {
			f.NexusResolvedEndpoint = &resolvedRefJSON{Name: t.ResolvedEndpoint.EndpointName, Line: t.ResolvedEndpoint.Line, Column: t.ResolvedEndpoint.Column}
		}
		if t.ResolvedEndpointNamespace != "" {
			f.NexusResolvedEndpointNamespace = t.ResolvedEndpointNamespace
		}
		if t.ResolvedService != nil {
			f.NexusResolvedService = &resolvedRefJSON{Name: t.ResolvedService.Name, Line: t.ResolvedService.Line, Column: t.ResolvedService.Column}
		}
		if t.ResolvedOperation != nil {
			f.NexusResolvedOperation = &resolvedRefJSON{Name: t.ResolvedOperation.Name, Line: t.ResolvedOperation.Line, Column: t.ResolvedOperation.Column}
		}
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
			WorkflowName: op.WorkflowName,
			Params:       op.Params,
			ReturnType:   op.ReturnType,
		}
		if op.OpType == NexusOpAsync {
			opj.OpType = "async"
		} else {
			opj.OpType = "sync"
		}
		for _, stmt := range op.Body {
			data, err := marshalStatement(stmt)
			if err != nil {
				return nil, err
			}
			opj.Body = append(opj.Body, data)
		}
		nj.Operations = append(nj.Operations, opj)
	}
	return json.Marshal(nj)
}
