package parser

import (
	"strings"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/token"
)

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
