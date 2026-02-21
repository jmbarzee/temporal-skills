package validator

import (
	"fmt"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
)

// ErrorKind classifies a validation error for structured handling.
type ErrorKind int

const (
	ErrEmptyWorkflow          ErrorKind = iota + 1
	ErrEmptyActivity
	ErrEmptyWorker
	ErrEmptyNamespace
	ErrMissingTaskQueue
	ErrMissingEndpointTaskQueue
	ErrUncoveredWorkflow
	ErrUncoveredActivity
	ErrUncoveredService
	ErrUninstantiatedWorker
	ErrTaskQueueIdentical
	ErrTaskQueueMismatch
	ErrExplicitRoutingMismatch
	ErrImplicitRoutingMismatch
	ErrEndpointServiceLinkage
)

// Error represents a validation error with position info.
type Error struct {
	Msg      string
	Line     int
	Column   int
	Severity string // "error" (default) or "warning"
	Kind     ErrorKind
	Name     string // primary entity referenced by this error
}

func (e *Error) Error() string {
	return fmt.Sprintf("validation error at %d:%d: %s", e.Line, e.Column, e.Msg)
}

// endpointInfo tracks which namespace defines a nexus endpoint.
type endpointInfo struct {
	namespaceName string
	endpoint      *ast.NamespaceEndpoint
}

type validationCtx struct {
	workflows     map[string]*ast.WorkflowDef
	activities    map[string]*ast.ActivityDef
	workers       map[string]*ast.WorkerDef
	namespaces    map[string]*ast.NamespaceDef
	nexusServices map[string]*ast.NexusServiceDef
	allEndpoints  map[string]*endpointInfo
	errs          []*Error
}

// Validate runs deployment/routing validation on a resolved AST.
// Call after resolver.Resolve().
func Validate(file *ast.File) []*Error {
	v := &validationCtx{
		workflows:     make(map[string]*ast.WorkflowDef),
		activities:    make(map[string]*ast.ActivityDef),
		workers:       make(map[string]*ast.WorkerDef),
		namespaces:    make(map[string]*ast.NamespaceDef),
		nexusServices: make(map[string]*ast.NexusServiceDef),
		allEndpoints:  make(map[string]*endpointInfo),
	}

	// Build definition maps from the AST.
	for _, def := range file.Definitions {
		switch d := def.(type) {
		case *ast.WorkflowDef:
			v.workflows[d.Name] = d
		case *ast.ActivityDef:
			v.activities[d.Name] = d
		case *ast.WorkerDef:
			v.workers[d.Name] = d
		case *ast.NamespaceDef:
			v.namespaces[d.Name] = d
		case *ast.NexusServiceDef:
			v.nexusServices[d.Name] = d
		}
	}

	// Build global endpoint map.
	for _, ns := range v.namespaces {
		for i := range ns.Endpoints {
			ep := &ns.Endpoints[i]
			v.allEndpoints[ep.EndpointName] = &endpointInfo{
				namespaceName: ns.Name,
				endpoint:      ep,
			}
		}
	}

	// 1. Empty definition warnings.
	v.checkEmptyDefinitions()

	// 2. Task queue requirements.
	v.checkTaskQueueRequirements()

	// 3. Coverage warnings.
	v.checkCoverage()

	// 4. Task queue coherence.
	v.checkTaskQueueCoherence()

	// 5-6. Call routing + endpoint-service linkage (walks resolved bodies).
	v.walkAllBodies()

	return v.errs
}

func (v *validationCtx) checkEmptyDefinitions() {
	for _, wf := range v.workflows {
		if !hasNonCommentStmts(wf.Body) && len(wf.Signals) == 0 && len(wf.Queries) == 0 && len(wf.Updates) == 0 && wf.State == nil {
			v.errs = append(v.errs, &Error{
				Msg:      fmt.Sprintf("workflow %s has an empty body", wf.Name),
				Line:     wf.Line,
				Column:   wf.Column,
				Severity: "warning",
				Kind:     ErrEmptyWorkflow,
				Name:     wf.Name,
			})
		}
	}
	for _, act := range v.activities {
		if !hasNonCommentStmts(act.Body) {
			v.errs = append(v.errs, &Error{
				Msg:      fmt.Sprintf("activity %s has an empty body", act.Name),
				Line:     act.Line,
				Column:   act.Column,
				Severity: "warning",
				Kind:     ErrEmptyActivity,
				Name:     act.Name,
			})
		}
	}
	for _, w := range v.workers {
		if len(w.Workflows) == 0 && len(w.Activities) == 0 && len(w.Services) == 0 {
			v.errs = append(v.errs, &Error{
				Msg:      fmt.Sprintf("worker %s has no workflow, activity, or nexus service registrations", w.Name),
				Line:     w.Line,
				Column:   w.Column,
				Severity: "warning",
				Kind:     ErrEmptyWorker,
				Name:     w.Name,
			})
		}
	}
	for _, ns := range v.namespaces {
		if len(ns.Workers) == 0 && len(ns.Endpoints) == 0 {
			v.errs = append(v.errs, &Error{
				Msg:      fmt.Sprintf("namespace %s has no worker or endpoint instantiations", ns.Name),
				Line:     ns.Line,
				Column:   ns.Column,
				Severity: "warning",
				Kind:     ErrEmptyNamespace,
				Name:     ns.Name,
			})
		}
	}
}

func (v *validationCtx) checkTaskQueueRequirements() {
	for _, ns := range v.namespaces {
		for _, nw := range ns.Workers {
			tq := extractTaskQueue(nw.Options)
			if tq == "" {
				v.errs = append(v.errs, &Error{
					Msg:    fmt.Sprintf("worker %s in namespace %s missing required task_queue option", nw.Worker.Name, ns.Name),
					Line:   nw.Line,
					Column: nw.Column,
					Kind:   ErrMissingTaskQueue,
					Name:   nw.Worker.Name,
				})
			}
		}
		for _, ep := range ns.Endpoints {
			tq := extractTaskQueue(ep.Options)
			if tq == "" {
				v.errs = append(v.errs, &Error{
					Msg:    fmt.Sprintf("nexus endpoint %s in namespace %s missing required task_queue option", ep.EndpointName, ns.Name),
					Line:   ep.Line,
					Column: ep.Column,
					Kind:   ErrMissingEndpointTaskQueue,
					Name:   ep.EndpointName,
				})
			}
		}
	}
}

func (v *validationCtx) checkCoverage() {
	if len(v.namespaces) == 0 {
		return
	}

	coveredWorkflows := make(map[string]bool)
	coveredActivities := make(map[string]bool)
	coveredServices := make(map[string]bool)
	instantiatedWorkers := make(map[string]bool)

	for _, ns := range v.namespaces {
		for _, nw := range ns.Workers {
			instantiatedWorkers[nw.Worker.Name] = true
			if w, ok := v.workers[nw.Worker.Name]; ok {
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

	checkUncovered(v.workflows, coveredWorkflows, "workflow %s is not registered on any instantiated worker", ErrUncoveredWorkflow, &v.errs)
	checkUncovered(v.activities, coveredActivities, "activity %s is not registered on any instantiated worker", ErrUncoveredActivity, &v.errs)
	checkUncovered(v.nexusServices, coveredServices, "nexus service %s is not referenced by any worker", ErrUncoveredService, &v.errs)
	checkUncovered(v.workers, instantiatedWorkers, "worker %s is not instantiated in any namespace", ErrUninstantiatedWorker, &v.errs)
}

func (v *validationCtx) checkTaskQueueCoherence() {
	type queueInfo struct {
		workerName string
		workflows  map[string]bool
		activities map[string]bool
	}
	for _, ns := range v.namespaces {
		queueWorkers := make(map[string][]queueInfo)
		for _, nw := range ns.Workers {
			tq := extractTaskQueue(nw.Options)
			if tq == "" {
				continue
			}
			w, ok := v.workers[nw.Worker.Name]
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
				workerName: nw.Worker.Name,
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
					v.errs = append(v.errs, &Error{
						Msg:      fmt.Sprintf("workers %s and %s on task queue %q in namespace %s have identical type sets (redundant)", first.workerName, other.workerName, queue, ns.Name),
						Severity: "warning",
						Kind:     ErrTaskQueueIdentical,
						Name:     queue,
					})
				} else {
					v.errs = append(v.errs, &Error{
						Msg:  fmt.Sprintf("workers %s and %s on task queue %q in namespace %s have different type sets", first.workerName, other.workerName, queue, ns.Name),
						Kind: ErrTaskQueueMismatch,
						Name: queue,
					})
				}
			}
		}
	}
}

// walkAllBodies walks all workflow and nexus service sync op bodies,
// checking call routing and endpoint-service linkage.
func (v *validationCtx) walkAllBodies() {
	if len(v.namespaces) == 0 {
		return
	}

	for _, wf := range v.workflows {
		v.walkStatements(wf.Body, wf.Name)
		for _, s := range wf.Signals {
			v.walkStatements(s.Body, wf.Name)
		}
		for _, q := range wf.Queries {
			v.walkStatements(q.Body, wf.Name)
		}
		for _, u := range wf.Updates {
			v.walkStatements(u.Body, wf.Name)
		}
	}

	for _, svc := range v.nexusServices {
		for _, op := range svc.Operations {
			if op.OpType == ast.NexusOpSync {
				v.walkStatements(op.Body, "")
			}
		}
	}
}

func (v *validationCtx) walkStatements(stmts []ast.Statement, callingWorkflow string) {
	ast.WalkStatements(stmts, func(s ast.Statement) bool {
		switch n := s.(type) {
		case *ast.ActivityCall:
			v.checkCallRouting("activity", n.Activity.Name, n.Options, callingWorkflow, n.Line, n.Column)
		case *ast.WorkflowCall:
			v.checkCallRouting("workflow", n.Workflow.Name, n.Options, callingWorkflow, n.Line, n.Column)
		case *ast.NexusCall:
			v.checkEndpointServiceLinkage(n.Endpoint, n.Service, n.Line, n.Column)
		case *ast.AwaitStmt:
			v.walkAsyncTarget(n.Target, n.Line, n.Column)
		case *ast.AwaitOneCase:
			if n.Target != nil {
				v.walkAsyncTarget(n.Target, n.Line, n.Column)
			}
		case *ast.PromiseStmt:
			v.walkAsyncTarget(n.Target, n.Line, n.Column)
		}
		return true
	})
}

func (v *validationCtx) walkAsyncTarget(target ast.AsyncTarget, line, column int) {
	if nt, ok := target.(*ast.NexusTarget); ok {
		v.checkEndpointServiceLinkage(nt.Endpoint, nt.Service, line, column)
	}
}

// checkCallRouting validates that an activity or workflow call can reach its target
// via task queue routing.
func (v *validationCtx) checkCallRouting(kind, targetName string, opts *ast.OptionsBlock, callingWorkflow string, line, column int) {
	if len(v.namespaces) == 0 {
		return
	}

	explicitTQ := extractTaskQueue(opts)

	if explicitTQ != "" {
		if v.typeOnQueue(kind, targetName, explicitTQ) {
			return
		}
		v.errs = append(v.errs, &Error{
			Msg:    fmt.Sprintf("%s %s has task_queue %q, but no worker on that queue registers it", kind, targetName, explicitTQ),
			Line:   line,
			Column: column,
			Kind:   ErrExplicitRoutingMismatch,
			Name:   targetName,
		})
		return
	}

	// Implicit routing: the call inherits the calling workflow's task queue.
	if callingWorkflow == "" {
		return
	}
	callerQueues := v.taskQueuesForType("workflow", callingWorkflow)
	if len(callerQueues) == 0 {
		return
	}

	for _, tq := range callerQueues {
		if !v.typeOnQueue(kind, targetName, tq) {
			v.errs = append(v.errs, &Error{
				Msg:    fmt.Sprintf("%s %s is not on any worker polling task queue %q (inherited from workflow %s)", kind, targetName, tq, callingWorkflow),
				Line:   line,
				Column: column,
				Kind:   ErrImplicitRoutingMismatch,
				Name:   targetName,
			})
		}
	}
}

// typeOnQueue checks if a workflow or activity is registered on any worker
// instantiated on the given task queue.
func (v *validationCtx) typeOnQueue(kind, name, taskQueue string) bool {
	for _, ns := range v.namespaces {
		for _, nw := range ns.Workers {
			nwTQ := extractTaskQueue(nw.Options)
			if nwTQ != taskQueue {
				continue
			}
			w, ok := v.workers[nw.Worker.Name]
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
func (v *validationCtx) taskQueuesForType(kind, name string) []string {
	seen := make(map[string]bool)
	var queues []string
	for _, ns := range v.namespaces {
		for _, nw := range ns.Workers {
			w, ok := v.workers[nw.Worker.Name]
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
func (v *validationCtx) checkEndpointServiceLinkage(endpoint, service string, line, column int) {
	epInfo, ok := v.allEndpoints[endpoint]
	if !ok {
		return // endpoint not found — already reported by resolver
	}
	tq := extractTaskQueue(epInfo.endpoint.Options)
	if tq == "" {
		return // missing task_queue — already reported in checkTaskQueueRequirements
	}

	for _, ns := range v.namespaces {
		for _, nw := range ns.Workers {
			nwTQ := extractTaskQueue(nw.Options)
			if nwTQ != tq {
				continue
			}
			w, ok := v.workers[nw.Worker.Name]
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

	v.errs = append(v.errs, &Error{
		Msg:    fmt.Sprintf("nexus endpoint %s routes to task queue %q, but no worker on that queue has service %s", endpoint, tq, service),
		Line:   line,
		Column: column,
		Kind:   ErrEndpointServiceLinkage,
		Name:   endpoint,
	})
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

// checkUncovered reports a warning for each definition in defs that is not
// present in the covered set.
func checkUncovered[T ast.Node](defs map[string]T, covered map[string]bool, msgFmt string, kind ErrorKind, errs *[]*Error) {
	for name, node := range defs {
		if !covered[name] {
			*errs = append(*errs, &Error{
				Msg:      fmt.Sprintf(msgFmt, name),
				Line:     node.NodeLine(),
				Column:   node.NodeColumn(),
				Severity: "warning",
				Kind:     kind,
				Name:     name,
			})
		}
	}
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
