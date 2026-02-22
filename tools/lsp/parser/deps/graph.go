package deps

import (
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
)

// Node represents a definition in the dependency graph.
type Node struct {
	Name       string `json:"name"`
	Kind       string `json:"kind"` // workflow, activity, nexusService, worker, namespace
	SourceFile string `json:"sourceFile,omitempty"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
}

// Edge represents a dependency from one definition to another.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"` // activityCall, workflowCall, nexusCall
	Line int    `json:"line"` // source line of the call
}

// UnresolvedRef represents a reference that could not be resolved.
type UnresolvedRef struct {
	From string `json:"from"`
	Name string `json:"name"`
	Kind string `json:"kind"`
	Line int    `json:"line"`
}

// CoarsenedEdge is an edge projected to a higher containment level.
type CoarsenedEdge struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Weight      int    `json:"weight"`
	DerivedFrom []int  `json:"derivedFrom"` // indices into Graph.Edges
}

// Summary counts definitions by type.
type Summary struct {
	Namespaces    int `json:"namespaces"`
	Workers       int `json:"workers"`
	Workflows     int `json:"workflows"`
	Activities    int `json:"activities"`
	NexusServices int `json:"nexusServices"`
	Edges         int `json:"edges"`
	Unresolved    int `json:"unresolved"`
}

// Graph is the full dependency graph output.
type Graph struct {
	Nodes       []Node            `json:"nodes"`
	Edges       []Edge            `json:"edges"`
	Containment map[string][]string `json:"containment"`
	Coarsened   *CoarsenedGraph   `json:"coarsened"`
	Unresolved  []UnresolvedRef   `json:"unresolved"`
	Summary     Summary           `json:"summary"`
}

// CoarsenedGraph holds edges projected to worker and namespace levels.
type CoarsenedGraph struct {
	WorkerEdges    []CoarsenedEdge `json:"workerEdges"`
	NamespaceEdges []CoarsenedEdge `json:"namespaceEdges"`
}

// Extract builds a dependency graph from a resolved AST.
func Extract(file *ast.File) *Graph {
	g := &Graph{
		Containment: make(map[string][]string),
		Coarsened:   &CoarsenedGraph{},
	}

	// childToParent maps a definition name to its containing worker/namespace.
	childToWorker := make(map[string]string)
	workerToNamespace := make(map[string]string)

	// Pass 1: Build nodes and containment.
	for _, def := range file.Definitions {
		switch d := def.(type) {
		case *ast.WorkflowDef:
			g.addNode(d.Name, "workflow", d.SourceFile, d.Line, d.Column)
		case *ast.ActivityDef:
			g.addNode(d.Name, "activity", d.SourceFile, d.Line, d.Column)
		case *ast.NexusServiceDef:
			g.addNode(d.Name, "nexusService", d.SourceFile, d.Line, d.Column)
		case *ast.WorkerDef:
			g.addNode(d.Name, "worker", d.SourceFile, d.Line, d.Column)
			var children []string
			for _, ref := range d.Workflows {
				children = append(children, ref.Name)
				childToWorker[ref.Name] = d.Name
			}
			for _, ref := range d.Activities {
				children = append(children, ref.Name)
				childToWorker[ref.Name] = d.Name
			}
			for _, ref := range d.Services {
				children = append(children, ref.Name)
				childToWorker[ref.Name] = d.Name
			}
			if len(children) > 0 {
				g.Containment[d.Name] = children
			}
		case *ast.NamespaceDef:
			g.addNode(d.Name, "namespace", d.SourceFile, d.Line, d.Column)
			var children []string
			for _, nw := range d.Workers {
				children = append(children, nw.Worker.Name)
				workerToNamespace[nw.Worker.Name] = d.Name
			}
			if len(children) > 0 {
				g.Containment[d.Name] = children
			}
		}
	}

	// Pass 2: Extract edges from definition bodies.
	for _, def := range file.Definitions {
		switch d := def.(type) {
		case *ast.WorkflowDef:
			g.extractFromBody(d.Name, d.Body)
			for _, s := range d.Signals {
				g.extractFromBody(d.Name, s.Body)
			}
			for _, q := range d.Queries {
				g.extractFromBody(d.Name, q.Body)
			}
			for _, u := range d.Updates {
				g.extractFromBody(d.Name, u.Body)
			}
		case *ast.ActivityDef:
			g.extractFromBody(d.Name, d.Body)
		case *ast.NexusServiceDef:
			for _, op := range d.Operations {
				if op.OpType == ast.NexusOpSync {
					g.extractFromBody(d.Name, op.Body)
				}
			}
		}
	}

	// Pass 3: Coarsen edges.
	g.coarsen(childToWorker, workerToNamespace)

	// Summary.
	for _, n := range g.Nodes {
		switch n.Kind {
		case "namespace":
			g.Summary.Namespaces++
		case "worker":
			g.Summary.Workers++
		case "workflow":
			g.Summary.Workflows++
		case "activity":
			g.Summary.Activities++
		case "nexusService":
			g.Summary.NexusServices++
		}
	}
	g.Summary.Edges = len(g.Edges)
	g.Summary.Unresolved = len(g.Unresolved)

	return g
}

func (g *Graph) addNode(name, kind, sourceFile string, line, column int) {
	g.Nodes = append(g.Nodes, Node{
		Name:       name,
		Kind:       kind,
		SourceFile: sourceFile,
		Line:       line,
		Column:     column,
	})
}

func (g *Graph) extractFromBody(from string, stmts []ast.Statement) {
	ast.WalkStatements(stmts, func(s ast.Statement) bool {
		switch stmt := s.(type) {
		case *ast.ActivityCall:
			g.addCallEdge(from, stmt.Activity.Name, "activityCall", stmt.Line, stmt.Activity.Resolved != nil)
		case *ast.WorkflowCall:
			g.addCallEdge(from, stmt.Workflow.Name, "workflowCall", stmt.Line, stmt.Workflow.Resolved != nil)
		case *ast.NexusCall:
			g.addCallEdge(from, stmt.Service.Name+"."+stmt.Operation.Name, "nexusCall", stmt.Line, stmt.Operation.Resolved != nil)
		}
		return true
	}, ast.WithAsyncTargets(func(target ast.AsyncTarget, parent ast.Statement) bool {
		switch t := target.(type) {
		case *ast.ActivityTarget:
			g.addCallEdge(from, t.Activity.Name, "activityCall", parent.NodeLine(), t.Activity.Resolved != nil)
		case *ast.WorkflowTarget:
			g.addCallEdge(from, t.Workflow.Name, "workflowCall", parent.NodeLine(), t.Workflow.Resolved != nil)
		case *ast.NexusTarget:
			g.addCallEdge(from, t.Service.Name+"."+t.Operation.Name, "nexusCall", parent.NodeLine(), t.Operation.Resolved != nil)
		}
		return true
	}))
}

func (g *Graph) addCallEdge(from, to, kind string, line int, resolved bool) {
	if !resolved {
		g.Unresolved = append(g.Unresolved, UnresolvedRef{
			From: from,
			Name: to,
			Kind: kind,
			Line: line,
		})
		return
	}
	g.Edges = append(g.Edges, Edge{
		From: from,
		To:   to,
		Kind: kind,
		Line: line,
	})
}

// coarsen projects edges to worker-level and namespace-level.
func (g *Graph) coarsen(childToWorker, workerToNamespace map[string]string) {
	type edgeKey struct{ from, to string }

	// Worker-level coarsening.
	workerAgg := make(map[edgeKey]*CoarsenedEdge)
	for i, e := range g.Edges {
		fromWorker := childToWorker[e.From]
		toWorker := childToWorker[e.To]
		if fromWorker == "" || toWorker == "" || fromWorker == toWorker {
			continue
		}
		key := edgeKey{fromWorker, toWorker}
		if ce, ok := workerAgg[key]; ok {
			ce.Weight++
			ce.DerivedFrom = append(ce.DerivedFrom, i)
		} else {
			workerAgg[key] = &CoarsenedEdge{
				From:        fromWorker,
				To:          toWorker,
				Weight:      1,
				DerivedFrom: []int{i},
			}
		}
	}
	for _, ce := range workerAgg {
		g.Coarsened.WorkerEdges = append(g.Coarsened.WorkerEdges, *ce)
	}

	// Namespace-level coarsening.
	nsAgg := make(map[edgeKey]*CoarsenedEdge)
	for i, e := range g.Edges {
		fromWorker := childToWorker[e.From]
		toWorker := childToWorker[e.To]
		fromNS := workerToNamespace[fromWorker]
		toNS := workerToNamespace[toWorker]
		if fromNS == "" || toNS == "" || fromNS == toNS {
			continue
		}
		key := edgeKey{fromNS, toNS}
		if ce, ok := nsAgg[key]; ok {
			ce.Weight++
			ce.DerivedFrom = append(ce.DerivedFrom, i)
		} else {
			nsAgg[key] = &CoarsenedEdge{
				From:        fromNS,
				To:          toNS,
				Weight:      1,
				DerivedFrom: []int{i},
			}
		}
	}
	for _, ce := range nsAgg {
		g.Coarsened.NamespaceEdges = append(g.Coarsened.NamespaceEdges, *ce)
	}
}
