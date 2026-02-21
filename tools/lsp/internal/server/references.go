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
	case *ast.WorkerRef:
		// WorkerRef can resolve to workflow, activity, or nexus service.
		if n.Resolved != nil {
			switch d := n.Resolved.(type) {
			case *ast.WorkflowDef:
				return d.Name, "workflow"
			case *ast.ActivityDef:
				return d.Name, "activity"
			case *ast.NexusServiceDef:
				return d.Name, "nexus_service"
			}
		}
		return n.Name, "workflow" // best-effort fallback
	case *ast.NamespaceWorker:
		return n.WorkerName, "worker"
	case *ast.NamespaceDef:
		return n.Name, "namespace"
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
					if d.Workers[i].WorkerName == name {
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
	case *ast.NexusCall:
		if kind == "nexus_service" && s.Service == name {
			refs = append(refs, s)
		}
		if kind == "nexus_endpoint" && s.Endpoint == name {
			refs = append(refs, s)
		}
	case *ast.AwaitStmt:
		if matchesAsyncTarget(s.Target, name, kind) {
			refs = append(refs, s)
		}
	case *ast.AwaitAllBlock:
		refs = collectRefsInStmts(s.Body, name, kind, refs)
	case *ast.AwaitOneBlock:
		for _, c := range s.Cases {
			if matchesAsyncTarget(c.Target, name, kind) {
				refs = append(refs, c)
			}
			if c.AwaitAll != nil {
				refs = collectRefsInStmts(c.AwaitAll.Body, name, kind, refs)
			}
			refs = collectRefsInStmts(c.Body, name, kind, refs)
		}
	case *ast.PromiseStmt:
		if matchesAsyncTarget(s.Target, name, kind) {
			refs = append(refs, s)
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

// nameOfAsyncTarget returns the name and kind for an async target node.
func nameOfAsyncTarget(target ast.AsyncTarget) (name, kind string) {
	switch t := target.(type) {
	case *ast.SignalTarget:
		return t.Name, "signal"
	case *ast.UpdateTarget:
		return t.Name, "update"
	case *ast.ActivityTarget:
		return t.Name, "activity"
	case *ast.WorkflowTarget:
		return t.Name, "workflow"
	case *ast.NexusTarget:
		return t.Service, "nexus_service"
	}
	return "", ""
}

// matchesAsyncTarget reports whether an async target matches the given name and kind.
func matchesAsyncTarget(target ast.AsyncTarget, name, kind string) bool {
	switch t := target.(type) {
	case *ast.SignalTarget:
		return kind == "signal" && t.Name == name
	case *ast.UpdateTarget:
		return kind == "update" && t.Name == name
	case *ast.ActivityTarget:
		return kind == "activity" && t.Name == name
	case *ast.WorkflowTarget:
		return kind == "workflow" && t.Name == name
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
