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
		wj.Signals = append(wj.Signals, &SignalDeclJSON{
			Type:   "signalDecl",
			Line:   s.Line,
			Column: s.Column,
			Name:   s.Name,
			Params: s.Params,
		})
	}
	for _, q := range w.Queries {
		wj.Queries = append(wj.Queries, &QueryDeclJSON{
			Type:       "queryDecl",
			Line:       q.Line,
			Column:     q.Column,
			Name:       q.Name,
			Params:     q.Params,
			ReturnType: q.ReturnType,
		})
	}
	for _, u := range w.Updates {
		wj.Updates = append(wj.Updates, &UpdateDeclJSON{
			Type:       "updateDecl",
			Line:       u.Line,
			Column:     u.Column,
			Name:       u.Name,
			Params:     u.Params,
			ReturnType: u.ReturnType,
		})
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
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Name   string `json:"name"`
	Params string `json:"params"`
}

type QueryDeclJSON struct {
	Type       string `json:"type"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Name       string `json:"name"`
	Params     string `json:"params"`
	ReturnType string `json:"returnType,omitempty"`
}

type UpdateDeclJSON struct {
	Type       string `json:"type"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Name       string `json:"name"`
	Params     string `json:"params"`
	ReturnType string `json:"returnType,omitempty"`
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
	case *AwaitStmt:
		targets := make([]awaitTargetJSON, 0, len(s.Targets))
		for _, t := range s.Targets {
			targets = append(targets, awaitTargetJSON{
				Kind: t.Kind,
				Name: t.Name,
				Args: t.Args,
			})
		}
		return json.Marshal(awaitStmtJSON{
			Type:    "await",
			Line:    s.Line,
			Column:  s.Column,
			Targets: targets,
		})
	case *ParallelBlock:
		body := make([]json.RawMessage, 0, len(s.Body))
		for _, stmt := range s.Body {
			data, err := marshalStatement(stmt)
			if err != nil {
				return nil, err
			}
			body = append(body, data)
		}
		return json.Marshal(parallelBlockJSON{
			Type:   "parallel",
			Line:   s.Line,
			Column: s.Column,
			Body:   body,
		})
	case *SelectBlock:
		cases := make([]selectCaseJSON, 0, len(s.Cases))
		for _, c := range s.Cases {
			caseBody := make([]json.RawMessage, 0, len(c.Body))
			for _, stmt := range c.Body {
				data, err := marshalStatement(stmt)
				if err != nil {
					return nil, err
				}
				caseBody = append(caseBody, data)
			}
			cases = append(cases, selectCaseJSON{
				Kind:              c.CaseKind(),
				WorkflowMode:      workflowCallModeString(c.WorkflowMode),
				WorkflowNamespace: c.WorkflowNamespace,
				WorkflowName:      c.WorkflowName,
				WorkflowArgs:      c.WorkflowArgs,
				WorkflowResult:    c.WorkflowResult,
				ActivityName:      c.ActivityName,
				ActivityArgs:      c.ActivityArgs,
				ActivityResult:    c.ActivityResult,
				SignalName:        c.SignalName,
				SignalArgs:        c.SignalArgs,
				UpdateName:        c.UpdateName,
				UpdateArgs:        c.UpdateArgs,
				TimerDuration:     c.TimerDuration,
				Body:              caseBody,
			})
		}
		return json.Marshal(selectBlockJSON{
			Type:   "select",
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

type awaitTargetJSON struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	Args string `json:"args,omitempty"`
}

type awaitStmtJSON struct {
	Type    string            `json:"type"`
	Line    int               `json:"line"`
	Column  int               `json:"column"`
	Targets []awaitTargetJSON `json:"targets"`
}

type parallelBlockJSON struct {
	Type   string            `json:"type"`
	Line   int               `json:"line"`
	Column int               `json:"column"`
	Body   []json.RawMessage `json:"body"`
}

type selectCaseJSON struct {
	Kind              string            `json:"kind"`
	WorkflowMode      string            `json:"workflowMode,omitempty"`
	WorkflowNamespace string            `json:"workflowNamespace,omitempty"`
	WorkflowName      string            `json:"workflowName,omitempty"`
	WorkflowArgs      string            `json:"workflowArgs,omitempty"`
	WorkflowResult    string            `json:"workflowResult,omitempty"`
	ActivityName      string            `json:"activityName,omitempty"`
	ActivityArgs      string            `json:"activityArgs,omitempty"`
	ActivityResult    string            `json:"activityResult,omitempty"`
	SignalName        string            `json:"signalName,omitempty"`
	SignalArgs        string            `json:"signalArgs,omitempty"`
	UpdateName        string            `json:"updateName,omitempty"`
	UpdateArgs        string            `json:"updateArgs,omitempty"`
	TimerDuration     string            `json:"timerDuration,omitempty"`
	Body              []json.RawMessage `json:"body"`
}

type selectBlockJSON struct {
	Type   string           `json:"type"`
	Line   int              `json:"line"`
	Column int              `json:"column"`
	Cases  []selectCaseJSON `json:"cases"`
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
