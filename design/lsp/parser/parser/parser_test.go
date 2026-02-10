package parser

import (
	"os"
	"strings"
	"testing"

	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
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

func TestHintStmt(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    signal Cancel(reason: string):
        return Result{cancelled: true}
    update ChangeAddr(addr: Addr) -> (Result):
        return Result{ok: true}

    hint signal Cancel
    hint update ChangeAddr
    return Result{}
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	if len(wf.Body) != 3 {
		t.Fatalf("expected 3 body statements, got %d", len(wf.Body))
	}
	hint1, ok := wf.Body[0].(*ast.HintStmt)
	if !ok {
		t.Fatalf("expected HintStmt, got %T", wf.Body[0])
	}
	if hint1.Kind != "signal" {
		t.Errorf("expected kind 'signal', got %q", hint1.Kind)
	}
	if hint1.Name != "Cancel" {
		t.Errorf("expected name 'Cancel', got %q", hint1.Name)
	}

	hint2, ok := wf.Body[1].(*ast.HintStmt)
	if !ok {
		t.Fatalf("expected HintStmt, got %T", wf.Body[1])
	}
	if hint2.Kind != "update" {
		t.Errorf("expected kind 'update', got %q", hint2.Kind)
	}
	if hint2.Name != "ChangeAddr" {
		t.Errorf("expected name 'ChangeAddr', got %q", hint2.Name)
	}
}

func TestHintInvalidTarget(t *testing.T) {
	// hint now supports signal, query, and update - all are valid.
	// Test that hint requires a valid target keyword.
	input := `workflow Foo(x: int) -> (Result):
    hint something Invalid
`
	_, err := ParseFile(input)
	if err == nil {
		t.Fatal("expected error for invalid hint target, got nil")
	}
}

func TestAwaitOneCaseError(t *testing.T) {
	// await one cases only support 'timer' and nested 'await all', not signal/update.
	input := `workflow Foo(x: int) -> (Result):
    await one:
        signal PaymentReceived:
            return x
`
	_, err := ParseFile(input)
	if err == nil {
		t.Fatal("expected error for signal in await one case, got nil")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected ParseError, got %T: %v", err, err)
	}
	// The error message should indicate signal is not allowed in await one cases.
	if !strings.Contains(pe.Msg, "unexpected token SIGNAL") {
		t.Errorf("unexpected error message: %q", pe.Msg)
	}
}

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
	if call.Name != "GetOrder" {
		t.Errorf("expected name 'GetOrder', got %q", call.Name)
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
	if call.Name != "ShipOrder" {
		t.Errorf("expected name 'ShipOrder', got %q", call.Name)
	}
	if call.Result != "shipResult" {
		t.Errorf("expected result 'shipResult', got %q", call.Result)
	}
}

func TestWorkflowCallSpawn(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    spawn workflow ProcessAsync(data) -> result
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call := wf.Body[0].(*ast.WorkflowCall)
	if call.Mode != ast.CallSpawn {
		t.Errorf("expected CallSpawn, got %d", call.Mode)
	}
	if call.Result != "result" {
		t.Errorf("expected result 'result', got %q", call.Result)
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

func TestWorkflowCallNexus(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    nexus "payments" workflow Charge(card) -> chargeResult
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call := wf.Body[0].(*ast.WorkflowCall)
	if call.Mode != ast.CallChild {
		t.Errorf("expected CallChild, got %d", call.Mode)
	}
	if call.Namespace != "payments" {
		t.Errorf("expected namespace 'payments', got %q", call.Namespace)
	}
	if call.Name != "Charge" {
		t.Errorf("expected name 'Charge', got %q", call.Name)
	}
	if call.Result != "chargeResult" {
		t.Errorf("expected result 'chargeResult', got %q", call.Result)
	}
}

func TestSpawnNexusWorkflowCall(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    spawn nexus "shipping" workflow Ship(order) -> shipment
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call := wf.Body[0].(*ast.WorkflowCall)
	if call.Mode != ast.CallSpawn {
		t.Errorf("expected CallSpawn, got %d", call.Mode)
	}
	if call.Namespace != "shipping" {
		t.Errorf("expected namespace 'shipping', got %q", call.Namespace)
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
        options(startToCloseTimeout: 30s, retryPolicy: {maxAttempts: 3})
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call := wf.Body[0].(*ast.ActivityCall)
	if call.Options != "startToCloseTimeout: 30s, retryPolicy: {maxAttempts: 3}" {
		t.Errorf("unexpected options: %q", call.Options)
	}
}

func TestOptionsOnWorkflowCall(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    workflow ShipOrder(order) -> result
        options(workflowExecutionTimeout: 24h)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	call := wf.Body[0].(*ast.WorkflowCall)
	if call.Options != "workflowExecutionTimeout: 24h" {
		t.Errorf("unexpected options: %q", call.Options)
	}
}

func TestOptionsOnDefinition(t *testing.T) {
	input := `activity GetOrder(orderId: string) -> (Order):
    options(startToCloseTimeout: 10s)
    order = db.get(orderId)
    return order
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	act := file.Definitions[0].(*ast.ActivityDef)
	if act.Options != "startToCloseTimeout: 10s" {
		t.Errorf("unexpected options: %q", act.Options)
	}
	if len(act.Body) != 2 {
		t.Fatalf("expected 2 body statements, got %d", len(act.Body))
	}
}

func TestTimer(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    timer 1h
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	timer, ok := wf.Body[0].(*ast.TimerStmt)
	if !ok {
		t.Fatalf("expected TimerStmt, got %T", wf.Body[0])
	}
	if timer.Duration != "1h" {
		t.Errorf("expected duration '1h', got %q", timer.Duration)
	}
}

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
	// await one cases support watch, timer (with bodies), and nested await all.
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
	if awaitOne.Cases[0].CaseKind() != "timer" {
		t.Errorf("case[0]: expected timer, got %q", awaitOne.Cases[0].CaseKind())
	}
	if awaitOne.Cases[0].TimerDuration != "1h" {
		t.Errorf("case[0] timer: expected '1h', got %q", awaitOne.Cases[0].TimerDuration)
	}
	if len(awaitOne.Cases[0].Body) != 1 {
		t.Errorf("case[0] body: expected 1 statement, got %d", len(awaitOne.Cases[0].Body))
	}
	if awaitOne.Cases[1].CaseKind() != "timer" {
		t.Errorf("case[1]: expected timer, got %q", awaitOne.Cases[1].CaseKind())
	}
	if awaitOne.Cases[1].TimerDuration != "24h" {
		t.Errorf("case[1] timer: expected '24h', got %q", awaitOne.Cases[1].TimerDuration)
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
    continue_as_new(newArgs)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	can, ok := wf.Body[0].(*ast.ContinueAsNewStmt)
	if !ok {
		t.Fatalf("expected ContinueAsNewStmt, got %T", wf.Body[0])
	}
	if can.Args != "newArgs" {
		t.Errorf("expected args 'newArgs', got %q", can.Args)
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

func TestSelectDetachError(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    await one:
        detach workflow Send(x):
            return x
`
	_, err := ParseFile(input)
	if err == nil {
		t.Fatal("expected error for detach in select case, got nil")
	}
}

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
        options(startToCloseTimeout: 30s)
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
	if call1.Options != "startToCloseTimeout: 30s" {
		t.Errorf("expected options, got %q", call1.Options)
	}
	call2 := wf.Body[1].(*ast.ActivityCall)
	if call2.Options != "" {
		t.Errorf("expected no options, got %q", call2.Options)
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

func TestWorkflowDefWithOptions(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    options(workflowExecutionTimeout: 24h)
    return x
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	if wf.Options != "workflowExecutionTimeout: 24h" {
		t.Errorf("expected options, got %q", wf.Options)
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

    hint signal PaymentReceived
    hint update ChangeAddress

    await one:
        timer (1h):
            activity CheckPayment(orderId) -> checkResult
        timer (24h):
            activity CancelOrder(orderId)
            close failed OrderResult{status: "timeout"}

    workflow ShipOrder(order) -> shipResult
    spawn workflow ProcessAsync(data) -> result
    detach workflow SendNotification(order.customer)
    nexus "payments" workflow Charge(card) -> chargeResult

    switch (order.type):
        case "invoice":
            activity ProcessInvoice(order)
        else:
            activity HandleDefault(order)

    continue_as_new(orderId)
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

func TestWatchCase(t *testing.T) {
	input := `workflow Test() -> (Result):
    signal Done():
        done = true

    done = false

    await one:
        watch (done):
            hint signal Done
            close Result{success: true}
        timer (7d):
            close failed Result{success: false}
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	awaitOne, ok := wf.Body[1].(*ast.AwaitOneBlock)
	if !ok {
		t.Fatalf("expected AwaitOneBlock, got %T", wf.Body[1])
	}
	if len(awaitOne.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(awaitOne.Cases))
	}

	// Check watch case
	if awaitOne.Cases[0].CaseKind() != "watch" {
		t.Errorf("case[0]: expected watch, got %q", awaitOne.Cases[0].CaseKind())
	}
	if awaitOne.Cases[0].WatchVariable != "done" {
		t.Errorf("case[0] watch: expected 'done', got %q", awaitOne.Cases[0].WatchVariable)
	}
	if len(awaitOne.Cases[0].Body) != 2 {
		t.Errorf("case[0] body: expected 2 statements, got %d", len(awaitOne.Cases[0].Body))
	}

	// Check timer case
	if awaitOne.Cases[1].CaseKind() != "timer" {
		t.Errorf("case[1]: expected timer, got %q", awaitOne.Cases[1].CaseKind())
	}
	if awaitOne.Cases[1].TimerDuration != "7d" {
		t.Errorf("case[1] timer: expected '7d', got %q", awaitOne.Cases[1].TimerDuration)
	}
	if len(awaitOne.Cases[1].Body) != 1 {
		t.Errorf("case[1] body: expected 1 statement, got %d", len(awaitOne.Cases[1].Body))
	}
}

func TestTimerCaseWithBody(t *testing.T) {
	input := `workflow Test():
    await one:
        timer (5m):
            activity SendReminder()
            close
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
	if c.CaseKind() != "timer" {
		t.Errorf("expected timer case, got %q", c.CaseKind())
	}
	if c.TimerDuration != "5m" {
		t.Errorf("timer duration: expected '5m', got %q", c.TimerDuration)
	}
	if len(c.Body) != 2 {
		t.Errorf("expected 2 statements in timer body, got %d", len(c.Body))
	}
}

func TestCloseStatement(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		reason string
		value  string
	}{
		{
			name:   "plain close",
			input:  "workflow Test():\n    close\n",
			reason: "",
			value:  "",
		},
		{
			name:   "close completed",
			input:  "workflow Test():\n    close completed\n",
			reason: "completed",
			value:  "",
		},
		{
			name:   "close failed",
			input:  "workflow Test():\n    close failed\n",
			reason: "failed",
			value:  "",
		},
		{
			name:   "close with value",
			input:  "workflow Test():\n    close Result{success: true}\n",
			reason: "",
			value:  "Result{success: true}",
		},
		{
			name:   "close failed with value",
			input:  "workflow Test():\n    close failed Error{message: \"timeout\"}\n",
			reason: "failed",
			value:  "Error{message: timeout  }",
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
				t.Errorf("reason: expected %q, got %q", tt.reason, closeStmt.Reason)
			}
			if closeStmt.Value != tt.value {
				t.Errorf("value: expected %q, got %q", tt.value, closeStmt.Value)
			}
		})
	}
}

func TestWatchWithMultipleHints(t *testing.T) {
	input := `workflow Approval():
    signal Approved():
        approved = true

    signal AdminOverride():
        approved = true

    approved = false

    await one:
        watch (approved):
            hint signal Approved
            hint signal AdminOverride
            close
        timer (7d):
            close failed
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	awaitOne, ok := wf.Body[1].(*ast.AwaitOneBlock)
	if !ok {
		t.Fatalf("expected AwaitOneBlock, got %T", wf.Body[1])
	}

	watchCase := awaitOne.Cases[0]
	if watchCase.CaseKind() != "watch" {
		t.Errorf("expected watch case, got %q", watchCase.CaseKind())
	}
	if len(watchCase.Body) != 3 {
		t.Errorf("expected 3 statements (2 hints + close), got %d", len(watchCase.Body))
	}
}

func TestMultipleWatchCases(t *testing.T) {
	input := `workflow Test():
    signal Approved():
        approved = true

    signal Rejected():
        rejected = true

    approved = false
    rejected = false

    await one:
        watch (approved):
            hint signal Approved
            close
        watch (rejected):
            hint signal Rejected
            close failed
        timer (7d):
            close failed "timeout"
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	awaitOne, ok := wf.Body[2].(*ast.AwaitOneBlock)
	if !ok {
		t.Fatalf("expected AwaitOneBlock, got %T", wf.Body[2])
	}

	if len(awaitOne.Cases) != 3 {
		t.Fatalf("expected 3 cases, got %d", len(awaitOne.Cases))
	}

	// Verify all case kinds
	if awaitOne.Cases[0].CaseKind() != "watch" {
		t.Errorf("case[0]: expected watch, got %q", awaitOne.Cases[0].CaseKind())
	}
	if awaitOne.Cases[0].WatchVariable != "approved" {
		t.Errorf("case[0]: expected watch 'approved', got %q", awaitOne.Cases[0].WatchVariable)
	}

	if awaitOne.Cases[1].CaseKind() != "watch" {
		t.Errorf("case[1]: expected watch, got %q", awaitOne.Cases[1].CaseKind())
	}
	if awaitOne.Cases[1].WatchVariable != "rejected" {
		t.Errorf("case[1]: expected watch 'rejected', got %q", awaitOne.Cases[1].WatchVariable)
	}

	if awaitOne.Cases[2].CaseKind() != "timer" {
		t.Errorf("case[2]: expected timer, got %q", awaitOne.Cases[2].CaseKind())
	}
}
