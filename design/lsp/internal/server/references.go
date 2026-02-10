package server

import (
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func referencesHandler(store *DocumentStore) protocol.TextDocumentReferencesFunc {
	return func(context *glsp.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
		doc := store.Get(params.TextDocument.URI)
		if doc == nil || doc.File == nil {
			return nil, nil
		}

		line := int(params.Position.Line) + 1

		node := findNodeAtLine(doc.File, line)
		if node == nil {
			return nil, nil
		}

		name, kind := nameOfNode(node)
		if name == "" {
			return nil, nil
		}

		refs := collectReferences(doc.File, name, kind, params.Context.IncludeDeclaration)
		if len(refs) == 0 {
			return nil, nil
		}

		var locs []protocol.Location
		for _, ref := range refs {
			locs = append(locs, protocol.Location{
				URI:   params.TextDocument.URI,
				Range: nameRange(ref),
			})
		}
		return locs, nil
	}
}

// nameOfNode returns the name and kind ("workflow", "activity", "signal",
// "query", "update") for an AST node. For references it follows the Resolved
// pointer to normalize to the definition identity.
func nameOfNode(node ast.Node) (name, kind string) {
	switch n := node.(type) {
	case *ast.WorkflowDef:
		return n.Name, "workflow"
	case *ast.ActivityDef:
		return n.Name, "activity"
	case *ast.SignalDecl:
		return n.Name, "signal"
	case *ast.QueryDecl:
		return n.Name, "query"
	case *ast.UpdateDecl:
		return n.Name, "update"
	case *ast.ActivityCall:
		if n.Resolved != nil {
			return n.Resolved.Name, "activity"
		}
		return n.Name, "activity"
	case *ast.WorkflowCall:
		if n.Resolved != nil {
			return n.Resolved.Name, "workflow"
		}
		return n.Name, "workflow"
	case *ast.HintStmt:
		return n.Name, n.Kind
	}
	return "", ""
}

// collectReferences walks the file and returns every node whose name and kind
// match the target. When includeDecl is true the definition node is included.
func collectReferences(file *ast.File, name, kind string, includeDecl bool) []ast.Node {
	var refs []ast.Node

	for _, def := range file.Definitions {
		switch d := def.(type) {
		case *ast.WorkflowDef:
			if includeDecl && kind == "workflow" && d.Name == name {
				refs = append(refs, d)
			}
			// Check embedded declarations.
			if includeDecl {
				for _, s := range d.Signals {
					if kind == "signal" && s.Name == name {
						refs = append(refs, s)
					}
				}
				for _, q := range d.Queries {
					if kind == "query" && q.Name == name {
						refs = append(refs, q)
					}
				}
				for _, u := range d.Updates {
					if kind == "update" && u.Name == name {
						refs = append(refs, u)
					}
				}
			}
			// Walk handler bodies.
			for _, s := range d.Signals {
				refs = collectRefsInStmts(s.Body, name, kind, refs)
			}
			for _, q := range d.Queries {
				refs = collectRefsInStmts(q.Body, name, kind, refs)
			}
			for _, u := range d.Updates {
				refs = collectRefsInStmts(u.Body, name, kind, refs)
			}
			refs = collectRefsInStmts(d.Body, name, kind, refs)

		case *ast.ActivityDef:
			if includeDecl && kind == "activity" && d.Name == name {
				refs = append(refs, d)
			}
			refs = collectRefsInStmts(d.Body, name, kind, refs)
		}
	}
	return refs
}

func collectRefsInStmts(stmts []ast.Statement, name, kind string, refs []ast.Node) []ast.Node {
	for _, stmt := range stmts {
		refs = collectRefsInStmt(stmt, name, kind, refs)
	}
	return refs
}

func collectRefsInStmt(stmt ast.Statement, name, kind string, refs []ast.Node) []ast.Node {
	switch s := stmt.(type) {
	case *ast.ActivityCall:
		if kind == "activity" && s.Name == name {
			refs = append(refs, s)
		}
	case *ast.WorkflowCall:
		if kind == "workflow" && s.Name == name {
			refs = append(refs, s)
		}
	case *ast.HintStmt:
		if s.Kind == kind && s.Name == name {
			refs = append(refs, s)
		}
	case *ast.AwaitAllBlock:
		refs = collectRefsInStmts(s.Body, name, kind, refs)
	case *ast.AwaitOneBlock:
		for _, c := range s.Cases {
			// Check nested await all block.
			if c.AwaitAll != nil {
				refs = collectRefsInStmts(c.AwaitAll.Body, name, kind, refs)
			}
			refs = collectRefsInStmts(c.Body, name, kind, refs)
		}
	case *ast.SwitchBlock:
		for _, c := range s.Cases {
			refs = collectRefsInStmts(c.Body, name, kind, refs)
		}
		refs = collectRefsInStmts(s.Default, name, kind, refs)
	case *ast.IfStmt:
		refs = collectRefsInStmts(s.Body, name, kind, refs)
		refs = collectRefsInStmts(s.ElseBody, name, kind, refs)
	case *ast.ForStmt:
		refs = collectRefsInStmts(s.Body, name, kind, refs)
	}
	return refs
}

// nameRange returns an LSP range covering just the name portion of a node.
func nameRange(node ast.Node) protocol.Range {
	n, _ := nameOfNode(node)
	line := uint32(0)
	if node.NodeLine() > 0 {
		line = uint32(node.NodeLine() - 1)
	}
	col := uint32(0)
	if node.NodeColumn() > 0 {
		col = uint32(node.NodeColumn() - 1)
	}
	return protocol.Range{
		Start: protocol.Position{Line: line, Character: col},
		End:   protocol.Position{Line: line, Character: col + uint32(len(n))},
	}
}
