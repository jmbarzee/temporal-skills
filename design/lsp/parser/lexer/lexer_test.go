package lexer

import (
	"testing"

	"github.com/jmbarzee/temporal-skills/design/lsp/parser/token"
)

func TestKeywords(t *testing.T) {
	input := "workflow activity signal query update spawn detach nexus timer options await or parallel select switch case if else for in return continue_as_new break continue"
	expected := []token.TokenType{
		token.WORKFLOW, token.ACTIVITY, token.SIGNAL, token.QUERY, token.UPDATE,
		token.SPAWN, token.DETACH, token.NEXUS, token.TIMER, token.OPTIONS,
		token.AWAIT, token.OR, token.PARALLEL, token.SELECT, token.SWITCH,
		token.CASE, token.IF, token.ELSE, token.FOR, token.IN,
		token.RETURN, token.CONTINUE_AS_NEW, token.BREAK, token.CONTINUE,
		token.NEWLINE, token.EOF,
	}
	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("token[%d]: expected %s, got %s (%q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestContinueVsContinueAsNew(t *testing.T) {
	input := "continue\ncontinue_as_new"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.CONTINUE {
		t.Fatalf("expected CONTINUE, got %s", tok.Type)
	}
	l.NextToken() // NEWLINE
	tok = l.NextToken()
	if tok.Type != token.CONTINUE_AS_NEW {
		t.Fatalf("expected CONTINUE_AS_NEW, got %s", tok.Type)
	}
}

func TestIdentifier(t *testing.T) {
	input := "OrderFulfillment"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.IDENT {
		t.Fatalf("expected IDENT, got %s", tok.Type)
	}
	if tok.Literal != "OrderFulfillment" {
		t.Fatalf("expected literal 'OrderFulfillment', got %q", tok.Literal)
	}
}

func TestSingleLevelIndent(t *testing.T) {
	input := "workflow:\n    body\n"
	expected := []token.TokenType{
		token.WORKFLOW, token.COLON, token.NEWLINE,
		token.INDENT, token.IDENT, token.NEWLINE,
		token.DEDENT, token.EOF,
	}
	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("token[%d]: expected %s, got %s (%q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestNestedIndent(t *testing.T) {
	input := "a:\n    b:\n        c\n"
	expected := []token.TokenType{
		token.IDENT, token.COLON, token.NEWLINE,
		token.INDENT, token.IDENT, token.COLON, token.NEWLINE,
		token.INDENT, token.IDENT, token.NEWLINE,
		token.DEDENT, token.DEDENT, token.EOF,
	}
	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("token[%d]: expected %s, got %s (%q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestMultiLevelDedent(t *testing.T) {
	input := "a:\n    b:\n        c\nd\n"
	expected := []token.TokenType{
		token.IDENT, token.COLON, token.NEWLINE,
		token.INDENT, token.IDENT, token.COLON, token.NEWLINE,
		token.INDENT, token.IDENT, token.NEWLINE,
		token.DEDENT, token.DEDENT,
		token.IDENT, token.NEWLINE,
		token.EOF,
	}
	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("token[%d]: expected %s, got %s (%q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestBlankLineSkipping(t *testing.T) {
	input := "a:\n    b\n\n    c\n"
	expected := []token.TokenType{
		token.IDENT, token.COLON, token.NEWLINE,
		token.INDENT, token.IDENT, token.NEWLINE,
		// blank line skipped
		token.IDENT, token.NEWLINE,
		token.DEDENT, token.EOF,
	}
	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("token[%d]: expected %s, got %s (%q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestBlankLineWithSpacesSkipping(t *testing.T) {
	input := "a:\n    b\n    \n    c\n"
	expected := []token.TokenType{
		token.IDENT, token.COLON, token.NEWLINE,
		token.INDENT, token.IDENT, token.NEWLINE,
		// blank line (spaces only) skipped
		token.IDENT, token.NEWLINE,
		token.DEDENT, token.EOF,
	}
	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("token[%d]: expected %s, got %s (%q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestArgs(t *testing.T) {
	input := "foo(bar, baz)"
	l := New(input)
	tok := l.NextToken() // IDENT "foo"
	if tok.Type != token.IDENT {
		t.Fatalf("expected IDENT, got %s", tok.Type)
	}
	tok = l.NextToken() // ARGS
	if tok.Type != token.ARGS {
		t.Fatalf("expected ARGS, got %s", tok.Type)
	}
	if tok.Literal != "bar, baz" {
		t.Fatalf("expected args literal 'bar, baz', got %q", tok.Literal)
	}
}

func TestArgsNoNestedParens(t *testing.T) {
	// First ) closes, so (a(b) captures "a(b" and the remaining ) is raw text.
	input := "(a(b)"
	l := New(input)
	tok := l.NextToken() // ARGS
	if tok.Type != token.ARGS {
		t.Fatalf("expected ARGS, got %s", tok.Type)
	}
	// First ) closes: content is "a(b"
	// Actually, we don't track nested parens — the first ) closes.
	// Input: ( a ( b )
	// The ( at pos 2 is part of the content. The ) at pos 4 closes.
	// Content = "a(b"
	if tok.Literal != "a(b" {
		t.Fatalf("expected 'a(b', got %q", tok.Literal)
	}
}

func TestString(t *testing.T) {
	input := `"payments"`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.STRING {
		t.Fatalf("expected STRING, got %s", tok.Type)
	}
	if tok.Literal != "payments" {
		t.Fatalf("expected 'payments', got %q", tok.Literal)
	}
}

func TestArrow(t *testing.T) {
	input := "a -> b"
	expected := []token.TokenType{
		token.IDENT, token.ARROW, token.IDENT, token.NEWLINE, token.EOF,
	}
	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("token[%d]: expected %s, got %s (%q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestComment(t *testing.T) {
	input := "# this is a comment\n"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.COMMENT {
		t.Fatalf("expected COMMENT, got %s", tok.Type)
	}
	if tok.Literal != " this is a comment" {
		t.Fatalf("expected ' this is a comment', got %q", tok.Literal)
	}
}

func TestColon(t *testing.T) {
	input := "workflow:"
	l := New(input)
	l.NextToken() // WORKFLOW
	tok := l.NextToken()
	if tok.Type != token.COLON {
		t.Fatalf("expected COLON, got %s", tok.Type)
	}
}

func TestEOFDedentEmission(t *testing.T) {
	input := "a:\n    b:\n        c"
	expected := []token.TokenType{
		token.IDENT, token.COLON, token.NEWLINE,
		token.INDENT, token.IDENT, token.COLON, token.NEWLINE,
		token.INDENT, token.IDENT,
		// no trailing newline, but should still get dedents + EOF
		token.NEWLINE, token.DEDENT, token.DEDENT, token.EOF,
	}
	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("token[%d]: expected %s, got %s (%q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestWorkflowHeaderTokenStream(t *testing.T) {
	input := `workflow OrderFulfillment(orderId: string) -> (OrderResult):
    activity GetOrder(orderId) -> order
`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.WORKFLOW, "workflow"},
		{token.IDENT, "OrderFulfillment"},
		{token.ARGS, "orderId: string"},
		{token.ARROW, "->"},
		{token.ARGS, "OrderResult"},
		{token.COLON, ":"},
		{token.NEWLINE, ""},
		{token.INDENT, ""},
		{token.ACTIVITY, "activity"},
		{token.IDENT, "GetOrder"},
		{token.ARGS, "orderId"},
		{token.ARROW, "->"},
		{token.IDENT, "order"},
		{token.NEWLINE, ""},
		{token.DEDENT, ""},
		{token.EOF, ""},
	}
	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Fatalf("token[%d]: expected type %s, got %s (%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if exp.lit != "" && tok.Literal != exp.lit {
			t.Fatalf("token[%d]: expected literal %q, got %q", i, exp.lit, tok.Literal)
		}
	}
}

func TestRawText(t *testing.T) {
	input := "= +"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.RAW_TEXT {
		t.Fatalf("expected RAW_TEXT, got %s", tok.Type)
	}
	if tok.Literal != "=" {
		t.Fatalf("expected '=', got %q", tok.Literal)
	}
}

func TestLineNumbers(t *testing.T) {
	input := "a\nb\nc\n"
	l := New(input)

	tok := l.NextToken() // a
	if tok.Line != 1 {
		t.Fatalf("expected line 1, got %d", tok.Line)
	}
	l.NextToken() // NEWLINE

	tok = l.NextToken() // b
	if tok.Line != 2 {
		t.Fatalf("expected line 2, got %d", tok.Line)
	}
	l.NextToken() // NEWLINE

	tok = l.NextToken() // c
	if tok.Line != 3 {
		t.Fatalf("expected line 3, got %d", tok.Line)
	}
}

func TestSpawnKeyword(t *testing.T) {
	input := "spawn workflow Foo(x)"
	expected := []token.TokenType{
		token.SPAWN, token.WORKFLOW, token.IDENT, token.ARGS,
		token.NEWLINE, token.EOF,
	}
	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("token[%d]: expected %s, got %s (%q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestOptionsKeyword(t *testing.T) {
	input := "options(timeout: 30s)"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.OPTIONS {
		t.Fatalf("expected OPTIONS, got %s", tok.Type)
	}
	tok = l.NextToken()
	if tok.Type != token.ARGS {
		t.Fatalf("expected ARGS, got %s", tok.Type)
	}
	if tok.Literal != "timeout: 30s" {
		t.Fatalf("expected 'timeout: 30s', got %q", tok.Literal)
	}
}
