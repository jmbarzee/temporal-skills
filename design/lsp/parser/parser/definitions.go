package parser

import (
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/token"
)

// parseWorkflowDef parses:
// WORKFLOW IDENT ARGS [ ARROW ARGS ] COLON NEWLINE
// INDENT [ options_stmt ] { signal_def | query_def | update_def } workflow_body DEDENT
func parseWorkflowDef(p *Parser) (ast.Definition, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume WORKFLOW

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	params, err := p.expect(token.ARGS)
	if err != nil {
		return nil, err
	}

	var returnType string
	if p.current.Type == token.ARROW {
		p.advance() // consume ARROW
		rt, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		returnType = rt.Literal
	}

	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.NEWLINE); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.INDENT); err != nil {
		return nil, err
	}

	// Optional options statement at start of body.
	options, err := p.parseOptionalOptionsStmt()
	if err != nil {
		return nil, err
	}

	// Parse signal/query/update declarations (must come before body stmts).
	var signals []*ast.SignalDecl
	var queries []*ast.QueryDecl
	var updates []*ast.UpdateDecl

	for {
		// Skip blank lines between declarations.
		p.skipNewlines()

		switch p.current.Type {
		case token.SIGNAL:
			sig, err := parseSignalDecl(p)
			if err != nil {
				return nil, err
			}
			signals = append(signals, sig)
		case token.QUERY:
			q, err := parseQueryDecl(p)
			if err != nil {
				return nil, err
			}
			queries = append(queries, q)
		case token.UPDATE:
			u, err := parseUpdateDecl(p)
			if err != nil {
				return nil, err
			}
			updates = append(updates, u)
		default:
			goto parseBody
		}
	}

parseBody:
	// Parse workflow body.
	prevWorkflow := p.inWorkflow
	prevActivity := p.inActivity
	p.inWorkflow = true
	p.inActivity = false
	body, err := p.parseBody()
	p.inWorkflow = prevWorkflow
	p.inActivity = prevActivity
	if err != nil {
		return nil, err
	}

	return &ast.WorkflowDef{
		Pos:        pos,
		Name:       name.Literal,
		Params:     params.Literal,
		ReturnType: returnType,
		Options:    options,
		Signals:    signals,
		Queries:    queries,
		Updates:    updates,
		Body:       body,
	}, nil
}

// parseActivityDef parses:
// ACTIVITY IDENT ARGS [ ARROW ARGS ] COLON NEWLINE
// INDENT [ options_stmt ] activity_body DEDENT
func parseActivityDef(p *Parser) (ast.Definition, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume ACTIVITY

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	params, err := p.expect(token.ARGS)
	if err != nil {
		return nil, err
	}

	var returnType string
	if p.current.Type == token.ARROW {
		p.advance()
		rt, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		returnType = rt.Literal
	}

	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.NEWLINE); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.INDENT); err != nil {
		return nil, err
	}

	options, err := p.parseOptionalOptionsStmt()
	if err != nil {
		return nil, err
	}

	prevWorkflow := p.inWorkflow
	prevActivity := p.inActivity
	p.inWorkflow = false
	p.inActivity = true
	body, err := p.parseBody()
	p.inWorkflow = prevWorkflow
	p.inActivity = prevActivity
	if err != nil {
		return nil, err
	}

	return &ast.ActivityDef{
		Pos:        pos,
		Name:       name.Literal,
		Params:     params.Literal,
		ReturnType: returnType,
		Options:    options,
		Body:       body,
	}, nil
}

// parseSignalDecl parses: SIGNAL IDENT ARGS NEWLINE
func parseSignalDecl(p *Parser) (*ast.SignalDecl, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume SIGNAL

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	var params string
	if p.current.Type == token.ARGS {
		params = p.current.Literal
		p.advance()
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.SignalDecl{
		Pos:    pos,
		Name:   name.Literal,
		Params: params,
	}, nil
}

// parseQueryDecl parses: QUERY IDENT ARGS [ ARROW ARGS ] NEWLINE
func parseQueryDecl(p *Parser) (*ast.QueryDecl, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume QUERY

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	params, err := p.expect(token.ARGS)
	if err != nil {
		return nil, err
	}

	var returnType string
	if p.current.Type == token.ARROW {
		p.advance()
		rt, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		returnType = rt.Literal
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.QueryDecl{
		Pos:        pos,
		Name:       name.Literal,
		Params:     params.Literal,
		ReturnType: returnType,
	}, nil
}

// parseUpdateDecl parses: UPDATE IDENT ARGS [ ARROW ARGS ] NEWLINE
func parseUpdateDecl(p *Parser) (*ast.UpdateDecl, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume UPDATE

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	params, err := p.expect(token.ARGS)
	if err != nil {
		return nil, err
	}

	var returnType string
	if p.current.Type == token.ARROW {
		p.advance()
		rt, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		returnType = rt.Literal
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.UpdateDecl{
		Pos:        pos,
		Name:       name.Literal,
		Params:     params.Literal,
		ReturnType: returnType,
	}, nil
}
