package lexer

import (
	"testing"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/token"
)

func TestKeywords(t *testing.T) {
	input := "workflow activity signal query update detach nexus promise condition set unset state timer options await all one switch case if else for in close complete fail return continue_as_new break continue"
	expected := []token.TokenType{
		token.WORKFLOW, token.ACTIVITY, token.SIGNAL, token.QUERY, token.UPDATE,
		token.DETACH, token.NEXUS, token.PROMISE, token.CONDITION, token.SET, token.UNSET, token.STATE,
		token.TIMER, token.OPTIONS,
		token.AWAIT, token.ALL, token.ONE, token.SWITCH,
		token.CASE, token.IF, token.ELSE, token.FOR, token.IN,
		token.CLOSE, token.COMPLETE, token.FAIL,
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

func TestLeftArrow(t *testing.T) {
	input := "promise p <- activity Foo(x)"
	expected := []token.TokenType{
		token.PROMISE, token.IDENT, token.LEFT_ARROW, token.ACTIVITY, token.IDENT, token.ARGS,
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

func TestNewKeywords(t *testing.T) {
	input := "promise condition set unset state"
	expected := []token.TokenType{
		token.PROMISE, token.CONDITION, token.SET, token.UNSET, token.STATE,
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
	input := "options:"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.OPTIONS {
		t.Fatalf("expected OPTIONS, got %s", tok.Type)
	}
	tok = l.NextToken()
	if tok.Type != token.COLON {
		t.Fatalf("expected COLON, got %s", tok.Type)
	}
}

func TestDurationToken(t *testing.T) {
	tests := []struct {
		input string
		lit   string
	}{
		{"60s", "60s"},
		{"5m", "5m"},
		{"1h", "1h"},
		{"500ms", "500ms"},
		{"7d", "7d"},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != token.DURATION {
			t.Errorf("input %q: expected DURATION, got %s", tt.input, tok.Type)
		}
		if tok.Literal != tt.lit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.lit, tok.Literal)
		}
	}
}

func TestNumberToken(t *testing.T) {
	tests := []struct {
		input string
		lit   string
	}{
		{"3", "3"},
		{"2.0", "2.0"},
		{"100", "100"},
		{"1.5", "1.5"},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != token.NUMBER {
			t.Errorf("input %q: expected NUMBER, got %s", tt.input, tok.Type)
		}
		if tok.Literal != tt.lit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.lit, tok.Literal)
		}
	}
}

func TestEmitEOFIdempotent(t *testing.T) {
	input := "a:\n    b"
	l := New(input)
	_ = l.AllTokens()

	// Calling NextToken after AllTokens should return EOF, not panic or garbage.
	tok := l.NextToken()
	if tok.Type != token.EOF {
		t.Fatalf("expected EOF after AllTokens, got %s (%q)", tok.Type, tok.Literal)
	}
	// And again.
	tok = l.NextToken()
	if tok.Type != token.EOF {
		t.Fatalf("expected EOF on third call, got %s (%q)", tok.Type, tok.Literal)
	}
}

func TestInconsistentDedent(t *testing.T) {
	// Indent stack will be [0, 4] after the indent. Dedenting to column 3
	// doesn't match any stack level, so an ILLEGAL token should appear.
	input := "a:\n    b\n   c\n"
	l := New(input)
	tokens := l.AllTokens()

	foundIllegal := false
	for _, tok := range tokens {
		if tok.Type == token.ILLEGAL {
			foundIllegal = true
			if tok.Literal != "inconsistent indentation" {
				t.Fatalf("expected ILLEGAL literal 'inconsistent indentation', got %q", tok.Literal)
			}
			break
		}
	}
	if !foundIllegal {
		t.Fatalf("expected ILLEGAL token for inconsistent dedent, got tokens: %v", tokens)
	}
}

func TestOptionsBlockTokenStream(t *testing.T) {
	input := "activity Foo(x) -> y\n    options:\n        task_queue: \"workers\"\n        start_to_close: 60s\n"
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.ACTIVITY, "activity"},
		{token.IDENT, "Foo"},
		{token.ARGS, "x"},
		{token.ARROW, "->"},
		{token.IDENT, "y"},
		{token.NEWLINE, ""},
		{token.INDENT, ""},
		{token.OPTIONS, "options"},
		{token.COLON, ":"},
		{token.NEWLINE, ""},
		{token.INDENT, ""},
		{token.TASK_QUEUE, "task_queue"},
		{token.COLON, ":"},
		{token.STRING, "workers"},
		{token.NEWLINE, ""},
		{token.IDENT, "start_to_close"},
		{token.COLON, ":"},
		{token.DURATION, "60s"},
		{token.NEWLINE, ""},
		{token.DEDENT, ""},
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
