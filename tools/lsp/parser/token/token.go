package token

import (
	"fmt"
	"strings"
)

// TokenType represents the type of a lexical token.
type TokenType int

const (
	// Structural
	EOF TokenType = iota
	ILLEGAL
	NEWLINE
	INDENT
	DEDENT

	// Keywords -- top-level defs
	WORKFLOW
	ACTIVITY
	WORKER

	// Keywords -- worker-level declarations
	NAMESPACE
	TASK_QUEUE

	// Keywords -- workflow-level declarations
	SIGNAL
	QUERY
	UPDATE

	// Keywords -- workflow call modifiers
	DETACH
	NEXUS
	SYNC
	ASYNC

	// Keywords -- promises and conditions
	PROMISE
	CONDITION
	SET
	UNSET
	STATE

	// Keywords -- calls and primitives
	TIMER
	OPTIONS

	// Keywords -- async
	AWAIT
	ALL
	ONE

	// Keywords -- blocks
	SWITCH
	CASE

	// Keywords -- control flow
	IF
	ELSE
	FOR
	IN

	// Keywords -- simple statements
	CLOSE
	COMPLETE
	FAIL
	RETURN
	CONTINUE_AS_NEW
	BREAK
	CONTINUE

	// Symbols
	COLON      // :
	ARROW      // ->
	LEFT_ARROW // <-
	DOT        // .

	// Literals
	NUMBER   // numeric literal (e.g. 3, 2.0)
	DURATION // numeric with duration suffix (e.g. 60s, 5m, 1h, 500ms)

	// Values
	IDENT    // non-keyword identifiers
	STRING   // quoted string
	ARGS     // raw content between ( and ), no nested parens
	COMMENT  // text after #
	RAW_TEXT // anything else

	tokenCount // sentinel: must be last — used for compile-time table size check
)

// Compile-time assertion: tokenTable has exactly tokenCount entries.
// If a new TokenType const is added without a table entry (or vice versa), this fails.
var _ [tokenCount]struct{} = [len(tokenTable)]struct{}{}

// tokenInfo describes a token type's display name and whether it is a keyword.
type tokenInfo struct {
	name      string
	isKeyword bool
}

// tokenTable is the single source of truth for token names and keyword status.
// Indexed by TokenType iota values.
var tokenTable = [...]tokenInfo{
	EOF:             {"EOF", false},
	ILLEGAL:         {"ILLEGAL", false},
	NEWLINE:         {"NEWLINE", false},
	INDENT:          {"INDENT", false},
	DEDENT:          {"DEDENT", false},
	WORKFLOW:        {"WORKFLOW", true},
	ACTIVITY:        {"ACTIVITY", true},
	WORKER:          {"WORKER", true},
	NAMESPACE:       {"NAMESPACE", true},
	TASK_QUEUE:      {"TASK_QUEUE", true},
	SIGNAL:          {"SIGNAL", true},
	QUERY:           {"QUERY", true},
	UPDATE:          {"UPDATE", true},
	DETACH:          {"DETACH", true},
	NEXUS:           {"NEXUS", true},
	SYNC:            {"SYNC", true},
	ASYNC:           {"ASYNC", true},
	PROMISE:         {"PROMISE", true},
	CONDITION:       {"CONDITION", true},
	SET:             {"SET", true},
	UNSET:           {"UNSET", true},
	STATE:           {"STATE", true},
	TIMER:           {"TIMER", true},
	OPTIONS:         {"OPTIONS", true},
	AWAIT:           {"AWAIT", true},
	ALL:             {"ALL", true},
	ONE:             {"ONE", true},
	SWITCH:          {"SWITCH", true},
	CASE:            {"CASE", true},
	IF:              {"IF", true},
	ELSE:            {"ELSE", true},
	FOR:             {"FOR", true},
	IN:              {"IN", true},
	CLOSE:           {"CLOSE", true},
	COMPLETE:        {"COMPLETE", true},
	FAIL:            {"FAIL", true},
	RETURN:          {"RETURN", true},
	CONTINUE_AS_NEW: {"CONTINUE_AS_NEW", true},
	BREAK:           {"BREAK", true},
	CONTINUE:        {"CONTINUE", true},
	COLON:           {"COLON", false},
	ARROW:           {"ARROW", false},
	LEFT_ARROW:      {"LEFT_ARROW", false},
	DOT:             {"DOT", false},
	NUMBER:          {"NUMBER", false},
	DURATION:        {"DURATION", false},
	IDENT:           {"IDENT", false},
	STRING:          {"STRING", false},
	ARGS:            {"ARGS", false},
	COMMENT:         {"COMMENT", false},
	RAW_TEXT:        {"RAW_TEXT", false},
}

// keywords maps keyword strings to their token types.
// Built from tokenTable in init().
var keywords map[string]TokenType

func init() {
	keywords = make(map[string]TokenType)
	for i, info := range tokenTable {
		if info.isKeyword {
			keywords[strings.ToLower(info.name)] = TokenType(i)
		}
	}
}

func (t TokenType) String() string {
	if int(t) >= 0 && int(t) < len(tokenTable) && tokenTable[t].name != "" {
		return tokenTable[t].name
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

// LookupIdent returns the TokenType for an identifier string.
// If the identifier is a keyword, the keyword token type is returned.
// Otherwise, IDENT is returned.
// Note: lookup is case-sensitive. Keywords are lowercase, so "Workflow" is
// treated as an IDENT, not a keyword. This is intentional — the DSL is
// case-sensitive.
func LookupIdent(ident string) TokenType {
	if tt, ok := keywords[ident]; ok {
		return tt
	}
	return IDENT
}
