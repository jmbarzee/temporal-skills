package parser

import (
	"testing"

	"github.com/jmbarzee/temporal-skills/design/parser/ast"
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
    signal PaymentReceived(transactionId: string, amount: decimal)
    query GetStatus() -> (OrderStatus)
    update ChangeAddress(addr: Address) -> (UpdateResult)

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
	if len(wf.Queries) != 1 {
		t.Fatalf("expected 1 query, got %d", len(wf.Queries))
	}
	if wf.Queries[0].ReturnType != "OrderStatus" {
		t.Errorf("expected query return type 'OrderStatus', got %q", wf.Queries[0].ReturnType)
	}
	if len(wf.Updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(wf.Updates))
	}
	if wf.Updates[0].Name != "ChangeAddress" {
		t.Errorf("expected update name 'ChangeAddress', got %q", wf.Updates[0].Name)
	}
	if len(wf.Body) != 2 {
		t.Fatalf("expected 2 body statements, got %d", len(wf.Body))
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

func TestAwaitSingle(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    await signal PaymentReceived
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	aw, ok := wf.Body[0].(*ast.AwaitStmt)
	if !ok {
		t.Fatalf("expected AwaitStmt, got %T", wf.Body[0])
	}
	if len(aw.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(aw.Targets))
	}
	if aw.Targets[0].Kind != "signal" {
		t.Errorf("expected kind 'signal', got %q", aw.Targets[0].Kind)
	}
	if aw.Targets[0].Name != "PaymentReceived" {
		t.Errorf("expected name 'PaymentReceived', got %q", aw.Targets[0].Name)
	}
}

func TestAwaitMultiTarget(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    await signal PaymentReceived or update ChangeAddress
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	aw := wf.Body[0].(*ast.AwaitStmt)
	if len(aw.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(aw.Targets))
	}
	if aw.Targets[0].Kind != "signal" {
		t.Errorf("target[0] kind: expected 'signal', got %q", aw.Targets[0].Kind)
	}
	if aw.Targets[1].Kind != "update" {
		t.Errorf("target[1] kind: expected 'update', got %q", aw.Targets[1].Kind)
	}
	if aw.Targets[1].Name != "ChangeAddress" {
		t.Errorf("target[1] name: expected 'ChangeAddress', got %q", aw.Targets[1].Name)
	}
}

func TestParallelBlock(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    parallel:
        activity ReserveInventory(order) -> inventory
        activity ProcessPayment(order) -> payment
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	par, ok := wf.Body[0].(*ast.ParallelBlock)
	if !ok {
		t.Fatalf("expected ParallelBlock, got %T", wf.Body[0])
	}
	if len(par.Body) != 2 {
		t.Fatalf("expected 2 body statements, got %d", len(par.Body))
	}
}

func TestSelectBlock(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    select:
        workflow ProcessPayment(order) -> paymentResult:
            activity HandlePayment(paymentResult)
        signal PaymentReceived:
            activity FulfillOrder(order)
        timer 24h:
            activity CancelOrder(orderId)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	sel, ok := wf.Body[0].(*ast.SelectBlock)
	if !ok {
		t.Fatalf("expected SelectBlock, got %T", wf.Body[0])
	}
	if len(sel.Cases) != 3 {
		t.Fatalf("expected 3 cases, got %d", len(sel.Cases))
	}
	if sel.Cases[0].CaseKind() != "workflow" {
		t.Errorf("case[0]: expected workflow, got %q", sel.Cases[0].CaseKind())
	}
	if sel.Cases[1].CaseKind() != "signal" {
		t.Errorf("case[1]: expected signal, got %q", sel.Cases[1].CaseKind())
	}
	if sel.Cases[2].CaseKind() != "timer" {
		t.Errorf("case[2]: expected timer, got %q", sel.Cases[2].CaseKind())
	}
	if sel.Cases[2].TimerDuration != "24h" {
		t.Errorf("case[2] timer: expected '24h', got %q", sel.Cases[2].TimerDuration)
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
	if len(wf.Body) != 2 {
		t.Fatalf("expected 2 body statements, got %d", len(wf.Body))
	}
	comment, ok := wf.Body[0].(*ast.Comment)
	if !ok {
		t.Fatalf("expected Comment, got %T", wf.Body[0])
	}
	if comment.Text != " this is a comment" {
		t.Errorf("unexpected comment text: %q", comment.Text)
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
    select:
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

func TestSelectWithActivityCase(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    select:
        activity GetOrder(orderId) -> order:
            return order
        update ChangeAddress(addr):
            return addr
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	sel := wf.Body[0].(*ast.SelectBlock)
	if len(sel.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(sel.Cases))
	}
	if sel.Cases[0].CaseKind() != "activity" {
		t.Errorf("case[0]: expected activity, got %q", sel.Cases[0].CaseKind())
	}
	if sel.Cases[1].CaseKind() != "update" {
		t.Errorf("case[1]: expected update, got %q", sel.Cases[1].CaseKind())
	}
}

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

func TestAwaitWithArgs(t *testing.T) {
	input := `workflow Foo(x: int) -> (Result):
    await signal PaymentReceived(txId)
`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := file.Definitions[0].(*ast.WorkflowDef)
	aw := wf.Body[0].(*ast.AwaitStmt)
	if aw.Targets[0].Args != "txId" {
		t.Errorf("expected args 'txId', got %q", aw.Targets[0].Args)
	}
}

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

func TestFullWorkflow(t *testing.T) {
	input := `workflow OrderFulfillment(orderId: string) -> (OrderResult):
    signal PaymentReceived(transactionId: string, amount: decimal)
    signal OrderCancelled(reason: string)
    query GetStatus() -> (OrderStatus)
    update ChangeAddress(addr: Address) -> (UpdateResult)

    activity GetOrder(orderId) -> order

    if (order.priority == "high"):
        activity ExpediteOrder(order)
    else:
        activity StandardProcessing(order)

    parallel:
        activity ReserveInventory(order) -> inventory
        activity ProcessPayment(order) -> payment

    for (item in order.items):
        activity ProcessItem(item)

    select:
        workflow ProcessPayment(order) -> paymentResult:
            activity HandlePayment(paymentResult)
        signal PaymentReceived:
            activity FulfillOrder(order)
        timer 24h:
            activity CancelOrder(orderId)

    timer 1h
    await signal PaymentReceived or update ChangeAddress
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
