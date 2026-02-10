package server

import (
	"fmt"
	"strings"

	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func hoverHandler(store *DocumentStore) protocol.TextDocumentHoverFunc {
	return func(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
		doc := store.Get(params.TextDocument.URI)
		if doc == nil || doc.File == nil {
			return nil, nil
		}

		line := int(params.Position.Line) + 1 // LSP 0-based → parser 1-based

		node := findNodeAtLine(doc.File, line)
		if node == nil {
			return nil, nil
		}

		sig := signatureFor(node)
		if sig == "" {
			return nil, nil
		}

		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.MarkupKindMarkdown,
				Value: fmt.Sprintf("```twf\n%s\n```", sig),
			},
		}, nil
	}
}

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
		}
	}
	return nil
}

// findNodeInStmts searches statements recursively for a node on the given line.
func findNodeInStmts(stmts []ast.Statement, line int) ast.Node {
	for _, stmt := range stmts {
		if n := findNodeInStmt(stmt, line); n != nil {
			return n
		}
	}
	return nil
}

func findNodeInStmt(stmt ast.Statement, line int) ast.Node {
	switch s := stmt.(type) {
	case *ast.ActivityCall:
		if s.Line == line {
			return s
		}
	case *ast.WorkflowCall:
		if s.Line == line {
			return s
		}
	case *ast.AwaitStmt:
		if s.Line == line {
			return s
		}
	case *ast.AwaitAllBlock:
		if n := findNodeInStmts(s.Body, line); n != nil {
			return n
		}
	case *ast.AwaitOneBlock:
		for _, c := range s.Cases {
			if c.Line == line {
				return c
			}
			// Check nested await all block.
			if c.AwaitAll != nil {
				if n := findNodeInStmts(c.AwaitAll.Body, line); n != nil {
					return n
				}
			}
			if n := findNodeInStmts(c.Body, line); n != nil {
				return n
			}
		}
	case *ast.SwitchBlock:
		for _, c := range s.Cases {
			if n := findNodeInStmts(c.Body, line); n != nil {
				return n
			}
		}
		if n := findNodeInStmts(s.Default, line); n != nil {
			return n
		}
	case *ast.IfStmt:
		if n := findNodeInStmts(s.Body, line); n != nil {
			return n
		}
		if n := findNodeInStmts(s.ElseBody, line); n != nil {
			return n
		}
	case *ast.ForStmt:
		if n := findNodeInStmts(s.Body, line); n != nil {
			return n
		}
	case *ast.ReturnStmt:
		if s.Line == line {
			return s
		}
	}
	return nil
}

// signatureFor builds a human-readable signature for a node.
func signatureFor(node ast.Node) string {
	switch n := node.(type) {
	case *ast.WorkflowDef:
		return workflowSig(n)
	case *ast.ActivityDef:
		return activitySig(n)
	case *ast.SignalDecl:
		return fmt.Sprintf("signal %s(%s)", n.Name, n.Params)
	case *ast.QueryDecl:
		sig := fmt.Sprintf("query %s(%s)", n.Name, n.Params)
		if n.ReturnType != "" {
			sig += " -> (" + n.ReturnType + ")"
		}
		return sig
	case *ast.UpdateDecl:
		sig := fmt.Sprintf("update %s(%s)", n.Name, n.Params)
		if n.ReturnType != "" {
			sig += " -> (" + n.ReturnType + ")"
		}
		return sig
	case *ast.ActivityCall:
		if n.Resolved != nil {
			return activitySig(n.Resolved)
		}
		return fmt.Sprintf("activity %s(%s)", n.Name, n.Args)
	case *ast.WorkflowCall:
		if n.Resolved != nil {
			return workflowSig(n.Resolved)
		}
		prefix := "workflow"
		switch n.Mode {
		case ast.CallSpawn:
			prefix = "spawn workflow"
		case ast.CallDetach:
			prefix = "detach workflow"
		}
		if n.Namespace != "" {
			prefix += " (nexus " + n.Namespace + ")"
		}
		return fmt.Sprintf("%s %s(%s)", prefix, n.Name, n.Args)
	case *ast.AwaitStmt:
		switch n.AwaitKind() {
		case "timer":
			return fmt.Sprintf("await timer(%s)", n.Timer)
		case "signal":
			if n.SignalParams != "" {
				return fmt.Sprintf("await signal %s -> %s", n.Signal, n.SignalParams)
			}
			return fmt.Sprintf("await signal %s", n.Signal)
		case "update":
			if n.UpdateParams != "" {
				return fmt.Sprintf("await update %s -> %s", n.Update, n.UpdateParams)
			}
			return fmt.Sprintf("await update %s", n.Update)
		case "activity":
			if n.ActivityResult != "" {
				return fmt.Sprintf("await activity %s(%s) -> %s", n.Activity, n.ActivityArgs, n.ActivityResult)
			}
			return fmt.Sprintf("await activity %s(%s)", n.Activity, n.ActivityArgs)
		case "workflow":
			prefix := "await workflow"
			if n.WorkflowMode == ast.CallSpawn {
				prefix = "await spawn workflow"
			} else if n.WorkflowMode == ast.CallDetach {
				prefix = "await detach workflow"
			}
			if n.WorkflowResult != "" {
				return fmt.Sprintf("%s %s(%s) -> %s", prefix, n.Workflow, n.WorkflowArgs, n.WorkflowResult)
			}
			return fmt.Sprintf("%s %s(%s)", prefix, n.Workflow, n.WorkflowArgs)
		}
		return "await"
	default:
		return ""
	}
}

func workflowSig(w *ast.WorkflowDef) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("workflow %s(%s)", w.Name, w.Params))
	if w.ReturnType != "" {
		parts = append(parts, "-> ("+w.ReturnType+")")
	}
	return strings.Join(parts, " ")
}

func activitySig(a *ast.ActivityDef) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("activity %s(%s)", a.Name, a.Params))
	if a.ReturnType != "" {
		parts = append(parts, "-> ("+a.ReturnType+")")
	}
	return strings.Join(parts, " ")
}
