package parser

import (
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/token"
)

// parseSetStmt parses: SET IDENT NEWLINE
func parseSetStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume SET

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.SetStmt{
		Pos:       pos,
		Condition: ast.Ref[*ast.ConditionDecl]{Pos: pos, Name: name.Literal},
	}, nil
}

// parseUnsetStmt parses: UNSET IDENT NEWLINE
func parseUnsetStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume UNSET

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.UnsetStmt{
		Pos:       pos,
		Condition: ast.Ref[*ast.ConditionDecl]{Pos: pos, Name: name.Literal},
	}, nil
}

// parseReturnStmt parses: RETURN [ raw_expr ] NEWLINE
func parseReturnStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume RETURN

	var value string
	if p.current.Type != token.NEWLINE && p.current.Type != token.DEDENT && p.current.Type != token.EOF {
		value = p.collectRawUntil(token.NEWLINE)
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.ReturnStmt{
		Pos:   pos,
		Value: value,
	}, nil
}

// parseCloseStmt parses: CLOSE (COMPLETE | FAIL | CONTINUE_AS_NEW) [ARGS] NEWLINE
func parseCloseStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume CLOSE

	var reason ast.CloseReason
	switch p.current.Type {
	case token.COMPLETE:
		reason = ast.CloseComplete
		p.advance()
	case token.FAIL:
		reason = ast.CloseFailWorkflow
		p.advance()
	case token.CONTINUE_AS_NEW:
		reason = ast.CloseContinueAsNew
		p.advance()
	default:
		return nil, p.errorf("expected complete, fail, or continue_as_new after 'close', got %s", p.current.Type)
	}

	var args string
	if p.current.Type == token.ARGS {
		args = p.current.Literal
		p.advance()
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.CloseStmt{
		Pos:    pos,
		Reason: reason,
		Args:   args,
	}, nil
}

// parseBreakStmt parses: BREAK NEWLINE
func parseBreakStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume BREAK

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.BreakStmt{Pos: pos}, nil
}

// parseContinueStmt parses: CONTINUE NEWLINE
func parseContinueStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume CONTINUE

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.ContinueStmt{Pos: pos}, nil
}

// parseRawStmt captures the rest of the line as a raw statement.
func parseRawStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	text := p.collectRawUntil(token.NEWLINE)

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.RawStmt{
		Pos:  pos,
		Text: text,
	}, nil
}
