package validator

import (
	"strings"
	"testing"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/parser"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/resolver"
)

func mustParseAndResolve(t *testing.T, input string) *ast.File {
	t.Helper()
	file, err := parser.ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	errs := resolver.Resolve(file)
	for _, e := range errs {
		if e.Severity != "warning" {
			t.Fatalf("unexpected resolve error: %v", e)
		}
	}
	return file
}

// hasError checks if any non-warning error contains the given substring.
func hasError(errs []*Error, substr string) bool {
	for _, e := range errs {
		if e.Severity != "warning" && strings.Contains(e.Msg, substr) {
			return true
		}
	}
	return false
}

// hasWarning checks if any warning contains the given substring.
func hasWarning(errs []*Error, substr string) bool {
	for _, e := range errs {
		if e.Severity == "warning" && strings.Contains(e.Msg, substr) {
			return true
		}
	}
	return false
}

// ===== EMPTY DEFINITION WARNING TESTS =====

func TestEmptyWorkflowWarning(t *testing.T) {
	input := `workflow EmptyWorkflow(x: int) -> (int):
    # nothing here
`
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	if !hasWarning(errs, "workflow EmptyWorkflow has an empty body") {
		t.Error("expected warning about empty workflow body")
	}
}

func TestEmptyActivityWarning(t *testing.T) {
	input := `activity EmptyActivity(x: int) -> (int):
    # nothing here
`
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	if !hasWarning(errs, "activity EmptyActivity has an empty body") {
		t.Error("expected warning about empty activity body")
	}
}

func TestEmptyWorkerWarning(t *testing.T) {
	input := `workflow Foo(x: int) -> (int):
    return x

worker emptyWorker:
    # no registrations

namespace ns:
    worker emptyWorker
        options:
            task_queue: "q"
`
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	if !hasWarning(errs, "worker emptyWorker has no workflow, activity, or nexus service registrations") {
		t.Error("expected warning about empty worker")
	}
}

func TestEmptyNamespaceWarning(t *testing.T) {
	input := `namespace emptyNs:
    # no workers
`
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	if !hasWarning(errs, "namespace emptyNs has no worker or endpoint instantiations") {
		t.Error("expected warning about empty namespace")
	}
}

// ===== TASK QUEUE TESTS =====

func TestWorkerMissingTaskQueue(t *testing.T) {
	input := `workflow Foo(x: int) -> (int):
    return x

worker w:
    workflow Foo

namespace orders:
    worker w
`
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	if !hasError(errs, "missing required task_queue") {
		t.Error("expected error about missing task_queue option")
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
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	if !hasError(errs, "nexus endpoint Ep in namespace ns missing required task_queue option") {
		t.Error("expected error about endpoint missing task_queue")
	}
}

// ===== COVERAGE TESTS =====

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
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	if !hasWarning(errs, "worker unusedWorker is not instantiated") {
		t.Error("expected warning about worker not instantiated in any namespace")
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
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	if !hasWarning(errs, "nexus service UnusedService is not referenced by any worker") {
		t.Error("expected warning about nexus service not on any worker")
	}
}

// ===== TASK QUEUE COHERENCE TESTS =====

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
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	if !hasError(errs, "different type sets") {
		t.Error("expected error about different type sets on same task queue")
	}
}

// ===== ROUTING REACHABILITY TESTS =====

func TestExplicitTaskQueueRouting(t *testing.T) {
	input := `workflow Caller(x: int) -> (int):
    activity Target(x) -> y
        options:
            task_queue: "other-queue"
    return y

activity Target(x: int) -> (int):
    return x

worker callerWorker:
    workflow Caller

worker targetWorker:
    activity Target

namespace ns:
    worker callerWorker
        options:
            task_queue: "main-queue"
    worker targetWorker
        options:
            task_queue: "other-queue"
`
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	for _, e := range errs {
		if e.Severity != "warning" {
			t.Errorf("unexpected error: %v", e)
		}
	}
}

func TestExplicitTaskQueueRoutingMismatch(t *testing.T) {
	input := `workflow Caller(x: int) -> (int):
    activity Target(x) -> y
        options:
            task_queue: "wrong-queue"
    return y

activity Target(x: int) -> (int):
    return x

worker callerWorker:
    workflow Caller

worker targetWorker:
    activity Target

namespace ns:
    worker callerWorker
        options:
            task_queue: "main-queue"
    worker targetWorker
        options:
            task_queue: "other-queue"
`
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	if !hasError(errs, `activity Target has task_queue "wrong-queue", but no worker on that queue registers it`) {
		t.Error("expected error about explicit task_queue routing mismatch")
	}
}

func TestImplicitTaskQueueRouting(t *testing.T) {
	input := `workflow Caller(x: int) -> (int):
    activity Target(x) -> y
    return y

activity Target(x: int) -> (int):
    return x

worker w:
    workflow Caller
    activity Target

namespace ns:
    worker w
        options:
            task_queue: "shared-queue"
`
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	for _, e := range errs {
		if e.Severity != "warning" {
			t.Errorf("unexpected error: %v", e)
		}
	}
}

func TestImplicitTaskQueueRoutingMismatch(t *testing.T) {
	input := `workflow Caller(x: int) -> (int):
    activity Unreachable(x) -> y
    return y

activity Unreachable(x: int) -> (int):
    return x

worker callerWorker:
    workflow Caller

worker otherWorker:
    activity Unreachable

namespace ns:
    worker callerWorker
        options:
            task_queue: "main-queue"
    worker otherWorker
        options:
            task_queue: "other-queue"
`
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	if !hasError(errs, `activity Unreachable is not on any worker polling task queue "main-queue"`) {
		t.Error("expected error about implicit task queue routing mismatch")
	}
}

func TestImplicitTaskQueueChildWorkflowRouting(t *testing.T) {
	input := `workflow Parent(x: int) -> (int):
    workflow Child(x) -> y
    return y

workflow Child(x: int) -> (int):
    return x

worker w:
    workflow Parent
    workflow Child

namespace ns:
    worker w
        options:
            task_queue: "shared-queue"
`
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	for _, e := range errs {
		if e.Severity != "warning" {
			t.Errorf("unexpected error: %v", e)
		}
	}
}

func TestExplicitTaskQueueWorkflowRouting(t *testing.T) {
	input := `workflow Parent(x: int) -> (int):
    workflow Child(x) -> y
        options:
            task_queue: "child-queue"
    return y

workflow Child(x: int) -> (int):
    return x

worker parentWorker:
    workflow Parent

worker childWorker:
    workflow Child

namespace ns:
    worker parentWorker
        options:
            task_queue: "parent-queue"
    worker childWorker
        options:
            task_queue: "child-queue"
`
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	for _, e := range errs {
		if e.Severity != "warning" {
			t.Errorf("unexpected error: %v", e)
		}
	}
}

// ===== ENDPOINT-SERVICE LINKAGE TESTS =====

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
	file := mustParseAndResolve(t, input)
	errs := Validate(file)
	if !hasError(errs, "no worker on that queue has service Svc") {
		t.Error("expected error about endpoint-service linkage")
	}
}
