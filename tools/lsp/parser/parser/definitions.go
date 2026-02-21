package parser

import (
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/token"
)

// parseWorkflowDef parses:
// WORKFLOW IDENT ARGS [ ARROW ARGS ] COLON NEWLINE
// INDENT { signal_def | query_def | update_def } workflow_body DEDENT
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

	if err := p.expectBlock(); err != nil {
		return nil, err
	}

	// Optional state block (must come before handlers and body).
	p.skipBlankLinesAndComments()
	var stateBlock *ast.StateBlock
	if p.current.Type == token.STATE {
		stateBlock, err = parseStateBlock(p)
		if err != nil {
			return nil, err
		}
	}

	// Parse signal/query/update declarations (must come before body stmts).
	var signals []*ast.SignalDecl
	var queries []*ast.QueryDecl
	var updates []*ast.UpdateDecl

	for {
		// Skip blank lines and comments between declarations.
		p.skipBlankLinesAndComments()

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
	body, err := p.parseBodyAs(bodyWorkflow)
	if err != nil {
		return nil, err
	}

	return &ast.WorkflowDef{
		Pos:        pos,
		Name:       name.Literal,
		Params:     params.Literal,
		ReturnType: returnType,
		State:      stateBlock,
		Signals:    signals,
		Queries:    queries,
		Updates:    updates,
		Body:       body,
	}, nil
}

// parseActivityDef parses:
// ACTIVITY IDENT ARGS [ ARROW ARGS ] COLON NEWLINE
// INDENT activity_body DEDENT
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

	if err := p.expectBlock(); err != nil {
		return nil, err
	}

	body, err := p.parseBodyAs(bodyActivity)
	if err != nil {
		return nil, err
	}

	return &ast.ActivityDef{
		Pos:        pos,
		Name:       name.Literal,
		Params:     params.Literal,
		ReturnType: returnType,
		Body:       body,
	}, nil
}

// parseSignalDecl parses: SIGNAL IDENT [ ARGS ] COLON NEWLINE INDENT body DEDENT
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

	if err := p.expectBlock(); err != nil {
		return nil, err
	}

	body, err := p.parseBodyAs(bodyWorkflow)
	if err != nil {
		return nil, err
	}

	return &ast.SignalDecl{
		Pos:    pos,
		Name:   name.Literal,
		Params: params,
		Body:   body,
	}, nil
}

// parseQueryDecl parses: QUERY IDENT ARGS [ ARROW ARGS ] COLON NEWLINE INDENT body DEDENT
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

	if err := p.expectBlock(); err != nil {
		return nil, err
	}

	// Query bodies are restricted like activity bodies (no temporal primitives).
	body, err := p.parseBodyAs(bodyActivity)
	if err != nil {
		return nil, err
	}

	return &ast.QueryDecl{
		Pos:        pos,
		Name:       name.Literal,
		Params:     params.Literal,
		ReturnType: returnType,
		Body:       body,
	}, nil
}

// parseUpdateDecl parses: UPDATE IDENT ARGS [ ARROW ARGS ] COLON NEWLINE INDENT body DEDENT
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

	if err := p.expectBlock(); err != nil {
		return nil, err
	}

	body, err := p.parseBodyAs(bodyWorkflow)
	if err != nil {
		return nil, err
	}

	return &ast.UpdateDecl{
		Pos:        pos,
		Name:       name.Literal,
		Params:     params.Literal,
		ReturnType: returnType,
		Body:       body,
	}, nil
}

// parseStateBlock parses: STATE COLON NEWLINE INDENT (condition_decl | raw_stmt)* DEDENT
func parseStateBlock(p *Parser) (*ast.StateBlock, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume STATE

	if err := p.expectBlock(); err != nil {
		return nil, err
	}

	block := &ast.StateBlock{Pos: pos}

	for p.current.Type != token.DEDENT && p.current.Type != token.EOF {
		if p.current.Type == token.NEWLINE {
			p.advance()
			continue
		}
		if p.current.Type == token.COMMENT {
			p.advance()
			if p.current.Type == token.NEWLINE {
				p.advance()
			}
			continue
		}

		if p.current.Type == token.CONDITION {
			cond, err := parseConditionDecl(p)
			if err != nil {
				return nil, err
			}
			block.Conditions = append(block.Conditions, cond)
		} else {
			// Raw statement (variable initialization, etc.)
			stmt, err := parseRawStmt(p)
			if err != nil {
				return nil, err
			}
			if raw, ok := stmt.(*ast.RawStmt); ok {
				block.RawStmts = append(block.RawStmts, raw)
			}
		}
	}

	if p.current.Type == token.DEDENT {
		p.advance()
	}

	return block, nil
}

// parseConditionDecl parses: CONDITION IDENT NEWLINE
func parseConditionDecl(p *Parser) (*ast.ConditionDecl, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume CONDITION

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.ConditionDecl{
		Pos:  pos,
		Name: name.Literal,
	}, nil
}
