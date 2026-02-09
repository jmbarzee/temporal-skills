package server

import (
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func documentSymbolHandler(store *DocumentStore) protocol.TextDocumentDocumentSymbolFunc {
	return func(context *glsp.Context, params *protocol.DocumentSymbolParams) (any, error) {
		doc := store.Get(params.TextDocument.URI)
		if doc == nil || doc.File == nil {
			return nil, nil
		}

		var symbols []protocol.DocumentSymbol
		for _, def := range doc.File.Definitions {
			switch d := def.(type) {
			case *ast.WorkflowDef:
				symbols = append(symbols, workflowSymbol(d))
			case *ast.ActivityDef:
				symbols = append(symbols, activitySymbol(d))
			}
		}

		return symbols, nil
	}
}

func workflowSymbol(wf *ast.WorkflowDef) protocol.DocumentSymbol {
	r := defRange(wf)
	sym := protocol.DocumentSymbol{
		Name:           wf.Name,
		Kind:           protocol.SymbolKindFunction,
		Range:          r,
		SelectionRange: posToRange(wf.Line, wf.Column),
	}

	var children []protocol.DocumentSymbol

	for _, s := range wf.Signals {
		sr := posToRange(s.Line, s.Column)
		children = append(children, protocol.DocumentSymbol{
			Name:           s.Name,
			Kind:           protocol.SymbolKindEvent,
			Range:          sr,
			SelectionRange: sr,
		})
	}
	for _, q := range wf.Queries {
		qr := posToRange(q.Line, q.Column)
		children = append(children, protocol.DocumentSymbol{
			Name:           q.Name,
			Kind:           protocol.SymbolKindMethod,
			Range:          qr,
			SelectionRange: qr,
		})
	}
	for _, u := range wf.Updates {
		ur := posToRange(u.Line, u.Column)
		children = append(children, protocol.DocumentSymbol{
			Name:           u.Name,
			Kind:           protocol.SymbolKindMethod,
			Range:          ur,
			SelectionRange: ur,
		})
	}

	if len(children) > 0 {
		sym.Children = children
	}

	return sym
}

func activitySymbol(act *ast.ActivityDef) protocol.DocumentSymbol {
	r := defRange(act)
	return protocol.DocumentSymbol{
		Name:           act.Name,
		Kind:           protocol.SymbolKindFunction,
		Range:          r,
		SelectionRange: posToRange(act.Line, act.Column),
	}
}

// defRange estimates the full range of a definition by scanning its body
// statements for the last line number, since the AST does not store end positions.
func defRange(def ast.Definition) protocol.Range {
	startLine := def.NodeLine()
	endLine := startLine

	switch d := def.(type) {
	case *ast.WorkflowDef:
		endLine = lastLineInStmts(d.Body, endLine)
		for _, s := range d.Signals {
			if s.Line > endLine {
				endLine = s.Line
			}
		}
		for _, q := range d.Queries {
			if q.Line > endLine {
				endLine = q.Line
			}
		}
		for _, u := range d.Updates {
			if u.Line > endLine {
				endLine = u.Line
			}
		}
	case *ast.ActivityDef:
		endLine = lastLineInStmts(d.Body, endLine)
	}

	start := protocol.Position{}
	if startLine > 0 {
		start.Line = uint32(startLine - 1)
	}
	end := protocol.Position{Line: uint32(endLine), Character: 0} // line after the last statement

	return protocol.Range{Start: start, End: end}
}

func lastLineInStmts(stmts []ast.Statement, current int) int {
	for _, s := range stmts {
		if l := lastLineInStmt(s); l > current {
			current = l
		}
	}
	return current
}

func lastLineInStmt(stmt ast.Statement) int {
	line := stmt.NodeLine()
	switch s := stmt.(type) {
	case *ast.ParallelBlock:
		line = lastLineInStmts(s.Body, line)
	case *ast.SelectBlock:
		for _, c := range s.Cases {
			if c.Line > line {
				line = c.Line
			}
			line = lastLineInStmts(c.Body, line)
		}
	case *ast.SwitchBlock:
		for _, c := range s.Cases {
			line = lastLineInStmts(c.Body, line)
		}
		line = lastLineInStmts(s.Default, line)
	case *ast.IfStmt:
		line = lastLineInStmts(s.Body, line)
		line = lastLineInStmts(s.ElseBody, line)
	case *ast.ForStmt:
		line = lastLineInStmts(s.Body, line)
	}
	return line
}
