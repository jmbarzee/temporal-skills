package resolver

import (
	"fmt"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
)

// ErrorKind classifies a resolve error for structured handling.
type ErrorKind int

const (
	// --- Duplicate definition errors ---

	// ErrDuplicateWorkflow: a workflow name appears more than once.
	ErrDuplicateWorkflow ErrorKind = iota + 1
	// ErrDuplicateActivity: an activity name appears more than once.
	ErrDuplicateActivity
	// ErrDuplicateWorker: a worker name appears more than once.
	ErrDuplicateWorker
	// ErrDuplicateNamespace: a namespace name appears more than once.
	ErrDuplicateNamespace
	// ErrDuplicateNexusService: a nexus service name appears more than once.
	ErrDuplicateNexusService
	// ErrDuplicateEndpoint: a nexus endpoint name appears in more than one namespace.
	ErrDuplicateEndpoint

	// --- Undefined reference errors ---

	// ErrUndefinedActivity: an activity call references a name with no definition.
	ErrUndefinedActivity
	// ErrUndefinedWorkflow: a workflow call references a name with no definition.
	ErrUndefinedWorkflow
	// ErrUndefinedSignal: an await/promise target references an undefined signal.
	ErrUndefinedSignal
	// ErrUndefinedUpdate: an await/promise target references an undefined update.
	ErrUndefinedUpdate
	// ErrUndefinedCondition: a set/unset statement references an undefined condition.
	ErrUndefinedCondition
	// ErrUndefinedPromiseOrCondition: an ident target matches neither a promise nor a condition.
	ErrUndefinedPromiseOrCondition
	// ErrConditionResultBinding: a condition target has a result binding (-> identifier), which is not allowed.
	ErrConditionResultBinding

	// --- Nexus resolution errors ---

	// ErrNexusAsyncUndefinedWorkflow: an async nexus operation references an undefined workflow.
	ErrNexusAsyncUndefinedWorkflow
	// ErrNexusUndefinedEndpoint: a nexus call references an endpoint not defined in any namespace.
	ErrNexusUndefinedEndpoint
	// ErrNexusUnresolvedEndpoint: no namespaces define any endpoints (may be external). Warning severity.
	ErrNexusUnresolvedEndpoint
	// ErrNexusUndefinedService: a nexus call references a service name with no definition.
	ErrNexusUndefinedService
	// ErrNexusUnresolvedService: no nexus services are defined (may be external). Warning severity.
	ErrNexusUnresolvedService
	// ErrNexusNoOperation: a nexus call references an operation not found on the resolved service.
	ErrNexusNoOperation

	// --- Worker reference errors ---

	// ErrWorkerUndefinedWorkflow: a worker registers an undefined workflow.
	ErrWorkerUndefinedWorkflow
	// ErrWorkerUndefinedActivity: a worker registers an undefined activity.
	ErrWorkerUndefinedActivity
	// ErrWorkerUndefinedNexusService: a worker registers an undefined nexus service.
	ErrWorkerUndefinedNexusService

	// --- Namespace reference errors ---

	// ErrNamespaceUndefinedWorker: a namespace references an undefined worker.
	ErrNamespaceUndefinedWorker
)

// ResolveError represents a resolution error with position info.
type ResolveError struct {
	Msg      string
	Line     int
	Column   int
	Severity string // "error" (default) or "warning"
	Kind     ErrorKind
	Name     string // primary entity referenced by this error
}

func (e *ResolveError) Error() string {
	return fmt.Sprintf("resolve error at %d:%d: %s", e.Line, e.Column, e.Msg)
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
			collectDef(workflows, d.Name, d, "workflow", ErrDuplicateWorkflow, d.Line, d.Column, &errs)
		case *ast.ActivityDef:
			collectDef(activities, d.Name, d, "activity", ErrDuplicateActivity, d.Line, d.Column, &errs)
		case *ast.WorkerDef:
			collectDef(workers, d.Name, d, "worker", ErrDuplicateWorker, d.Line, d.Column, &errs)
		case *ast.NamespaceDef:
			collectDef(namespaces, d.Name, d, "namespace", ErrDuplicateNamespace, d.Line, d.Column, &errs)
		case *ast.NexusServiceDef:
			collectDef(nexusServices, d.Name, d, "nexus service", ErrDuplicateNexusService, d.Line, d.Column, &errs)
		}
	}

	// Build global endpoint map across all namespaces.
	allEndpoints := make(map[string]*ast.NamespaceEndpoint)
	for _, ns := range namespaces {
		for i := range ns.Endpoints {
			ep := &ns.Endpoints[i]
			ep.Namespace = ns.Name
			if existing, exists := allEndpoints[ep.EndpointName]; exists {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("duplicate nexus endpoint name %q: defined in namespace %s and namespace %s", ep.EndpointName, existing.Namespace, ns.Name),
					Line:   ep.Line,
					Column: ep.Column,
					Kind:   ErrDuplicateEndpoint,
					Name:   ep.EndpointName,
				})
			}
			allEndpoints[ep.EndpointName] = ep
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
				if wf, ok := workflows[op.Workflow.Name]; ok {
					op.Workflow.Resolved = wf
				} else {
					errs = append(errs, &ResolveError{
						Msg:    fmt.Sprintf("nexus service %s: async operation %s references undefined workflow: %s", svc.Name, op.Name, op.Workflow.Name),
						Line:   op.Line,
						Column: op.Column,
						Kind:   ErrNexusAsyncUndefinedWorkflow,
						Name:   op.Workflow.Name,
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
		resolveWorkerRefs(w.Workflows, workflows, "workflow", ErrWorkerUndefinedWorkflow, &errs)
		resolveWorkerRefs(w.Activities, activities, "activity", ErrWorkerUndefinedActivity, &errs)
		resolveWorkerRefs(w.Services, nexusServices, "nexus service", ErrWorkerUndefinedNexusService, &errs)
	}

	for _, ns := range namespaces {
		for i := range ns.Workers {
			nw := &ns.Workers[i]
			if def, ok := workers[nw.Worker.Name]; ok {
				nw.Worker.Resolved = def
			} else {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("namespace %s references undefined worker: %s", ns.Name, nw.Worker.Name),
					Line:   nw.Line,
					Column: nw.Column,
					Kind:   ErrNamespaceUndefinedWorker,
					Name:   nw.Worker.Name,
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
	allEndpoints  map[string]*ast.NamespaceEndpoint
	errs          []*ResolveError
}

func (c *resolveCtx) resolveStatements(stmts []ast.Statement) {
	ast.WalkStatements(stmts, func(s ast.Statement) bool {
		switch s := s.(type) {
		case *ast.ActivityCall:
			resolveRef(&s.Activity, c.activities, "activity", ErrUndefinedActivity, &c.errs)
		case *ast.WorkflowCall:
			resolveRef(&s.Workflow, c.workflows, "workflow", ErrUndefinedWorkflow, &c.errs)
		case *ast.NexusCall:
			c.resolveNexusRefs(&s.Endpoint, &s.Service, &s.Operation)
		case *ast.SetStmt:
			resolveRef(&s.Condition, c.conditions, "condition", ErrUndefinedCondition, &c.errs)
		case *ast.UnsetStmt:
			resolveRef(&s.Condition, c.conditions, "condition", ErrUndefinedCondition, &c.errs)
		}
		return true
	}, ast.WithAsyncTargets(func(target ast.AsyncTarget, parent ast.Statement) bool {
		c.resolveAsyncTarget(target, parent.NodeLine(), parent.NodeColumn())
		return true
	}))
}

// resolveNexusRefs validates and resolves a nexus call site's endpoint, service,
// and operation Ref fields.
func (c *resolveCtx) resolveNexusRefs(endpoint *ast.Ref[*ast.NamespaceEndpoint], service *ast.Ref[*ast.NexusServiceDef], operation *ast.Ref[*ast.NexusOperation]) {
	resolveRefWithWarn(endpoint, c.allEndpoints, "endpoint", ErrNexusUndefinedEndpoint, ErrNexusUnresolvedEndpoint, &c.errs)
	if resolveRefWithWarn(service, c.nexusServices, "service", ErrNexusUndefinedService, ErrNexusUnresolvedService, &c.errs) {
		c.resolveNexusOperation(service.Resolved, operation)
	}
}

// resolveNexusOperation resolves an operation name against a service's operation list.
func (c *resolveCtx) resolveNexusOperation(svc *ast.NexusServiceDef, operation *ast.Ref[*ast.NexusOperation]) {
	for _, op := range svc.Operations {
		if op.Name == operation.Name {
			operation.Resolved = op
			return
		}
	}
	c.errs = append(c.errs, &ResolveError{
		Msg:    fmt.Sprintf("nexus service %s has no operation %s", svc.Name, operation.Name),
		Line:   operation.Line,
		Column: operation.Column,
		Kind:   ErrNexusNoOperation,
		Name:   operation.Name,
	})
}

// resolveRefWithWarn resolves a Ref against a definition map with special handling
// for the case where no definitions exist (emits a warning instead of an error).
func resolveRefWithWarn[T any](ref *ast.Ref[T], defs map[string]T, kind string, errUndef, errUnresolved ErrorKind, errs *[]*ResolveError) bool {
	if len(defs) == 0 {
		*errs = append(*errs, &ResolveError{
			Msg:      fmt.Sprintf("unresolved nexus %s: %s (no %ss defined — may be external)", kind, ref.Name, kind),
			Severity: "warning",
			Line:     ref.Line,
			Column:   ref.Column,
			Kind:     errUnresolved,
			Name:     ref.Name,
		})
		return false
	}
	if def, ok := defs[ref.Name]; ok {
		ref.Resolved = def
		return true
	}
	*errs = append(*errs, &ResolveError{
		Msg:    fmt.Sprintf("undefined nexus %s: %s", kind, ref.Name),
		Line:   ref.Line,
		Column: ref.Column,
		Kind:   errUndef,
		Name:   ref.Name,
	})
	return false
}

// resolveAsyncTarget resolves references inside an async target.
func (c *resolveCtx) resolveAsyncTarget(target ast.AsyncTarget, line, column int) {
	switch t := target.(type) {
	case *ast.SignalTarget:
		resolveRef(&t.Signal, c.signals, "signal", ErrUndefinedSignal, &c.errs)
	case *ast.UpdateTarget:
		resolveRef(&t.Update, c.updates, "update", ErrUndefinedUpdate, &c.errs)
	case *ast.ActivityTarget:
		resolveRef(&t.Activity, c.activities, "activity", ErrUndefinedActivity, &c.errs)
	case *ast.WorkflowTarget:
		resolveRef(&t.Workflow, c.workflows, "workflow", ErrUndefinedWorkflow, &c.errs)
	case *ast.NexusTarget:
		c.resolveNexusRefs(&t.Endpoint, &t.Service, &t.Operation)
	case *ast.IdentTarget:
		promise, isPromise := c.promises[t.Name]
		condition, isCondition := c.conditions[t.Name]
		if !isPromise && !isCondition {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined promise or condition: %s", t.Name),
				Line:   line,
				Column: column,
				Kind:   ErrUndefinedPromiseOrCondition,
				Name:   t.Name,
			})
		}
		if isPromise {
			t.Resolved.Promise = promise
		}
		if isCondition {
			t.Resolved.Condition = condition
		}
		if isCondition && t.Result != "" {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("condition %q cannot have a result binding (-> identifier)", t.Name),
				Line:   line,
				Column: column,
				Kind:   ErrConditionResultBinding,
				Name:   t.Name,
			})
		}
	case *ast.TimerTarget:
		// No resolution needed for timers
	}
}

// collectDef registers a definition in the map, appending a duplicate error if
// the name already exists.
func collectDef[T any](m map[string]T, name string, def T, kind string, errKind ErrorKind, line, column int, errs *[]*ResolveError) {
	if _, exists := m[name]; exists {
		*errs = append(*errs, &ResolveError{
			Msg:    fmt.Sprintf("duplicate %s definition: %s", kind, name),
			Line:   line,
			Column: column,
			Kind:   errKind,
			Name:   name,
		})
	}
	m[name] = def
}

// resolveRef resolves a single Ref against a definition map, setting Resolved on
// match or appending a ResolveError on miss.
func resolveRef[T any](ref *ast.Ref[T], defs map[string]T, kind string, errKind ErrorKind, errs *[]*ResolveError) {
	if def, ok := defs[ref.Name]; ok {
		ref.Resolved = def
	} else {
		*errs = append(*errs, &ResolveError{
			Msg:    fmt.Sprintf("undefined %s: %s", kind, ref.Name),
			Line:   ref.Line,
			Column: ref.Column,
			Kind:   errKind,
			Name:   ref.Name,
		})
	}
}

// resolveWorkerRefs resolves a slice of worker references against a definition map.
func resolveWorkerRefs[T any](refs []ast.Ref[T], defs map[string]T, kind string, errKind ErrorKind, errs *[]*ResolveError) {
	for i := range refs {
		resolveRef(&refs[i], defs, kind, errKind, errs)
	}
}
