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

	// Pass 1b: Empty definition warnings.
	for _, wf := range workflows {
		if !hasNonCommentStmts(wf.Body) && len(wf.Signals) == 0 && len(wf.Queries) == 0 && len(wf.Updates) == 0 && wf.State == nil {
			errs = append(errs, &ResolveError{
				Msg:      fmt.Sprintf("workflow %s has an empty body", wf.Name),
				Line:     wf.Line,
				Column:   wf.Column,
				Severity: "warning",
			})
		}
	}
	for _, act := range activities {
		if !hasNonCommentStmts(act.Body) {
			errs = append(errs, &ResolveError{
				Msg:      fmt.Sprintf("activity %s has an empty body", act.Name),
				Line:     act.Line,
				Column:   act.Column,
				Severity: "warning",
			})
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
			workflows:       workflows,
			activities:      activities,
			signals:         signals,
			queries:         queries,
			updates:         updates,
			conditions:      conditions,
			promises:        promises,
			nexusServices:   nexusServices,
			allEndpoints:    allEndpoints,
			workers:         workers,
			namespaces:      namespaces,
			callingWorkflow: wf.Name,
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
					workflows:     workflows,
					activities:    activities,
					signals:       make(map[string]*ast.SignalDecl),
					queries:       make(map[string]*ast.QueryDecl),
					updates:       make(map[string]*ast.UpdateDecl),
					conditions:    make(map[string]*ast.ConditionDecl),
					promises:      make(map[string]*ast.PromiseStmt),
					nexusServices: nexusServices,
					allEndpoints:  allEndpoints,
					workers:       workers,
					namespaces:    namespaces,
				}
				syncCtx.resolveStatements(op.Body)
				errs = append(errs, syncCtx.errs...)
			}
		}
	}

	// Pass 3: Worker and namespace validation.
	errs = append(errs, resolveWorkersAndNamespaces(namespaces, workers, workflows, activities, nexusServices, allEndpoints)...)

	return errs
}

// resolveWorkersAndNamespaces validates worker type sets and namespace instantiations.
func resolveWorkersAndNamespaces(namespaces map[string]*ast.NamespaceDef, workers map[string]*ast.WorkerDef, workflows map[string]*ast.WorkflowDef, activities map[string]*ast.ActivityDef, nexusServices map[string]*ast.NexusServiceDef, allEndpoints map[string]*endpointInfo) []*ResolveError {
	var errs []*ResolveError

	// 0. Empty definition warnings.
	for _, w := range workers {
		if len(w.Workflows) == 0 && len(w.Activities) == 0 && len(w.Services) == 0 {
			errs = append(errs, &ResolveError{
				Msg:      fmt.Sprintf("worker %s has no workflow, activity, or nexus service registrations", w.Name),
				Line:     w.Line,
				Column:   w.Column,
				Severity: "warning",
			})
		}
	}
	for _, ns := range namespaces {
		if len(ns.Workers) == 0 && len(ns.Endpoints) == 0 {
			errs = append(errs, &ResolveError{
				Msg:      fmt.Sprintf("namespace %s has no worker or endpoint instantiations", ns.Name),
				Line:     ns.Line,
				Column:   ns.Column,
				Severity: "warning",
			})
		}
	}

	// 1. Worker type set validation: refs must point to defined workflows/activities/services.
	//    Store resolution links on each WorkerRef.
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

	// 2. Namespace validation: worker instantiations must ref defined workers,
	//    and each instantiation must have a task_queue option.
	//    Endpoint instantiations must have a task_queue option.
	//    Store resolution links on each NamespaceWorker.
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
			tq := extractTaskQueue(nw.Options)
			if tq == "" {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("worker %s in namespace %s missing required task_queue option", nw.WorkerName, ns.Name),
					Line:   nw.Line,
					Column: nw.Column,
				})
			}
		}
		for _, ep := range ns.Endpoints {
			tq := extractTaskQueue(ep.Options)
			if tq == "" {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("nexus endpoint %s in namespace %s missing required task_queue option", ep.EndpointName, ns.Name),
					Line:   ep.Line,
					Column: ep.Column,
				})
			}
		}
	}

	// 3. Coverage warnings (only when namespaces exist).
	if len(namespaces) > 0 {
		// Track which workflows/activities/services are covered by instantiated workers.
		coveredWorkflows := make(map[string]bool)
		coveredActivities := make(map[string]bool)
		coveredServices := make(map[string]bool)
		instantiatedWorkers := make(map[string]bool)

		for _, ns := range namespaces {
			for _, nw := range ns.Workers {
				instantiatedWorkers[nw.WorkerName] = true
				if w, ok := workers[nw.WorkerName]; ok {
					for _, ref := range w.Workflows {
						coveredWorkflows[ref.Name] = true
					}
					for _, ref := range w.Activities {
						coveredActivities[ref.Name] = true
					}
					for _, ref := range w.Services {
						coveredServices[ref.Name] = true
					}
				}
			}
		}

		for name, wf := range workflows {
			if !coveredWorkflows[name] {
				errs = append(errs, &ResolveError{
					Msg:      fmt.Sprintf("workflow %s is not registered on any instantiated worker", name),
					Line:     wf.Line,
					Column:   wf.Column,
					Severity: "warning",
				})
			}
		}
		for name, act := range activities {
			if !coveredActivities[name] {
				errs = append(errs, &ResolveError{
					Msg:      fmt.Sprintf("activity %s is not registered on any instantiated worker", name),
					Line:     act.Line,
					Column:   act.Column,
					Severity: "warning",
				})
			}
		}
		for name, svc := range nexusServices {
			if !coveredServices[name] {
				errs = append(errs, &ResolveError{
					Msg:      fmt.Sprintf("nexus service %s is not referenced by any worker", name),
					Line:     svc.Line,
					Column:   svc.Column,
					Severity: "warning",
				})
			}
		}
		for name, w := range workers {
			if !instantiatedWorkers[name] {
				errs = append(errs, &ResolveError{
					Msg:      fmt.Sprintf("worker %s is not instantiated in any namespace", name),
					Line:     w.Line,
					Column:   w.Column,
					Severity: "warning",
				})
			}
		}
	}

	// 4. Task queue coherence (per namespace): different worker type sets on same queue → error.
	type queueInfo struct {
		workerName string
		workflows  map[string]bool
		activities map[string]bool
	}
	for _, ns := range namespaces {
		queueWorkers := make(map[string][]queueInfo)
		for _, nw := range ns.Workers {
			tq := extractTaskQueue(nw.Options)
			if tq == "" {
				continue
			}
			w, ok := workers[nw.WorkerName]
			if !ok {
				continue
			}
			wfSet := make(map[string]bool)
			for _, ref := range w.Workflows {
				wfSet[ref.Name] = true
			}
			actSet := make(map[string]bool)
			for _, ref := range w.Activities {
				actSet[ref.Name] = true
			}
			queueWorkers[tq] = append(queueWorkers[tq], queueInfo{
				workerName: nw.WorkerName,
				workflows:  wfSet,
				activities: actSet,
			})
		}
		for queue, infos := range queueWorkers {
			if len(infos) < 2 {
				continue
			}
			first := infos[0]
			for _, other := range infos[1:] {
				if sameStringSet(first.workflows, other.workflows) && sameStringSet(first.activities, other.activities) {
					errs = append(errs, &ResolveError{
						Msg:      fmt.Sprintf("workers %s and %s on task queue %q in namespace %s have identical type sets (redundant)", first.workerName, other.workerName, queue, ns.Name),
						Severity: "warning",
					})
				} else {
					errs = append(errs, &ResolveError{
						Msg: fmt.Sprintf("workers %s and %s on task queue %q in namespace %s have different type sets", first.workerName, other.workerName, queue, ns.Name),
					})
				}
			}
		}
	}

	return errs
}

// hasNonCommentStmts returns true if the statement slice has at least one
// statement that is not a Comment.
func hasNonCommentStmts(stmts []ast.Statement) bool {
	for _, s := range stmts {
		if _, isComment := s.(*ast.Comment); !isComment {
			return true
		}
	}
	return false
}

// extractTaskQueue walks an OptionsBlock to find the task_queue key.
func extractTaskQueue(opts *ast.OptionsBlock) string {
	if opts == nil {
		return ""
	}
	for _, e := range opts.Entries {
		if e.Key == "task_queue" {
			return e.Value
		}
	}
	return ""
}

func sameStringSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

type resolveCtx struct {
	workflows       map[string]*ast.WorkflowDef
	activities      map[string]*ast.ActivityDef
	signals         map[string]*ast.SignalDecl
	queries         map[string]*ast.QueryDecl
	updates         map[string]*ast.UpdateDecl
	conditions      map[string]*ast.ConditionDecl
	promises        map[string]*ast.PromiseStmt
	nexusServices   map[string]*ast.NexusServiceDef
	allEndpoints    map[string]*endpointInfo
	workers         map[string]*ast.WorkerDef
	namespaces      map[string]*ast.NamespaceDef
	callingWorkflow string // name of the workflow being resolved (for routing checks)
	errs            []*ResolveError
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
		c.checkCallRouting("activity", s.Name, s.Options, s.Line, s.Column)

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
		c.checkCallRouting("workflow", s.Name, s.Options, s.Line, s.Column)

	case *ast.NexusCall:
		res := c.resolveNexusRef(s.Endpoint, s.Service, s.Operation, s.Detach, s.Result, s.Line, s.Column)
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
		// Resolve signal/update/activity/workflow references
		if s.Signal != "" {
			if def, ok := c.signals[s.Signal]; ok {
				s.SignalResolved = def
			} else {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("undefined signal: %s", s.Signal),
					Line:   s.Line,
					Column: s.Column,
				})
			}
		}
		if s.Update != "" {
			if def, ok := c.updates[s.Update]; ok {
				s.UpdateResolved = def
			} else {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("undefined update: %s", s.Update),
					Line:   s.Line,
					Column: s.Column,
				})
			}
		}
		if s.Activity != "" {
			if def, ok := c.activities[s.Activity]; ok {
				s.ActivityResolved = def
			} else {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("undefined activity: %s", s.Activity),
					Line:   s.Line,
					Column: s.Column,
				})
			}
		}
		if s.Workflow != "" {
			if def, ok := c.workflows[s.Workflow]; ok {
				s.WorkflowResolved = def
			} else {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("undefined workflow: %s", s.Workflow),
					Line:   s.Line,
					Column: s.Column,
				})
			}
		}
		if s.Nexus != "" {
			res := c.resolveNexusRef(s.Nexus, s.NexusService, s.NexusOperation, s.NexusDetach, s.NexusResult, s.Line, s.Column)
			s.NexusResolvedEndpoint = res.endpoint
			s.NexusResolvedEndpointNamespace = res.endpointNamespace
			s.NexusResolvedService = res.service
			s.NexusResolvedOperation = res.operation
		}
		if s.Ident != "" {
			_, isPromise := c.promises[s.Ident]
			_, isCondition := c.conditions[s.Ident]
			if !isPromise && !isCondition {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("undefined promise or condition: %s", s.Ident),
					Line:   s.Line,
					Column: s.Column,
				})
			}
			// Conditions cannot have result bindings
			if isCondition && s.IdentResult != "" {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("condition %q cannot have a result binding (-> identifier)", s.Ident),
					Line:   s.Line,
					Column: s.Column,
				})
			}
		}

	case *ast.PromiseStmt:
		// Resolve the async target references
		if s.Activity != "" {
			if _, ok := c.activities[s.Activity]; !ok {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("undefined activity: %s", s.Activity),
					Line:   s.Line,
					Column: s.Column,
				})
			}
		}
		if s.Workflow != "" {
			if _, ok := c.workflows[s.Workflow]; !ok {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("undefined workflow: %s", s.Workflow),
					Line:   s.Line,
					Column: s.Column,
				})
			}
		}
		if s.Signal != "" {
			if _, ok := c.signals[s.Signal]; !ok {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("undefined signal: %s", s.Signal),
					Line:   s.Line,
					Column: s.Column,
				})
			}
		}
		if s.Update != "" {
			if _, ok := c.updates[s.Update]; !ok {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("undefined update: %s", s.Update),
					Line:   s.Line,
					Column: s.Column,
				})
			}
		}
		if s.Nexus != "" {
			res := c.resolveNexusRef(s.Nexus, s.NexusService, s.NexusOperation, false, "", s.Line, s.Column)
			s.NexusResolvedEndpoint = res.endpoint
			s.NexusResolvedEndpointNamespace = res.endpointNamespace
			s.NexusResolvedService = res.service
			s.NexusResolvedOperation = res.operation
		}

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
func (c *resolveCtx) resolveNexusRef(endpoint, service, operation string, detach bool, result string, line, column int) nexusResolution {
	var res nexusResolution

	// Detach + result is invalid.
	if detach && result != "" {
		c.errs = append(c.errs, &ResolveError{
			Msg:    "detach nexus call cannot have a result (-> identifier)",
			Line:   line,
			Column: column,
		})
	}

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
			} else {
				// Check endpoint→task_queue→worker→service linkage.
				c.checkEndpointServiceLinkage(endpoint, service, line, column)
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

// checkCallRouting validates that an activity or workflow call can reach its target
// via task queue routing. Two cases:
//   - Explicit task_queue option: target type must be on a worker polling that queue.
//   - No task_queue option (implicit inheritance): target type must be on workers
//     that share a task queue with the calling workflow.
func (c *resolveCtx) checkCallRouting(kind, targetName string, opts *ast.OptionsBlock, line, column int) {
	// Only validate when deployment topology exists.
	if len(c.namespaces) == 0 {
		return
	}

	explicitTQ := extractTaskQueue(opts)

	if explicitTQ != "" {
		// Explicit routing: check the target type is on a worker polling that queue.
		if c.typeOnQueue(kind, targetName, explicitTQ) {
			return
		}
		c.errs = append(c.errs, &ResolveError{
			Msg:    fmt.Sprintf("%s %s has task_queue %q, but no worker on that queue registers it", kind, targetName, explicitTQ),
			Line:   line,
			Column: column,
		})
		return
	}

	// Implicit routing: the call inherits the calling workflow's task queue.
	// Find all task queues that the calling workflow is instantiated on.
	if c.callingWorkflow == "" {
		return
	}
	callerQueues := c.taskQueuesForType("workflow", c.callingWorkflow)
	if len(callerQueues) == 0 {
		// Calling workflow not on any worker — already warned about in coverage checks.
		return
	}

	for _, tq := range callerQueues {
		if !c.typeOnQueue(kind, targetName, tq) {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("%s %s is not on any worker polling task queue %q (inherited from workflow %s)", kind, targetName, tq, c.callingWorkflow),
				Line:   line,
				Column: column,
			})
		}
	}
}

// typeOnQueue checks if a workflow or activity is registered on any worker
// instantiated on the given task queue.
func (c *resolveCtx) typeOnQueue(kind, name, taskQueue string) bool {
	for _, ns := range c.namespaces {
		for _, nw := range ns.Workers {
			nwTQ := extractTaskQueue(nw.Options)
			if nwTQ != taskQueue {
				continue
			}
			w, ok := c.workers[nw.WorkerName]
			if !ok {
				continue
			}
			switch kind {
			case "activity":
				for _, ref := range w.Activities {
					if ref.Name == name {
						return true
					}
				}
			case "workflow":
				for _, ref := range w.Workflows {
					if ref.Name == name {
						return true
					}
				}
			}
		}
	}
	return false
}

// taskQueuesForType returns all task queues that a given workflow or activity
// is instantiated on across all namespaces.
func (c *resolveCtx) taskQueuesForType(kind, name string) []string {
	seen := make(map[string]bool)
	var queues []string
	for _, ns := range c.namespaces {
		for _, nw := range ns.Workers {
			w, ok := c.workers[nw.WorkerName]
			if !ok {
				continue
			}
			var found bool
			switch kind {
			case "workflow":
				for _, ref := range w.Workflows {
					if ref.Name == name {
						found = true
						break
					}
				}
			case "activity":
				for _, ref := range w.Activities {
					if ref.Name == name {
						found = true
						break
					}
				}
			}
			if found {
				tq := extractTaskQueue(nw.Options)
				if tq != "" && !seen[tq] {
					seen[tq] = true
					queues = append(queues, tq)
				}
			}
		}
	}
	return queues
}

// checkEndpointServiceLinkage verifies that the endpoint's task queue has a worker
// that registers the given service.
func (c *resolveCtx) checkEndpointServiceLinkage(endpoint, service string, line, column int) {
	epInfo, ok := c.allEndpoints[endpoint]
	if !ok {
		return // endpoint not found — already reported
	}
	tq := extractTaskQueue(epInfo.endpoint.Options)
	if tq == "" {
		return // missing task_queue — already reported in Pass 3
	}

	// Find all workers instantiated on this task queue across all namespaces.
	for _, ns := range c.namespaces {
		for _, nw := range ns.Workers {
			nwTQ := extractTaskQueue(nw.Options)
			if nwTQ != tq {
				continue
			}
			w, ok := c.workers[nw.WorkerName]
			if !ok {
				continue
			}
			for _, ref := range w.Services {
				if ref.Name == service {
					return // found a worker on the right queue with this service
				}
			}
		}
	}

	c.errs = append(c.errs, &ResolveError{
		Msg:    fmt.Sprintf("nexus endpoint %s routes to task queue %q, but no worker on that queue has service %s", endpoint, tq, service),
		Line:   line,
		Column: column,
	})
}

func (c *resolveCtx) resolveAwaitOneCase(awaitCase *ast.AwaitOneCase) {
	// Resolve signal/update/activity/workflow references
	if awaitCase.Signal != "" {
		if def, ok := c.signals[awaitCase.Signal]; ok {
			awaitCase.SignalResolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined signal: %s", awaitCase.Signal),
				Line:   awaitCase.Line,
				Column: awaitCase.Column,
			})
		}
	}
	if awaitCase.Update != "" {
		if def, ok := c.updates[awaitCase.Update]; ok {
			awaitCase.UpdateResolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined update: %s", awaitCase.Update),
				Line:   awaitCase.Line,
				Column: awaitCase.Column,
			})
		}
	}
	if awaitCase.Activity != "" {
		if def, ok := c.activities[awaitCase.Activity]; ok {
			awaitCase.ActivityResolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined activity: %s", awaitCase.Activity),
				Line:   awaitCase.Line,
				Column: awaitCase.Column,
			})
		}
	}
	if awaitCase.Workflow != "" {
		if def, ok := c.workflows[awaitCase.Workflow]; ok {
			awaitCase.WorkflowResolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined workflow: %s", awaitCase.Workflow),
				Line:   awaitCase.Line,
				Column: awaitCase.Column,
			})
		}
	}

	if awaitCase.Nexus != "" {
		res := c.resolveNexusRef(awaitCase.Nexus, awaitCase.NexusService, awaitCase.NexusOperation, awaitCase.NexusDetach, awaitCase.NexusResult, awaitCase.Line, awaitCase.Column)
		awaitCase.NexusResolvedEndpoint = res.endpoint
		awaitCase.NexusResolvedEndpointNamespace = res.endpointNamespace
		awaitCase.NexusResolvedService = res.service
		awaitCase.NexusResolvedOperation = res.operation
	}

	if awaitCase.Ident != "" {
		_, isPromise := c.promises[awaitCase.Ident]
		_, isCondition := c.conditions[awaitCase.Ident]
		if !isPromise && !isCondition {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined promise or condition: %s", awaitCase.Ident),
				Line:   awaitCase.Line,
				Column: awaitCase.Column,
			})
		}
		// Conditions cannot have result bindings
		if isCondition && awaitCase.IdentResult != "" {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("condition %q cannot have a result binding (-> identifier)", awaitCase.Ident),
				Line:   awaitCase.Line,
				Column: awaitCase.Column,
			})
		}
	}

	// Resolve nested await all block if present.
	if awaitCase.AwaitAll != nil {
		c.resolveStatements(awaitCase.AwaitAll.Body)
	}
	// Resolve the case body.
	c.resolveStatements(awaitCase.Body)
}
