package server

import (
	"fmt"
	"strings"

	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func codeActionHandler(store *DocumentStore) protocol.TextDocumentCodeActionFunc {
	return func(context *glsp.Context, params *protocol.CodeActionParams) (any, error) {
		doc := store.Get(params.TextDocument.URI)
		if doc == nil || doc.File == nil {
			return nil, nil
		}

		var actions []protocol.CodeAction

		// Collect all available code actions based on diagnostics and context
		actions = append(actions, addMissingDefinitionActions(doc, params)...)
		actions = append(actions, convertReturnToCloseActions(doc, params)...)

		// Return as interface slice for JSON encoding
		result := make([]any, len(actions))
		for i, a := range actions {
			result[i] = a
		}
		return result, nil
	}
}

// addMissingDefinitionActions creates code actions to add missing activity/workflow definitions
func addMissingDefinitionActions(doc *Document, params *protocol.CodeActionParams) []protocol.CodeAction {
	var actions []protocol.CodeAction

	// Check for "undefined" errors in the range
	for _, err := range doc.ResolveErrs {
		if !strings.Contains(err.Msg, "undefined") {
			continue
		}

		errRange := posToRange(err.Line, err.Column)
		if !rangesOverlap(params.Range, errRange) {
			continue
		}

		// Extract the name from error message like "undefined activity: Foo"
		var name string
		var kind string
		if strings.Contains(err.Msg, "undefined activity:") {
			parts := strings.SplitN(err.Msg, "undefined activity:", 2)
			if len(parts) == 2 {
				name = strings.TrimSpace(parts[1])
				kind = "activity"
			}
		} else if strings.Contains(err.Msg, "undefined workflow:") {
			parts := strings.SplitN(err.Msg, "undefined workflow:", 2)
			if len(parts) == 2 {
				name = strings.TrimSpace(parts[1])
				kind = "workflow"
			}
		}

		if name != "" && kind != "" {
			// Find the call to extract parameters
			call := findCallByName(doc.File, name)
			params := ""
			returnType := ""
			if call != nil {
				params = call.Args
				if call.Result != "" {
					returnType = "Result"
				}
			}

			// Generate the definition
			var def string
			if kind == "activity" {
				if returnType != "" {
					def = fmt.Sprintf("\nactivity %s(%s) -> (%s):\n    # TODO: implement\n    return result\n", name, params, returnType)
				} else {
					def = fmt.Sprintf("\nactivity %s(%s):\n    # TODO: implement\n", name, params)
				}
			} else {
				if returnType != "" {
					def = fmt.Sprintf("\nworkflow %s(%s) -> (%s):\n    # TODO: implement\n    close result\n", name, params, returnType)
				} else {
					def = fmt.Sprintf("\nworkflow %s(%s):\n    # TODO: implement\n    close\n", name, params)
				}
			}

			// Insert at end of file
			lines := strings.Split(doc.Content, "\n")
			endLine := uint32(len(lines))

			action := protocol.CodeAction{
				Title: fmt.Sprintf("Add missing %s '%s'", kind, name),
				Kind:  ptrTo(protocol.CodeActionKindQuickFix),
				Edit: &protocol.WorkspaceEdit{
					Changes: map[string][]protocol.TextEdit{
						doc.URI: {
							{
								Range: protocol.Range{
									Start: protocol.Position{Line: endLine, Character: 0},
									End:   protocol.Position{Line: endLine, Character: 0},
								},
								NewText: def,
							},
						},
					},
				},
			}
			actions = append(actions, action)
		}
	}

	return actions
}

// convertReturnToCloseActions suggests converting old return statements to close
func convertReturnToCloseActions(doc *Document, params *protocol.CodeActionParams) []protocol.CodeAction {
	var actions []protocol.CodeAction

	// Look for return statements in workflow bodies (not in signal/query/update handlers)
	for _, def := range doc.File.Definitions {
		wf, ok := def.(*ast.WorkflowDef)
		if !ok {
			continue
		}

		// Check workflow body for return statements
		returns := findReturnStatements(wf.Body)
		for _, ret := range returns {
			retRange := lineRange(ret.Line, ret.Line)
			if !rangesOverlap(params.Range, retRange) {
				continue
			}

			// Suggest converting to close
			var newText string
			if ret.Value != "" {
				// return Foo{} -> close Foo{}
				newText = fmt.Sprintf("    close %s", ret.Value)
			} else {
				newText = "    close"
			}

			action := protocol.CodeAction{
				Title: "Convert 'return' to 'close'",
				Kind:  ptrTo(protocol.CodeActionKindRefactor),
				Edit: &protocol.WorkspaceEdit{
					Changes: map[string][]protocol.TextEdit{
						doc.URI: {
							{
								Range: protocol.Range{
									Start: protocol.Position{Line: uint32(ret.Line - 1), Character: 0},
									End:   protocol.Position{Line: uint32(ret.Line - 1), Character: 1000},
								},
								NewText: newText,
							},
						},
					},
				},
			}
			actions = append(actions, action)
		}
	}

	return actions
}

// Helper functions

func rangesOverlap(a, b protocol.Range) bool {
	// Check if ranges overlap at all
	return !(a.End.Line < b.Start.Line || b.End.Line < a.Start.Line)
}

func findCallByName(file *ast.File, name string) *ast.ActivityCall {
	for _, def := range file.Definitions {
		if wf, ok := def.(*ast.WorkflowDef); ok {
			if call := findCallInStatements(wf.Body, name); call != nil {
				return call
			}
		}
	}
	return nil
}

func findCallInStatements(stmts []ast.Statement, name string) *ast.ActivityCall {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.ActivityCall:
			if s.Name == name {
				return s
			}
		case *ast.IfStmt:
			if call := findCallInStatements(s.Body, name); call != nil {
				return call
			}
			if call := findCallInStatements(s.ElseBody, name); call != nil {
				return call
			}
		case *ast.ForStmt:
			if call := findCallInStatements(s.Body, name); call != nil {
				return call
			}
		case *ast.AwaitAllBlock:
			if call := findCallInStatements(s.Body, name); call != nil {
				return call
			}
		case *ast.AwaitOneBlock:
			for _, c := range s.Cases {
				if call := findCallInStatements(c.Body, name); call != nil {
					return call
				}
			}
		}
	}
	return nil
}

func findAwaitOneBlocks(wf *ast.WorkflowDef) []*ast.AwaitOneBlock {
	var blocks []*ast.AwaitOneBlock
	var visit func([]ast.Statement)
	visit = func(stmts []ast.Statement) {
		for _, stmt := range stmts {
			switch s := stmt.(type) {
			case *ast.AwaitOneBlock:
				blocks = append(blocks, s)
			case *ast.IfStmt:
				visit(s.Body)
				visit(s.ElseBody)
			case *ast.ForStmt:
				visit(s.Body)
			}
		}
	}
	visit(wf.Body)
	return blocks
}

func findSignalsThatModify(wf *ast.WorkflowDef, varName string) []string {
	var signals []string
	for _, sig := range wf.Signals {
		// Check if signal body contains assignment to varName
		if containsAssignment(sig.Body, varName) {
			signals = append(signals, sig.Name)
		}
	}
	return signals
}

func containsAssignment(stmts []ast.Statement, varName string) bool {
	for _, stmt := range stmts {
		if raw, ok := stmt.(*ast.RawStmt); ok {
			// Check if raw statement contains assignment like "varName = ..."
			if strings.Contains(raw.Text, varName+" =") || strings.Contains(raw.Text, varName+"=") {
				return true
			}
		}
	}
	return false
}

func findReturnStatements(stmts []ast.Statement) []*ast.ReturnStmt {
	var returns []*ast.ReturnStmt
	var visit func([]ast.Statement)
	visit = func(stmts []ast.Statement) {
		for _, stmt := range stmts {
			switch s := stmt.(type) {
			case *ast.ReturnStmt:
				returns = append(returns, s)
			case *ast.IfStmt:
				visit(s.Body)
				visit(s.ElseBody)
			case *ast.ForStmt:
				visit(s.Body)
			case *ast.AwaitOneBlock:
				for _, c := range s.Cases {
					visit(c.Body)
				}
			}
		}
	}
	visit(stmts)
	return returns
}
