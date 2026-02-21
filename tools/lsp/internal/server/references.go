package server

import (
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func referencesHandler(store *DocumentStore) protocol.TextDocumentReferencesFunc {
	return func(context *glsp.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
		doc, ok := store.Get(params.TextDocument.URI)
		if !ok || doc.File == nil {
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
// "query", "update", "nexus_service", "nexus_endpoint", "worker") for an AST node.
// For references it follows the Resolved pointer to normalize to the definition identity.
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
	case *ast.NexusServiceDef:
		return n.Name, "nexus_service"
	case *ast.NamespaceEndpoint:
		return n.EndpointName, "nexus_endpoint"
	case *ast.WorkerDef:
		return n.Name, "worker"
	case *ast.Ref[*ast.WorkflowDef]:
		return n.Name, "workflow"
	case *ast.Ref[*ast.ActivityDef]:
		return n.Name, "activity"
	case *ast.Ref[*ast.NexusServiceDef]:
		return n.Name, "nexus_service"
	case *ast.NamespaceWorker:
		return n.Worker.Name, "worker"
	case *ast.NamespaceDef:
		return n.Name, "namespace"
	case *ast.ActivityCall:
		if n.Activity.Resolved != nil {
			return n.Activity.Resolved.Name, "activity"
		}
		return n.Activity.Name, "activity"
	case *ast.WorkflowCall:
		if n.Workflow.Resolved != nil {
			return n.Workflow.Resolved.Name, "workflow"
		}
		return n.Workflow.Name, "workflow"
	case *ast.NexusCall:
		if n.ResolvedService != nil {
			return n.ResolvedService.Name, "nexus_service"
		}
		return n.Service, "nexus_service"
	case *ast.AwaitStmt:
		if n.Target != nil {
			return nameOfAsyncTarget(n.Target)
		}
	case *ast.AwaitOneCase:
		if n.Target != nil {
			return nameOfAsyncTarget(n.Target)
		}
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

		case *ast.NexusServiceDef:
			if includeDecl && kind == "nexus_service" && d.Name == name {
				refs = append(refs, d)
			}
			// Walk sync operation bodies for nested references.
			for _, op := range d.Operations {
				if op.OpType == ast.NexusOpSync {
					refs = collectRefsInStmts(op.Body, name, kind, refs)
				}
			}

		case *ast.WorkerDef:
			if includeDecl && kind == "worker" && d.Name == name {
				refs = append(refs, d)
			}
			// Worker refs reference workflows, activities, and nexus services.
			for i := range d.Workflows {
				ref := &d.Workflows[i]
				if kind == "workflow" && ref.Name == name {
					refs = append(refs, ref)
				}
			}
			for i := range d.Activities {
				ref := &d.Activities[i]
				if kind == "activity" && ref.Name == name {
					refs = append(refs, ref)
				}
			}
			for i := range d.Services {
				ref := &d.Services[i]
				if kind == "nexus_service" && ref.Name == name {
					refs = append(refs, ref)
				}
			}

		case *ast.NamespaceDef:
			if includeDecl && kind == "namespace" && d.Name == name {
				refs = append(refs, d)
			}
			// Check worker references.
			if kind == "worker" {
				for i := range d.Workers {
					if d.Workers[i].Worker.Name == name {
						refs = append(refs, &d.Workers[i])
					}
				}
			}
			// Check endpoint references.
			if kind == "nexus_endpoint" {
				for i := range d.Endpoints {
					if d.Endpoints[i].EndpointName == name {
						refs = append(refs, &d.Endpoints[i])
					}
				}
			}
		}
	}
	return refs
}

func collectRefsInStmts(stmts []ast.Statement, name, kind string, refs []ast.Node) []ast.Node {
	ast.WalkStatements(stmts, func(s ast.Statement) bool {
		switch n := s.(type) {
		case *ast.ActivityCall:
			if kind == "activity" && n.Activity.Name == name {
				refs = append(refs, n)
			}
		case *ast.WorkflowCall:
			if kind == "workflow" && n.Workflow.Name == name {
				refs = append(refs, n)
			}
		case *ast.NexusCall:
			if kind == "nexus_service" && n.Service == name {
				refs = append(refs, n)
			}
			if kind == "nexus_endpoint" && n.Endpoint == name {
				refs = append(refs, n)
			}
		case *ast.AwaitStmt:
			if matchesAsyncTarget(n.Target, name, kind) {
				refs = append(refs, n)
			}
		case *ast.AwaitOneCase:
			if matchesAsyncTarget(n.Target, name, kind) {
				refs = append(refs, n)
			}
		case *ast.PromiseStmt:
			if matchesAsyncTarget(n.Target, name, kind) {
				refs = append(refs, n)
			}
		}
		return true
	})
	return refs
}

// nameOfAsyncTarget returns the name and kind for an async target node.
func nameOfAsyncTarget(target ast.AsyncTarget) (name, kind string) {
	switch t := target.(type) {
	case *ast.SignalTarget:
		return t.Signal.Name, "signal"
	case *ast.UpdateTarget:
		return t.Update.Name, "update"
	case *ast.ActivityTarget:
		return t.Activity.Name, "activity"
	case *ast.WorkflowTarget:
		return t.Workflow.Name, "workflow"
	case *ast.NexusTarget:
		return t.Service, "nexus_service"
	}
	return "", ""
}

// matchesAsyncTarget reports whether an async target matches the given name and kind.
func matchesAsyncTarget(target ast.AsyncTarget, name, kind string) bool {
	switch t := target.(type) {
	case *ast.SignalTarget:
		return kind == "signal" && t.Signal.Name == name
	case *ast.UpdateTarget:
		return kind == "update" && t.Update.Name == name
	case *ast.ActivityTarget:
		return kind == "activity" && t.Activity.Name == name
	case *ast.WorkflowTarget:
		return kind == "workflow" && t.Workflow.Name == name
	case *ast.NexusTarget:
		return (kind == "nexus_service" && t.Service == name) ||
			(kind == "nexus_endpoint" && t.Endpoint == name)
	}
	return false
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
