package token

import "fmt"

// TokenType represents the type of a lexical token.
type TokenType int

const (
	// Structural
	EOF TokenType = iota
	NEWLINE
	INDENT
	DEDENT

	// Keywords -- top-level defs
	WORKFLOW
	ACTIVITY

	// Keywords -- workflow-level declarations
	SIGNAL
	QUERY
	UPDATE

	// Keywords -- workflow call modifiers
	SPAWN
	DETACH
	NEXUS

	// Keywords -- calls and primitives
	TIMER
	OPTIONS

	// Keywords -- async
	AWAIT
	OR

	// Keywords -- blocks
	PARALLEL
	SELECT
	SWITCH
	CASE

	// Keywords -- control flow
	IF
	ELSE
	FOR
	IN

	// Keywords -- simple statements
	RETURN
	CONTINUE_AS_NEW
	BREAK
	CONTINUE

	// Symbols
	COLON    // :
	ARROW    // ->

	// Values
	IDENT    // non-keyword identifiers
	STRING   // quoted string (for nexus namespaces)
	ARGS     // raw content between ( and ), no nested parens
	COMMENT  // text after #
	RAW_TEXT // anything else
)

var tokenNames = map[TokenType]string{
	EOF:             "EOF",
	NEWLINE:         "NEWLINE",
	INDENT:          "INDENT",
	DEDENT:          "DEDENT",
	WORKFLOW:        "WORKFLOW",
	ACTIVITY:        "ACTIVITY",
	SIGNAL:          "SIGNAL",
	QUERY:           "QUERY",
	UPDATE:          "UPDATE",
	SPAWN:           "SPAWN",
	DETACH:          "DETACH",
	NEXUS:           "NEXUS",
	TIMER:           "TIMER",
	OPTIONS:         "OPTIONS",
	AWAIT:           "AWAIT",
	OR:              "OR",
	PARALLEL:        "PARALLEL",
	SELECT:          "SELECT",
	SWITCH:          "SWITCH",
	CASE:            "CASE",
	IF:              "IF",
	ELSE:            "ELSE",
	FOR:             "FOR",
	IN:              "IN",
	RETURN:          "RETURN",
	CONTINUE_AS_NEW: "CONTINUE_AS_NEW",
	BREAK:           "BREAK",
	CONTINUE:        "CONTINUE",
	COLON:           "COLON",
	ARROW:           "ARROW",
	IDENT:           "IDENT",
	STRING:          "STRING",
	ARGS:            "ARGS",
	COMMENT:         "COMMENT",
	RAW_TEXT:        "RAW_TEXT",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("TokenType(%d)", int(t))
}

// Token represents a single lexical token.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

func (t Token) String() string {
	if t.Literal == "" {
		return fmt.Sprintf("%s@%d:%d", t.Type, t.Line, t.Column)
	}
	return fmt.Sprintf("%s(%q)@%d:%d", t.Type, t.Literal, t.Line, t.Column)
}

var keywords = map[string]TokenType{
	"workflow":        WORKFLOW,
	"activity":        ACTIVITY,
	"signal":          SIGNAL,
	"query":           QUERY,
	"update":          UPDATE,
	"spawn":           SPAWN,
	"detach":          DETACH,
	"nexus":           NEXUS,
	"timer":           TIMER,
	"options":         OPTIONS,
	"await":           AWAIT,
	"or":              OR,
	"parallel":        PARALLEL,
	"select":          SELECT,
	"switch":          SWITCH,
	"case":            CASE,
	"if":              IF,
	"else":            ELSE,
	"for":             FOR,
	"in":              IN,
	"return":          RETURN,
	"continue_as_new": CONTINUE_AS_NEW,
	"break":           BREAK,
	"continue":        CONTINUE,
}

// LookupIdent returns the TokenType for an identifier string.
// If the identifier is a keyword, the keyword token type is returned.
// Otherwise, IDENT is returned.
func LookupIdent(ident string) TokenType {
	if tt, ok := keywords[ident]; ok {
		return tt
	}
	return IDENT
}
