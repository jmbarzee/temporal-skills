package server

import (
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func foldingRangeHandler(store *DocumentStore) protocol.TextDocumentFoldingRangeFunc {
	return func(context *glsp.Context, params *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
		doc, ok := store.Get(params.TextDocument.URI)
		if !ok || doc.File == nil {
			return nil, nil
		}

		var ranges []protocol.FoldingRange
		for _, def := range doc.File.Definitions {
			switch d := def.(type) {
			case *ast.WorkflowDef:
				startLine := d.Line
				endLine := lastLineInStmts(d.Body, startLine)
				for _, s := range d.Signals {
					if s.Line > endLine {
						endLine = s.Line
					}
					sEnd := lastLineInStmts(s.Body, s.Line)
					if sEnd > endLine {
						endLine = sEnd
					}
					addFold(&ranges, s.Line, sEnd)
					foldStmts(s.Body, &ranges)
				}
				for _, q := range d.Queries {
					if q.Line > endLine {
						endLine = q.Line
					}
					qEnd := lastLineInStmts(q.Body, q.Line)
					if qEnd > endLine {
						endLine = qEnd
					}
					addFold(&ranges, q.Line, qEnd)
					foldStmts(q.Body, &ranges)
				}
				for _, u := range d.Updates {
					if u.Line > endLine {
						endLine = u.Line
					}
					uEnd := lastLineInStmts(u.Body, u.Line)
					if uEnd > endLine {
						endLine = uEnd
					}
					addFold(&ranges, u.Line, uEnd)
					foldStmts(u.Body, &ranges)
				}
				addFold(&ranges, startLine, endLine)
				foldStmts(d.Body, &ranges)

			case *ast.ActivityDef:
				startLine := d.Line
				endLine := lastLineInStmts(d.Body, startLine)
				addFold(&ranges, startLine, endLine)
				foldStmts(d.Body, &ranges)
			}
		}

		return ranges, nil
	}
}

func foldStmts(stmts []ast.Statement, ranges *[]protocol.FoldingRange) {
	ast.WalkStatements(stmts, func(s ast.Statement) bool {
		switch n := s.(type) {
		case *ast.AwaitAllBlock:
			addFold(ranges, n.Line, lastLineInStmts(n.Body, n.Line))
		case *ast.AwaitOneBlock:
			endLine := n.Line
			for _, c := range n.Cases {
				if c.AwaitAll != nil {
					aaEnd := lastLineInStmts(c.AwaitAll.Body, c.Line)
					if aaEnd > endLine {
						endLine = aaEnd
					}
					addFold(ranges, c.Line, aaEnd)
				}
				cEnd := lastLineInStmts(c.Body, c.Line)
				if cEnd > endLine {
					endLine = cEnd
				}
				addFold(ranges, c.Line, lastLineInStmts(c.Body, c.Line))
			}
			addFold(ranges, n.Line, endLine)
		case *ast.SwitchBlock:
			endLine := n.Line
			for _, c := range n.Cases {
				cEnd := lastLineInStmts(c.Body, c.Line)
				if cEnd > endLine {
					endLine = cEnd
				}
				addFold(ranges, c.Line, lastLineInStmts(c.Body, c.Line))
			}
			defEnd := lastLineInStmts(n.Default, endLine)
			if defEnd > endLine {
				endLine = defEnd
			}
			addFold(ranges, n.Line, endLine)
		case *ast.IfStmt:
			endLine := lastLineInStmts(n.Body, n.Line)
			endLine = lastLineInStmts(n.ElseBody, endLine)
			addFold(ranges, n.Line, endLine)
		case *ast.ForStmt:
			endLine := lastLineInStmts(n.Body, n.Line)
			addFold(ranges, n.Line, endLine)
		}
		return true
	})
}

// addFold appends a FoldingRange converting 1-based lines to 0-based.
// It skips zero-length folds (start == end).
func addFold(ranges *[]protocol.FoldingRange, startLine, endLine int) {
	if startLine >= endLine {
		return
	}
	s := uint32(startLine - 1)
	e := uint32(endLine - 1)
	*ranges = append(*ranges, protocol.FoldingRange{
		StartLine: s,
		EndLine:   e,
	})
}
