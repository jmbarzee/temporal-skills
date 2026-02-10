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
		// await timer(duration)
		p.advance()
		duration, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		stmt.Timer = duration.Literal

	case token.SIGNAL:
		// await signal Name [-> params]
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		stmt.Signal = name.Literal

		if p.current.Type == token.ARROW {
			p.advance()
			params, err := parseParamBinding(p)
			if err != nil {
				return nil, err
			}
			stmt.SignalParams = params
		}

	case token.UPDATE:
		// await update Name [-> params]
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		stmt.Update = name.Literal

		if p.current.Type == token.ARROW {
			p.advance()
			params, err := parseParamBinding(p)
			if err != nil {
				return nil, err
			}
			stmt.UpdateParams = params
		}

	case token.ACTIVITY:
		// await activity Name(args) [-> result]
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		stmt.Activity = name.Literal

		args, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		stmt.ActivityArgs = args.Literal

		if p.current.Type == token.ARROW {
			p.advance()
			result, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			stmt.ActivityResult = result.Literal
		}

	case token.SPAWN, token.DETACH, token.WORKFLOW:
		// await [spawn|detach] workflow Name(args) [-> result]
		mode := ast.CallChild
		if p.current.Type == token.SPAWN {
			mode = ast.CallSpawn
			p.advance()
		} else if p.current.Type == token.DETACH {
			mode = ast.CallDetach
			p.advance()
		}
		stmt.WorkflowMode = mode

		// Optional nexus namespace
		if p.current.Type == token.NEXUS {
			p.advance()
			ns, err := p.expect(token.STRING)
			if err != nil {
				return nil, err
			}
			stmt.WorkflowNamespace = ns.Literal
		}

		if _, err := p.expect(token.WORKFLOW); err != nil {
			return nil, err
		}

		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		stmt.Workflow = name.Literal

		args, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		stmt.WorkflowArgs = args.Literal

		if p.current.Type == token.ARROW {
			p.advance()
			result, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			stmt.WorkflowResult = result.Literal
		}

		// Validate: detach + arrow = error
		if mode == ast.CallDetach && stmt.WorkflowResult != "" {
			return nil, &ParseError{
				Msg:    "detach workflow cannot have a result (-> identifier)",
				Line:   pos.Line,
				Column: pos.Column,
			}
		}

	default:
		return nil, p.errorf("expected timer, signal, update, activity, or workflow after 'await', got %s", p.current.Type)
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
		// signal Name [-> params]: [body]
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		c.Signal = name.Literal

		if p.current.Type == token.ARROW {
			p.advance()
			params, err := parseParamBinding(p)
			if err != nil {
				return nil, err
			}
			c.SignalParams = params
		}

		if _, err := p.expect(token.COLON); err != nil {
			return nil, err
		}

		// Parse optional body
		body, err := parseOptionalCaseBody(p)
		if err != nil {
			return nil, err
		}
		c.Body = body
		return c, nil

	case token.UPDATE:
		// update Name [-> params]: [body]
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		c.Update = name.Literal

		if p.current.Type == token.ARROW {
			p.advance()
			params, err := parseParamBinding(p)
			if err != nil {
				return nil, err
			}
			c.UpdateParams = params
		}

		if _, err := p.expect(token.COLON); err != nil {
			return nil, err
		}

		// Parse optional body
		body, err := parseOptionalCaseBody(p)
		if err != nil {
			return nil, err
		}
		c.Body = body
		return c, nil

	case token.TIMER:
		// timer(duration): [body]
		p.advance()
		duration, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		c.Timer = duration.Literal

		if _, err := p.expect(token.COLON); err != nil {
			return nil, err
		}

		// Parse optional body
		body, err := parseOptionalCaseBody(p)
		if err != nil {
			return nil, err
		}
		c.Body = body
		return c, nil

	case token.ACTIVITY:
		// activity Name(args) [-> result]: [body]
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		c.Activity = name.Literal

		args, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		c.ActivityArgs = args.Literal

		if p.current.Type == token.ARROW {
			p.advance()
			result, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			c.ActivityResult = result.Literal
		}

		if _, err := p.expect(token.COLON); err != nil {
			return nil, err
		}

		// Parse optional body
		body, err := parseOptionalCaseBody(p)
		if err != nil {
			return nil, err
		}
		c.Body = body
		return c, nil

	case token.SPAWN, token.DETACH, token.WORKFLOW:
		// [spawn|detach] workflow Name(args) [-> result]: [body]
		mode := ast.CallChild
		if p.current.Type == token.SPAWN {
			mode = ast.CallSpawn
			p.advance()
		} else if p.current.Type == token.DETACH {
			mode = ast.CallDetach
			p.advance()
		}
		c.WorkflowMode = mode

		// Optional nexus namespace
		if p.current.Type == token.NEXUS {
			p.advance()
			ns, err := p.expect(token.STRING)
			if err != nil {
				return nil, err
			}
			c.WorkflowNamespace = ns.Literal
		}

		if _, err := p.expect(token.WORKFLOW); err != nil {
			return nil, err
		}

		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		c.Workflow = name.Literal

		args, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		c.WorkflowArgs = args.Literal

		if p.current.Type == token.ARROW {
			p.advance()
			result, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			c.WorkflowResult = result.Literal
		}

		// Validate: detach + arrow = error
		if mode == ast.CallDetach && c.WorkflowResult != "" {
			return nil, &ParseError{
				Msg:    "detach workflow cannot have a result (-> identifier)",
				Line:   pos.Line,
				Column: pos.Column,
			}
		}

		if _, err := p.expect(token.COLON); err != nil {
			return nil, err
		}

		// Parse optional body
		body, err := parseOptionalCaseBody(p)
		if err != nil {
			return nil, err
		}
		c.Body = body
		return c, nil

	case token.AWAIT:
		// Nested await all case: await all: ...
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
		return nil, p.errorf("unexpected token %s in await one case (expected signal, update, timer, activity, workflow, or await)", p.current.Type)
	}
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

// isCaseKeyword checks if a token type is a valid await one case keyword.
func isCaseKeyword(t token.TokenType) bool {
	return t == token.SIGNAL || t == token.UPDATE || t == token.TIMER ||
		t == token.ACTIVITY || t == token.WORKFLOW || t == token.SPAWN ||
		t == token.DETACH || t == token.AWAIT
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

// parseCloseStmt parses: CLOSE [ COMPLETED | FAILED ] [ raw_expr ] NEWLINE
func parseCloseStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume CLOSE

	var reason string
	if p.current.Type == token.COMPLETED {
		reason = "completed"
		p.advance()
	} else if p.current.Type == token.FAILED {
		reason = "failed"
		p.advance()
	}

	var value string
	if p.current.Type != token.NEWLINE && p.current.Type != token.DEDENT && p.current.Type != token.EOF {
		value = p.collectRawUntil(token.NEWLINE)
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.CloseStmt{
		Pos:    pos,
		Reason: reason,
		Value:  value,
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
