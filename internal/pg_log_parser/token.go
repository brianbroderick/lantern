package pgLogParser

import "strings"

// Sample log line:
// 2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from users

// This comes from PostgreSQL's postgresql.conf file. It is the standard prefix used by RDS
// log_line_prefix = '%t:%r:%u@%d:[%p]:'   # special values:
// %t = timestamp without milliseconds: 2023-07-10 09:52:46 MDT
// %r = remote host and port: 127.0.0.1(50032)
// %u = user name: postgres
// %d = database name: sampledb
// %p = process ID: 24649

type Token int

const (
	ILLEGAL Token = iota
	EOF
	WS // whitespace
	NIL

	// Literals
	// literalBeg // This may be used if we want to do something with the literal
	IDENT     // main
	NUMBER    // 12345.67
	STRING    // "abc"
	INTEGER   // 12345
	BADSTRING // "abc
	BADESCAPE // \qwe
	TIMESTAMP // 2023-07-10 09:52:46 MDT
	DATE      // 2023-07-10
	TIME      // 09:52:46
	TIMEZONE  // MDT
	IPADDR    // 127.0.0.1
	// literalEnd

	// Misc characters
	LPAREN     // (
	RPAREN     // )
	LBRACKET   // [
	RBRACKET   // ]
	COMMA      // ,
	SEMI       // ;
	COLON      // :
	APOSTROPHE // '
	DOT        // .
	ATSYMBOL   // @
	LT         // <
	GT         // >

	// placeholder for keywords, when we need some
	keywordBeg
	keywordEnd
)

// These are how a string is mapped to the token
var Tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	WS:      "WS",
	NIL:     "NIL",

	IDENT:     "IDENT",
	NUMBER:    "NUMBER",
	STRING:    "STRING",
	INTEGER:   "INTEGER",
	BADSTRING: "BADSTRING",
	BADESCAPE: "BADESCAPE",
	TIMESTAMP: "TIMESTAMP",
	DATE:      "DATE",
	TIME:      "TIME",
	TIMEZONE:  "TIMEZONE",
	IPADDR:    "IPADDR",

	LPAREN:     "(",
	RPAREN:     ")",
	LBRACKET:   "[",
	RBRACKET:   "]",
	COMMA:      ",",
	SEMI:       ";",
	COLON:      ":",
	APOSTROPHE: "'",
	DOT:        ".",
	ATSYMBOL:   "@",
	LT:         "<",
	GT:         ">",
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for tok := keywordBeg + 1; tok < keywordEnd; tok++ {
		keywords[strings.ToLower(Tokens[tok])] = tok
	}

}

// String returns the string representation of the token.
func (tok Token) String() string {
	if tok >= 0 && tok < Token(len(Tokens)) {
		return Tokens[tok]
	}
	return ""
}

// tokstr returns a literal if provided, otherwise returns the token string.
func tokstr(tok Token, lit string) string {
	if lit != "" {
		return lit
	}
	return tok.String()
}

// Lookup returns the token associated with a given string.
func Lookup(ident string) Token {
	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}

	return IDENT
}

// Pos specifies the line and character position of a token.
// The Char and Line are both zero-based indexes.
type Pos struct {
	Line int
	Char int
}
