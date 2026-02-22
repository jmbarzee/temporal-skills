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
		if n.Activity.Resolved != nil {
			return n.Activity.Resolved
		}
	case *ast.WorkflowCall:
		if n.Workflow.Resolved != nil {
			return n.Workflow.Resolved
		}
	case *ast.NexusCall:
		// Prefer service definition as the primary go-to-definition target.
		if n.Service.Resolved != nil {
			return n.Service.Resolved
		}
		if n.Endpoint.Resolved != nil {
			return n.Endpoint.Resolved
		}
	case *ast.AwaitStmt:
		if n.Target != nil {
			return resolvedTargetFromAsync(n.Target)
		}
	case *ast.Ref[*ast.WorkflowDef]:
		if n.Resolved != nil {
			return n.Resolved
		}
	case *ast.Ref[*ast.ActivityDef]:
		if n.Resolved != nil {
			return n.Resolved
		}
	case *ast.Ref[*ast.NexusServiceDef]:
		if n.Resolved != nil {
			return n.Resolved
		}
	case *ast.NamespaceWorker:
		if n.Worker.Resolved != nil {
			return n.Worker.Resolved
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
		if t.Signal.Resolved != nil {
			return t.Signal.Resolved
		}
	case *ast.UpdateTarget:
		if t.Update.Resolved != nil {
			return t.Update.Resolved
		}
	case *ast.ActivityTarget:
		if t.Activity.Resolved != nil {
			return t.Activity.Resolved
		}
	case *ast.WorkflowTarget:
		if t.Workflow.Resolved != nil {
			return t.Workflow.Resolved
		}
	case *ast.NexusTarget:
		if t.Service.Resolved != nil {
			return t.Service.Resolved
		}
		if t.Endpoint.Resolved != nil {
			return t.Endpoint.Resolved
		}
	}
	return nil
}
