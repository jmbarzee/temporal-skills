package parser

import (
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/token"
)

// parseNexusTopLevel dispatches nexus top-level definitions.
// Current token is NEXUS. Peek: if IDENT "service" → parseNexusServiceDef
func parseNexusTopLevel(p *Parser) (ast.Definition, error) {
	// Current = NEXUS. Check if next is IDENT "service".
	if p.peek.Type == token.IDENT && p.peek.Literal == "service" {
		return parseNexusServiceDef(p)
	}
	return nil, p.errorf("expected 'service' after 'nexus' at top level, got %s", p.peek.Type)
}

// parseNexusServiceDef parses:
// NEXUS "service" IDENT COLON NEWLINE INDENT operations DEDENT
func parseNexusServiceDef(p *Parser) (ast.Definition, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume NEXUS
	p.advance() // consume "service" IDENT

	name, err := p.expect(token.IDENT)
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

	svc := &ast.NexusServiceDef{
		Pos:  pos,
		Name: name.Literal,
	}

	for p.current.Type != token.DEDENT && p.current.Type != token.EOF {
		switch p.current.Type {
		case token.NEWLINE:
			p.advance()
			continue
		case token.COMMENT:
			p.advance()
			if p.current.Type == token.NEWLINE {
				p.advance()
			}
			continue
		case token.ASYNC:
			op, err := parseAsyncOperation(p)
			if err != nil {
				return nil, err
			}
			svc.Operations = append(svc.Operations, op)
		case token.SYNC:
			op, err := parseSyncOperation(p)
			if err != nil {
				return nil, err
			}
			svc.Operations = append(svc.Operations, op)
		default:
			return nil, p.errorf("expected 'async' or 'sync' in nexus service body, got %s", p.current.Type)
		}
	}

	if p.current.Type == token.DEDENT {
		p.advance()
	}

	return svc, nil
}

// parseAsyncOperation parses: ASYNC IDENT WORKFLOW IDENT NEWLINE
func parseAsyncOperation(p *Parser) (*ast.NexusOperation, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume ASYNC

	opName, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(token.WORKFLOW); err != nil {
		return nil, err
	}

	wfName, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return &ast.NexusOperation{
		Pos:          pos,
		OpType:       ast.NexusOpAsync,
		Name:         opName.Literal,
		WorkflowName: wfName.Literal,
	}, nil
}

// parseSyncOperation parses: SYNC IDENT ARGS ARROW ARGS COLON NEWLINE INDENT body DEDENT
func parseSyncOperation(p *Parser) (*ast.NexusOperation, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume SYNC

	opName, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	params, err := p.expect(token.ARGS)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(token.ARROW); err != nil {
		return nil, err
	}

	retType, err := p.expect(token.ARGS)
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

	// Parse body with workflow statement set (sync ops can use temporal primitives)
	prevInWorkflow := p.inWorkflow
	p.inWorkflow = true
	body, err := p.parseBody()
	p.inWorkflow = prevInWorkflow
	if err != nil {
		return nil, err
	}

	return &ast.NexusOperation{
		Pos:        pos,
		OpType:     ast.NexusOpSync,
		Name:       opName.Literal,
		Params:     params.Literal,
		ReturnType: retType.Literal,
		Body:       body,
	}, nil
}

// parseNexusCall parses: NEXUS IDENT IDENT DOT IDENT ARGS [ARROW IDENT] NEWLINE [options]
// Called when current token is NEXUS inside a workflow body.
func parseNexusCall(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	return parseNexusCallInner(p, pos, false)
}

// parseNexusCallInner is the shared parser for nexus calls (direct and detach).
func parseNexusCallInner(p *Parser, pos ast.Pos, detach bool) (ast.Statement, error) {
	p.advance() // consume NEXUS

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
	if detach && result != "" {
		return nil, &ParseError{
			Msg:    "detach nexus call cannot have a result (-> identifier)",
			Line:   pos.Line,
			Column: pos.Column,
		}
	}

	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	options, err := p.parseOptionalOptionsLine(OptionsContextNexusCall)
	if err != nil {
		return nil, err
	}

	return &ast.NexusCall{
		Pos:       pos,
		Detach:    detach,
		Endpoint:  endpoint.Literal,
		Service:   service.Literal,
		Operation: operation.Literal,
		Args:      args.Literal,
		Result:    result,
		Options:   options,
	}, nil
}

// parseWorkflowCallOrNexus handles DETACH dispatch: detach workflow ... or detach nexus ...
func parseWorkflowCallOrNexus(p *Parser) (ast.Statement, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume DETACH

	if p.current.Type == token.NEXUS {
		return parseNexusCallInner(p, pos, true)
	}

	// Fall through to workflow call (detach workflow ...)
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
	if result != "" {
		return nil, &ParseError{
			Msg:    "detach workflow call cannot have a result (-> identifier)",
			Line:   pos.Line,
			Column: pos.Column,
		}
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
		Mode:    ast.CallDetach,
		Name:    name.Literal,
		Args:    args.Literal,
		Result:  result,
		Options: options,
	}, nil
}
