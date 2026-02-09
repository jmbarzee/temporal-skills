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

// parseAwaitStmt parses: AWAIT await_target { OR await_target } NEWLINE
func parseAwaitStmt(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume AWAIT

	var targets []*ast.AwaitTarget

	target, err := parseAwaitTarget(p)
	if err != nil {
		return nil, err
	}
	targets = append(targets, target)

	for p.current.Type == token.OR {
		p.advance() // consume OR
		target, err := parseAwaitTarget(p)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.AwaitStmt{
		Pos:     pos,
		Targets: targets,
	}, nil
}

// parseAwaitTarget parses: SIGNAL IDENT [ ARGS ] | UPDATE IDENT [ ARGS ]
func parseAwaitTarget(p *Parser) (*ast.AwaitTarget, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}

	var kind string
	switch p.current.Type {
	case token.SIGNAL:
		kind = "signal"
	case token.UPDATE:
		kind = "update"
	default:
		return nil, p.errorf("await target must be signal or update, got %s", p.current.Type)
	}
	p.advance()

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	var args string
	if p.current.Type == token.ARGS {
		args = p.current.Literal
		p.advance()
	}

	return &ast.AwaitTarget{
		Pos:  pos,
		Kind: kind,
		Name: name.Literal,
		Args: args,
	}, nil
}

// parseParallelBlock parses: PARALLEL COLON NEWLINE INDENT workflow_body DEDENT
func parseParallelBlock(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume PARALLEL

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

	return &ast.ParallelBlock{
		Pos:  pos,
		Body: body,
	}, nil
}

// parseSelectBlock parses: SELECT COLON NEWLINE INDENT { select_case } DEDENT
func parseSelectBlock(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume SELECT

	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.NEWLINE); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.INDENT); err != nil {
		return nil, err
	}

	var cases []*ast.SelectCase
	for p.current.Type != token.DEDENT && p.current.Type != token.EOF {
		if p.current.Type == token.NEWLINE {
			p.advance()
			continue
		}

		sc, err := parseSelectCase(p)
		if err != nil {
			return nil, err
		}
		cases = append(cases, sc)
	}

	if p.current.Type == token.DEDENT {
		p.advance()
	}

	return &ast.SelectBlock{
		Pos:   pos,
		Cases: cases,
	}, nil
}

// parseSelectCase parses a single select case.
func parseSelectCase(p *Parser) (*ast.SelectCase, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	sc := &ast.SelectCase{Pos: pos}

	switch p.current.Type {
	case token.DETACH:
		return nil, p.errorf("detach is not allowed in select cases")

	case token.SPAWN, token.NEXUS, token.WORKFLOW:
		// Workflow case: [ SPAWN ] [ NEXUS STRING ] WORKFLOW IDENT ARGS [ ARROW IDENT ] COLON
		mode := ast.CallChild
		if p.current.Type == token.SPAWN {
			mode = ast.CallSpawn
			p.advance()
		}
		sc.WorkflowMode = mode

		if p.current.Type == token.NEXUS {
			p.advance()
			ns, err := p.expect(token.STRING)
			if err != nil {
				return nil, err
			}
			sc.WorkflowNamespace = ns.Literal
		}

		if _, err := p.expect(token.WORKFLOW); err != nil {
			return nil, err
		}
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		sc.WorkflowName = name.Literal

		args, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		sc.WorkflowArgs = args.Literal

		if p.current.Type == token.ARROW {
			p.advance()
			res, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			sc.WorkflowResult = res.Literal
		}

	case token.ACTIVITY:
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		sc.ActivityName = name.Literal

		args, err := p.expect(token.ARGS)
		if err != nil {
			return nil, err
		}
		sc.ActivityArgs = args.Literal

		if p.current.Type == token.ARROW {
			p.advance()
			res, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			sc.ActivityResult = res.Literal
		}

	case token.SIGNAL:
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		sc.SignalName = name.Literal
		if p.current.Type == token.ARGS {
			sc.SignalArgs = p.current.Literal
			p.advance()
		}

	case token.UPDATE:
		p.advance()
		name, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		sc.UpdateName = name.Literal
		if p.current.Type == token.ARGS {
			sc.UpdateArgs = p.current.Literal
			p.advance()
		}

	case token.TIMER:
		p.advance()
		sc.TimerDuration = p.collectRawUntil(token.COLON, token.NEWLINE)

	default:
		return nil, p.errorf("unexpected token %s in select case", p.current.Type)
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
	sc.Body = body

	return sc, nil
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
