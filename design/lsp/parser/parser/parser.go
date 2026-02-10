package parser

import (
	"fmt"

	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/lexer"
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/token"
)

// ParseError represents a parse error with position info.
type ParseError struct {
	Msg    string
	Line   int
	Column int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at %d:%d: %s", e.Line, e.Column, e.Msg)
}

type defParser func(p *Parser) (ast.Definition, error)
type stmtParser func(p *Parser) (ast.Statement, error)

// Parser is a recursive descent parser for .twf files.
type Parser struct {
	lex     *lexer.Lexer
	current token.Token
	peek    token.Token

	inWorkflow bool
	inActivity bool

	collecting bool          // true when collecting errors instead of bailing
	errors     []*ParseError // accumulated errors in collecting mode
}

// Registration maps for keyword dispatch.
var (
	topLevelParsers     map[token.TokenType]defParser
	workflowStmtParsers map[token.TokenType]stmtParser
	activityStmtParsers map[token.TokenType]stmtParser
)

func init() {
	topLevelParsers = map[token.TokenType]defParser{
		token.WORKFLOW: parseWorkflowDef,
		token.ACTIVITY: parseActivityDef,
	}

	workflowStmtParsers = map[token.TokenType]stmtParser{
		token.ACTIVITY:        parseActivityCall,
		token.WORKFLOW:        parseWorkflowCall,
		token.SPAWN:           parseWorkflowCall,
		token.DETACH:          parseWorkflowCall,
		token.NEXUS:           parseWorkflowCall,
		token.AWAIT:           parseAwaitStmt, // handles both single await and await blocks
		token.SWITCH:          parseSwitchBlock,
		token.IF:              parseIfStmt,
		token.FOR:             parseForStmt,
		token.CLOSE:           parseCloseStmt,
		token.RETURN:          parseReturnStmt,
		token.CONTINUE_AS_NEW: parseContinueAsNewStmt,
		token.BREAK:           parseBreakStmt,
		token.CONTINUE:        parseContinueStmt,
	}

	activityStmtParsers = map[token.TokenType]stmtParser{
		token.SWITCH:   parseSwitchBlock,
		token.IF:       parseIfStmt,
		token.FOR:      parseForStmt,
		token.RETURN:   parseReturnStmt,
		token.BREAK:    parseBreakStmt,
		token.CONTINUE: parseContinueStmt,
	}
}

// temporalKeywords are keywords that are not allowed in activity bodies.
var temporalKeywords = map[token.TokenType]bool{
	token.WORKFLOW:        true,
	token.ACTIVITY:        true,
	token.SIGNAL:          true,
	token.QUERY:           true,
	token.UPDATE:          true,
	token.SPAWN:           true,
	token.DETACH:          true,
	token.NEXUS:           true,
	token.TIMER:           true,
	token.AWAIT:           true,
	token.ALL:             true,
	token.ONE:             true,
	token.CONTINUE_AS_NEW: true,
	token.CLOSE:           true,
}

// ParseFile parses a .twf source string into an AST File.
func ParseFile(input string) (*ast.File, error) {
	l := lexer.New(input)
	p := &Parser{lex: l}
	p.advance() // fill current
	p.advance() // fill peek

	file := &ast.File{}

	for p.current.Type != token.EOF {
		switch {
		case p.current.Type == token.NEWLINE:
			p.advance()
			continue
		case p.current.Type == token.COMMENT:
			p.advance()
			continue
		default:
			parser, ok := topLevelParsers[p.current.Type]
			if !ok {
				return nil, p.errorf("unexpected token %s at top level", p.current.Type)
			}
			def, err := parser(p)
			if err != nil {
				return nil, err
			}
			file.Definitions = append(file.Definitions, def)
		}
	}

	return file, nil
}

// ParseFileAll parses a .twf source string, collecting as many errors as
// possible instead of stopping at the first one. It returns a partial AST
// (which may have successfully parsed definitions) alongside all parse errors.
func ParseFileAll(input string) (*ast.File, []*ParseError) {
	l := lexer.New(input)
	p := &Parser{lex: l, collecting: true}
	p.advance() // fill current
	p.advance() // fill peek

	file := &ast.File{}

	for p.current.Type != token.EOF {
		switch {
		case p.current.Type == token.NEWLINE:
			p.advance()
			continue
		case p.current.Type == token.COMMENT:
			p.advance()
			continue
		default:
			parser, ok := topLevelParsers[p.current.Type]
			if !ok {
				p.addError(p.errorf("unexpected token %s at top level", p.current.Type).(*ParseError))
				p.recoverTopLevel()
				continue
			}
			def, err := parser(p)
			if err != nil {
				if pe, ok := err.(*ParseError); ok {
					p.addError(pe)
				}
				p.recoverTopLevel()
				continue
			}
			file.Definitions = append(file.Definitions, def)
		}
	}

	return file, p.errors
}

// parseBody parses statements inside an indented block (after INDENT, until DEDENT).
func (p *Parser) parseBody() ([]ast.Statement, error) {
	var stmts []ast.Statement
	for p.current.Type != token.DEDENT && p.current.Type != token.EOF {
		if p.current.Type == token.NEWLINE {
			p.advance()
			continue
		}
		if p.current.Type == token.COMMENT {
			stmts = append(stmts, &ast.Comment{
				Pos:  ast.Pos{Line: p.current.Line, Column: p.current.Column},
				Text: p.current.Literal,
			})
			p.advance()
			if p.current.Type == token.NEWLINE {
				p.advance()
			}
			continue
		}

		var parseFn stmtParser
		var ok bool
		if p.inWorkflow {
			parseFn, ok = workflowStmtParsers[p.current.Type]
		} else if p.inActivity {
			// Check for temporal keywords that aren't allowed.
			if temporalKeywords[p.current.Type] {
				return nil, p.errorf("%s is not allowed in activity body", p.current.Literal)
			}
			parseFn, ok = activityStmtParsers[p.current.Type]
		}

		if !ok {
			// Fallback to raw statement.
			stmt, err := parseRawStmt(p)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, stmt)
			continue
		}

		stmt, err := parseFn(p)
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}

	if p.current.Type == token.DEDENT {
		p.advance()
	}

	return stmts, nil
}
