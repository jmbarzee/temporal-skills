package parser

import (
	"fmt"
	"strings"

	"github.com/jmbarzee/temporal-skills/design/lsp/parser/token"
)

// advance moves to the next token.
func (p *Parser) advance() {
	p.current = p.peek
	p.peek = p.lex.NextToken()
}

// expect consumes the current token if it matches the expected type.
// Returns the consumed token or an error.
func (p *Parser) expect(tt token.TokenType) (token.Token, error) {
	if p.current.Type != tt {
		return token.Token{}, p.errorf("expected %s, got %s (%q)", tt, p.current.Type, p.current.Literal)
	}
	tok := p.current
	p.advance()
	return tok, nil
}

// errorf creates a ParseError at the current token position.
func (p *Parser) errorf(format string, args ...interface{}) error {
	return &ParseError{
		Msg:    fmt.Sprintf(format, args...),
		Line:   p.current.Line,
		Column: p.current.Column,
	}
}

// addError appends a parse error to the accumulated error list.
func (p *Parser) addError(err *ParseError) {
	p.errors = append(p.errors, err)
}

// recoverTopLevel skips tokens until the parser reaches a WORKFLOW or ACTIVITY
// keyword at column 1 (top-level boundary) or EOF.
func (p *Parser) recoverTopLevel() {
	for p.current.Type != token.EOF {
		if (p.current.Type == token.WORKFLOW || p.current.Type == token.ACTIVITY) && p.current.Column == 1 {
			return
		}
		p.advance()
	}
}

// collectRawUntil reads and concatenates token literals until one of the
// terminator token types is found. The terminator is NOT consumed.
// Uses token positions to preserve original spacing.
func (p *Parser) collectRawUntil(terminators ...token.TokenType) string {
	var b strings.Builder
	lastEnd := 0 // column after last token
	for {
		for _, t := range terminators {
			if p.current.Type == t {
				return strings.TrimSpace(b.String())
			}
		}
		if p.current.Type == token.EOF {
			return strings.TrimSpace(b.String())
		}
		// Reconstruct spacing: if this token starts after last token ended,
		// insert spaces. For the first token or same-line adjacent tokens.
		if lastEnd > 0 && p.current.Column > lastEnd {
			for i := 0; i < p.current.Column-lastEnd; i++ {
				b.WriteByte(' ')
			}
		}
		b.WriteString(p.current.Literal)
		lastEnd = p.current.Column + len(p.current.Literal)
		p.advance()
	}
}

// skipBlankLinesAndComments consumes any NEWLINE and COMMENT tokens at the current position.
// This allows comments to appear between declarations without breaking parsing.
func (p *Parser) skipBlankLinesAndComments() {
	for p.current.Type == token.NEWLINE || p.current.Type == token.COMMENT {
		p.advance()
	}
}

// parseOptionalOptionsLine checks for an options line after a call:
// INDENT OPTIONS ARGS NEWLINE DEDENT
// Returns the options args string (empty if no options found).
func (p *Parser) parseOptionalOptionsLine() (string, error) {
	if p.current.Type != token.INDENT {
		return "", nil
	}
	if p.peek.Type != token.OPTIONS {
		return "", nil
	}
	// Consume INDENT
	p.advance()
	// Consume OPTIONS
	p.advance()
	// Expect ARGS
	args, err := p.expect(token.ARGS)
	if err != nil {
		return "", err
	}
	// Consume NEWLINE if present
	if p.current.Type == token.NEWLINE {
		p.advance()
	}
	// Expect DEDENT
	if _, err := p.expect(token.DEDENT); err != nil {
		return "", err
	}
	return args.Literal, nil
}

// parseOptionalOptionsStmt checks for an options statement at the start of a body:
// OPTIONS ARGS NEWLINE
// Returns the options args string (empty if no options found).
func (p *Parser) parseOptionalOptionsStmt() (string, error) {
	if p.current.Type != token.OPTIONS {
		return "", nil
	}
	p.advance() // consume OPTIONS
	args, err := p.expect(token.ARGS)
	if err != nil {
		return "", err
	}
	if p.current.Type == token.NEWLINE {
		p.advance()
	}
	return args.Literal, nil
}
