package lexer

import (
	"fmt"

	"github.com/jmbarzee/temporal-skills/design/lsp/parser/token"
)

// Lexer tokenizes .twf source input with indentation-aware INDENT/DEDENT emission.
type Lexer struct {
	input   []byte
	pos     int // current position in input
	line    int // 1-based line number
	col     int // 1-based column number
	atBOL   bool // at beginning of line (for indent processing)
	pending []token.Token // queued tokens (INDENT/DEDENT)

	indentStack []int // stack of indent levels, starts at [0]
}

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	return &Lexer{
		input:       []byte(input),
		pos:         0,
		line:        1,
		col:         1,
		atBOL:       true,
		indentStack: []int{0},
	}
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() token.Token {
	for {
		// Return pending tokens first (INDENT/DEDENT).
		if len(l.pending) > 0 {
			tok := l.pending[0]
			l.pending = l.pending[1:]
			return tok
		}

		// At beginning of line, handle indentation.
		if l.atBOL {
			l.atBOL = false
			tok, hasTok := l.handleIndent()
			if hasTok {
				return tok
			}
			// handleIndent may have queued tokens or set atBOL again (blank line).
			if len(l.pending) > 0 || l.atBOL {
				continue
			}
		}

		// Skip spaces (not newlines) within a line.
		l.skipSpaces()

		if l.pos >= len(l.input) {
			return l.emitEOF()
		}

		ch := l.input[l.pos]

		var tok token.Token
		switch {
		case ch == '\n':
			tok = l.makeToken(token.NEWLINE, "")
			l.advance()
			l.line++
			l.col = 1
			l.atBOL = true

		case ch == '#':
			tok = l.scanComment()

		case ch == '(':
			tok = l.scanArgs()

		case ch == '"':
			tok = l.scanString()

		case ch == ':':
			tok = l.makeToken(token.COLON, ":")
			l.advance()

		case ch == '-' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '>':
			tok = l.makeToken(token.ARROW, "->")
			l.advance()
			l.advance()

		case isIdentStart(ch):
			tok = l.scanIdentifier()

		default:
			tok = l.scanRawText()
		}

		return tok
	}
}

// AllTokens returns all tokens from the input.
func (l *Lexer) AllTokens() []token.Token {
	var tokens []token.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}

// handleIndent processes whitespace at the beginning of a line.
// Returns a token and true if a single token should be returned immediately,
// or zero-value and false if tokens were queued into pending (or blank line skipped).
func (l *Lexer) handleIndent() (token.Token, bool) {
	// Count leading spaces.
	indent := 0
	for l.pos < len(l.input) && l.input[l.pos] == ' ' {
		indent++
		l.pos++
		l.col++
	}

	// Check for blank line (only whitespace then newline or EOF).
	if l.pos >= len(l.input) {
		// EOF after spaces — emit dedents to 0 (handled by emitEOF).
		l.emitDedentsTo(0)
		return token.Token{}, false
	}
	if l.input[l.pos] == '\n' {
		// Blank line — skip entirely.
		l.advance() // consume the newline
		l.line++
		l.col = 1
		l.atBOL = true
		return token.Token{}, false
	}

	top := l.indentStack[len(l.indentStack)-1]

	if indent > top {
		l.indentStack = append(l.indentStack, indent)
		return l.makeToken(token.INDENT, ""), true
	}

	if indent < top {
		l.emitDedentsTo(indent)
		return token.Token{}, false
	}

	// indent == top: no change
	return token.Token{}, false
}

// emitDedentsTo emits DEDENT tokens until the stack top matches the target level.
func (l *Lexer) emitDedentsTo(target int) {
	for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > target {
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
		l.pending = append(l.pending, l.makeToken(token.DEDENT, ""))
	}
}

// emitEOF emits a synthetic NEWLINE (if needed), remaining DEDENTs, then EOF.
func (l *Lexer) emitEOF() token.Token {
	// If the input doesn't end with a newline, emit a synthetic one so
	// the parser always sees NEWLINE before DEDENT/EOF.
	if len(l.input) > 0 && l.input[len(l.input)-1] != '\n' {
		// Rewrite the input to appear as if it ended with \n so this
		// branch only fires once.
		l.input = append(l.input, '\n')
		nl := l.makeToken(token.NEWLINE, "")
		// Queue remaining dedents and EOF after the newline.
		for len(l.indentStack) > 1 {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			l.pending = append(l.pending, l.makeToken(token.DEDENT, ""))
		}
		l.pending = append(l.pending, l.makeToken(token.EOF, ""))
		return nl
	}

	if len(l.indentStack) > 1 {
		for len(l.indentStack) > 1 {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			l.pending = append(l.pending, l.makeToken(token.DEDENT, ""))
		}
		l.pending = append(l.pending, l.makeToken(token.EOF, ""))
		tok := l.pending[0]
		l.pending = l.pending[1:]
		return tok
	}
	return l.makeToken(token.EOF, "")
}

func (l *Lexer) scanComment() token.Token {
	tok := l.makeToken(token.COMMENT, "")
	l.advance() // consume '#'
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.advance()
	}
	tok.Literal = string(l.input[start:l.pos])
	return tok
}

func (l *Lexer) scanArgs() token.Token {
	tok := l.makeToken(token.ARGS, "")
	l.advance() // consume '('
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != ')' {
		if l.input[l.pos] == '\n' {
			l.line++
			l.col = 0 // advance will set to 1
		}
		l.advance()
	}
	tok.Literal = string(l.input[start:l.pos])
	if l.pos < len(l.input) {
		l.advance() // consume ')'
	}
	return tok
}

func (l *Lexer) scanString() token.Token {
	tok := l.makeToken(token.STRING, "")
	l.advance() // consume opening '"'
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != '"' {
		if l.input[l.pos] == '\n' {
			l.line++
			l.col = 0
		}
		l.advance()
	}
	tok.Literal = string(l.input[start:l.pos])
	if l.pos < len(l.input) {
		l.advance() // consume closing '"'
	}
	return tok
}

func (l *Lexer) scanIdentifier() token.Token {
	tok := l.makeToken(token.IDENT, "")
	start := l.pos
	for l.pos < len(l.input) && isIdentContinue(l.input[l.pos]) {
		l.advance()
	}
	literal := string(l.input[start:l.pos])
	tok.Literal = literal
	tok.Type = token.LookupIdent(literal)
	return tok
}

func (l *Lexer) scanRawText() token.Token {
	tok := l.makeToken(token.RAW_TEXT, "")
	start := l.pos
	l.advance()
	tok.Literal = string(l.input[start:l.pos])
	return tok
}

func (l *Lexer) makeToken(tt token.TokenType, literal string) token.Token {
	return token.Token{
		Type:    tt,
		Literal: literal,
		Line:    l.line,
		Column:  l.col,
	}
}

func (l *Lexer) advance() {
	l.pos++
	l.col++
}

func (l *Lexer) skipSpaces() {
	for l.pos < len(l.input) && l.input[l.pos] == ' ' {
		l.advance()
	}
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentContinue(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9')
}

// LexError represents a lexer error with position.
type LexError struct {
	Msg    string
	Line   int
	Column int
}

func (e *LexError) Error() string {
	return fmt.Sprintf("lexer error at %d:%d: %s", e.Line, e.Column, e.Msg)
}
