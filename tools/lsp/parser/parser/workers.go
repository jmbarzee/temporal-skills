package parser

import (
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/token"
)

// parseWorkerDef parses:
// WORKER IDENT COLON NEWLINE INDENT worker_entries DEDENT
func parseWorkerDef(p *Parser) (ast.Definition, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
	p.advance() // consume WORKER

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

	worker := &ast.WorkerDef{
		Pos:  pos,
		Name: name.Literal,
	}

	hasNamespace := false
	hasTaskQueue := false

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

		case token.NAMESPACE:
			if hasNamespace {
				return nil, p.errorf("duplicate namespace declaration in worker %s", worker.Name)
			}
			p.advance() // consume NAMESPACE
			ns, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			worker.Namespace = ns.Literal
			hasNamespace = true
			if p.current.Type == token.NEWLINE {
				p.advance()
			}

		case token.TASK_QUEUE:
			if hasTaskQueue {
				return nil, p.errorf("duplicate task_queue declaration in worker %s", worker.Name)
			}
			p.advance() // consume TASK_QUEUE
			tq, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			worker.TaskQueue = tq.Literal
			hasTaskQueue = true
			if p.current.Type == token.NEWLINE {
				p.advance()
			}

		case token.WORKFLOW:
			refPos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
			p.advance() // consume WORKFLOW
			wfName, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			worker.Workflows = append(worker.Workflows, ast.WorkerRef{
				Pos:  refPos,
				Name: wfName.Literal,
			})
			if p.current.Type == token.NEWLINE {
				p.advance()
			}

		case token.ACTIVITY:
			refPos := ast.Pos{Line: p.current.Line, Column: p.current.Column}
			p.advance() // consume ACTIVITY
			actName, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			worker.Activities = append(worker.Activities, ast.WorkerRef{
				Pos:  refPos,
				Name: actName.Literal,
			})
			if p.current.Type == token.NEWLINE {
				p.advance()
			}

		default:
			return nil, p.errorf("unexpected %s in worker block", p.current.Type)
		}
	}

	if p.current.Type == token.DEDENT {
		p.advance()
	}

	if !hasNamespace {
		return nil, &ParseError{
			Msg:    "worker " + worker.Name + " missing required namespace declaration",
			Line:   pos.Line,
			Column: pos.Column,
		}
	}
	if !hasTaskQueue {
		return nil, &ParseError{
			Msg:    "worker " + worker.Name + " missing required task_queue declaration",
			Line:   pos.Line,
			Column: pos.Column,
		}
	}

	return worker, nil
}
