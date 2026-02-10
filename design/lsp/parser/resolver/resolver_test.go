package resolver

import (
	"strings"
	"testing"

	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/parser"
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
    hint signal PaymentReceived
    hint update ChangeAddress

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

	hintSig := wf.Body[2].(*ast.HintStmt)
	if hintSig.Resolved == nil {
		t.Error("hint signal not resolved")
	}

	hintUpd := wf.Body[3].(*ast.HintStmt)
	if hintUpd.Resolved == nil {
		t.Error("hint update not resolved")
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

func TestUndefinedSignal(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    hint signal Nonexistent
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Msg, "undefined signal: Nonexistent") {
		t.Errorf("unexpected error: %q", errs[0].Msg)
	}
}

func TestUndefinedUpdate(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    hint update Nonexistent
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Msg, "undefined update: Nonexistent") {
		t.Errorf("unexpected error: %q", errs[0].Msg)
	}
}

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

func TestHintResolution(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    signal Cancel(reason: string):
        return Result{cancelled: true}
    update ChangeAddr(addr: Addr) -> (Result):
        return Result{ok: true}

    hint signal Cancel
    hint update ChangeAddr
    return Result{}
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
	hint1 := wf.Body[0].(*ast.HintStmt)
	if hint1.Resolved == nil {
		t.Error("hint signal not resolved")
	}

	hint2 := wf.Body[1].(*ast.HintStmt)
	if hint2.Resolved == nil {
		t.Error("hint update not resolved")
	}
}

func TestHintUndefinedSignal(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    hint signal Missing
`
	file := mustParse(t, input)
	errs := Resolve(file)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Msg, "undefined signal: Missing") {
		t.Errorf("unexpected error: %q", errs[0].Msg)
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
