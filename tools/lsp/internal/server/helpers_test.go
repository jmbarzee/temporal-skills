package server

import (
	"testing"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/parser"
)

// mustParseWorkflowBody parses a workflow with the given body and returns
// the body statements.
func mustParseWorkflowBody(t *testing.T, body string) []ast.Statement {
	t.Helper()
	input := "workflow Test():\n" + body
	file, err := parser.ParseFile(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(file.Definitions) == 0 {
		t.Fatal("no definitions parsed")
	}
	wf, ok := file.Definitions[0].(*ast.WorkflowDef)
	if !ok {
		t.Fatalf("expected WorkflowDef, got %T", file.Definitions[0])
	}
	return wf.Body
}

func TestLastLineInStmts(t *testing.T) {
	body := mustParseWorkflowBody(t, "    activity Foo()\n    activity Bar()\n")
	result := lastLineInStmts(body, 0)
	if result != 3 {
		t.Errorf("expected last line 3, got %d", result)
	}
}

func TestLastLineInStmtsNested(t *testing.T) {
	body := mustParseWorkflowBody(t,
		"    if (cond):\n"+
			"        activity Foo()\n"+
			"        activity Bar()\n")
	result := lastLineInStmts(body, 0)
	if result != 4 {
		t.Errorf("expected last line 4, got %d", result)
	}
}

func TestLastLineInStmtsEmpty(t *testing.T) {
	result := lastLineInStmts(nil, 5)
	if result != 5 {
		t.Errorf("expected 5 (unchanged), got %d", result)
	}
}

func TestFindNodeInStmts(t *testing.T) {
	body := mustParseWorkflowBody(t,
		"    activity Foo()\n"+
			"    activity Bar()\n")

	// Line 2 = "activity Foo()"
	node := findNodeInStmts(body, 2)
	if node == nil {
		t.Fatal("expected to find node at line 2")
	}
	call, ok := node.(*ast.ActivityCall)
	if !ok {
		t.Fatalf("expected ActivityCall, got %T", node)
	}
	if call.Name != "Foo" {
		t.Errorf("expected name 'Foo', got %q", call.Name)
	}

	// Line 3 = "activity Bar()"
	node = findNodeInStmts(body, 3)
	if node == nil {
		t.Fatal("expected to find node at line 3")
	}
	call, ok = node.(*ast.ActivityCall)
	if !ok {
		t.Fatalf("expected ActivityCall, got %T", node)
	}
	if call.Name != "Bar" {
		t.Errorf("expected name 'Bar', got %q", call.Name)
	}
}

func TestFindNodeInStmtsNested(t *testing.T) {
	body := mustParseWorkflowBody(t,
		"    if (cond):\n"+
			"        activity Foo()\n")

	// Line 3 = "activity Foo()" inside the if body
	node := findNodeInStmts(body, 3)
	if node == nil {
		t.Fatal("expected to find node at line 3")
	}
	if _, ok := node.(*ast.ActivityCall); !ok {
		t.Fatalf("expected ActivityCall, got %T", node)
	}
}

func TestFindNodeInStmtsNotFound(t *testing.T) {
	body := mustParseWorkflowBody(t, "    activity Foo()\n")
	node := findNodeInStmts(body, 999)
	if node != nil {
		t.Errorf("expected nil for non-existent line, got %T", node)
	}
}

func TestFindCallInStatements(t *testing.T) {
	body := mustParseWorkflowBody(t,
		"    if (cond):\n"+
			"        activity Foo(x)\n"+
			"    activity Bar(y)\n")

	// Should find Foo nested in if body.
	call := findCallInStatements(body, "Foo")
	if call == nil {
		t.Fatal("expected to find activity call 'Foo'")
	}
	if call.Name != "Foo" {
		t.Errorf("expected 'Foo', got %q", call.Name)
	}

	// Should find Bar at top level.
	call = findCallInStatements(body, "Bar")
	if call == nil {
		t.Fatal("expected to find activity call 'Bar'")
	}

	// Should return nil for non-existent call.
	call = findCallInStatements(body, "Baz")
	if call != nil {
		t.Errorf("expected nil for non-existent call, got %v", call)
	}
}

func TestFindCallInStatementsSwitch(t *testing.T) {
	// Verify the bug fix: findCallInStatements now searches SwitchBlock.
	body := mustParseWorkflowBody(t,
		"    switch (x):\n"+
			"        case a:\n"+
			"            activity Foo()\n")

	call := findCallInStatements(body, "Foo")
	if call == nil {
		t.Fatal("expected to find activity call 'Foo' inside switch block")
	}
}

func TestFindReturnStatements(t *testing.T) {
	body := mustParseWorkflowBody(t,
		"    if (cond):\n"+
			"        return x\n"+
			"    return y\n")

	returns := findReturnStatements(body)
	if len(returns) != 2 {
		t.Fatalf("expected 2 return statements, got %d", len(returns))
	}
}

func TestFindReturnStatementsNone(t *testing.T) {
	body := mustParseWorkflowBody(t, "    activity Foo()\n")
	returns := findReturnStatements(body)
	if len(returns) != 0 {
		t.Errorf("expected 0 return statements, got %d", len(returns))
	}
}

func TestCollectRefsInStmts(t *testing.T) {
	body := mustParseWorkflowBody(t,
		"    activity Foo()\n"+
			"    activity Bar()\n"+
			"    activity Foo()\n")

	refs := collectRefsInStmts(body, "Foo", "activity", nil)
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs for 'Foo', got %d", len(refs))
	}
}

func TestCollectRefsInStmtsNested(t *testing.T) {
	body := mustParseWorkflowBody(t,
		"    if (cond):\n"+
			"        activity Foo()\n"+
			"    for (true):\n"+
			"        activity Foo()\n")

	refs := collectRefsInStmts(body, "Foo", "activity", nil)
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs for 'Foo' across nested blocks, got %d", len(refs))
	}
}

func TestCollectRefsInStmtsWorkflow(t *testing.T) {
	body := mustParseWorkflowBody(t,
		"    workflow Child(x)\n"+
			"    activity Foo()\n")

	refs := collectRefsInStmts(body, "Child", "workflow", nil)
	if len(refs) != 1 {
		t.Fatalf("expected 1 ref for workflow 'Child', got %d", len(refs))
	}

	// Should not find it when looking for activity kind.
	refs = collectRefsInStmts(body, "Child", "activity", nil)
	if len(refs) != 0 {
		t.Fatalf("expected 0 refs for activity 'Child', got %d", len(refs))
	}
}
