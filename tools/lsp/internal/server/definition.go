package server

import (
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func definitionHandler(store *DocumentStore) protocol.TextDocumentDefinitionFunc {
	return func(context *glsp.Context, params *protocol.DefinitionParams) (any, error) {
		doc, ok := store.Get(params.TextDocument.URI)
		if !ok || doc.File == nil {
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
	case *ast.NexusCall:
		// Prefer service definition as the primary go-to-definition target.
		if n.ResolvedService != nil {
			return n.ResolvedService
		}
		if n.ResolvedEndpoint != nil {
			return n.ResolvedEndpoint
		}
	case *ast.AwaitStmt:
		if n.Target != nil {
			return resolvedTargetFromAsync(n.Target)
		}
	case *ast.WorkerRef:
		if n.Resolved != nil {
			return n.Resolved
		}
	case *ast.NamespaceWorker:
		if n.ResolvedWorker != nil {
			return n.ResolvedWorker
		}
	case *ast.AwaitOneCase:
		if n.Target != nil {
			return resolvedTargetFromAsync(n.Target)
		}
	}
	return nil
}

// resolvedTargetFromAsync returns the resolved definition from an async target.
func resolvedTargetFromAsync(target ast.AsyncTarget) ast.Node {
	switch t := target.(type) {
	case *ast.SignalTarget:
		if t.Resolved != nil {
			return t.Resolved
		}
	case *ast.UpdateTarget:
		if t.Resolved != nil {
			return t.Resolved
		}
	case *ast.ActivityTarget:
		if t.Resolved != nil {
			return t.Resolved
		}
	case *ast.WorkflowTarget:
		if t.Resolved != nil {
			return t.Resolved
		}
	case *ast.NexusTarget:
		if t.ResolvedService != nil {
			return t.ResolvedService
		}
		if t.ResolvedEndpoint != nil {
			return t.ResolvedEndpoint
		}
	}
	return nil
}
