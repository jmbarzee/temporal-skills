package ast

import "encoding/json"

// JSON serialization with type discriminators for interface types.

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
	default:
		return json.Marshal(def)
	}
}

// WorkflowDefJSON is the JSON representation of WorkflowDef.
type WorkflowDefJSON struct {
	Type       string             `json:"type"`
	Line       int                `json:"line"`
	Column     int                `json:"column"`
	Name       string             `json:"name"`
	Params     string             `json:"params"`
	ReturnType string             `json:"returnType,omitempty"`
	Options    string             `json:"options,omitempty"`
	Signals    []*SignalDeclJSON  `json:"signals,omitempty"`
	Queries    []*QueryDeclJSON   `json:"queries,omitempty"`
	Updates    []*UpdateDeclJSON  `json:"updates,omitempty"`
	Body       []json.RawMessage  `json:"body"`
}

func (w *WorkflowDef) MarshalJSON() ([]byte, error) {
	wj := WorkflowDefJSON{
		Type:       "workflowDef",
		Line:       w.Line,
		Column:     w.Column,
		Name:       w.Name,
		Params:     w.Params,
		ReturnType: w.ReturnType,
		Options:    w.Options,
		Signals:    make([]*SignalDeclJSON, 0, len(w.Signals)),
		Queries:    make([]*QueryDeclJSON, 0, len(w.Queries)),
		Updates:    make([]*UpdateDeclJSON, 0, len(w.Updates)),
		Body:       make([]json.RawMessage, 0, len(w.Body)),
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
	Options    string            `json:"options,omitempty"`
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
		Options:    a.Options,
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
			Options: s.Options,
		})
	case *WorkflowCall:
		return json.Marshal(workflowCallJSON{
			Type:      "workflowCall",
			Line:      s.Line,
			Column:    s.Column,
			Mode:      workflowCallModeString(s.Mode),
			Namespace: s.Namespace,
			Name:      s.Name,
			Args:      s.Args,
			Result:    s.Result,
			Options:   s.Options,
		})
	case *TimerStmt:
		return json.Marshal(timerStmtJSON{
			Type:     "timer",
			Line:     s.Line,
			Column:   s.Column,
			Duration: s.Duration,
		})
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
			var awaitAllData json.RawMessage
			if c.AwaitAll != nil {
				data, err := marshalStatement(c.AwaitAll)
				if err != nil {
					return nil, err
				}
				awaitAllData = data
			}
			cases = append(cases, awaitOneCaseJSON{
				Kind:          c.CaseKind(),
				TimerDuration: c.TimerDuration,
				AwaitAll:      awaitAllData,
				Body:          caseBody,
			})
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
				Value: c.Value,
				Body:  caseBody,
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
	case *ContinueAsNewStmt:
		return json.Marshal(continueAsNewStmtJSON{
			Type:   "continueAsNew",
			Line:   s.Line,
			Column: s.Column,
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
	case *HintStmt:
		return json.Marshal(hintStmtJSON{
			Type:   "hint",
			Line:   s.Line,
			Column: s.Column,
			Kind:   s.Kind,
			Name:   s.Name,
		})
	case *Comment:
		return json.Marshal(commentJSON{
			Type:   "comment",
			Line:   s.Line,
			Column: s.Column,
			Text:   s.Text,
		})
	default:
		return json.Marshal(stmt)
	}
}

func workflowCallModeString(mode WorkflowCallMode) string {
	switch mode {
	case CallChild:
		return "child"
	case CallSpawn:
		return "spawn"
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
	Type    string `json:"type"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Name    string `json:"name"`
	Args    string `json:"args"`
	Result  string `json:"result,omitempty"`
	Options string `json:"options,omitempty"`
}

type workflowCallJSON struct {
	Type      string `json:"type"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	Mode      string `json:"mode"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
	Args      string `json:"args"`
	Result    string `json:"result,omitempty"`
	Options   string `json:"options,omitempty"`
}

type timerStmtJSON struct {
	Type     string `json:"type"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Duration string `json:"duration"`
}

type awaitAllBlockJSON struct {
	Type   string            `json:"type"`
	Line   int               `json:"line"`
	Column int               `json:"column"`
	Body   []json.RawMessage `json:"body"`
}

type awaitOneCaseJSON struct {
	Kind          string          `json:"kind"`
	TimerDuration string          `json:"timerDuration,omitempty"`
	AwaitAll      json.RawMessage `json:"awaitAll,omitempty"`
	Body          []json.RawMessage `json:"body"`
}

type awaitOneBlockJSON struct {
	Type   string             `json:"type"`
	Line   int                `json:"line"`
	Column int                `json:"column"`
	Cases  []awaitOneCaseJSON `json:"cases"`
}

type switchCaseJSON struct {
	Value string            `json:"value"`
	Body  []json.RawMessage `json:"body"`
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

type continueAsNewStmtJSON struct {
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Args   string `json:"args"`
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

type hintStmtJSON struct {
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Kind   string `json:"kind"`
	Name   string `json:"name"`
}

type rawStmtJSON struct {
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
