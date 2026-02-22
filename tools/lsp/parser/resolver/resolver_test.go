package resolver

import (
	"strings"
	"testing"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/parser"
)

func mustParse(t *testing.T, input string) *ast.File {
	t.Helper()
	file, err := parser.ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	return file
}

func TestSuccessfulResolution(t *testing.T) {
	input := `workflow OrderWorkflow(orderId: string) -> (OrderResult):
    signal PaymentReceived(txId: string):
        status = "paid"
    update ChangeAddress(addr: Address) -> (Result):
        address = addr
        return Result{ok: true}

    activity GetOrder(orderId) -> order
    workflow ShipOrder(order) -> shipResult

activity GetOrder(orderId: string) -> (Order):
    return db.get(orderId)

workflow ShipOrder(order: Order) -> (ShipResult):
    return ship(order)
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("unexpected error: %v", e)
		}
		t.FailNow()
	}

	// Verify resolution links.
	wf := file.Definitions[0].(*ast.WorkflowDef)
	actCall := wf.Body[0].(*ast.ActivityCall)
	if actCall.Activity.Resolved == nil {
		t.Error("activity call not resolved")
	} else if actCall.Activity.Resolved.Name != "GetOrder" {
		t.Errorf("activity resolved to %q, expected 'GetOrder'", actCall.Activity.Resolved.Name)
	}

	wfCall := wf.Body[1].(*ast.WorkflowCall)
	if wfCall.Workflow.Resolved == nil {
		t.Error("workflow call not resolved")
	} else if wfCall.Workflow.Resolved.Name != "ShipOrder" {
		t.Errorf("workflow resolved to %q, expected 'ShipOrder'", wfCall.Workflow.Resolved.Name)
	}
}

func TestUndefinedActivity(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    activity NonExistent(x) -> y
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Msg, "undefined activity: NonExistent") {
		t.Errorf("unexpected error: %q", errs[0].Msg)
	}
}

func TestUndefinedWorkflow(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    workflow Missing(x) -> y
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Msg, "undefined workflow: Missing") {
		t.Errorf("unexpected error: %q", errs[0].Msg)
	}
}

// REMOVED: TestUndefinedSignal - hint statements are no longer supported.
// REMOVED: TestUndefinedUpdate - hint statements are no longer supported.

func TestDuplicateWorkflow(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    return x

workflow Foo(y: int) -> (Result):
    return y
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Msg, "duplicate workflow definition: Foo") {
		t.Errorf("unexpected error: %q", errs[0].Msg)
	}
}

func TestDuplicateActivity(t *testing.T) {
	input := `activity Foo(x: int) -> (Result):
    return x

activity Foo(y: int) -> (Result):
    return y
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Msg, "duplicate activity definition: Foo") {
		t.Errorf("unexpected error: %q", errs[0].Msg)
	}
}

func TestNestedResolution(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    if (x > 0):
        await all:
            activity Bar(x)
            workflow Baz(x) -> y

activity Bar(x: int) -> (int):
    return x

workflow Baz(x: int) -> (int):
    return x
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("unexpected error: %v", e)
		}
	}
}

// REMOVED: TestSelectCaseResolution - workflow/activity cases in await one are no longer supported.
// await one now only supports timer and nested await all cases.

// REMOVED: TestSelectCaseUndefinedWorkflow - workflow cases in await one are no longer supported.
// await one now only supports timer and nested await all cases.

func TestResolutionInsideForLoop(t *testing.T) {
	input := `workflow Foo(items: []string) -> (Result):
    for (item in items):
        activity Process(item) -> result

activity Process(item: string) -> (Result):
    return item
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("unexpected error: %v", e)
		}
	}
}

func TestResolutionInsideSwitchCase(t *testing.T) {
	input := `workflow Foo(x: string) -> (Result):
    switch (x):
        case "a":
            activity DoA(x)
        else:
            activity DoB(x)

activity DoA(x: string) -> (string):
    return x

activity DoB(x: string) -> (string):
    return x
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("unexpected error: %v", e)
		}
	}
}

func TestMultipleUndefinedErrors(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    activity Missing1(x)
    activity Missing2(x)
    workflow Missing3(x)
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(errs))
	}
}

// REMOVED: TestHintResolution - hint statements are no longer supported.
// REMOVED: TestHintUndefinedSignal - hint statements are no longer supported.

func TestStructuredOptionsResolution(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    activity Bar(x) -> y
        options:
            start_to_close_timeout: 30s
            retry_policy:
                maximum_attempts: 3
                initial_interval: 1s

    workflow Child(y) -> z
        options:
            workflow_run_timeout: 1h

    return z

activity Bar(x: int) -> (int):
    return x

workflow Child(y: int) -> (int):
    return y
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("unexpected error: %v", e)
		}
	}

	// Verify resolution links still work with structured options.
	wf := file.Definitions[0].(*ast.WorkflowDef)
	actCall := wf.Body[0].(*ast.ActivityCall)
	if actCall.Activity.Resolved == nil {
		t.Error("activity call not resolved")
	}
	if actCall.Options == nil || len(actCall.Options.Entries) != 2 {
		t.Error("expected 2 option entries on activity call")
	}
	wfCall := wf.Body[1].(*ast.WorkflowCall)
	if wfCall.Workflow.Resolved == nil {
		t.Error("workflow call not resolved")
	}
}

func TestWorkerResolution(t *testing.T) {
	input := `workflow ProcessOrder(orderId: string) -> (Result):
    activity ChargePayment(orderId) -> payment
    return Result{payment: payment}

activity ChargePayment(orderId: string) -> (Payment):
    return charge(orderId)

worker orderWorker:
    workflow ProcessOrder
    activity ChargePayment

namespace orders:
    worker orderWorker
        options:
            task_queue: "orderProcessing"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("unexpected error: %v", e)
		}
	}
}

func TestWorkerUndefinedWorkflow(t *testing.T) {
	input := `activity ChargePayment(orderId: string) -> (Payment):
    return charge(orderId)

worker badWorker:
    workflow NonExistent
    activity ChargePayment
`
	file := mustParse(t, input)
	errs := Resolve(file)
	var found bool
	for _, e := range errs {
		if strings.Contains(e.Msg, "undefined workflow: NonExistent") {
			found = true
		}
	}
	if !found {
		t.Error("expected error about undefined workflow NonExistent")
	}
}

func TestWorkerUndefinedActivity(t *testing.T) {
	input := `workflow ProcessOrder(orderId: string) -> (Result):
    return Result{}

worker badWorker:
    workflow ProcessOrder
    activity NonExistent
`
	file := mustParse(t, input)
	errs := Resolve(file)
	var found bool
	for _, e := range errs {
		if strings.Contains(e.Msg, "undefined activity: NonExistent") {
			found = true
		}
	}
	if !found {
		t.Error("expected error about undefined activity NonExistent")
	}
}

func TestDuplicateWorker(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    return x

worker myWorker:
    workflow Foo

worker myWorker:
    workflow Foo
`
	file := mustParse(t, input)
	errs := Resolve(file)
	var found bool
	for _, e := range errs {
		if strings.Contains(e.Msg, "duplicate worker definition: myWorker") {
			found = true
		}
	}
	if !found {
		t.Error("expected error about duplicate worker definition")
	}
}

func TestDuplicateNamespace(t *testing.T) {
	input := `worker w:
    workflow Foo

workflow Foo(x: int) -> (int):
    return x

namespace myNs:
    worker w
        options:
            task_queue: "q"

namespace myNs:
    worker w
        options:
            task_queue: "q"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	var found bool
	for _, e := range errs {
		if strings.Contains(e.Msg, "duplicate namespace definition: myNs") {
			found = true
		}
	}
	if !found {
		t.Error("expected error about duplicate namespace definition")
	}
}

func TestNamespaceUndefinedWorker(t *testing.T) {
	input := `namespace orders:
    worker nonExistent
        options:
            task_queue: "q"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	var found bool
	for _, e := range errs {
		if strings.Contains(e.Msg, "undefined worker: nonExistent") {
			found = true
		}
	}
	if !found {
		t.Error("expected error about undefined worker nonExistent")
	}
}

func TestNamespaceResolution(t *testing.T) {
	input := `workflow ProcessOrder(orderId: string) -> (Result):
    activity ChargePayment(orderId) -> payment
    return Result{payment: payment}

activity ChargePayment(orderId: string) -> (Payment):
    return charge(orderId)

worker orderWorker:
    workflow ProcessOrder
    activity ChargePayment

namespace orders:
    worker orderWorker
        options:
            task_queue: "orderProcessing"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("unexpected error: %v", e)
		}
	}
}

func TestHandlerBodyResolution(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    signal Cancel(reason: string):
        activity LogCancel(reason)
    update ChangeAddr(addr: Addr) -> (Result):
        activity ValidateAddr(addr) -> valid
        return Result{ok: valid}

    return Result{}

activity LogCancel(reason: string) -> (int):
    return 0

activity ValidateAddr(addr: string) -> (bool):
    return true
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("unexpected error: %v", e)
		}
		t.FailNow()
	}

	wf := file.Definitions[0].(*ast.WorkflowDef)

	// Check signal body resolution
	sigBody := wf.Signals[0].Body[0].(*ast.ActivityCall)
	if sigBody.Activity.Resolved == nil {
		t.Error("signal body activity call not resolved")
	} else if sigBody.Activity.Resolved.Name != "LogCancel" {
		t.Errorf("resolved to %q, expected 'LogCancel'", sigBody.Activity.Resolved.Name)
	}

	// Check update body resolution
	updBody := wf.Updates[0].Body[0].(*ast.ActivityCall)
	if updBody.Activity.Resolved == nil {
		t.Error("update body activity call not resolved")
	} else if updBody.Activity.Resolved.Name != "ValidateAddr" {
		t.Errorf("resolved to %q, expected 'ValidateAddr'", updBody.Activity.Resolved.Name)
	}
}

// ===== NEXUS RESOLVER TESTS =====

func TestNexusResolutionSuccess(t *testing.T) {
	input := `nexus service OrderService:
    async PlaceOrder workflow ProcessOrder

workflow ProcessOrder(order: Order) -> (Result):
    close complete(Result{})

workflow Caller():
    nexus OrderEndpoint OrderService.PlaceOrder(order) -> result
    close complete(result)

activity Dummy():
    pass()

worker w:
    workflow ProcessOrder
    workflow Caller
    activity Dummy
    nexus service OrderService

namespace ns:
    worker w
        options:
            task_queue: "q"
    nexus endpoint OrderEndpoint
        options:
            task_queue: "q"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	for _, e := range errs {
		t.Errorf("unexpected error: %v (severity: %s)", e, e.Severity)
	}

	// Verify resolution links on NexusCall.
	caller := file.Definitions[2].(*ast.WorkflowDef)
	call := caller.Body[0].(*ast.NexusCall)
	if call.Service.Resolved == nil {
		t.Error("nexus call service not resolved")
	} else if call.Service.Resolved.Name != "OrderService" {
		t.Errorf("resolved service %q, expected 'OrderService'", call.Service.Resolved.Name)
	}
	if call.Operation.Resolved == nil {
		t.Error("nexus call operation not resolved")
	} else if call.Operation.Resolved.Name != "PlaceOrder" {
		t.Errorf("resolved operation %q, expected 'PlaceOrder'", call.Operation.Resolved.Name)
	}
	if call.Endpoint.Resolved == nil {
		t.Error("nexus call endpoint not resolved")
	} else if call.Endpoint.Resolved.EndpointName != "OrderEndpoint" {
		t.Errorf("resolved endpoint %q, expected 'OrderEndpoint'", call.Endpoint.Resolved.EndpointName)
	}
	if call.Endpoint.Resolved.Namespace != "ns" {
		t.Errorf("resolved endpoint namespace %q, expected 'ns'", call.Endpoint.Resolved.Namespace)
	}
}

func TestNexusDuplicateService(t *testing.T) {
	input := `nexus service Svc:
    async Op workflow W

nexus service Svc:
    async Op workflow W

workflow W():
    close complete(Result{})
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if !hasError(errs, "duplicate nexus service definition: Svc") {
		t.Error("expected error about duplicate nexus service definition")
	}
}

func TestNexusDuplicateEndpoint(t *testing.T) {
	input := `worker w:
    workflow W

workflow W():
    close complete(Result{})

namespace ns1:
    worker w
        options:
            task_queue: "q1"
    nexus endpoint Ep
        options:
            task_queue: "q1"

namespace ns2:
    worker w
        options:
            task_queue: "q2"
    nexus endpoint Ep
        options:
            task_queue: "q2"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if !hasError(errs, "duplicate nexus endpoint name") {
		t.Error("expected error about duplicate nexus endpoint name")
	}
}

func TestNexusUndefinedEndpoint(t *testing.T) {
	input := `nexus service Svc:
    async Op workflow W

workflow W():
    nexus MissingEndpoint Svc.Op(x) -> result
    close complete(result)

worker w:
    workflow W
    nexus service Svc

namespace ns:
    worker w
        options:
            task_queue: "q"
    nexus endpoint RealEndpoint
        options:
            task_queue: "q"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if !hasError(errs, "undefined nexus endpoint: MissingEndpoint") {
		t.Error("expected error about undefined nexus endpoint")
	}
}

func TestNexusUndefinedService(t *testing.T) {
	input := `nexus service RealService:
    async Op workflow W

workflow W():
    nexus Ep MissingService.Op(x) -> result
    close complete(result)

worker w:
    workflow W
    nexus service RealService

namespace ns:
    worker w
        options:
            task_queue: "q"
    nexus endpoint Ep
        options:
            task_queue: "q"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if !hasError(errs, "undefined nexus service: MissingService") {
		t.Error("expected error about undefined nexus service")
	}
}

func TestNexusUndefinedOperation(t *testing.T) {
	input := `nexus service Svc:
    async RealOp workflow W

workflow W():
    nexus Ep Svc.MissingOp(x) -> result
    close complete(result)

worker w:
    workflow W
    nexus service Svc

namespace ns:
    worker w
        options:
            task_queue: "q"
    nexus endpoint Ep
        options:
            task_queue: "q"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if !hasError(errs, "nexus service Svc has no operation MissingOp") {
		t.Error("expected error about missing operation")
	}
}

func TestNexusDetachWithResult(t *testing.T) {
	// The parser itself rejects detach nexus with a result binding.
	input := `workflow W():
    detach nexus Ep Svc.Op(x) -> result
`
	_, err := parser.ParseFile(input)
	if err == nil {
		t.Fatal("expected parse error for detach nexus with result")
	}
	if !strings.Contains(err.Error(), "detach nexus call cannot have a result") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNexusAsyncUndefinedWorkflow(t *testing.T) {
	input := `nexus service Svc:
    async Op workflow MissingWorkflow

workflow W():
    close complete(Result{})
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if !hasError(errs, "async operation Op references undefined workflow: MissingWorkflow") {
		t.Error("expected error about async op referencing undefined workflow")
	}
}

func TestNexusWorkerUndefinedService(t *testing.T) {
	input := `workflow W():
    close complete(Result{})

worker w:
    workflow W
    nexus service MissingService
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if !hasError(errs, "undefined nexus service: MissingService") {
		t.Error("expected error about worker referencing undefined nexus service")
	}
}

func TestNexusCallNoEndpointsDefined(t *testing.T) {
	input := `workflow W():
    nexus Ep Svc.Op(x) -> result
    close complete(result)
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if !hasWarning(errs, "unresolved nexus endpoint: Ep") {
		t.Error("expected warning about unresolved endpoint (no endpoints defined)")
	}
}

func TestNexusCallNoServicesDefined(t *testing.T) {
	input := `workflow W():
    nexus Ep Svc.Op(x) -> result
    close complete(result)
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if !hasWarning(errs, "unresolved nexus service: Svc") {
		t.Error("expected warning about unresolved service (no services defined)")
	}
}

// ===== PRIORITY SCHEMA TESTS =====

func TestPriorityNestedSchema(t *testing.T) {
	input := `workflow Foo(x: int) -> (int):
    activity Bar(x) -> y
        options:
            priority:
                priority_key: 1
                fairness_key: "high"
                fairness_weight: 9.0
    return y

activity Bar(x: int) -> (int):
    return x
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("unexpected error: %v", e)
		}
	}
}

func TestNexusEndpointResolutionOnAwaitAndPromise(t *testing.T) {
	input := `nexus service Svc:
    async Op workflow W

workflow W():
    close complete(Result{})

workflow Caller():
    await nexus Ep Svc.Op(x) -> result
    promise p <- nexus Ep Svc.Op(x)
    close complete(result)

activity Dummy():
    pass()

worker w:
    workflow W
    workflow Caller
    activity Dummy
    nexus service Svc

namespace ns:
    worker w
        options:
            task_queue: "q"
    nexus endpoint Ep
        options:
            task_queue: "q"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	for _, e := range errs {
		t.Errorf("unexpected error: %v (severity: %s)", e, e.Severity)
	}

	caller := file.Definitions[2].(*ast.WorkflowDef)

	// Check await nexus resolution.
	awaitStmt := caller.Body[0].(*ast.AwaitStmt)
	awaitNexus, ok := awaitStmt.Target.(*ast.NexusTarget)
	if !ok {
		t.Fatalf("expected NexusTarget, got %T", awaitStmt.Target)
	}
	if awaitNexus.Endpoint.Resolved == nil {
		t.Error("await nexus endpoint not resolved")
	} else if awaitNexus.Endpoint.Resolved.EndpointName != "Ep" {
		t.Errorf("await resolved endpoint %q, expected 'Ep'", awaitNexus.Endpoint.Resolved.EndpointName)
	}
	if awaitNexus.Service.Resolved == nil {
		t.Error("await nexus service not resolved")
	}
	if awaitNexus.Operation.Resolved == nil {
		t.Error("await nexus operation not resolved")
	}
	if awaitNexus.Endpoint.Resolved.Namespace != "ns" {
		t.Errorf("await resolved endpoint namespace %q, expected 'ns'", awaitNexus.Endpoint.Resolved.Namespace)
	}

	// Check promise nexus resolution.
	promiseStmt := caller.Body[1].(*ast.PromiseStmt)
	promiseNexus, ok := promiseStmt.Target.(*ast.NexusTarget)
	if !ok {
		t.Fatalf("expected NexusTarget, got %T", promiseStmt.Target)
	}
	if promiseNexus.Endpoint.Resolved == nil {
		t.Error("promise nexus endpoint not resolved")
	} else if promiseNexus.Endpoint.Resolved.EndpointName != "Ep" {
		t.Errorf("promise resolved endpoint %q, expected 'Ep'", promiseNexus.Endpoint.Resolved.EndpointName)
	}
	if promiseNexus.Service.Resolved == nil {
		t.Error("promise nexus service not resolved")
	}
	if promiseNexus.Operation.Resolved == nil {
		t.Error("promise nexus operation not resolved")
	}
}

func TestWorkerRefResolution(t *testing.T) {
	input := `workflow ProcessOrder(orderId: string) -> (Result):
    activity ChargePayment(orderId) -> payment
    return Result{payment: payment}

activity ChargePayment(orderId: string) -> (Payment):
    return charge(orderId)

nexus service OrderService:
    async PlaceOrder workflow ProcessOrder

worker orderWorker:
    workflow ProcessOrder
    activity ChargePayment
    nexus service OrderService

namespace orders:
    worker orderWorker
        options:
            task_queue: "orderProcessing"
    nexus endpoint OrderEndpoint
        options:
            task_queue: "orderProcessing"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("unexpected error: %v", e)
		}
	}

	// Verify WorkerRef resolution links.
	worker := file.Definitions[3].(*ast.WorkerDef)
	if worker.Name != "orderWorker" {
		t.Fatalf("expected worker 'orderWorker', got %q", worker.Name)
	}

	// Workflow ref should resolve to ProcessOrder.
	if len(worker.Workflows) != 1 {
		t.Fatalf("expected 1 workflow ref, got %d", len(worker.Workflows))
	}
	wfRef := worker.Workflows[0]
	if wfRef.Resolved == nil {
		t.Error("workflow ref not resolved")
	} else if wfRef.Resolved.Name != "ProcessOrder" {
		t.Errorf("workflow ref resolved to %q, expected 'ProcessOrder'", wfRef.Resolved.Name)
	}

	// Activity ref should resolve to ChargePayment.
	if len(worker.Activities) != 1 {
		t.Fatalf("expected 1 activity ref, got %d", len(worker.Activities))
	}
	actRef := worker.Activities[0]
	if actRef.Resolved == nil {
		t.Error("activity ref not resolved")
	} else if actRef.Resolved.Name != "ChargePayment" {
		t.Errorf("activity ref resolved to %q, expected 'ChargePayment'", actRef.Resolved.Name)
	}

	// Service ref should resolve to OrderService.
	if len(worker.Services) != 1 {
		t.Fatalf("expected 1 service ref, got %d", len(worker.Services))
	}
	svcRef := worker.Services[0]
	if svcRef.Resolved == nil {
		t.Error("service ref not resolved")
	} else if svcRef.Resolved.Name != "OrderService" {
		t.Errorf("service ref resolved to %q, expected 'OrderService'", svcRef.Resolved.Name)
	}
}

func TestNamespaceWorkerResolution(t *testing.T) {
	input := `workflow Foo(x: int) -> (int):
    return x

worker myWorker:
    workflow Foo

namespace myNs:
    worker myWorker
        options:
            task_queue: "q"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("unexpected error: %v", e)
		}
	}

	// Verify NamespaceWorker resolution link.
	ns := file.Definitions[2].(*ast.NamespaceDef)
	if ns.Name != "myNs" {
		t.Fatalf("expected namespace 'myNs', got %q", ns.Name)
	}
	if len(ns.Workers) != 1 {
		t.Fatalf("expected 1 worker in namespace, got %d", len(ns.Workers))
	}

	nw := ns.Workers[0]
	if nw.Worker.Resolved == nil {
		t.Error("namespace worker not resolved")
	} else if nw.Worker.Resolved.Name != "myWorker" {
		t.Errorf("namespace worker resolved to %q, expected 'myWorker'", nw.Worker.Resolved.Name)
	}
}

// hasError checks if any non-warning error contains the given substring.
func hasError(errs []*ResolveError, substr string) bool {
	for _, e := range errs {
		if e.Severity != "warning" && strings.Contains(e.Msg, substr) {
			return true
		}
	}
	return false
}

// hasWarning checks if any warning contains the given substring.
func hasWarning(errs []*ResolveError, substr string) bool {
	for _, e := range errs {
		if e.Severity == "warning" && strings.Contains(e.Msg, substr) {
			return true
		}
	}
	return false
}
