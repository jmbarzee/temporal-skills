package server

import (
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/lexer"
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/token"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// Semantic token type indices (must match legend order).
const (
	semKeyword   = 0
	semFunction  = 1
	semMethod    = 2
	semEvent     = 3
	semString    = 4
	semComment   = 5
	semOperator  = 6
	semParameter = 7
)

// Semantic token modifier bits.
const (
	modDeclaration = 1 << 0
)

func semanticTokensHandler(store *DocumentStore) protocol.TextDocumentSemanticTokensFullFunc {
	return func(context *glsp.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
		doc := store.Get(params.TextDocument.URI)
		if doc == nil {
			return nil, nil
		}

		data := buildSemanticTokens(doc.Content)
		return &protocol.SemanticTokens{
			Data: data,
		}, nil
	}
}

// buildSemanticTokens lexes the content and returns delta-encoded semantic token data.
func buildSemanticTokens(content string) []uint32 {
	tokens := lexer.New(content).AllTokens()

	var data []uint32
	var prevLine, prevCol uint32
	var prevType token.TokenType
	indentLevel := 0

	for _, tok := range tokens {
		switch tok.Type {
		case token.INDENT:
			indentLevel++
			prevType = tok.Type
			continue
		case token.DEDENT:
			indentLevel--
			if indentLevel < 0 {
				indentLevel = 0
			}
			prevType = tok.Type
			continue
		}

		tokenType, modifiers, shouldEmit := classifyToken(tok, prevType, indentLevel)
		if !shouldEmit {
			if !isStructural(tok.Type) {
				prevType = tok.Type
			}
			continue
		}

		length := tokenLength(tok)
		line := uint32(tok.Line - 1)   // LSP 0-based
		col := uint32(tok.Column - 1)  // LSP 0-based

		deltaLine := line - prevLine
		var deltaCol uint32
		if deltaLine == 0 {
			deltaCol = col - prevCol
		} else {
			deltaCol = col
		}

		data = append(data, deltaLine, deltaCol, length, tokenType, modifiers)
		prevLine = line
		prevCol = col

		if !isStructural(tok.Type) {
			prevType = tok.Type
		}
	}

	return data
}

// isStructural returns true for tokens that don't affect classification context.
func isStructural(tt token.TokenType) bool {
	switch tt {
	case token.NEWLINE, token.EOF:
		return true
	default:
		return false
	}
}

// classifyToken determines the semantic token type and modifiers for a token.
func classifyToken(tok token.Token, prevType token.TokenType, indentLevel int) (tokenType uint32, modifiers uint32, shouldEmit bool) {
	switch tok.Type {
	case token.WORKFLOW, token.ACTIVITY,
		token.SIGNAL, token.QUERY, token.UPDATE,
		token.SPAWN, token.DETACH, token.NEXUS,
		token.TIMER, token.OPTIONS,
		token.AWAIT, token.ALL, token.ONE,
		token.SWITCH, token.CASE,
		token.IF, token.ELSE, token.FOR, token.IN,
		token.RETURN, token.CONTINUE_AS_NEW, token.BREAK, token.CONTINUE:
		return semKeyword, 0, true

	case token.IDENT:
		return classifyIdent(prevType, indentLevel)

	case token.STRING:
		return semString, 0, true

	case token.COMMENT:
		return semComment, 0, true

	case token.COLON, token.ARROW:
		return semOperator, 0, true

	case token.ARGS:
		return semParameter, 0, true

	default:
		return 0, 0, false
	}
}

// classifyIdent determines the semantic type for an IDENT based on context.
func classifyIdent(prevType token.TokenType, indentLevel int) (tokenType uint32, modifiers uint32, shouldEmit bool) {
	switch prevType {
	case token.WORKFLOW, token.ACTIVITY:
		if indentLevel == 0 {
			return semFunction, modDeclaration, true
		}
		return semFunction, 0, true

	case token.SIGNAL:
		return semEvent, modDeclaration, true

	case token.QUERY, token.UPDATE:
		return semMethod, modDeclaration, true

	default:
		// Bare ident in body — treat as call site.
		return semFunction, 0, true
	}
}

// tokenLength returns the display length of a token.
func tokenLength(tok token.Token) uint32 {
	switch tok.Type {
	case token.ARGS:
		return uint32(len(tok.Literal)) + 2 // parens
	case token.STRING:
		return uint32(len(tok.Literal)) + 2 // quotes
	case token.COMMENT:
		return uint32(len(tok.Literal)) + 1 // #
	default:
		return uint32(len(tok.Literal))
	}
}
