package server

import (
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func definitionHandler(store *DocumentStore) protocol.TextDocumentDefinitionFunc {
	return func(context *glsp.Context, params *protocol.DefinitionParams) (any, error) {
		doc := store.Get(params.TextDocument.URI)
		if doc == nil || doc.File == nil {
			return nil, nil
		}

		line := int(params.Position.Line) + 1

		node := findNodeAtLine(doc.File, line)
		if node == nil {
			return nil, nil
		}

		target := resolvedTarget(node)
		if target == nil {
			return nil, nil
		}

		return protocol.Location{
			URI:   params.TextDocument.URI,
			Range: posToRange(target.NodeLine(), target.NodeColumn()),
		}, nil
	}
}

// resolvedTarget returns the definition node that a call/reference resolves to.
func resolvedTarget(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.ActivityCall:
		if n.Resolved != nil {
			return n.Resolved
		}
	case *ast.WorkflowCall:
		if n.Resolved != nil {
			return n.Resolved
		}
	case *ast.AwaitStmt:
		// Check which type of await and return appropriate resolved reference
		if n.Signal != "" && n.SignalResolved != nil {
			return n.SignalResolved
		}
		if n.Update != "" && n.UpdateResolved != nil {
			return n.UpdateResolved
		}
		if n.Activity != "" && n.ActivityResolved != nil {
			return n.ActivityResolved
		}
		if n.Workflow != "" && n.WorkflowResolved != nil {
			return n.WorkflowResolved
		}
	}
	return nil
}
