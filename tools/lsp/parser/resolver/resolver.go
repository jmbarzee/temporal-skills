package resolver

import (
	"fmt"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
)

// ResolveError represents a resolution error with position info.
type ResolveError struct {
	Msg      string
	Line     int
	Column   int
	Severity string // "error" (default) or "warning"
}

func (e *ResolveError) Error() string {
	return fmt.Sprintf("resolve error at %d:%d: %s", e.Line, e.Column, e.Msg)
}

// endpointInfo tracks which namespace defines a nexus endpoint.
type endpointInfo struct {
	namespaceName string
	endpoint      *ast.NamespaceEndpoint
}

// Resolve walks the AST, linking calls to their definitions.
// Returns a list of errors (empty on success).
func Resolve(file *ast.File) []*ResolveError {
	workflows := make(map[string]*ast.WorkflowDef)
	activities := make(map[string]*ast.ActivityDef)
	workers := make(map[string]*ast.WorkerDef)
	namespaces := make(map[string]*ast.NamespaceDef)
	nexusServices := make(map[string]*ast.NexusServiceDef)
	var errs []*ResolveError

	// Pass 1: Collect all definitions.
	for _, def := range file.Definitions {
		switch d := def.(type) {
		case *ast.WorkflowDef:
			if _, exists := workflows[d.Name]; exists {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("duplicate workflow definition: %s", d.Name),
					Line:   d.Line,
					Column: d.Column,
				})
			}
			workflows[d.Name] = d
		case *ast.ActivityDef:
			if _, exists := activities[d.Name]; exists {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("duplicate activity definition: %s", d.Name),
					Line:   d.Line,
					Column: d.Column,
				})
			}
			activities[d.Name] = d
		case *ast.WorkerDef:
			if _, exists := workers[d.Name]; exists {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("duplicate worker definition: %s", d.Name),
					Line:   d.Line,
					Column: d.Column,
				})
			}
			workers[d.Name] = d
		case *ast.NamespaceDef:
			if _, exists := namespaces[d.Name]; exists {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("duplicate namespace definition: %s", d.Name),
					Line:   d.Line,
					Column: d.Column,
				})
			}
			namespaces[d.Name] = d
		case *ast.NexusServiceDef:
			if _, exists := nexusServices[d.Name]; exists {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("duplicate nexus service definition: %s", d.Name),
					Line:   d.Line,
					Column: d.Column,
				})
			}
			nexusServices[d.Name] = d
		}
	}

	// Build global endpoint map across all namespaces.
	allEndpoints := make(map[string]*endpointInfo)
	for _, ns := range namespaces {
		for i := range ns.Endpoints {
			ep := &ns.Endpoints[i]
			if existing, exists := allEndpoints[ep.EndpointName]; exists {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("duplicate nexus endpoint name %q: defined in namespace %s and namespace %s", ep.EndpointName, existing.namespaceName, ns.Name),
					Line:   ep.Line,
					Column: ep.Column,
				})
			}
			allEndpoints[ep.EndpointName] = &endpointInfo{namespaceName: ns.Name, endpoint: ep}
		}
	}

	// Continue to Pass 2 even if there are duplicate definition errors.
	// This provides better diagnostics by also reporting undefined references.

	// Pass 2: Walk workflow bodies, resolving references.
	for _, def := range file.Definitions {
		wf, ok := def.(*ast.WorkflowDef)
		if !ok {
			continue
		}

		// Build signal, query, and update maps for this workflow.
		signals := make(map[string]*ast.SignalDecl)
		queries := make(map[string]*ast.QueryDecl)
		updates := make(map[string]*ast.UpdateDecl)
		for _, s := range wf.Signals {
			signals[s.Name] = s
		}
		for _, q := range wf.Queries {
			queries[q.Name] = q
		}
		for _, u := range wf.Updates {
			updates[u.Name] = u
		}

		// Build condition map from state block.
		conditions := make(map[string]*ast.ConditionDecl)
		if wf.State != nil {
			for _, c := range wf.State.Conditions {
				conditions[c.Name] = c
			}
		}

		// Build promise set from workflow body.
		promises := make(map[string]*ast.PromiseStmt)
		for _, stmt := range wf.Body {
			if p, ok := stmt.(*ast.PromiseStmt); ok {
				promises[p.Name] = p
			}
		}

		ctx := &resolveCtx{
			workflows:    workflows,
			activities:   activities,
			signals:      signals,
			queries:      queries,
			updates:      updates,
			conditions:   conditions,
			promises:     promises,
			nexusServices: nexusServices,
			allEndpoints: allEndpoints,
		}

		// Resolve handler bodies.
		for _, s := range wf.Signals {
			ctx.resolveStatements(s.Body)
		}
		for _, q := range wf.Queries {
			ctx.resolveStatements(q.Body)
		}
		for _, u := range wf.Updates {
			ctx.resolveStatements(u.Body)
		}

		ctx.resolveStatements(wf.Body)
		errs = append(errs, ctx.errs...)
	}

	// Pass 2b: Resolve nexus service operation bodies.
	for _, def := range file.Definitions {
		svc, ok := def.(*ast.NexusServiceDef)
		if !ok {
			continue
		}
		for _, op := range svc.Operations {
			if op.OpType == ast.NexusOpAsync {
				// Async operations reference a workflow by name.
				if _, ok := workflows[op.WorkflowName]; !ok {
					errs = append(errs, &ResolveError{
						Msg:    fmt.Sprintf("nexus service %s: async operation %s references undefined workflow: %s", svc.Name, op.Name, op.WorkflowName),
						Line:   op.Line,
						Column: op.Column,
					})
				}
			} else if op.OpType == ast.NexusOpSync {
				// Sync operations have a body — resolve like a workflow body.
				syncCtx := &resolveCtx{
					workflows:    workflows,
					activities:   activities,
					signals:      make(map[string]*ast.SignalDecl),
					queries:      make(map[string]*ast.QueryDecl),
					updates:      make(map[string]*ast.UpdateDecl),
					conditions:   make(map[string]*ast.ConditionDecl),
					promises:     make(map[string]*ast.PromiseStmt),
					nexusServices: nexusServices,
					allEndpoints: allEndpoints,
				}
				syncCtx.resolveStatements(op.Body)
				errs = append(errs, syncCtx.errs...)
			}
		}
	}

	// Pass 3: Resolve worker and namespace references.
	for _, w := range workers {
		for i := range w.Workflows {
			ref := &w.Workflows[i]
			if def, ok := workflows[ref.Name]; ok {
				ref.Resolved = def
			} else {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("worker %s references undefined workflow: %s", w.Name, ref.Name),
					Line:   ref.Line,
					Column: ref.Column,
				})
			}
		}
		for i := range w.Activities {
			ref := &w.Activities[i]
			if def, ok := activities[ref.Name]; ok {
				ref.Resolved = def
			} else {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("worker %s references undefined activity: %s", w.Name, ref.Name),
					Line:   ref.Line,
					Column: ref.Column,
				})
			}
		}
		for i := range w.Services {
			ref := &w.Services[i]
			if def, ok := nexusServices[ref.Name]; ok {
				ref.Resolved = def
			} else {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("worker %s references undefined nexus service: %s", w.Name, ref.Name),
					Line:   ref.Line,
					Column: ref.Column,
				})
			}
		}
	}

	for _, ns := range namespaces {
		for i := range ns.Workers {
			nw := &ns.Workers[i]
			if def, ok := workers[nw.WorkerName]; ok {
				nw.ResolvedWorker = def
			} else {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("namespace %s references undefined worker: %s", ns.Name, nw.WorkerName),
					Line:   nw.Line,
					Column: nw.Column,
				})
			}
		}
	}

	return errs
}

type resolveCtx struct {
	workflows     map[string]*ast.WorkflowDef
	activities    map[string]*ast.ActivityDef
	signals       map[string]*ast.SignalDecl
	queries       map[string]*ast.QueryDecl
	updates       map[string]*ast.UpdateDecl
	conditions    map[string]*ast.ConditionDecl
	promises      map[string]*ast.PromiseStmt
	nexusServices map[string]*ast.NexusServiceDef
	allEndpoints  map[string]*endpointInfo
	errs          []*ResolveError
}

func (c *resolveCtx) resolveStatements(stmts []ast.Statement) {
	for _, stmt := range stmts {
		c.resolveStatement(stmt)
	}
}

func (c *resolveCtx) resolveStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ActivityCall:
		if def, ok := c.activities[s.Name]; ok {
			s.Resolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined activity: %s", s.Name),
				Line:   s.Line,
				Column: s.Column,
			})
		}

	case *ast.WorkflowCall:
		if def, ok := c.workflows[s.Name]; ok {
			s.Resolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined workflow: %s", s.Name),
				Line:   s.Line,
				Column: s.Column,
			})
		}

	case *ast.NexusCall:
		res := c.resolveNexusRef(s.Endpoint, s.Service, s.Operation, s.Line, s.Column)
		s.ResolvedEndpoint = res.endpoint
		s.ResolvedEndpointNamespace = res.endpointNamespace
		s.ResolvedService = res.service
		s.ResolvedOperation = res.operation

	case *ast.AwaitAllBlock:
		c.resolveStatements(s.Body)

	case *ast.AwaitOneBlock:
		for _, awaitCase := range s.Cases {
			c.resolveAwaitOneCase(awaitCase)
		}

	case *ast.SwitchBlock:
		for _, sc := range s.Cases {
			c.resolveStatements(sc.Body)
		}
		if s.Default != nil {
			c.resolveStatements(s.Default)
		}

	case *ast.IfStmt:
		c.resolveStatements(s.Body)
		if s.ElseBody != nil {
			c.resolveStatements(s.ElseBody)
		}

	case *ast.ForStmt:
		c.resolveStatements(s.Body)

	case *ast.AwaitStmt:
		c.resolveAsyncTarget(s.Target, s.Line, s.Column)

	case *ast.PromiseStmt:
		c.resolveAsyncTarget(s.Target, s.Line, s.Column)

	case *ast.SetStmt:
		if _, ok := c.conditions[s.Name]; !ok {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined condition: %s", s.Name),
				Line:   s.Line,
				Column: s.Column,
			})
		}

	case *ast.UnsetStmt:
		if _, ok := c.conditions[s.Name]; !ok {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined condition: %s", s.Name),
				Line:   s.Line,
				Column: s.Column,
			})
		}
	}
}

// nexusResolution holds the resolved links from a nexus call site.
type nexusResolution struct {
	endpoint          *ast.NamespaceEndpoint
	endpointNamespace string // namespace that owns the endpoint
	service           *ast.NexusServiceDef
	operation         *ast.NexusOperation
}

// resolveNexusRef validates a nexus call site (endpoint, service, operation).
// Used by NexusCall, AwaitStmt nexus, AwaitOneCase nexus, and PromiseStmt nexus.
func (c *resolveCtx) resolveNexusRef(endpoint, service, operation string, line, column int) nexusResolution {
	var res nexusResolution

	// Endpoint resolution.
	if len(c.allEndpoints) > 0 {
		if epInfo, ok := c.allEndpoints[endpoint]; ok {
			res.endpoint = epInfo.endpoint
			res.endpointNamespace = epInfo.namespaceName
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined nexus endpoint: %s", endpoint),
				Line:   line,
				Column: column,
			})
		}
	} else {
		c.errs = append(c.errs, &ResolveError{
			Msg:      fmt.Sprintf("unresolved nexus endpoint: %s (no endpoints defined — may be external)", endpoint),
			Line:     line,
			Column:   column,
			Severity: "warning",
		})
	}

	// Service resolution.
	if len(c.nexusServices) > 0 {
		svc, ok := c.nexusServices[service]
		if !ok {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined nexus service: %s", service),
				Line:   line,
				Column: column,
			})
		} else {
			res.service = svc
			// Operation resolution (only when service was found).
			for _, op := range svc.Operations {
				if op.Name == operation {
					res.operation = op
					break
				}
			}
			if res.operation == nil {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("nexus service %s has no operation %s", service, operation),
					Line:   line,
					Column: column,
				})
			}
		}
	} else {
		c.errs = append(c.errs, &ResolveError{
			Msg:      fmt.Sprintf("unresolved nexus service: %s (no nexus services defined — may be external)", service),
			Line:     line,
			Column:   column,
			Severity: "warning",
		})
	}

	return res
}

func (c *resolveCtx) resolveAwaitOneCase(awaitCase *ast.AwaitOneCase) {
	if awaitCase.Target != nil {
		c.resolveAsyncTarget(awaitCase.Target, awaitCase.Line, awaitCase.Column)
	}

	// Resolve nested await all block if present.
	if awaitCase.AwaitAll != nil {
		c.resolveStatements(awaitCase.AwaitAll.Body)
	}
	// Resolve the case body.
	c.resolveStatements(awaitCase.Body)
}

// resolveAsyncTarget resolves references inside an async target.
func (c *resolveCtx) resolveAsyncTarget(target ast.AsyncTarget, line, column int) {
	switch t := target.(type) {
	case *ast.SignalTarget:
		if def, ok := c.signals[t.Name]; ok {
			t.Resolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined signal: %s", t.Name),
				Line:   line,
				Column: column,
			})
		}
	case *ast.UpdateTarget:
		if def, ok := c.updates[t.Name]; ok {
			t.Resolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined update: %s", t.Name),
				Line:   line,
				Column: column,
			})
		}
	case *ast.ActivityTarget:
		if def, ok := c.activities[t.Name]; ok {
			t.Resolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined activity: %s", t.Name),
				Line:   line,
				Column: column,
			})
		}
	case *ast.WorkflowTarget:
		if def, ok := c.workflows[t.Name]; ok {
			t.Resolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined workflow: %s", t.Name),
				Line:   line,
				Column: column,
			})
		}
	case *ast.NexusTarget:
		res := c.resolveNexusRef(t.Endpoint, t.Service, t.Operation, line, column)
		t.ResolvedEndpoint = res.endpoint
		t.ResolvedEndpointNamespace = res.endpointNamespace
		t.ResolvedService = res.service
		t.ResolvedOperation = res.operation
	case *ast.IdentTarget:
		_, isPromise := c.promises[t.Name]
		_, isCondition := c.conditions[t.Name]
		if !isPromise && !isCondition {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined promise or condition: %s", t.Name),
				Line:   line,
				Column: column,
			})
		}
		if isCondition && t.Result != "" {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("condition %q cannot have a result binding (-> identifier)", t.Name),
				Line:   line,
				Column: column,
			})
		}
	case *ast.TimerTarget:
		// No resolution needed for timers
	}
}
