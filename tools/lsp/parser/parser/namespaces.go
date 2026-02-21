package parser

import (
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/token"
)

// parseNamespaceDef parses:
// NAMESPACE IDENT COLON NEWLINE INDENT namespace_entries DEDENT
func parseNamespaceDef(p *Parser) (ast.Definition, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume NAMESPACE

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	if err := p.expectBlock(); err != nil {
		return nil, err
	}

	ns := &ast.NamespaceDef{
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

		case token.WORKER:
			workerPos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
			p.advance() // consume WORKER
			workerName, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			if p.current.Type == token.NEWLINE {
				p.advance()
			}
			opts, err := p.parseOptionalOptionsLine(OptionsContextWorker)
			if err != nil {
				return nil, err
			}
			ns.Workers = append(ns.Workers, ast.NamespaceWorker{
				Pos:        workerPos,
				WorkerName: workerName.Literal,
				Options:    opts,
			})

		case token.NEXUS:
			epPos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
			p.advance() // consume NEXUS
			// Expect IDENT "endpoint"
			if p.current.Type != token.IDENT || p.current.Literal != "endpoint" {
				return nil, p.errorf("expected 'endpoint' after 'nexus' in namespace block, got %s %q", p.current.Type, p.current.Literal)
			}
			p.advance() // consume "endpoint"
			epName, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			if p.current.Type == token.NEWLINE {
				p.advance()
			}
			opts, err := p.parseOptionalOptionsLine(OptionsContextEndpoint)
			if err != nil {
				return nil, err
			}
			ns.Endpoints = append(ns.Endpoints, ast.NamespaceEndpoint{
				Pos:          epPos,
				EndpointName: epName.Literal,
				Options:      opts,
			})

		default:
			return nil, p.errorf("unexpected %s in namespace block", p.current.Type)
		}
	}

	if p.current.Type == token.DEDENT {
		p.advance()
	}

	return ns, nil
}
