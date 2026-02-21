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
)

var tokenNames = map[TokenType]string{
	EOF:             "EOF",
	NEWLINE:         "NEWLINE",
	INDENT:          "INDENT",
	DEDENT:          "DEDENT",
	WORKFLOW:        "WORKFLOW",
	ACTIVITY:        "ACTIVITY",
	WORKER:          "WORKER",
	NAMESPACE:       "NAMESPACE",
	TASK_QUEUE:      "TASK_QUEUE",
	SIGNAL:          "SIGNAL",
	QUERY:           "QUERY",
	UPDATE:          "UPDATE",
	DETACH:          "DETACH",
	NEXUS:           "NEXUS",
	SYNC:            "SYNC",
	ASYNC:           "ASYNC",
	PROMISE:         "PROMISE",
	CONDITION:       "CONDITION",
	SET:             "SET",
	UNSET:           "UNSET",
	STATE:           "STATE",
	TIMER:           "TIMER",
	OPTIONS:         "OPTIONS",
	AWAIT:           "AWAIT",
	ALL:             "ALL",
	ONE:             "ONE",
	SWITCH:          "SWITCH",
	CASE:            "CASE",
	IF:              "IF",
	ELSE:            "ELSE",
	FOR:             "FOR",
	IN:              "IN",
	CLOSE:           "CLOSE",
	COMPLETE:        "COMPLETE",
	FAIL:            "FAIL",
	RETURN:          "RETURN",
	CONTINUE_AS_NEW: "CONTINUE_AS_NEW",
	BREAK:           "BREAK",
	CONTINUE:        "CONTINUE",
	COLON:           "COLON",
	ARROW:           "ARROW",
	LEFT_ARROW:      "LEFT_ARROW",
	DOT:             "DOT",
	NUMBER:          "NUMBER",
	DURATION:        "DURATION",
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
	"worker":          WORKER,
	"namespace":       NAMESPACE,
	"task_queue":      TASK_QUEUE,
	"signal":          SIGNAL,
	"query":           QUERY,
	"update":          UPDATE,
	"detach":          DETACH,
	"nexus":           NEXUS,
	"sync":            SYNC,
	"async":           ASYNC,
	"promise":         PROMISE,
	"condition":       CONDITION,
	"set":             SET,
	"unset":           UNSET,
	"state":           STATE,
	"timer":           TIMER,
	"options":         OPTIONS,
	"await":           AWAIT,
	"all":             ALL,
	"one":             ONE,
	"switch":          SWITCH,
	"case":            CASE,
	"if":              IF,
	"else":            ELSE,
	"for":             FOR,
	"in":              IN,
	"close":           CLOSE,
	"complete":        COMPLETE,
	"fail":            FAIL,
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
