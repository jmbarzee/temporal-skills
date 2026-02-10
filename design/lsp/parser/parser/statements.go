package parser

import (
	"strings"

	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/token"
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

	options, err := p.parseOptionalOptionsLine()
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

// parseWorkflowCall parses:
// [ SPAWN | DETACH ] [ NEXUS STRING ] WORKFLOW IDENT ARGS [ ARROW IDENT ] NEWLINE [ options_line ]
func parseWorkflowCall(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}

	mode := ast.CallChild
	switch p.current.Type {
	case token.SPAWN:
		mode = ast.CallSpawn
		p.advance()
	case token.DETACH:
		mode = ast.CallDetach
		p.advance()
	}

	var namespace string
	if p.current.Type == token.NEXUS {
		p.advance()
		ns, err := p.expect(token.STRING)
		if err != nil {
			return nil, err
		}
		namespace = ns.Literal
	}

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

	var result string
	if p.current.Type == token.ARROW {
		p.advance()
		res, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		result = res.Literal
	}

	// Validate: detach + arrow = error
	if mode == ast.CallDetach && result != "" {
		return nil, &ParseError{
			Msg:    "detach workflow call cannot have a result (-> identifier)",
			Line:   pos.Line,
			Column: pos.Column,
		}
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	options, err := p.parseOptionalOptionsLine()
	if err != nil {
		return nil, err
	}

	return &ast.WorkflowCall{
		Pos:       pos,
		Mode:      mode,
		Namespace: namespace,
		Name:      name.Literal,
		Args:      args.Literal,
		Result:    result,
		Options:   options,
	}, nil
}

// parseTimerStmt parses: TIMER duration_expr NEWLINE
func parseTimerStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume TIMER

	duration := p.collectRawUntil(token.NEWLINE, token.COLON)

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.TimerStmt{
		Pos:      pos,
		Duration: duration,
	}, nil
}

// parseAwaitBlock dispatches to parseAwaitAllBlock or parseAwaitOneBlock based on next token.
// Parses: AWAIT (ALL | ONE) COLON ...
func parseAwaitBlock(p *Parser) (ast.Statement, error) {
	p.advance() // consume AWAIT

	switch p.current.Type {
	case token.ALL:
		return parseAwaitAllBlock(p)
	case token.ONE:
		return parseAwaitOneBlock(p)
	default:
		return nil, p.errorf("expected 'all' or 'one' after 'await', got %s", p.current.Type)
	}
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

// parseAwaitOneCase parses a single await one case (timer or await all).
func parseAwaitOneCase(p *Parser) (*ast.AwaitOneCase, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	c := &ast.AwaitOneCase{Pos: pos}

	switch p.current.Type {
	case token.TIMER:
		// Timer case: TIMER duration COLON NEWLINE INDENT body DEDENT
		p.advance()
		c.TimerDuration = p.collectRawUntil(token.COLON, token.NEWLINE)

	case token.AWAIT:
		// Nested await all case: AWAIT ALL COLON NEWLINE INDENT body DEDENT
		stmt, err := parseAwaitBlock(p)
		if err != nil {
			return nil, err
		}
		awaitAll, ok := stmt.(*ast.AwaitAllBlock)
		if !ok {
			return nil, p.errorf("expected 'await all' in await one case, got await one")
		}
		c.AwaitAll = awaitAll
		return c, nil // await all case has no additional body

	default:
		return nil, p.errorf("unexpected token %s in await one case (expected 'timer' or 'await')", p.current.Type)
	}

	// Expect COLON NEWLINE INDENT body DEDENT
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
	c.Body = body

	return c, nil
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

// parseContinueAsNewStmt parses: CONTINUE_AS_NEW ARGS NEWLINE
func parseContinueAsNewStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume CONTINUE_AS_NEW

	args, err := p.expect(token.ARGS)
	if err != nil {
		return nil, err
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.ContinueAsNewStmt{
		Pos:  pos,
		Args: args.Literal,
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

// parseHintStmt parses: HINT (SIGNAL | UPDATE) IDENT NEWLINE
func parseHintStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume HINT

	var kind string
	switch p.current.Type {
	case token.SIGNAL:
		kind = "signal"
	case token.QUERY:
		kind = "query"
	case token.UPDATE:
		kind = "update"
	default:
		return nil, p.errorf("hint target must be signal, query, or update, got %s", p.current.Type)
	}
	p.advance()

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.HintStmt{
		Pos:  pos,
		Kind: kind,
		Name: name.Literal,
	}, nil
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
