package parser

import (
	"os"
	"strings"
	"testing"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
)

func TestMinimalWorkflowDef(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    return x
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Definitions) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(file.Definitions))
	}
	wf, ok := file.Definitions[0].(*ast.WorkflowDef)
	if !ok {
		t.Fatalf("expected WorkflowDef, got %T", file.Definitions[0])
	}
	if wf.Name != "Foo" {
		t.Errorf("expected name 'Foo', got %q", wf.Name)
	}
	if wf.Params != "x: int" {
		t.Errorf("expected params 'x: int', got %q", wf.Params)
	}
	if wf.ReturnType != "Result" {
		t.Errorf("expected return type 'Result', got %q", wf.ReturnType)
	}
	if len(wf.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(wf.Body))
	}
	ret, ok := wf.Body[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("expected ReturnStmt, got %T", wf.Body[0])
	}
	if ret.Value != "x" {
		t.Errorf("expected return value 'x', got %q", ret.Value)
	}
}

func TestMinimalActivityDef(t *testing.T) {
	input := `activity GetOrder(orderId: string) -> (Order):
    order = db.get(orderId)
    return order
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Definitions) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(file.Definitions))
	}
	act, ok := file.Definitions[0].(*ast.ActivityDef)
	if !ok {
		t.Fatalf("expected ActivityDef, got %T", file.Definitions[0])
	}
	if act.Name != "GetOrder" {
		t.Errorf("expected name 'GetOrder', got %q", act.Name)
	}
	if act.ReturnType != "Order" {
		t.Errorf("expected return type 'Order', got %q", act.ReturnType)
	}
	if len(act.Body) != 2 {
		t.Fatalf("expected 2 body statements, got %d", len(act.Body))
	}
}

func TestWorkflowWithDeclarations(t *testing.T) {
	input := `workflow OrderWorkflow(orderId: string) -> (OrderResult):
    signal PaymentReceived(transactionId: string, amount: decimal):
        status = "paid"
    query GetStatus() -> (OrderStatus):
        return OrderStatus{phase: currentPhase}
    update ChangeAddress(addr: Address) -> (UpdateResult):
        address = addr
        return UpdateResult{ok: true}

    activity GetOrder(orderId) -> order
    return OrderResult{status: "completed"}
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	if len(wf.Signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(wf.Signals))
	}
	if wf.Signals[0].Name != "PaymentReceived" {
		t.Errorf("expected signal name 'PaymentReceived', got %q", wf.Signals[0].Name)
	}
	if len(wf.Signals[0].Body) != 1 {
		t.Fatalf("expected 1 signal body statement, got %d", len(wf.Signals[0].Body))
	}
	if len(wf.Queries) != 1 {
		t.Fatalf("expected 1 query, got %d", len(wf.Queries))
	}
	if wf.Queries[0].ReturnType != "OrderStatus" {
		t.Errorf("expected query return type 'OrderStatus', got %q", wf.Queries[0].ReturnType)
	}
	if len(wf.Queries[0].Body) != 1 {
		t.Fatalf("expected 1 query body statement, got %d", len(wf.Queries[0].Body))
	}
	if len(wf.Updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(wf.Updates))
	}
	if wf.Updates[0].Name != "ChangeAddress" {
		t.Errorf("expected update name 'ChangeAddress', got %q", wf.Updates[0].Name)
	}
	if len(wf.Updates[0].Body) != 2 {
		t.Fatalf("expected 2 update body statements, got %d", len(wf.Updates[0].Body))
	}
	if len(wf.Body) != 2 {
		t.Fatalf("expected 2 body statements, got %d", len(wf.Body))
	}
}

func TestSignalDeclWithBody(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    signal Cancel(reason: string):
        cancelled = true
        return Result{cancelled: true}

    return Result{}
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	if len(wf.Signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(wf.Signals))
	}
	sig := wf.Signals[0]
	if sig.Name != "Cancel" {
		t.Errorf("expected signal name 'Cancel', got %q", sig.Name)
	}
	if sig.Params != "reason: string" {
		t.Errorf("expected params 'reason: string', got %q", sig.Params)
	}
	if len(sig.Body) != 2 {
		t.Fatalf("expected 2 signal body statements, got %d", len(sig.Body))
	}
}

func TestQueryDeclWithBody(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    query GetStatus() -> (Status):
        return Status{phase: currentPhase}

    return Result{}
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	if len(wf.Queries) != 1 {
		t.Fatalf("expected 1 query, got %d", len(wf.Queries))
	}
	q := wf.Queries[0]
	if q.Name != "GetStatus" {
		t.Errorf("expected query name 'GetStatus', got %q", q.Name)
	}
	if q.ReturnType != "Status" {
		t.Errorf("expected return type 'Status', got %q", q.ReturnType)
	}
	if len(q.Body) != 1 {
		t.Fatalf("expected 1 query body statement, got %d", len(q.Body))
	}
}

func TestUpdateDeclWithBody(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    update ChangeAddr(addr: Addr) -> (Result):
        address = addr
        return Result{ok: true}

    return Result{}
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	if len(wf.Updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(wf.Updates))
	}
	u := wf.Updates[0]
	if u.Name != "ChangeAddr" {
		t.Errorf("expected update name 'ChangeAddr', got %q", u.Name)
	}
	if u.ReturnType != "Result" {
		t.Errorf("expected return type 'Result', got %q", u.ReturnType)
	}
	if len(u.Body) != 2 {
		t.Fatalf("expected 2 update body statements, got %d", len(u.Body))
	}
}

// REMOVED: TestHintStmt - hint statements are no longer supported.
// REMOVED: TestHintInvalidTarget - hint statements are no longer supported.

// REMOVED: TestAwaitOneCaseError - signal cases are now supported in await one blocks.

func TestActivityCallWithResult(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    activity GetOrder(orderId) -> order
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call, ok := wf.Body[0].(*ast.ActivityCall)
	if !ok {
		t.Fatalf("expected ActivityCall, got %T", wf.Body[0])
	}
	if call.Activity.Name != "GetOrder" {
		t.Errorf("expected name 'GetOrder', got %q", call.Activity.Name)
	}
	if call.Args != "orderId" {
		t.Errorf("expected args 'orderId', got %q", call.Args)
	}
	if call.Result != "order" {
		t.Errorf("expected result 'order', got %q", call.Result)
	}
}

func TestWorkflowCallChild(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    workflow ShipOrder(order) -> shipResult
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call, ok := wf.Body[0].(*ast.WorkflowCall)
	if !ok {
		t.Fatalf("expected WorkflowCall, got %T", wf.Body[0])
	}
	if call.Mode != ast.CallChild {
		t.Errorf("expected CallChild, got %d", call.Mode)
	}
	if call.Workflow.Name != "ShipOrder" {
		t.Errorf("expected name 'ShipOrder', got %q", call.Workflow.Name)
	}
	if call.Result != "shipResult" {
		t.Errorf("expected result 'shipResult', got %q", call.Result)
	}
}

func TestPromiseActivity(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    promise p <- activity ProcessAsync(data)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	promise, ok := wf.Body[0].(*ast.PromiseStmt)
	if !ok {
		t.Fatalf("expected PromiseStmt, got %T", wf.Body[0])
	}
	if promise.Name != "p" {
		t.Errorf("expected name 'p', got %q", promise.Name)
	}
	at, ok := promise.Target.(*ast.ActivityTarget)
	if !ok {
		t.Fatalf("expected ActivityTarget, got %T", promise.Target)
	}
	if at.Activity.Name != "ProcessAsync" {
		t.Errorf("expected activity 'ProcessAsync', got %q", at.Activity.Name)
	}
	if at.Args != "data" {
		t.Errorf("expected args 'data', got %q", at.Args)
	}
}

func TestWorkflowCallDetach(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    detach workflow SendNotification(customer)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call := wf.Body[0].(*ast.WorkflowCall)
	if call.Mode != ast.CallDetach {
		t.Errorf("expected CallDetach, got %d", call.Mode)
	}
	if call.Result != "" {
		t.Errorf("expected no result, got %q", call.Result)
	}
}

func TestNexusCallBasic(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    nexus PaymentEndpoint PaymentService.Charge(card) -> chargeResult
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call, ok := wf.Body[0].(*ast.NexusCall)
	if !ok {
		t.Fatalf("expected NexusCall, got %T", wf.Body[0])
	}
	if call.Endpoint.Name != "PaymentEndpoint" {
		t.Errorf("expected endpoint 'PaymentEndpoint', got %q", call.Endpoint.Name)
	}
	if call.Service.Name != "PaymentService" {
		t.Errorf("expected service 'PaymentService', got %q", call.Service.Name)
	}
	if call.Operation.Name != "Charge" {
		t.Errorf("expected operation 'Charge', got %q", call.Operation.Name)
	}
	if call.Args != "card" {
		t.Errorf("expected args 'card', got %q", call.Args)
	}
	if call.Result != "chargeResult" {
		t.Errorf("expected result 'chargeResult', got %q", call.Result)
	}
	if call.Detach {
		t.Error("expected detach false")
	}
}

func TestPromiseWorkflow(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    promise p <- workflow ProcessAsync(data)

workflow ProcessAsync(data: Data) -> (Result):
    close complete(Result{})
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	promise, ok := wf.Body[0].(*ast.PromiseStmt)
	if !ok {
		t.Fatalf("expected PromiseStmt, got %T", wf.Body[0])
	}
	if promise.Name != "p" {
		t.Errorf("expected name 'p', got %q", promise.Name)
	}
	wt, ok := promise.Target.(*ast.WorkflowTarget)
	if !ok {
		t.Fatalf("expected WorkflowTarget, got %T", promise.Target)
	}
	if wt.Workflow.Name != "ProcessAsync" {
		t.Errorf("expected workflow 'ProcessAsync', got %q", wt.Workflow.Name)
	}
}

func TestPromiseTimer(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    promise timeout <- timer(5m)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	promise, ok := wf.Body[0].(*ast.PromiseStmt)
	if !ok {
		t.Fatalf("expected PromiseStmt, got %T", wf.Body[0])
	}
	if promise.Name != "timeout" {
		t.Errorf("expected name 'timeout', got %q", promise.Name)
	}
	tt, ok := promise.Target.(*ast.TimerTarget)
	if !ok {
		t.Fatalf("expected TimerTarget, got %T", promise.Target)
	}
	if tt.Duration != "5m" {
		t.Errorf("expected timer '5m', got %q", tt.Duration)
	}
}

func TestPromiseSignal(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    signal Approved():
        approved = true
    promise p <- signal Approved
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	promise, ok := wf.Body[0].(*ast.PromiseStmt)
	if !ok {
		t.Fatalf("expected PromiseStmt, got %T", wf.Body[0])
	}
	st, ok := promise.Target.(*ast.SignalTarget)
	if !ok {
		t.Fatalf("expected SignalTarget, got %T", promise.Target)
	}
	if st.Signal.Name != "Approved" {
		t.Errorf("expected signal 'Approved', got %q", st.Signal.Name)
	}
}

func TestStateBlockWithCondition(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    state:
        condition clusterStarted
        balance = 0

    close complete(Result{})
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	if wf.State == nil {
		t.Fatal("expected state block, got nil")
	}
	if len(wf.State.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(wf.State.Conditions))
	}
	if wf.State.Conditions[0].Name != "clusterStarted" {
		t.Errorf("expected condition 'clusterStarted', got %q", wf.State.Conditions[0].Name)
	}
	if len(wf.State.RawStmts) != 1 {
		t.Fatalf("expected 1 raw stmt, got %d", len(wf.State.RawStmts))
	}
}

func TestSetUnsetStatements(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    state:
        condition ready

    signal Activate():
        set ready

    set ready
    unset ready
    close complete(Result{})
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	if wf.State == nil {
		t.Fatal("expected state block")
	}

	setStmt, ok := wf.Body[0].(*ast.SetStmt)
	if !ok {
		t.Fatalf("expected SetStmt, got %T", wf.Body[0])
	}
	if setStmt.Condition.Name != "ready" {
		t.Errorf("expected name 'ready', got %q", setStmt.Condition.Name)
	}

	unsetStmt, ok := wf.Body[1].(*ast.UnsetStmt)
	if !ok {
		t.Fatalf("expected UnsetStmt, got %T", wf.Body[1])
	}
	if unsetStmt.Condition.Name != "ready" {
		t.Errorf("expected name 'ready', got %q", unsetStmt.Condition.Name)
	}
}

func TestAwaitIdent(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    state:
        condition ready

    await ready
    close complete(Result{})
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	awaitStmt, ok := wf.Body[0].(*ast.AwaitStmt)
	if !ok {
		t.Fatalf("expected AwaitStmt, got %T", wf.Body[0])
	}
	it, ok := awaitStmt.Target.(*ast.IdentTarget)
	if !ok {
		t.Fatalf("expected IdentTarget, got %T", awaitStmt.Target)
	}
	if it.Name != "ready" {
		t.Errorf("expected ident 'ready', got %q", it.Name)
	}
}

func TestAwaitIdentWithResult(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    promise p <- activity Process(data)
    await p -> result
    close complete(Result{})

activity Process(data: Data) -> (Result):
    return Result{}
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	awaitStmt, ok := wf.Body[1].(*ast.AwaitStmt)
	if !ok {
		t.Fatalf("expected AwaitStmt, got %T", wf.Body[1])
	}
	it, ok := awaitStmt.Target.(*ast.IdentTarget)
	if !ok {
		t.Fatalf("expected IdentTarget, got %T", awaitStmt.Target)
	}
	if it.Name != "p" {
		t.Errorf("expected ident 'p', got %q", it.Name)
	}
	if it.Result != "result" {
		t.Errorf("expected ident result 'result', got %q", it.Result)
	}
}

func TestAwaitOneIdentCase(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    state:
        condition ready

    await one:
        ready:
            close complete(Result{status: "ready"})
        timer(5m):
            close fail(Result{status: "timeout"})
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	awaitOne, ok := wf.Body[0].(*ast.AwaitOneBlock)
	if !ok {
		t.Fatalf("expected AwaitOneBlock, got %T", wf.Body[0])
	}
	if len(awaitOne.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(awaitOne.Cases))
	}
	if ast.AsyncTargetKind(awaitOne.Cases[0].Target) != "ident" {
		t.Errorf("case[0]: expected ident, got %q", ast.AsyncTargetKind(awaitOne.Cases[0].Target))
	}
	identCase, ok := awaitOne.Cases[0].Target.(*ast.IdentTarget)
	if !ok {
		t.Fatalf("expected IdentTarget, got %T", awaitOne.Cases[0].Target)
	}
	if identCase.Name != "ready" {
		t.Errorf("case[0] ident: expected 'ready', got %q", identCase.Name)
	}
}

func TestDetachWithArrowError(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    detach workflow Send(x) -> result
`
	_, err := ParseFile(input)
	if err == nil {
		t.Fatal("expected error for detach with arrow, got nil")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected ParseError, got %T: %v", err, err)
	}
	if pe.Msg != "detach workflow call cannot have a result (-> identifier)" {
		t.Errorf("unexpected error message: %q", pe.Msg)
	}
}

func TestOptionsOnCall(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    activity CreateShipment(order) -> shipment
        options:
            start_to_close_timeout: 30s
            retry_policy:
                maximum_attempts: 3
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call := wf.Body[0].(*ast.ActivityCall)
	if call.Options == nil {
		t.Fatal("expected options, got nil")
	}
	if len(call.Options.Entries) != 2 {
		t.Fatalf("expected 2 option entries, got %d", len(call.Options.Entries))
	}
	if call.Options.Entries[0].Key != "start_to_close_timeout" {
		t.Errorf("expected key 'start_to_close_timeout', got %q", call.Options.Entries[0].Key)
	}
	if call.Options.Entries[0].Value != "30s" {
		t.Errorf("expected value '30s', got %q", call.Options.Entries[0].Value)
	}
	if call.Options.Entries[1].Key != "retry_policy" {
		t.Errorf("expected key 'retry_policy', got %q", call.Options.Entries[1].Key)
	}
	if len(call.Options.Entries[1].Nested) != 1 {
		t.Fatalf("expected 1 nested entry, got %d", len(call.Options.Entries[1].Nested))
	}
}

func TestOptionsOnWorkflowCall(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    workflow ShipOrder(order) -> result
        options:
            workflow_execution_timeout: 24h
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call := wf.Body[0].(*ast.WorkflowCall)
	if call.Options == nil {
		t.Fatal("expected options, got nil")
	}
	if len(call.Options.Entries) != 1 {
		t.Fatalf("expected 1 option entry, got %d", len(call.Options.Entries))
	}
	if call.Options.Entries[0].Key != "workflow_execution_timeout" {
		t.Errorf("expected key 'workflow_execution_timeout', got %q", call.Options.Entries[0].Key)
	}
}

// REMOVED: TestOptionsOnDefinition - options on definitions are no longer supported.
// Options are only valid on activity/workflow calls.

// REMOVED: TestTimer - TimerStmt is no longer a standalone statement.
// Timers are now used via await timer(duration) or await one cases.

// REMOVED: TestAwaitSingle - 'await signal' syntax no longer supported.
// Signals are now referenced via 'hint signal' statements.

// REMOVED: TestAwaitMultiTarget - 'await signal/update' syntax no longer supported.
// Signals and updates are now referenced via 'hint' statements.

func TestAwaitAllBlock(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    await all:
        activity ReserveInventory(order) -> inventory
        activity ProcessPayment(order) -> payment
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	par, ok := wf.Body[0].(*ast.AwaitAllBlock)
	if !ok {
		t.Fatalf("expected AwaitAllBlock, got %T", wf.Body[0])
	}
	if len(par.Body) != 2 {
		t.Fatalf("expected 2 body statements, got %d", len(par.Body))
	}
}

func TestAwaitOneBlock(t *testing.T) {
	// await one cases support timer (with bodies), and nested await all.
	input := `workflow Foo(x: int) -> (Result):
    await one:
        timer (1h):
            activity HandleTimeout1()
        timer (24h):
            activity HandleTimeout2()
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	awaitOne, ok := wf.Body[0].(*ast.AwaitOneBlock)
	if !ok {
		t.Fatalf("expected AwaitOneBlock, got %T", wf.Body[0])
	}
	if len(awaitOne.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(awaitOne.Cases))
	}
	if ast.AsyncTargetKind(awaitOne.Cases[0].Target) != "timer" {
		t.Errorf("case[0]: expected timer, got %q", ast.AsyncTargetKind(awaitOne.Cases[0].Target))
	}
	timer0, ok := awaitOne.Cases[0].Target.(*ast.TimerTarget)
	if !ok {
		t.Fatalf("expected TimerTarget, got %T", awaitOne.Cases[0].Target)
	}
	if timer0.Duration != "1h" {
		t.Errorf("case[0] timer: expected '1h', got %q", timer0.Duration)
	}
	if len(awaitOne.Cases[0].Body) != 1 {
		t.Errorf("case[0] body: expected 1 statement, got %d", len(awaitOne.Cases[0].Body))
	}
	if ast.AsyncTargetKind(awaitOne.Cases[1].Target) != "timer" {
		t.Errorf("case[1]: expected timer, got %q", ast.AsyncTargetKind(awaitOne.Cases[1].Target))
	}
	timer1, ok := awaitOne.Cases[1].Target.(*ast.TimerTarget)
	if !ok {
		t.Fatalf("expected TimerTarget, got %T", awaitOne.Cases[1].Target)
	}
	if timer1.Duration != "24h" {
		t.Errorf("case[1] timer: expected '24h', got %q", timer1.Duration)
	}
	if len(awaitOne.Cases[1].Body) != 1 {
		t.Errorf("case[1] body: expected 1 statement, got %d", len(awaitOne.Cases[1].Body))
	}
}

func TestSwitchBlock(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    switch (batch.type):
        case "invoice":
            activity ProcessInvoice(batch)
        case "refund":
            activity ProcessRefund(batch)
        else:
            activity HandleUnknown(batch)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	sw, ok := wf.Body[0].(*ast.SwitchBlock)
	if !ok {
		t.Fatalf("expected SwitchBlock, got %T", wf.Body[0])
	}
	if sw.Expr != "batch.type" {
		t.Errorf("expected expr 'batch.type', got %q", sw.Expr)
	}
	if len(sw.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(sw.Cases))
	}
	if sw.Default == nil {
		t.Fatal("expected default body, got nil")
	}
}

func TestIfElse(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    if (validated.priority == "high"):
        activity ExpediteOrder(order)
    else:
        activity StandardProcessing(order)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	ifStmt, ok := wf.Body[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", wf.Body[0])
	}
	if ifStmt.Condition != `validated.priority == "high"` {
		t.Errorf("unexpected condition: %q", ifStmt.Condition)
	}
	if len(ifStmt.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(ifStmt.Body))
	}
	if len(ifStmt.ElseBody) != 1 {
		t.Errorf("expected 1 else body statement, got %d", len(ifStmt.ElseBody))
	}
}

func TestForInfinite(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    for:
        activity Poll(x)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	forStmt := wf.Body[0].(*ast.ForStmt)
	if forStmt.Variant != ast.ForInfinite {
		t.Errorf("expected ForInfinite, got %d", forStmt.Variant)
	}
}

func TestForConditional(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    for (retries < 3):
        activity TryOnce(x)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	forStmt := wf.Body[0].(*ast.ForStmt)
	if forStmt.Variant != ast.ForConditional {
		t.Errorf("expected ForConditional, got %d", forStmt.Variant)
	}
	if forStmt.Condition != "retries < 3" {
		t.Errorf("unexpected condition: %q", forStmt.Condition)
	}
}

func TestForIteration(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    for (item in order.items):
        activity ProcessItem(item)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	forStmt := wf.Body[0].(*ast.ForStmt)
	if forStmt.Variant != ast.ForIteration {
		t.Errorf("expected ForIteration, got %d", forStmt.Variant)
	}
	if forStmt.Variable != "item" {
		t.Errorf("expected variable 'item', got %q", forStmt.Variable)
	}
	if forStmt.Iterable != "order.items" {
		t.Errorf("expected iterable 'order.items', got %q", forStmt.Iterable)
	}
}

func TestContinueAsNew(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    close continue_as_new(newArgs)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	closeStmt, ok := wf.Body[0].(*ast.CloseStmt)
	if !ok {
		t.Fatalf("expected CloseStmt, got %T", wf.Body[0])
	}
	if closeStmt.Reason != ast.CloseContinueAsNew {
		t.Errorf("expected reason CloseContinueAsNew, got %v", closeStmt.Reason)
	}
	if closeStmt.Args != "newArgs" {
		t.Errorf("expected args 'newArgs', got %q", closeStmt.Args)
	}
}

func TestBreakContinue(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    for:
        if (done):
            break
        continue
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	forStmt := wf.Body[0].(*ast.ForStmt)
	if len(forStmt.Body) != 2 {
		t.Fatalf("expected 2 body statements, got %d", len(forStmt.Body))
	}
	ifStmt := forStmt.Body[0].(*ast.IfStmt)
	if _, ok := ifStmt.Body[0].(*ast.BreakStmt); !ok {
		t.Errorf("expected BreakStmt, got %T", ifStmt.Body[0])
	}
	if _, ok := forStmt.Body[1].(*ast.ContinueStmt); !ok {
		t.Errorf("expected ContinueStmt, got %T", forStmt.Body[1])
	}
}

func TestRawStmt(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    order.status = "completed"
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	raw, ok := wf.Body[0].(*ast.RawStmt)
	if !ok {
		t.Fatalf("expected RawStmt, got %T", wf.Body[0])
	}
	// Raw captures all tokens on the line.
	if raw.Text == "" {
		t.Error("expected non-empty raw text")
	}
}

func TestComment(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    # this is a comment
    return x
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	// Comments are included in the body when inside workflow bodies.
	// However, comments may be consumed by skipBlankLinesAndComments() in some contexts.
	if len(wf.Body) < 1 {
		t.Fatalf("expected at least 1 body statement, got %d", len(wf.Body))
	}
	// The body should have a return statement.
	var foundReturn bool
	for _, stmt := range wf.Body {
		if _, ok := stmt.(*ast.ReturnStmt); ok {
			foundReturn = true
		}
	}
	if !foundReturn {
		t.Fatal("expected return statement in body")
	}
}

func TestTemporalKeywordInActivityError(t *testing.T) {
	input := `activity Foo(x: int) -> (Result):
    timer 1h
`
	_, err := ParseFile(input)
	if err == nil {
		t.Fatal("expected error for temporal keyword in activity body, got nil")
	}
}

// REMOVED: TestSelectDetachError - detach workflow cases are now supported in await one blocks.

func TestMultipleDefinitions(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    activity Bar(x) -> y
    return y

activity Bar(x: int) -> (int):
    return x
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Definitions) != 2 {
		t.Fatalf("expected 2 definitions, got %d", len(file.Definitions))
	}
	if _, ok := file.Definitions[0].(*ast.WorkflowDef); !ok {
		t.Errorf("expected WorkflowDef, got %T", file.Definitions[0])
	}
	if _, ok := file.Definitions[1].(*ast.ActivityDef); !ok {
		t.Errorf("expected ActivityDef, got %T", file.Definitions[1])
	}
}

func TestActivityCallWithOptions(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    activity CreateShipment(order) -> shipment
        options:
            start_to_close_timeout: 30s
    activity GetOrder(orderId) -> order
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	if len(wf.Body) != 2 {
		t.Fatalf("expected 2 body statements, got %d", len(wf.Body))
	}
	call1 := wf.Body[0].(*ast.ActivityCall)
	if call1.Options == nil {
		t.Error("expected options, got nil")
	}
	call2 := wf.Body[1].(*ast.ActivityCall)
	if call2.Options != nil {
		t.Errorf("expected no options, got %v", call2.Options)
	}
}

// REMOVED: TestSelectWithActivityCase - activity/workflow cases no longer supported in await one.
// await one now only supports timer and nested await all cases.

func TestSwitchInActivity(t *testing.T) {
	input := `activity Foo(x: string) -> (Result):
    switch (x):
        case "a":
            doA()
        case "b":
            doB()
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	act := file.Definitions[0].(*ast.ActivityDef)
	sw, ok := act.Body[0].(*ast.SwitchBlock)
	if !ok {
		t.Fatalf("expected SwitchBlock, got %T", act.Body[0])
	}
	if len(sw.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(sw.Cases))
	}
}

func TestIfInActivity(t *testing.T) {
	input := `activity Foo(x: int) -> (Result):
    if (x > 0):
        return x
    else:
        return 0
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	act := file.Definitions[0].(*ast.ActivityDef)
	ifStmt, ok := act.Body[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", act.Body[0])
	}
	if len(ifStmt.ElseBody) != 1 {
		t.Errorf("expected 1 else body statement, got %d", len(ifStmt.ElseBody))
	}
}

func TestForInActivity(t *testing.T) {
	input := `activity Foo(items: []string) -> (Result):
    for (item in items):
        process(item)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	act := file.Definitions[0].(*ast.ActivityDef)
	forStmt, ok := act.Body[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", act.Body[0])
	}
	if forStmt.Variant != ast.ForIteration {
		t.Errorf("expected ForIteration, got %d", forStmt.Variant)
	}
}

// REMOVED: TestWorkflowDefWithOptions - options on definitions are no longer supported.
// Options are only valid on activity/workflow calls.

func TestOptionsNestedRetryPolicy(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    activity Bar(x) -> y
        options:
            start_to_close_timeout: 60s
            retry_policy:
                maximum_attempts: 3
                initial_interval: 1s
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call := wf.Body[0].(*ast.ActivityCall)
	if call.Options == nil {
		t.Fatal("expected options, got nil")
	}
	if len(call.Options.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(call.Options.Entries))
	}
	rp := call.Options.Entries[1]
	if rp.Key != "retry_policy" {
		t.Errorf("expected key 'retry_policy', got %q", rp.Key)
	}
	if len(rp.Nested) != 2 {
		t.Fatalf("expected 2 nested entries, got %d", len(rp.Nested))
	}
	if rp.Nested[0].Key != "maximum_attempts" {
		t.Errorf("expected nested key 'maximum_attempts', got %q", rp.Nested[0].Key)
	}
	if rp.Nested[0].Value != "3" {
		t.Errorf("expected value '3', got %q", rp.Nested[0].Value)
	}
}

func TestOptionsUnrecognizedKey(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    activity Bar(x) -> y
        options:
            bogus_key: 5s
`
	_, err := ParseFile(input)
	if err == nil {
		t.Fatal("expected error for unrecognized key, got nil")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected ParseError, got %T: %v", err, err)
	}
	if pe.Msg != "unknown option key: bogus_key" {
		t.Errorf("unexpected error message: %q", pe.Msg)
	}
}

func TestOptionsWrongValueType(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    activity Bar(x) -> y
        options:
            start_to_close_timeout: 3
`
	_, err := ParseFile(input)
	if err == nil {
		t.Fatal("expected error for wrong value type, got nil")
	}
}

func TestOptionsEnumValidation(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    workflow Child(x) -> y
        options:
            parent_close_policy: TERMINATE
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call := wf.Body[0].(*ast.WorkflowCall)
	if call.Options == nil || len(call.Options.Entries) != 1 {
		t.Fatal("expected 1 option entry")
	}
	if call.Options.Entries[0].Value != "TERMINATE" {
		t.Errorf("expected value 'TERMINATE', got %q", call.Options.Entries[0].Value)
	}
}

func TestOptionsInvalidEnum(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    workflow Child(x) -> y
        options:
            parent_close_policy: INVALID
`
	_, err := ParseFile(input)
	if err == nil {
		t.Fatal("expected error for invalid enum, got nil")
	}
}

func TestOptionsEmpty(t *testing.T) {
	// Empty options: with no indented content is a parse error
	// (indentation-based syntax requires at least one entry).
	input := `workflow Foo(x: int) -> (Result):
    activity Bar(x) -> y
        options:
`
	_, err := ParseFile(input)
	if err == nil {
		t.Fatal("expected error for empty options block, got nil")
	}
}

// REMOVED: TestAwaitWithArgs - 'await signal' syntax no longer supported.
// Signals are now referenced via 'hint signal' statements.

func TestParseFileAllFirstDefError(t *testing.T) {
	// First definition has a syntax error; second should still parse.
	input := `workflow Broken(x: int)
    return x

activity Bar(x: int) -> (int):
    return x
`
	file, errs := ParseFileAll(input)
	if len(errs) == 0 {
		t.Fatal("expected at least one error")
	}
	if len(file.Definitions) != 1 {
		t.Fatalf("expected 1 definition (the second one), got %d", len(file.Definitions))
	}
	if _, ok := file.Definitions[0].(*ast.ActivityDef); !ok {
		t.Errorf("expected ActivityDef, got %T", file.Definitions[0])
	}
}

func TestParseFileAllAllErrors(t *testing.T) {
	// All definitions have errors; expect empty definitions, non-empty errors.
	input := `workflow Bad1
workflow Bad2
`
	file, errs := ParseFileAll(input)
	if len(errs) == 0 {
		t.Fatal("expected errors, got none")
	}
	if len(file.Definitions) != 0 {
		t.Errorf("expected 0 definitions, got %d", len(file.Definitions))
	}
}

func TestParseFileAllCleanInput(t *testing.T) {
	// Clean input should produce zero errors and a complete AST.
	input := `workflow Foo(x: int) -> (Result):
    return x

activity Bar(y: string) -> (string):
    return y
`
	file, errs := ParseFileAll(input)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errs), errs)
	}
	if len(file.Definitions) != 2 {
		t.Fatalf("expected 2 definitions, got %d", len(file.Definitions))
	}
}

func TestParseAllTestdata(t *testing.T) {
	// Test that all files in testdata/ parse successfully.
	files, err := os.ReadDir("../testdata")
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".twf") {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			path := "../testdata/" + file.Name()
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", file.Name(), err)
			}

			_, err = ParseFile(string(content))
			if err != nil {
				t.Errorf("failed to parse %s: %v", file.Name(), err)
			}
		})
	}
}

func TestFullWorkflow(t *testing.T) {
	input := `workflow OrderFulfillment(orderId: string) -> (OrderResult):
    signal PaymentReceived(transactionId: string, amount: decimal):
        status = "paid"
    signal OrderCancelled(reason: string):
        cancelled = true
    query GetStatus() -> (OrderStatus):
        return OrderStatus{phase: currentPhase}
    update ChangeAddress(addr: Address) -> (UpdateResult):
        address = addr
        return UpdateResult{ok: true}

    activity GetOrder(orderId) -> order

    if (order.priority == "high"):
        activity ExpediteOrder(order)
    else:
        activity StandardProcessing(order)

    await all:
        activity ReserveInventory(order) -> inventory
        activity ProcessPayment(order) -> payment

    for (item in order.items):
        activity ProcessItem(item)

    await one:
        timer (1h):
            activity CheckPayment(orderId) -> checkResult
        timer (24h):
            activity CancelOrder(orderId)
            close fail(OrderResult{status: "timeout"})

    workflow ShipOrder(order) -> shipResult
    promise asyncResult <- workflow ProcessAsync(data)
    detach workflow SendNotification(order.customer)
    nexus PaymentEndpoint PaymentService.Charge(card) -> chargeResult

    switch (order.type):
        case "invoice":
            activity ProcessInvoice(order)
        else:
            activity HandleDefault(order)

    close continue_as_new(orderId)
    return OrderResult{status: "completed"}
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	if wf.Name != "OrderFulfillment" {
		t.Errorf("expected name 'OrderFulfillment', got %q", wf.Name)
	}
	if len(wf.Signals) != 2 {
		t.Errorf("expected 2 signals, got %d", len(wf.Signals))
	}
	if len(wf.Queries) != 1 {
		t.Errorf("expected 1 query, got %d", len(wf.Queries))
	}
	if len(wf.Updates) != 1 {
		t.Errorf("expected 1 update, got %d", len(wf.Updates))
	}
	// Count body statements (all the stuff after declarations).
	if len(wf.Body) < 10 {
		t.Errorf("expected at least 10 body statements, got %d", len(wf.Body))
	}
}

// === New Syntax Tests ===

// REMOVED: TestWatchCase - watch keyword is no longer supported.
// The watch functionality has been removed from the language.

func TestTimerCaseWithBody(t *testing.T) {
	input := `workflow Test():
    await one:
        timer (5m):
            activity SendReminder()
            close complete
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	awaitOne, ok := wf.Body[0].(*ast.AwaitOneBlock)
	if !ok {
		t.Fatalf("expected AwaitOneBlock, got %T", wf.Body[0])
	}
	if len(awaitOne.Cases) != 1 {
		t.Fatalf("expected 1 case, got %d", len(awaitOne.Cases))
	}

	c := awaitOne.Cases[0]
	if ast.AsyncTargetKind(c.Target) != "timer" {
		t.Errorf("expected timer case, got %q", ast.AsyncTargetKind(c.Target))
	}
	ct, ok := c.Target.(*ast.TimerTarget)
	if !ok {
		t.Fatalf("expected TimerTarget, got %T", c.Target)
	}
	if ct.Duration != "5m" {
		t.Errorf("timer duration: expected '5m', got %q", ct.Duration)
	}
	if len(c.Body) != 2 {
		t.Errorf("expected 2 statements in timer body, got %d", len(c.Body))
	}
}

func TestCloseStatement(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		reason ast.CloseReason
		args   string
	}{
		{
			name:   "close complete",
			input:  "workflow Test():\n    close complete\n",
			reason: ast.CloseComplete,
			args:   "",
		},
		{
			name:   "close complete with args",
			input:  "workflow Test():\n    close complete(Result{success: true})\n",
			reason: ast.CloseComplete,
			args:   "Result{success: true}",
		},
		{
			name:   "close fail",
			input:  "workflow Test():\n    close fail\n",
			reason: ast.CloseFailWorkflow,
			args:   "",
		},
		{
			name:   "close fail with args",
			input:  "workflow Test():\n    close fail(Error{message: \"timeout\"})\n",
			reason: ast.CloseFailWorkflow,
			args:   "Error{message: \"timeout\"}",
		},
		{
			name:   "close continue_as_new",
			input:  "workflow Test():\n    close continue_as_new(newArgs)\n",
			reason: ast.CloseContinueAsNew,
			args:   "newArgs",
		},
		{
			name:   "close continue_as_new no args",
			input:  "workflow Test():\n    close continue_as_new\n",
			reason: ast.CloseContinueAsNew,
			args:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := ParseFile(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			wf := file.Definitions[0].(*ast.WorkflowDef)
			if len(wf.Body) != 1 {
				t.Fatalf("expected 1 body statement, got %d", len(wf.Body))
			}
			closeStmt, ok := wf.Body[0].(*ast.CloseStmt)
			if !ok {
				t.Fatalf("expected CloseStmt, got %T", wf.Body[0])
			}
			if closeStmt.Reason != tt.reason {
				t.Errorf("reason: expected %v, got %v", tt.reason, closeStmt.Reason)
			}
			if closeStmt.Args != tt.args {
				t.Errorf("args: expected %q, got %q", tt.args, closeStmt.Args)
			}
		})
	}
}

func TestCloseRequiresReason(t *testing.T) {
	input := "workflow Test():\n    close\n"
	_, err := ParseFile(input)
	if err == nil {
		t.Fatal("expected error for bare 'close' without reason, got nil")
	}
}

// REMOVED: TestWatchWithMultipleHints - watch keyword and hint statements are no longer supported.
// REMOVED: TestMultipleWatchCases - watch keyword is no longer supported.

func TestWorkerDef(t *testing.T) {
	input := `workflow ProcessOrder(orderId: string) -> (Result):
    return Result{}

activity ChargePayment(orderId: string) -> (Payment):
    return charge(orderId)

worker orderWorker:
    workflow ProcessOrder
    activity ChargePayment
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Definitions) != 3 {
		t.Fatalf("expected 3 definitions, got %d", len(file.Definitions))
	}
	w, ok := file.Definitions[2].(*ast.WorkerDef)
	if !ok {
		t.Fatalf("expected WorkerDef, got %T", file.Definitions[2])
	}
	if w.Name != "orderWorker" {
		t.Errorf("expected name 'orderWorker', got %q", w.Name)
	}
	if len(w.Workflows) != 1 {
		t.Fatalf("expected 1 workflow ref, got %d", len(w.Workflows))
	}
	if w.Workflows[0].Name != "ProcessOrder" {
		t.Errorf("expected workflow ref 'ProcessOrder', got %q", w.Workflows[0].Name)
	}
	if len(w.Activities) != 1 {
		t.Fatalf("expected 1 activity ref, got %d", len(w.Activities))
	}
	if w.Activities[0].Name != "ChargePayment" {
		t.Errorf("expected activity ref 'ChargePayment', got %q", w.Activities[0].Name)
	}
}

func TestWorkerDefMultipleWorkflows(t *testing.T) {
	input := `workflow A(x: int) -> (int):
    return x

workflow B(x: int) -> (int):
    return x

activity C(x: int) -> (int):
    return x

activity D(x: int) -> (int):
    return x

worker multiWorker:
    workflow A
    workflow B
    activity C
    activity D
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w := file.Definitions[4].(*ast.WorkerDef)
	if len(w.Workflows) != 2 {
		t.Fatalf("expected 2 workflow refs, got %d", len(w.Workflows))
	}
	if len(w.Activities) != 2 {
		t.Fatalf("expected 2 activity refs, got %d", len(w.Activities))
	}
}

func TestWorkerDefEmptyBody(t *testing.T) {
	input := `worker emptyWorker:
    # just a comment
`
	// Worker with no workflow/activity refs should parse OK
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w := file.Definitions[0].(*ast.WorkerDef)
	if len(w.Workflows) != 0 {
		t.Errorf("expected 0 workflow refs, got %d", len(w.Workflows))
	}
	if len(w.Activities) != 0 {
		t.Errorf("expected 0 activity refs, got %d", len(w.Activities))
	}
}

func TestNamespaceDef(t *testing.T) {
	input := `worker orderTypes:
    workflow ProcessOrder
    activity ChargePayment

namespace orders:
    worker orderTypes
        options:
            task_queue: "orderProcessing"
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Definitions) != 2 {
		t.Fatalf("expected 2 definitions, got %d", len(file.Definitions))
	}
	ns, ok := file.Definitions[1].(*ast.NamespaceDef)
	if !ok {
		t.Fatalf("expected NamespaceDef, got %T", file.Definitions[1])
	}
	if ns.Name != "orders" {
		t.Errorf("expected name 'orders', got %q", ns.Name)
	}
	if len(ns.Workers) != 1 {
		t.Fatalf("expected 1 worker instantiation, got %d", len(ns.Workers))
	}
	if ns.Workers[0].Worker.Name != "orderTypes" {
		t.Errorf("expected worker ref 'orderTypes', got %q", ns.Workers[0].Worker.Name)
	}
	if ns.Workers[0].Options == nil {
		t.Fatal("expected options on worker instantiation")
	}
	if len(ns.Workers[0].Options.Entries) != 1 {
		t.Fatalf("expected 1 option entry, got %d", len(ns.Workers[0].Options.Entries))
	}
	if ns.Workers[0].Options.Entries[0].Key != "task_queue" {
		t.Errorf("expected option key 'task_queue', got %q", ns.Workers[0].Options.Entries[0].Key)
	}
	if ns.Workers[0].Options.Entries[0].Value != "orderProcessing" {
		t.Errorf("expected option value 'orderProcessing', got %q", ns.Workers[0].Options.Entries[0].Value)
	}
}

func TestNamespaceDefMultipleWorkers(t *testing.T) {
	input := `worker orderTypes:
    workflow ProcessOrder

worker paymentTypes:
    activity ChargePayment

namespace orders:
    worker orderTypes
        options:
            task_queue: "order-queue"
    worker paymentTypes
        options:
            task_queue: "payment-queue"
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ns := file.Definitions[2].(*ast.NamespaceDef)
	if len(ns.Workers) != 2 {
		t.Fatalf("expected 2 worker instantiations, got %d", len(ns.Workers))
	}
	if ns.Workers[0].Worker.Name != "orderTypes" {
		t.Errorf("expected first worker 'orderTypes', got %q", ns.Workers[0].Worker.Name)
	}
	if ns.Workers[1].Worker.Name != "paymentTypes" {
		t.Errorf("expected second worker 'paymentTypes', got %q", ns.Workers[1].Worker.Name)
	}
}

func TestNamespaceDefWorkerWithMultipleOptions(t *testing.T) {
	input := `worker orderTypes:
    workflow ProcessOrder

namespace orders:
    worker orderTypes
        options:
            task_queue: "order-queue"
            max_concurrent_activity_executions: 50
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ns := file.Definitions[1].(*ast.NamespaceDef)
	if ns.Workers[0].Options == nil {
		t.Fatal("expected options")
	}
	if len(ns.Workers[0].Options.Entries) != 2 {
		t.Fatalf("expected 2 option entries, got %d", len(ns.Workers[0].Options.Entries))
	}
	if ns.Workers[0].Options.Entries[1].Key != "max_concurrent_activity_executions" {
		t.Errorf("expected key 'max_concurrent_activity_executions', got %q", ns.Workers[0].Options.Entries[1].Key)
	}
}

func TestNamespaceDefNoEntries(t *testing.T) {
	input := `namespace empty:
    # empty namespace
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ns := file.Definitions[0].(*ast.NamespaceDef)
	if len(ns.Workers) != 0 {
		t.Errorf("expected 0 workers, got %d", len(ns.Workers))
	}
}

func TestNamespaceDefWorkerWithoutOptions(t *testing.T) {
	input := `worker orderTypes:
    workflow ProcessOrder

namespace orders:
    worker orderTypes
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ns := file.Definitions[1].(*ast.NamespaceDef)
	if len(ns.Workers) != 1 {
		t.Fatalf("expected 1 worker, got %d", len(ns.Workers))
	}
	if ns.Workers[0].Options != nil {
		t.Errorf("expected no options, got %v", ns.Workers[0].Options)
	}
}

// ===== NEXUS TESTS =====

func TestNexusServiceDef(t *testing.T) {
	input := `nexus service OrderService:
    async PlaceOrder workflow ProcessOrder
    sync GetStatus(orderId: string) -> (Status):
        activity FetchStatus(orderId) -> status
        close complete(status)

workflow ProcessOrder(order: Order) -> (Result):
    close complete(Result{})

activity FetchStatus(orderId: string) -> (Status):
    return getStatus(orderId)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	svc, ok := file.Definitions[0].(*ast.NexusServiceDef)
	if !ok {
		t.Fatalf("expected NexusServiceDef, got %T", file.Definitions[0])
	}
	if svc.Name != "OrderService" {
		t.Errorf("expected name 'OrderService', got %q", svc.Name)
	}
	if len(svc.Operations) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(svc.Operations))
	}

	asyncOp := svc.Operations[0]
	if asyncOp.OpType != ast.NexusOpAsync {
		t.Errorf("expected async operation, got %d", asyncOp.OpType)
	}
	if asyncOp.Name != "PlaceOrder" {
		t.Errorf("expected name 'PlaceOrder', got %q", asyncOp.Name)
	}
	if asyncOp.Workflow.Name != "ProcessOrder" {
		t.Errorf("expected workflow 'ProcessOrder', got %q", asyncOp.Workflow.Name)
	}

	syncOp := svc.Operations[1]
	if syncOp.OpType != ast.NexusOpSync {
		t.Errorf("expected sync operation, got %d", syncOp.OpType)
	}
	if syncOp.Name != "GetStatus" {
		t.Errorf("expected name 'GetStatus', got %q", syncOp.Name)
	}
	if syncOp.Params != "orderId: string" {
		t.Errorf("expected params 'orderId: string', got %q", syncOp.Params)
	}
	if syncOp.ReturnType != "Status" {
		t.Errorf("expected return type 'Status', got %q", syncOp.ReturnType)
	}
	if len(syncOp.Body) != 2 {
		t.Fatalf("expected 2 body statements, got %d", len(syncOp.Body))
	}
}

func TestNexusServiceDefEmptyBody(t *testing.T) {
	input := `nexus service EmptyService:
    # just a comment
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	svc := file.Definitions[0].(*ast.NexusServiceDef)
	if len(svc.Operations) != 0 {
		t.Errorf("expected 0 operations, got %d", len(svc.Operations))
	}
}

func TestNexusCallDetach(t *testing.T) {
	input := `workflow Foo():
    detach nexus Endpoint Svc.Op(args)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call, ok := wf.Body[0].(*ast.NexusCall)
	if !ok {
		t.Fatalf("expected NexusCall, got %T", wf.Body[0])
	}
	if !call.Detach {
		t.Error("expected detach true")
	}
	if call.Endpoint.Name != "Endpoint" {
		t.Errorf("expected endpoint 'Endpoint', got %q", call.Endpoint.Name)
	}
	if call.Service.Name != "Svc" {
		t.Errorf("expected service 'Svc', got %q", call.Service.Name)
	}
	if call.Operation.Name != "Op" {
		t.Errorf("expected operation 'Op', got %q", call.Operation.Name)
	}
	if call.Result != "" {
		t.Errorf("expected no result, got %q", call.Result)
	}
}

func TestNexusCallWithOptions(t *testing.T) {
	input := `workflow Foo():
    nexus Endpoint Svc.Op(args) -> result
        options:
            schedule_to_close_timeout: 30s
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call := wf.Body[0].(*ast.NexusCall)
	if call.Options == nil {
		t.Fatal("expected options block")
	}
	if len(call.Options.Entries) != 1 {
		t.Fatalf("expected 1 option entry, got %d", len(call.Options.Entries))
	}
	if call.Options.Entries[0].Key != "schedule_to_close_timeout" {
		t.Errorf("expected key 'schedule_to_close_timeout', got %q", call.Options.Entries[0].Key)
	}
}

func TestNexusCallPromise(t *testing.T) {
	input := `workflow Foo():
    promise p <- nexus Endpoint Svc.Op(args)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	promise, ok := wf.Body[0].(*ast.PromiseStmt)
	if !ok {
		t.Fatalf("expected PromiseStmt, got %T", wf.Body[0])
	}
	if promise.Name != "p" {
		t.Errorf("expected name 'p', got %q", promise.Name)
	}
	nt, ok := promise.Target.(*ast.NexusTarget)
	if !ok {
		t.Fatalf("expected NexusTarget, got %T", promise.Target)
	}
	if nt.Endpoint.Name != "Endpoint" {
		t.Errorf("expected nexus 'Endpoint', got %q", nt.Endpoint.Name)
	}
	if nt.Service.Name != "Svc" {
		t.Errorf("expected nexus service 'Svc', got %q", nt.Service.Name)
	}
	if nt.Operation.Name != "Op" {
		t.Errorf("expected nexus operation 'Op', got %q", nt.Operation.Name)
	}
	if nt.Args != "args" {
		t.Errorf("expected nexus args 'args', got %q", nt.Args)
	}
}

func TestNexusCallAwait(t *testing.T) {
	input := `workflow Foo():
    await nexus Endpoint Svc.Op(args) -> result
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	await, ok := wf.Body[0].(*ast.AwaitStmt)
	if !ok {
		t.Fatalf("expected AwaitStmt, got %T", wf.Body[0])
	}
	if ast.AsyncTargetKind(await.Target) != "nexus" {
		t.Errorf("expected kind 'nexus', got %q", ast.AsyncTargetKind(await.Target))
	}
	ant, ok := await.Target.(*ast.NexusTarget)
	if !ok {
		t.Fatalf("expected NexusTarget, got %T", await.Target)
	}
	if ant.Endpoint.Name != "Endpoint" {
		t.Errorf("expected nexus 'Endpoint', got %q", ant.Endpoint.Name)
	}
	if ant.Service.Name != "Svc" {
		t.Errorf("expected service 'Svc', got %q", ant.Service.Name)
	}
	if ant.Operation.Name != "Op" {
		t.Errorf("expected operation 'Op', got %q", ant.Operation.Name)
	}
	if ant.Result != "result" {
		t.Errorf("expected result 'result', got %q", ant.Result)
	}
}

func TestNexusCallAwaitOne(t *testing.T) {
	input := `workflow Foo():
    await one:
        nexus Endpoint Svc.Op(args) -> result:
            activity HandleResult(result)
        timer(5m):
            activity HandleTimeout()

activity HandleResult(result: Result):
    handle(result)

activity HandleTimeout():
    handleTimeout()
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	awaitOne, ok := wf.Body[0].(*ast.AwaitOneBlock)
	if !ok {
		t.Fatalf("expected AwaitOneBlock, got %T", wf.Body[0])
	}
	if len(awaitOne.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(awaitOne.Cases))
	}
	nexusCase := awaitOne.Cases[0]
	if ast.AsyncTargetKind(nexusCase.Target) != "nexus" {
		t.Errorf("expected kind 'nexus', got %q", ast.AsyncTargetKind(nexusCase.Target))
	}
	nct, ok := nexusCase.Target.(*ast.NexusTarget)
	if !ok {
		t.Fatalf("expected NexusTarget, got %T", nexusCase.Target)
	}
	if nct.Endpoint.Name != "Endpoint" {
		t.Errorf("expected nexus 'Endpoint', got %q", nct.Endpoint.Name)
	}
	if nct.Service.Name != "Svc" {
		t.Errorf("expected service 'Svc', got %q", nct.Service.Name)
	}
	if nct.Operation.Name != "Op" {
		t.Errorf("expected operation 'Op', got %q", nct.Operation.Name)
	}
	if nct.Result != "result" {
		t.Errorf("expected result 'result', got %q", nct.Result)
	}
	if len(nexusCase.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(nexusCase.Body))
	}
}

func TestWorkerDefWithNexusService(t *testing.T) {
	input := `worker OrderWorker:
    workflow ProcessOrder
    activity ChargePayment
    nexus service OrderService

workflow ProcessOrder(order: Order) -> (Result):
    close complete(Result{})

activity ChargePayment(order: Order) -> (Payment):
    return charge(order)

nexus service OrderService:
    async PlaceOrder workflow ProcessOrder
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	worker, ok := file.Definitions[0].(*ast.WorkerDef)
	if !ok {
		t.Fatalf("expected WorkerDef, got %T", file.Definitions[0])
	}
	if len(worker.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(worker.Services))
	}
	if worker.Services[0].Name != "OrderService" {
		t.Errorf("expected service 'OrderService', got %q", worker.Services[0].Name)
	}
}

func TestNamespaceDefWithEndpoint(t *testing.T) {
	input := `worker OrderWorker:
    workflow ProcessOrder

namespace Orders:
    worker OrderWorker
        options:
            task_queue: "orders"
    nexus endpoint OrderEndpoint
        options:
            task_queue: "orders"

workflow ProcessOrder(order: Order) -> (Result):
    close complete(Result{})
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ns := file.Definitions[1].(*ast.NamespaceDef)
	if len(ns.Endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(ns.Endpoints))
	}
	ep := ns.Endpoints[0]
	if ep.EndpointName != "OrderEndpoint" {
		t.Errorf("expected endpoint 'OrderEndpoint', got %q", ep.EndpointName)
	}
	if ep.Options == nil {
		t.Fatal("expected endpoint options")
	}
	if len(ep.Options.Entries) != 1 {
		t.Fatalf("expected 1 option entry, got %d", len(ep.Options.Entries))
	}
	if ep.Options.Entries[0].Key != "task_queue" {
		t.Errorf("expected key 'task_queue', got %q", ep.Options.Entries[0].Key)
	}
	if ep.Options.Entries[0].Value != "orders" {
		t.Errorf("expected value 'orders', got %q", ep.Options.Entries[0].Value)
	}
}
