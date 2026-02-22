package server

import (
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
)

// findNodeAtLine returns the most specific AST node on the given line.
func findNodeAtLine(file *ast.File, line int) ast.Node {
	for _, def := range file.Definitions {
		switch d := def.(type) {
		case *ast.WorkflowDef:
			if d.Line == line {
				return d
			}
			for _, s := range d.Signals {
				if s.Line == line {
					return s
				}
			}
			for _, q := range d.Queries {
				if q.Line == line {
					return q
				}
			}
			for _, u := range d.Updates {
				if u.Line == line {
					return u
				}
			}
			// Search handler bodies.
			for _, s := range d.Signals {
				if n := findNodeInStmts(s.Body, line); n != nil {
					return n
				}
			}
			for _, q := range d.Queries {
				if n := findNodeInStmts(q.Body, line); n != nil {
					return n
				}
			}
			for _, u := range d.Updates {
				if n := findNodeInStmts(u.Body, line); n != nil {
					return n
				}
			}
			if n := findNodeInStmts(d.Body, line); n != nil {
				return n
			}

		case *ast.ActivityDef:
			if d.Line == line {
				return d
			}
			if n := findNodeInStmts(d.Body, line); n != nil {
				return n
			}

		case *ast.NexusServiceDef:
			if d.Line == line {
				return d
			}
			// Check operations for sync operation bodies.
			for _, op := range d.Operations {
				if op.OpType == ast.NexusOpSync {
					if n := findNodeInStmts(op.Body, line); n != nil {
						return n
					}
				}
			}

		case *ast.WorkerDef:
			if d.Line == line {
				return d
			}
			for i := range d.Workflows {
				if d.Workflows[i].Line == line {
					return &d.Workflows[i]
				}
			}
			for i := range d.Activities {
				if d.Activities[i].Line == line {
					return &d.Activities[i]
				}
			}
			for i := range d.Services {
				if d.Services[i].Line == line {
					return &d.Services[i]
				}
			}

		case *ast.NamespaceDef:
			if d.Line == line {
				return d
			}
			// Check namespace worker entries.
			for i := range d.Workers {
				if d.Workers[i].Line == line {
					return &d.Workers[i]
				}
			}
			// Check namespace endpoint entries.
			for i := range d.Endpoints {
				if d.Endpoints[i].Line == line {
					return &d.Endpoints[i]
				}
			}
		}
	}
	return nil
}

// findNodeInStmts searches statements recursively for a node on the given line.
func findNodeInStmts(stmts []ast.Statement, line int) ast.Node {
	var found ast.Node
	ast.WalkStatements(stmts, func(s ast.Statement) bool {
		if s.NodeLine() == line {
			found = s
			return false
		}
		return true
	})
	return found
}
