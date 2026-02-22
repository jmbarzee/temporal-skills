package parser

import (
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/token"
)

// callParts holds the shared parsed components of an activity or workflow call.
type callParts struct {
	pos     ast.Pos
	name    string
	args    string
	result  string
	options *ast.OptionsBlock
}

// parseCallParts parses the shared IDENT ARGS [ ARROW IDENT ] NEWLINE [ options ] pattern.
func parseCallParts(p *Parser, optCtx OptionsContext) (*callParts, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume keyword

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

	options, err := p.parseOptionalOptionsLine(optCtx)
	if err != nil {
		return nil, err
	}

	return &callParts{pos: pos, name: name.Literal, args: args.Literal, result: result, options: options}, nil
}

// parseActivityCall parses: ACTIVITY IDENT ARGS [ ARROW IDENT ] NEWLINE [ options_line ]
func parseActivityCall(p *Parser) (ast.Statement, error) {
	cp, err := parseCallParts(p, OptionsContextActivity)
	if err != nil {
		return nil, err
	}
	return &ast.ActivityCall{
		Pos:      cp.pos,
		Activity: ast.Ref[*ast.ActivityDef]{Pos: cp.pos, Name: cp.name},
		Args:     cp.args,
		Result:   cp.result,
		Options:  cp.options,
	}, nil
}

// parseWorkflowCall parses: WORKFLOW IDENT ARGS [ ARROW IDENT ] NEWLINE [ options_line ]
func parseWorkflowCall(p *Parser) (ast.Statement, error) {
	cp, err := parseCallParts(p, OptionsContextWorkflow)
	if err != nil {
		return nil, err
	}
	return &ast.WorkflowCall{
		Pos:      cp.pos,
		Mode:     ast.CallChild,
		Workflow: ast.Ref[*ast.WorkflowDef]{Pos: cp.pos, Name: cp.name},
		Args:     cp.args,
		Result:   cp.result,
		Options:  cp.options,
	}, nil
}
