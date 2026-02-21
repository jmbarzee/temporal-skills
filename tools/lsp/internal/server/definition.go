package server

import (
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
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
	case *ast.NexusCall:
		// Prefer service definition as the primary go-to-definition target.
		if n.ResolvedService != nil {
			return n.ResolvedService
		}
		if n.ResolvedEndpoint != nil {
			return n.ResolvedEndpoint
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
		if n.Nexus != "" && n.NexusResolvedService != nil {
			return n.NexusResolvedService
		}
		if n.Nexus != "" && n.NexusResolvedEndpoint != nil {
			return n.NexusResolvedEndpoint
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
		// Check which type of case and return appropriate resolved reference
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
		if n.Nexus != "" && n.NexusResolvedService != nil {
			return n.NexusResolvedService
		}
		if n.Nexus != "" && n.NexusResolvedEndpoint != nil {
			return n.NexusResolvedEndpoint
		}
	}
	return nil
}
