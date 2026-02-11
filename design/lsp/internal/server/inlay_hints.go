package server

import (
	"strings"

	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// InlayHint types (LSP 3.17 - not in protocol_3_16)
type InlayHintParams struct {
	TextDocument protocol.TextDocumentIdentifier `json:"textDocument"`
	Range        protocol.Range                  `json:"range"`
}

type InlayHint struct {
	Position     protocol.Position     `json:"position"`
	Label        InlayHintLabelPart    `json:"label"`
	Kind         *InlayHintKind        `json:"kind,omitempty"`
	PaddingLeft  *bool                 `json:"paddingLeft,omitempty"`
	PaddingRight *bool                 `json:"paddingRight,omitempty"`
}

type InlayHintLabelPart struct {
	Value string `json:"value"`
}

type InlayHintKind uint32

const (
	InlayHintKindType InlayHintKind = 1
	InlayHintKindParameter InlayHintKind = 2
)

func inlayHintHandler(store *DocumentStore) func(*glsp.Context, *InlayHintParams) ([]InlayHint, error) {
	return func(context *glsp.Context, params *InlayHintParams) ([]InlayHint, error) {
		doc := store.Get(params.TextDocument.URI)
		if doc == nil || doc.File == nil {
			return nil, nil
		}

		var hints []InlayHint

		// Collect hints from all definitions
		for _, def := range doc.File.Definitions {
			if wf, ok := def.(*ast.WorkflowDef); ok {
				hints = append(hints, collectWorkflowHints(wf, params.Range)...)
			}
		}

		return hints, nil
	}
}

// collectWorkflowHints walks a workflow and collects all inlay hints
func collectWorkflowHints(wf *ast.WorkflowDef, visibleRange protocol.Range) []InlayHint {
	var hints []InlayHint

	// Walk all statements in workflow body
	hints = append(hints, collectStatementHints(wf.Body, visibleRange)...)

	// Walk signal/query/update handlers
	for _, sig := range wf.Signals {
		hints = append(hints, collectStatementHints(sig.Body, visibleRange)...)
	}
	for _, upd := range wf.Updates {
		hints = append(hints, collectStatementHints(upd.Body, visibleRange)...)
	}

	return hints
}

// collectStatementHints walks statements and collects hints
func collectStatementHints(stmts []ast.Statement, visibleRange protocol.Range) []InlayHint {
	var hints []InlayHint

	for _, stmt := range stmts {
		hints = append(hints, collectHintsFromStatement(stmt, visibleRange)...)
	}

	return hints
}

// collectHintsFromStatement collects hints from a single statement
func collectHintsFromStatement(stmt ast.Statement, visibleRange protocol.Range) []InlayHint {
	var hints []InlayHint

	// Check if statement is in visible range
	if !isInRange(stmt.NodeLine(), visibleRange) {
		return nil
	}

	switch s := stmt.(type) {
	case *ast.ActivityCall:
		hints = append(hints, collectActivityHints(s)...)
	case *ast.WorkflowCall:
		hints = append(hints, collectWorkflowCallHints(s)...)
	case *ast.AwaitStmt:
		hints = append(hints, collectAwaitHints(s)...)
	case *ast.AwaitOneBlock:
		for _, c := range s.Cases {
			hints = append(hints, collectCaseHints(c, visibleRange)...)
		}
	case *ast.AwaitAllBlock:
		hints = append(hints, collectStatementHints(s.Body, visibleRange)...)
	case *ast.IfStmt:
		hints = append(hints, collectStatementHints(s.Body, visibleRange)...)
		hints = append(hints, collectStatementHints(s.ElseBody, visibleRange)...)
	case *ast.ForStmt:
		hints = append(hints, collectStatementHints(s.Body, visibleRange)...)
	case *ast.SwitchBlock:
		for _, c := range s.Cases {
			hints = append(hints, collectStatementHints(c.Body, visibleRange)...)
		}
		hints = append(hints, collectStatementHints(s.Default, visibleRange)...)
	}

	return hints
}

// collectActivityHints shows parameter names for activity calls
func collectActivityHints(call *ast.ActivityCall) []InlayHint {
	if call.Resolved == nil || call.Args == "" {
		return nil
	}

	var hints []InlayHint

	// Parse parameters from the resolved activity definition
	params := parseParams(call.Resolved.Params)
	args := parseArgs(call.Args)

	// Show parameter name before each argument
	for i, arg := range args {
		if i >= len(params) {
			break // More args than params (error, but continue)
		}

		// Position hint at start of argument within the args string
		// Args start after "activity Name("
		pos := protocol.Position{
			Line:      uint32(call.Line - 1), // LSP is 0-indexed
			Character: uint32(call.Column + len("activity ") + len(call.Name) + len("(") + arg.Column - 1),
		}

		hints = append(hints, InlayHint{
			Position:     pos,
			Label:        InlayHintLabelPart{Value: params[i].Name + ": "},
			Kind:         ptrToKind(InlayHintKindParameter),
			PaddingLeft:  boolPtr(false),
			PaddingRight: boolPtr(true),
		})
	}

	// Show return type if result is bound
	if call.Result != "" && call.Resolved.ReturnType != "" {
		// Position after the result variable
		pos := protocol.Position{
			Line:      uint32(call.Line - 1),
			Character: uint32(call.Column + len(call.Result)),
		}

		hints = append(hints, InlayHint{
			Position:     pos,
			Label:        InlayHintLabelPart{Value: ": " + call.Resolved.ReturnType},
			Kind:         ptrToKind(InlayHintKindType),
			PaddingLeft:  boolPtr(true),
			PaddingRight: boolPtr(false),
		})
	}

	return hints
}

// collectWorkflowCallHints shows parameter names for workflow calls
func collectWorkflowCallHints(call *ast.WorkflowCall) []InlayHint {
	if call.Resolved == nil || call.Args == "" {
		return nil
	}

	var hints []InlayHint

	// Parse parameters from the resolved workflow definition
	params := parseParams(call.Resolved.Params)
	args := parseArgs(call.Args)

	// Show parameter name before each argument
	for i, arg := range args {
		if i >= len(params) {
			break
		}

		// Calculate position - need to account for spawn/detach keywords
		prefix := "workflow "
		if call.Mode == ast.CallSpawn {
			prefix = "spawn workflow "
		} else if call.Mode == ast.CallDetach {
			prefix = "detach workflow "
		}

		pos := protocol.Position{
			Line:      uint32(call.Line - 1),
			Character: uint32(call.Column + len(prefix) + len(call.Name) + len("(") + arg.Column - 1),
		}

		hints = append(hints, InlayHint{
			Position:     pos,
			Label:        InlayHintLabelPart{Value: params[i].Name + ": "},
			Kind:         ptrToKind(InlayHintKindParameter),
			PaddingLeft:  boolPtr(false),
			PaddingRight: boolPtr(true),
		})
	}

	// Show return type if result is bound (not for detach mode)
	if call.Result != "" && call.Resolved.ReturnType != "" && call.Mode != ast.CallDetach {
		pos := protocol.Position{
			Line:      uint32(call.Line - 1),
			Character: uint32(call.Column + len(call.Result)),
		}

		hints = append(hints, InlayHint{
			Position:     pos,
			Label:        InlayHintLabelPart{Value: ": " + call.Resolved.ReturnType},
			Kind:         ptrToKind(InlayHintKindType),
			PaddingLeft:  boolPtr(true),
			PaddingRight: boolPtr(false),
		})
	}

	return hints
}

// collectAwaitHints shows hints for single await statements
func collectAwaitHints(await *ast.AwaitStmt) []InlayHint {
	var hints []InlayHint

	switch await.AwaitKind() {
	case "timer":
		// Show human-readable duration
		readable := humanizeDuration(await.Timer)
		if readable != "" {
			pos := protocol.Position{
				Line:      uint32(await.Line - 1),
				Character: uint32(await.Column + len("await timer(")),
			}
			hints = append(hints, InlayHint{
				Position:     pos,
				Label:        InlayHintLabelPart{Value: readable},
				Kind:         ptrToKind(InlayHintKindType),
				PaddingLeft:  boolPtr(false),
				PaddingRight: boolPtr(false),
			})
		}

	case "signal":
		// Show signal parameters if bound
		if await.SignalParams != "" && await.SignalResolved != nil {
			params := parseParams(await.SignalResolved.Params)
			paramTypes := make([]string, len(params))
			for i, p := range params {
				paramTypes[i] = p.Type
			}

			pos := protocol.Position{
				Line:      uint32(await.Line - 1),
				Character: uint32(await.Column + len("await signal ") + len(await.Signal) + len(" -> ")),
			}

			hints = append(hints, InlayHint{
				Position:     pos,
				Label:        InlayHintLabelPart{Value: "(" + strings.Join(paramTypes, ", ") + ")"},
				Kind:         ptrToKind(InlayHintKindType),
				PaddingLeft:  boolPtr(false),
				PaddingRight: boolPtr(false),
			})
		}
	}

	return hints
}

// collectCaseHints shows hints for await one cases
func collectCaseHints(c *ast.AwaitOneCase, visibleRange protocol.Range) []InlayHint {
	var hints []InlayHint

	// Recurse into case body
	hints = append(hints, collectStatementHints(c.Body, visibleRange)...)

	// Add hint about what this case waits for
	switch c.CaseKind() {
	case "timer":
		readable := humanizeDuration(c.Timer)
		if readable != "" {
			pos := protocol.Position{
				Line:      uint32(c.Line - 1),
				Character: uint32(c.Column + len("timer(")),
			}
			hints = append(hints, InlayHint{
				Position:     pos,
				Label:        InlayHintLabelPart{Value: readable},
				Kind:         ptrToKind(InlayHintKindType),
				PaddingLeft:  boolPtr(false),
				PaddingRight: boolPtr(false),
			})
		}
	}

	return hints
}

// Helper functions

func isInRange(line int, r protocol.Range) bool {
	return uint32(line-1) >= r.Start.Line && uint32(line-1) <= r.End.Line
}

func humanizeDuration(dur string) string {
	// Parse duration like "5m", "1h", "7d"
	if len(dur) < 2 {
		return ""
	}

	unit := dur[len(dur)-1:]
	value := dur[:len(dur)-1]

	switch unit {
	case "s":
		if value == "1" {
			return "1 second"
		}
		return value + " seconds"
	case "m":
		if value == "1" {
			return "1 minute"
		}
		return value + " minutes"
	case "h":
		if value == "1" {
			return "1 hour"
		}
		return value + " hours"
	case "d":
		if value == "1" {
			return "1 day"
		}
		return value + " days"
	}
	return ""
}

type Param struct {
	Name string
	Type string
}

func parseParams(paramsStr string) []Param {
	// Parse "(name: type, name2: type2)" -> []Param
	paramsStr = strings.Trim(paramsStr, "()")
	if paramsStr == "" {
		return nil
	}

	parts := strings.Split(paramsStr, ",")
	params := make([]Param, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		colonIdx := strings.Index(part, ":")
		if colonIdx > 0 {
			name := strings.TrimSpace(part[:colonIdx])
			typ := strings.TrimSpace(part[colonIdx+1:])
			params = append(params, Param{Name: name, Type: typ})
		}
	}

	return params
}

type Arg struct {
	Text   string
	Column int
}

func parseArgs(argsStr string) []Arg {
	// Parse "arg1, arg2, arg3" -> []Arg with positions
	if argsStr == "" {
		return nil
	}

	parts := strings.Split(argsStr, ",")
	args := make([]Arg, 0, len(parts))
	col := 0

	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		// Find where the trimmed content starts in the original part
		startOffset := strings.Index(part, trimmed)
		if i == 0 {
			col = startOffset
		} else {
			col += len(parts[i-1]) + 1 + startOffset // +1 for comma
		}
		args = append(args, Arg{Text: trimmed, Column: col})
		col += len(trimmed)
	}

	return args
}

func ptrToKind(k InlayHintKind) *InlayHintKind {
	return &k
}
