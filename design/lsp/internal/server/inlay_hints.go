package server

import (
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
	"github.com/tliron/glsp"
	protocol316 "github.com/tliron/glsp/protocol_3_16"
	protocol "github.com/tliron/glsp/protocol_3_17"
)

func inlayHintHandler(store *DocumentStore) protocol.TextDocumentInlayHintFunc {
	return func(context *glsp.Context, params *protocol.InlayHintParams) ([]protocol.InlayHint, error) {
		doc := store.Get(params.TextDocument.URI)
		if doc == nil || doc.File == nil {
			return nil, nil
		}

		var hints []protocol.InlayHint

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
func collectWorkflowHints(wf *ast.WorkflowDef, visibleRange protocol316.Range) []protocol.InlayHint {
	var hints []protocol.InlayHint

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
func collectStatementHints(stmts []ast.Statement, visibleRange protocol316.Range) []protocol.InlayHint {
	var hints []protocol.InlayHint

	for _, stmt := range stmts {
		hints = append(hints, collectHintsFromStatement(stmt, visibleRange)...)
	}

	return hints
}

// collectHintsFromStatement collects hints from a single statement
func collectHintsFromStatement(stmt ast.Statement, visibleRange protocol316.Range) []protocol.InlayHint {
	var hints []protocol.InlayHint

	// Check if statement is in visible range
	if !isInRange(stmt.NodeLine(), visibleRange) {
		return nil
	}

	switch s := stmt.(type) {
	case *ast.AwaitOneBlock:
		for _, c := range s.Cases {
			hints = append(hints, collectStatementHints(c.Body, visibleRange)...)
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

// Helper functions

func isInRange(line int, r protocol316.Range) bool {
	return uint32(line-1) >= r.Start.Line && uint32(line-1) <= r.End.Line
}
