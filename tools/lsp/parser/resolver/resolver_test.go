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
	if actCall.Resolved == nil {
		t.Error("activity call not resolved")
	} else if actCall.Resolved.Name != "GetOrder" {
		t.Errorf("activity resolved to %q, expected 'GetOrder'", actCall.Resolved.Name)
	}

	wfCall := wf.Body[1].(*ast.WorkflowCall)
	if wfCall.Resolved == nil {
		t.Error("workflow call not resolved")
	} else if wfCall.Resolved.Name != "ShipOrder" {
		t.Errorf("workflow resolved to %q, expected 'ShipOrder'", wfCall.Resolved.Name)
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
	if actCall.Resolved == nil {
		t.Error("activity call not resolved")
	}
	if actCall.Options == nil || len(actCall.Options.Entries) != 2 {
		t.Error("expected 2 option entries on activity call")
	}
	wfCall := wf.Body[1].(*ast.WorkflowCall)
	if wfCall.Resolved == nil {
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
	for _, e := range errs {
		if e.Severity != "warning" {
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

func TestWorkerMissingTaskQueue(t *testing.T) {
	input := `workflow Foo(x: int) -> (int):
    return x

worker w:
    workflow Foo

namespace orders:
    worker w
`
	file := mustParse(t, input)
	errs := Resolve(file)
	var found bool
	for _, e := range errs {
		if strings.Contains(e.Msg, "missing required task_queue") {
			found = true
		}
	}
	if !found {
		t.Error("expected error about missing task_queue option")
	}
}

func TestWorkerNotInstantiated(t *testing.T) {
	input := `workflow Foo(x: int) -> (int):
    return x

worker usedWorker:
    workflow Foo

worker unusedWorker:
    workflow Foo

namespace orders:
    worker usedWorker
        options:
            task_queue: "q"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	var found bool
	for _, e := range errs {
		if strings.Contains(e.Msg, "worker unusedWorker is not instantiated") {
			found = true
		}
	}
	if !found {
		t.Error("expected warning about worker not instantiated in any namespace")
	}
}

func TestTaskQueueCoherence(t *testing.T) {
	input := `workflow A(x: int) -> (int):
    return x

workflow B(x: int) -> (int):
    return x

activity C(x: int) -> (int):
    return x

worker worker1:
    workflow A
    activity C

worker worker2:
    workflow B
    activity C

namespace ns:
    worker worker1
        options:
            task_queue: "sharedQueue"
    worker worker2
        options:
            task_queue: "sharedQueue"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	var found bool
	for _, e := range errs {
		if strings.Contains(e.Msg, "different type sets") {
			found = true
		}
	}
	if !found {
		t.Error("expected error about different type sets on same task queue")
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
	for _, e := range errs {
		if e.Severity != "warning" {
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
	if sigBody.Resolved == nil {
		t.Error("signal body activity call not resolved")
	} else if sigBody.Resolved.Name != "LogCancel" {
		t.Errorf("resolved to %q, expected 'LogCancel'", sigBody.Resolved.Name)
	}

	// Check update body resolution
	updBody := wf.Updates[0].Body[0].(*ast.ActivityCall)
	if updBody.Resolved == nil {
		t.Error("update body activity call not resolved")
	} else if updBody.Resolved.Name != "ValidateAddr" {
		t.Errorf("resolved to %q, expected 'ValidateAddr'", updBody.Resolved.Name)
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
	if call.ResolvedService == nil {
		t.Error("nexus call service not resolved")
	} else if call.ResolvedService.Name != "OrderService" {
		t.Errorf("resolved service %q, expected 'OrderService'", call.ResolvedService.Name)
	}
	if call.ResolvedOperation == nil {
		t.Error("nexus call operation not resolved")
	} else if call.ResolvedOperation.Name != "PlaceOrder" {
		t.Errorf("resolved operation %q, expected 'PlaceOrder'", call.ResolvedOperation.Name)
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

func TestNexusEndpointMissingTaskQueue(t *testing.T) {
	input := `worker w:
    workflow W

workflow W():
    close complete(Result{})

namespace ns:
    worker w
        options:
            task_queue: "q"
    nexus endpoint Ep
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if !hasError(errs, "nexus endpoint Ep in namespace ns missing required task_queue option") {
		t.Error("expected error about endpoint missing task_queue")
	}
}

func TestNexusEndpointServiceLinkage(t *testing.T) {
	input := `nexus service Svc:
    async Op workflow W

workflow W():
    nexus Ep Svc.Op(x) -> result
    close complete(result)

worker w:
    workflow W

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
	if !hasError(errs, "no worker on that queue has service Svc") {
		t.Error("expected error about endpoint-service linkage")
	}
}

func TestNexusServiceNotOnWorker(t *testing.T) {
	input := `nexus service UnusedService:
    async Op workflow W

workflow W():
    close complete(Result{})

worker w:
    workflow W

namespace ns:
    worker w
        options:
            task_queue: "q"
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if !hasWarning(errs, "nexus service UnusedService is not referenced by any worker") {
		t.Error("expected warning about nexus service not on any worker")
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
