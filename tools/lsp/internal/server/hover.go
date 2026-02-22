package server

import (
	"fmt"
	"strings"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func hoverHandler(store *DocumentStore) protocol.TextDocumentHoverFunc {
	return func(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
		doc, ok := store.Get(params.TextDocument.URI)
		if !ok || doc.File == nil {
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
		if n.Activity.Resolved != nil {
			return activitySig(n.Activity.Resolved)
		}
		return fmt.Sprintf("activity %s(%s)", n.Activity.Name, n.Args)
	case *ast.WorkflowCall:
		if n.Workflow.Resolved != nil {
			return workflowSig(n.Workflow.Resolved)
		}
		prefix := "workflow"
		if n.Mode == ast.CallDetach {
			prefix = "detach workflow"
		}
		return fmt.Sprintf("%s %s(%s)", prefix, n.Workflow.Name, n.Args)
	case *ast.WorkerDef:
		sig := fmt.Sprintf("worker %s", n.Name)
		if len(n.Workflows) > 0 || len(n.Activities) > 0 || len(n.Services) > 0 {
			var parts []string
			for _, ref := range n.Workflows {
				parts = append(parts, fmt.Sprintf("  workflow %s", ref.Name))
			}
			for _, ref := range n.Activities {
				parts = append(parts, fmt.Sprintf("  activity %s", ref.Name))
			}
			for _, ref := range n.Services {
				parts = append(parts, fmt.Sprintf("  nexus service %s", ref.Name))
			}
			sig += "\n" + strings.Join(parts, "\n")
		}
		return sig
	case *ast.Ref[*ast.WorkflowDef]:
		if n.Resolved != nil {
			return signatureFor(n.Resolved)
		}
		return fmt.Sprintf("ref %s (unresolved)", n.Name)
	case *ast.Ref[*ast.ActivityDef]:
		if n.Resolved != nil {
			return signatureFor(n.Resolved)
		}
		return fmt.Sprintf("ref %s (unresolved)", n.Name)
	case *ast.Ref[*ast.NexusServiceDef]:
		if n.Resolved != nil {
			return signatureFor(n.Resolved)
		}
		return fmt.Sprintf("ref %s (unresolved)", n.Name)
	case *ast.NamespaceWorker:
		sig := fmt.Sprintf("worker %s", n.Worker.Name)
		tq := extractWorkerTaskQueue(n)
		if tq != "" {
			sig += fmt.Sprintf("\n  task_queue: %s", tq)
		}
		if n.Worker.Resolved != nil {
			var parts []string
			for _, ref := range n.Worker.Resolved.Workflows {
				parts = append(parts, fmt.Sprintf("  workflow %s", ref.Name))
			}
			for _, ref := range n.Worker.Resolved.Activities {
				parts = append(parts, fmt.Sprintf("  activity %s", ref.Name))
			}
			for _, ref := range n.Worker.Resolved.Services {
				parts = append(parts, fmt.Sprintf("  nexus service %s", ref.Name))
			}
			if len(parts) > 0 {
				sig += "\n" + strings.Join(parts, "\n")
			}
		}
		return sig
	case *ast.NamespaceDef:
		sig := fmt.Sprintf("namespace %s", n.Name)
		var parts []string
		for _, w := range n.Workers {
			tq := extractWorkerTaskQueue(&w)
			if tq != "" {
				parts = append(parts, fmt.Sprintf("  worker %s (task_queue: %s)", w.Worker.Name, tq))
			} else {
				parts = append(parts, fmt.Sprintf("  worker %s", w.Worker.Name))
			}
		}
		for _, ep := range n.Endpoints {
			tq := extractEndpointTaskQueue(&ep)
			if tq != "" {
				parts = append(parts, fmt.Sprintf("  nexus endpoint %s (task_queue: %s)", ep.EndpointName, tq))
			} else {
				parts = append(parts, fmt.Sprintf("  nexus endpoint %s", ep.EndpointName))
			}
		}
		if len(parts) > 0 {
			sig += "\n" + strings.Join(parts, "\n")
		}
		return sig
	case *ast.NamespaceEndpoint:
		sig := fmt.Sprintf("nexus endpoint %s", n.EndpointName)
		tq := extractEndpointTaskQueue(n)
		if tq != "" {
			sig += fmt.Sprintf("\n  task_queue: %s", tq)
		}
		return sig
	case *ast.NexusCall:
		prefix := "nexus"
		if n.Detach {
			prefix = "detach nexus"
		}
		sig := fmt.Sprintf("%s %s %s.%s(%s)", prefix, n.Endpoint.Name, n.Service.Name, n.Operation.Name, n.Args)
		if n.Endpoint.Resolved != nil {
			tq := extractEndpointTaskQueue(n.Endpoint.Resolved)
			if tq != "" {
				sig += fmt.Sprintf("\n→ routes to task_queue %s (namespace %s)", tq, n.Endpoint.Resolved.Namespace)
			}
		}
		return sig
	case *ast.NexusServiceDef:
		var ops []string
		for _, op := range n.Operations {
			if op.OpType == ast.NexusOpAsync {
				ops = append(ops, fmt.Sprintf("  async %s workflow %s", op.Name, op.Workflow.Name))
			} else {
				ops = append(ops, fmt.Sprintf("  sync %s(%s) -> (%s)", op.Name, op.Params, op.ReturnType))
			}
		}
		sig := fmt.Sprintf("nexus service %s", n.Name)
		if len(ops) > 0 {
			sig += "\n" + strings.Join(ops, "\n")
		}
		return sig
	case *ast.AwaitStmt:
		if n.Target == nil {
			return "await"
		}
		switch t := n.Target.(type) {
		case *ast.TimerTarget:
			return fmt.Sprintf("await timer(%s)", t.Duration)
		case *ast.SignalTarget:
			if t.Params != "" {
				return fmt.Sprintf("await signal %s -> %s", t.Signal.Name, t.Params)
			}
			return fmt.Sprintf("await signal %s", t.Signal.Name)
		case *ast.UpdateTarget:
			if t.Params != "" {
				return fmt.Sprintf("await update %s -> %s", t.Update.Name, t.Params)
			}
			return fmt.Sprintf("await update %s", t.Update.Name)
		case *ast.ActivityTarget:
			if t.Result != "" {
				return fmt.Sprintf("await activity %s(%s) -> %s", t.Activity.Name, t.Args, t.Result)
			}
			return fmt.Sprintf("await activity %s(%s)", t.Activity.Name, t.Args)
		case *ast.WorkflowTarget:
			prefix := "await workflow"
			if t.Mode == ast.CallDetach {
				prefix = "await detach workflow"
			}
			if t.Result != "" {
				return fmt.Sprintf("%s %s(%s) -> %s", prefix, t.Workflow.Name, t.Args, t.Result)
			}
			return fmt.Sprintf("%s %s(%s)", prefix, t.Workflow.Name, t.Args)
		case *ast.NexusTarget:
			sig := fmt.Sprintf("await nexus %s %s.%s(%s)", t.Endpoint.Name, t.Service.Name, t.Operation.Name, t.Args)
			if t.Result != "" {
				sig += " -> " + t.Result
			}
			if t.Endpoint.Resolved != nil {
				tq := extractEndpointTaskQueue(t.Endpoint.Resolved)
				if tq != "" {
					sig += fmt.Sprintf("\n→ routes to task_queue %s (namespace %s)", tq, t.Endpoint.Resolved.Namespace)
				}
			}
			return sig
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

// extractEndpointTaskQueue returns the task_queue value from a namespace endpoint's options.
func extractEndpointTaskQueue(ep *ast.NamespaceEndpoint) string {
	if ep == nil || ep.Options == nil {
		return ""
	}
	for _, e := range ep.Options.Entries {
		if e.Key == "task_queue" {
			return e.Value
		}
	}
	return ""
}

// extractWorkerTaskQueue returns the task_queue value from a namespace worker's options.
func extractWorkerTaskQueue(nw *ast.NamespaceWorker) string {
	if nw == nil || nw.Options == nil {
		return ""
	}
	for _, e := range nw.Options.Entries {
		if e.Key == "task_queue" {
			return e.Value
		}
	}
	return ""
}
