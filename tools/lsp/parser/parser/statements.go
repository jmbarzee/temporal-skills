package parser

import (
	"strings"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/token"
)

// parseActivityCall parses: ACTIVITY IDENT ARGS [ ARROW IDENT ] NEWLINE [ options_line ]
func parseActivityCall(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume ACTIVITY

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	args, err := p.expect(token.ARGS)
	if err != nil {
		return nil, err
	}

	var result string
	if p.current.Type == token.ARROW {
		p.advance()
		res, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		result = res.Literal
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	options, err := p.parseOptionalOptionsLine(OptionsContextActivity)
	if err != nil {
		return nil, err
	}

	return &ast.ActivityCall{
		Pos:     pos,
		Name:    name.Literal,
		Args:    args.Literal,
		Result:  result,
		Options: options,
	}, nil
}

// parseWorkflowCall parses: WORKFLOW IDENT ARGS [ ARROW IDENT ] NEWLINE [ options_line ]
func parseWorkflowCall(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume WORKFLOW

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	args, err := p.expect(token.ARGS)
	if err != nil {
		return nil, err
	}

	var result string
	if p.current.Type == token.ARROW {
		p.advance()
		res, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		result = res.Literal
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	options, err := p.parseOptionalOptionsLine(OptionsContextWorkflow)
	if err != nil {
		return nil, err
	}

	return &ast.WorkflowCall{
		Pos:     pos,
		Mode:    ast.CallChild,
		Name:    name.Literal,
		Args:    args.Literal,
		Result:  result,
		Options: options,
	}, nil
}

// parseAwaitStmt handles both single await and await blocks (await all/one).
// Single await: await timer(5m), await signal Name -> params, etc.
// Block await: await all: ..., await one: ...
func parseAwaitStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume AWAIT

	// Check if this is a block form (await all/one) or single await
	switch p.current.Type {
	case token.ALL:
		return parseAwaitAllBlock(p)
	case token.ONE:
		return parseAwaitOneBlock(p)
	default:
		// Single await form
		return parseSingleAwait(p, pos)
	}
}

// parseSingleAwait parses a single await target:
// timer(duration), signal Name [-> params], update Name [-> params],
// activity Name(args) [-> result], workflow Name(args) [-> result]
func parseSingleAwait(p *Parser, pos ast.Pos) (*ast.AwaitStmt, error) {
	stmt := &ast.AwaitStmt{Pos: pos}

	switch p.current.Type {
	case token.TIMER:
		p.advance()
		duration, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		stmt.Target = &ast.TimerTarget{Duration: duration.Literal}

	case token.SIGNAL:
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		t := &ast.SignalTarget{Name: name.Literal}
		if p.current.Type == token.ARROW {
			p.advance()
			params, err := parseParamBinding(p)
			if err != nil {
				return nil, err
			}
			t.Params = params
		}
		stmt.Target = t

	case token.UPDATE:
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		t := &ast.UpdateTarget{Name: name.Literal}
		if p.current.Type == token.ARROW {
			p.advance()
			params, err := parseParamBinding(p)
			if err != nil {
				return nil, err
			}
			t.Params = params
		}
		stmt.Target = t

	case token.ACTIVITY:
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		t := &ast.ActivityTarget{Name: name.Literal}
		args, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		t.Args = args.Literal
		if p.current.Type == token.ARROW {
			p.advance()
			result, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			t.Result = result.Literal
		}
		stmt.Target = t

	case token.DETACH, token.WORKFLOW:
		mode := ast.CallChild
		if p.current.Type == token.DETACH {
			mode = ast.CallDetach
			p.advance()
		}

		if p.current.Type == token.NEXUS {
			p.advance()
			endpoint, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			service, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(token.DOT); err != nil {
				return nil, err
			}
			operation, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			args, err := p.expect(token.ARGS)
			if err != nil {
				return nil, err
			}
			t := &ast.NexusTarget{
				Endpoint:  endpoint.Literal,
				Service:   service.Literal,
				Operation: operation.Literal,
				Args:      args.Literal,
				Detach:    mode == ast.CallDetach,
			}
			if p.current.Type == token.ARROW {
				p.advance()
				result, err := p.expect(token.IDENT)
				if err != nil {
					return nil, err
				}
				t.Result = result.Literal
			}
			if t.Detach && t.Result != "" {
				return nil, &ParseError{
					Msg:    "detach nexus call cannot have a result (-> identifier)",
					Line:   pos.Line,
					Column: pos.Column,
				}
			}
			stmt.Target = t
		} else {
			if _, err := p.expect(token.WORKFLOW); err != nil {
				return nil, err
			}
			name, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			args, err := p.expect(token.ARGS)
			if err != nil {
				return nil, err
			}
			t := &ast.WorkflowTarget{
				Name: name.Literal,
				Mode: mode,
				Args: args.Literal,
			}
			if p.current.Type == token.ARROW {
				p.advance()
				result, err := p.expect(token.IDENT)
				if err != nil {
					return nil, err
				}
				t.Result = result.Literal
			}
			if mode == ast.CallDetach && t.Result != "" {
				return nil, &ParseError{
					Msg:    "detach workflow cannot have a result (-> identifier)",
					Line:   pos.Line,
					Column: pos.Column,
				}
			}
			stmt.Target = t
		}

	case token.NEXUS:
		p.advance()
		endpoint, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		service, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(token.DOT); err != nil {
			return nil, err
		}
		operation, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		args, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		t := &ast.NexusTarget{
			Endpoint:  endpoint.Literal,
			Service:   service.Literal,
			Operation: operation.Literal,
			Args:      args.Literal,
		}
		if p.current.Type == token.ARROW {
			p.advance()
			result, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			t.Result = result.Literal
		}
		stmt.Target = t

	case token.IDENT:
		t := &ast.IdentTarget{Name: p.current.Literal}
		p.advance()
		if p.current.Type == token.ARROW {
			p.advance()
			result, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			t.Result = result.Literal
		}
		stmt.Target = t

	default:
		return nil, p.errorf("expected timer, signal, update, activity, workflow, nexus, or identifier after 'await', got %s", p.current.Type)
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return stmt, nil
}

// parseParamBinding parses parameter binding after ARROW:
// either IDENT (single param) or ARGS (multiple params in parens)
func parseParamBinding(p *Parser) (string, error) {
	if p.current.Type == token.IDENT {
		result := p.current.Literal
		p.advance()
		return result, nil
	} else if p.current.Type == token.ARGS {
		result := p.current.Literal
		p.advance()
		return result, nil
	}
	return "", p.errorf("expected identifier or ( after ->, got %s", p.current.Type)
}

// parseAwaitAllBlock parses: ALL COLON NEWLINE INDENT workflow_body DEDENT
func parseAwaitAllBlock(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume ALL

	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.NEWLINE); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.INDENT); err != nil {
		return nil, err
	}

	body, err := p.parseBody()
	if err != nil {
		return nil, err
	}

	return &ast.AwaitAllBlock{
		Pos:  pos,
		Body: body,
	}, nil
}

// parseAwaitOneBlock parses: ONE COLON NEWLINE INDENT { await_one_case } DEDENT
func parseAwaitOneBlock(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume ONE

	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.NEWLINE); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.INDENT); err != nil {
		return nil, err
	}

	var cases []*ast.AwaitOneCase
	for p.current.Type != token.DEDENT && p.current.Type != token.EOF {
		if p.current.Type == token.NEWLINE {
			p.advance()
			continue
		}

		c, err := parseAwaitOneCase(p)
		if err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}

	if p.current.Type == token.DEDENT {
		p.advance()
	}

	return &ast.AwaitOneBlock{
		Pos:   pos,
		Cases: cases,
	}, nil
}

// parseAwaitOneCase parses a single await one case.
// Supports: signal Name [-> params]:, update Name [-> params]:,
// timer(duration):, activity Name(args) [-> result]:,
// workflow Name(args) [-> result]:, or await all:
// Case bodies are optional (can be empty after colon).
func parseAwaitOneCase(p *Parser) (*ast.AwaitOneCase, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	c := &ast.AwaitOneCase{Pos: pos}

	switch p.current.Type {
	case token.SIGNAL:
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		t := &ast.SignalTarget{Name: name.Literal}
		if p.current.Type == token.ARROW {
			p.advance()
			params, err := parseParamBinding(p)
			if err != nil {
				return nil, err
			}
			t.Params = params
		}
		c.Target = t

	case token.UPDATE:
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		t := &ast.UpdateTarget{Name: name.Literal}
		if p.current.Type == token.ARROW {
			p.advance()
			params, err := parseParamBinding(p)
			if err != nil {
				return nil, err
			}
			t.Params = params
		}
		c.Target = t

	case token.TIMER:
		p.advance()
		duration, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		c.Target = &ast.TimerTarget{Duration: duration.Literal}

	case token.ACTIVITY:
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		t := &ast.ActivityTarget{Name: name.Literal}
		args, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		t.Args = args.Literal
		if p.current.Type == token.ARROW {
			p.advance()
			result, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			t.Result = result.Literal
		}
		c.Target = t

	case token.DETACH, token.WORKFLOW:
		mode := ast.CallChild
		if p.current.Type == token.DETACH {
			mode = ast.CallDetach
			p.advance()
		}

		if p.current.Type == token.NEXUS {
			p.advance()
			endpoint, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			service, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(token.DOT); err != nil {
				return nil, err
			}
			operation, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			args, err := p.expect(token.ARGS)
			if err != nil {
				return nil, err
			}
			t := &ast.NexusTarget{
				Endpoint:  endpoint.Literal,
				Service:   service.Literal,
				Operation: operation.Literal,
				Args:      args.Literal,
				Detach:    mode == ast.CallDetach,
			}
			if p.current.Type == token.ARROW {
				p.advance()
				result, err := p.expect(token.IDENT)
				if err != nil {
					return nil, err
				}
				t.Result = result.Literal
			}
			if t.Detach && t.Result != "" {
				return nil, &ParseError{
					Msg:    "detach nexus call cannot have a result (-> identifier)",
					Line:   pos.Line,
					Column: pos.Column,
				}
			}
			c.Target = t
		} else {
			if _, err := p.expect(token.WORKFLOW); err != nil {
				return nil, err
			}
			name, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			args, err := p.expect(token.ARGS)
			if err != nil {
				return nil, err
			}
			t := &ast.WorkflowTarget{
				Name: name.Literal,
				Mode: mode,
				Args: args.Literal,
			}
			if p.current.Type == token.ARROW {
				p.advance()
				result, err := p.expect(token.IDENT)
				if err != nil {
					return nil, err
				}
				t.Result = result.Literal
			}
			if mode == ast.CallDetach && t.Result != "" {
				return nil, &ParseError{
					Msg:    "detach workflow cannot have a result (-> identifier)",
					Line:   pos.Line,
					Column: pos.Column,
				}
			}
			c.Target = t
		}

	case token.NEXUS:
		p.advance()
		endpoint, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		service, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(token.DOT); err != nil {
			return nil, err
		}
		operation, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		args, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		t := &ast.NexusTarget{
			Endpoint:  endpoint.Literal,
			Service:   service.Literal,
			Operation: operation.Literal,
			Args:      args.Literal,
		}
		if p.current.Type == token.ARROW {
			p.advance()
			result, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			t.Result = result.Literal
		}
		c.Target = t

	case token.IDENT:
		t := &ast.IdentTarget{Name: p.current.Literal}
		p.advance()
		if p.current.Type == token.ARROW {
			p.advance()
			result, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			t.Result = result.Literal
		}
		c.Target = t

	case token.AWAIT:
		stmt, err := parseAwaitStmt(p)
		if err != nil {
			return nil, err
		}
		awaitAll, ok := stmt.(*ast.AwaitAllBlock)
		if !ok {
			return nil, p.errorf("expected 'await all' in await one case, got %s", stmt)
		}
		c.AwaitAll = awaitAll
		return c, nil

	default:
		return nil, p.errorf("unexpected token %s in await one case (expected signal, update, timer, activity, workflow, nexus, identifier, or await)", p.current.Type)
	}

	// All target cases (not await all) need colon + optional body
	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	body, err := parseOptionalCaseBody(p)
	if err != nil {
		return nil, err
	}
	c.Body = body
	return c, nil
}

// parseOptionalCaseBody parses an optional case body after colon.
// If the next token after NEWLINE is DEDENT or another case keyword, the body is empty.
func parseOptionalCaseBody(p *Parser) ([]ast.Statement, error) {
	if _, err := p.expect(token.NEWLINE); err != nil {
		return nil, err
	}

	// Check if body is empty (next token is DEDENT or case keyword)
	if p.current.Type == token.DEDENT || isCaseKeyword(p.current.Type) {
		return nil, nil
	}

	if _, err := p.expect(token.INDENT); err != nil {
		return nil, err
	}

	body, err := p.parseBody()
	if err != nil {
		return nil, err
	}

	return body, nil
}

// isCaseKeyword checks if a token type is a valid await one case keyword or identifier.
func isCaseKeyword(t token.TokenType) bool {
	return t == token.SIGNAL || t == token.UPDATE || t == token.TIMER ||
		t == token.ACTIVITY || t == token.WORKFLOW ||
		t == token.DETACH || t == token.NEXUS || t == token.AWAIT || t == token.IDENT
}

// parsePromiseStmt parses: PROMISE IDENT LEFT_ARROW async_target NEWLINE
func parsePromiseStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume PROMISE

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(token.LEFT_ARROW); err != nil {
		return nil, err
	}

	stmt := &ast.PromiseStmt{
		Pos:  pos,
		Name: name.Literal,
	}

	// Parse async target
	switch p.current.Type {
	case token.TIMER:
		p.advance()
		duration, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		stmt.Target = &ast.TimerTarget{Duration: duration.Literal}

	case token.SIGNAL:
		p.advance()
		sigName, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		stmt.Target = &ast.SignalTarget{Name: sigName.Literal}

	case token.UPDATE:
		p.advance()
		updName, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		stmt.Target = &ast.UpdateTarget{Name: updName.Literal}

	case token.ACTIVITY:
		p.advance()
		actName, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		args, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		stmt.Target = &ast.ActivityTarget{Name: actName.Literal, Args: args.Literal}

	case token.WORKFLOW:
		p.advance()
		wfName, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		args, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		stmt.Target = &ast.WorkflowTarget{Name: wfName.Literal, Args: args.Literal}

	case token.NEXUS:
		p.advance()
		endpoint, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		service, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(token.DOT); err != nil {
			return nil, err
		}
		operation, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		args, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		stmt.Target = &ast.NexusTarget{
			Endpoint:  endpoint.Literal,
			Service:   service.Literal,
			Operation: operation.Literal,
			Args:      args.Literal,
		}

	default:
		return nil, p.errorf("expected timer, signal, update, activity, workflow, or nexus after '<-', got %s", p.current.Type)
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return stmt, nil
}

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
		Pos:  pos,
		Name: name.Literal,
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
		Pos:  pos,
		Name: name.Literal,
	}, nil
}

// parseSwitchBlock parses: SWITCH ARGS COLON NEWLINE INDENT { switch_case } [ else ] DEDENT
func parseSwitchBlock(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume SWITCH

	expr, err := p.expect(token.ARGS)
	if err != nil {
		return nil, err
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

	var cases []*ast.SwitchCase
	var defaultBody []ast.Statement

	for p.current.Type != token.DEDENT && p.current.Type != token.EOF {
		if p.current.Type == token.NEWLINE {
			p.advance()
			continue
		}

		if p.current.Type == token.ELSE {
			p.advance()
			if _, err := p.expect(token.COLON); err != nil {
				return nil, err
			}
			if _, err := p.expect(token.NEWLINE); err != nil {
				return nil, err
			}
			if _, err := p.expect(token.INDENT); err != nil {
				return nil, err
			}
			defaultBody, err = p.parseBody()
			if err != nil {
				return nil, err
			}
			continue
		}

		if p.current.Type != token.CASE {
			return nil, p.errorf("expected case or else in switch, got %s", p.current.Type)
		}

		casePos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
		p.advance() // consume CASE

		// Collect the case value expression until COLON.
		value := p.collectRawUntil(token.COLON)

		if _, err := p.expect(token.COLON); err != nil {
			return nil, err
		}
		if _, err := p.expect(token.NEWLINE); err != nil {
			return nil, err
		}
		if _, err := p.expect(token.INDENT); err != nil {
			return nil, err
		}

		body, err := p.parseBody()
		if err != nil {
			return nil, err
		}

		cases = append(cases, &ast.SwitchCase{
			Pos:   casePos,
			Value: value,
			Body:  body,
		})
	}

	if len(cases) == 0 {
		return nil, &ParseError{
			Msg:    "switch must have at least one case",
			Line:   pos.Line,
			Column: pos.Column,
		}
	}

	if p.current.Type == token.DEDENT {
		p.advance()
	}

	return &ast.SwitchBlock{
		Pos:     pos,
		Expr:    expr.Literal,
		Cases:   cases,
		Default: defaultBody,
	}, nil
}

// parseIfStmt parses: IF ARGS COLON NEWLINE INDENT body DEDENT [ ELSE COLON NEWLINE INDENT body DEDENT ]
func parseIfStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume IF

	cond, err := p.expect(token.ARGS)
	if err != nil {
		return nil, err
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

	body, err := p.parseBody()
	if err != nil {
		return nil, err
	}

	var elseBody []ast.Statement
	if p.current.Type == token.ELSE {
		p.advance()
		if _, err := p.expect(token.COLON); err != nil {
			return nil, err
		}
		if _, err := p.expect(token.NEWLINE); err != nil {
			return nil, err
		}
		if _, err := p.expect(token.INDENT); err != nil {
			return nil, err
		}
		elseBody, err = p.parseBody()
		if err != nil {
			return nil, err
		}
	}

	return &ast.IfStmt{
		Pos:       pos,
		Condition: cond.Literal,
		Body:      body,
		ElseBody:  elseBody,
	}, nil
}

// parseForStmt parses: FOR [ ARGS ] COLON NEWLINE INDENT body DEDENT
func parseForStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume FOR

	stmt := &ast.ForStmt{Pos: pos}

	if p.current.Type == token.COLON {
		// Infinite loop: for:
		stmt.Variant = ast.ForInfinite
	} else if p.current.Type == token.ARGS {
		content := p.current.Literal
		p.advance()

		// Check for "in" keyword using strings.Fields to find standalone word.
		fields := strings.Fields(content)
		inIdx := -1
		for i, f := range fields {
			if f == "in" {
				inIdx = i
				break
			}
		}

		if inIdx > 0 {
			// Iteration: for (var in collection):
			stmt.Variant = ast.ForIteration
			stmt.Variable = strings.Join(fields[:inIdx], " ")
			stmt.Iterable = strings.Join(fields[inIdx+1:], " ")
		} else {
			// Conditional: for (condition):
			stmt.Variant = ast.ForConditional
			stmt.Condition = content
		}
	} else {
		return nil, p.errorf("expected ( or : after for, got %s", p.current.Type)
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

	body, err := p.parseBody()
	if err != nil {
		return nil, err
	}
	stmt.Body = body

	return stmt, nil
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

	var reason string
	switch p.current.Type {
	case token.COMPLETE:
		reason = "complete"
		p.advance()
	case token.FAIL:
		reason = "fail"
		p.advance()
	case token.CONTINUE_AS_NEW:
		reason = "continue_as_new"
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
