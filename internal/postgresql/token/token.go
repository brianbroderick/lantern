package token

import "strings"

type TokenType int

type Token struct {
	Type TokenType
	Lit  string
}

// PG Log entry:
// 2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from foo where bar = $1

const (
	ILLEGAL TokenType = iota
	EOF
	WS // whitespace
	NIL

	literalBeg // Literals
	IDENT      // identity: add, foobar, x, y, my_var, ...
	INT        // 12345
	NUMBER     // 12345.67
	STRING     // "foobar"
	DATE       // 2023-07-10
	TIME       // 09:52:46
	TIMEZONE   // MDT
	IPADDR     //	127.0.0.1
	QUERY      // select * from users
	literalEnd

	// Delimiters
	LPAREN   // (
	RPAREN   // )
	LBRACKET // [
	RBRACKET // ]
	DOT      // .
	COLON    // :
	ATSYMBOL // @
	GT       // >
	LT       // <

	// Keywords
	keywordBeg
	keywordEnd
)

// These are how a string is mapped to the token
var Tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	WS:      "WS",
	NIL:     "NIL",

	// Literals
	IDENT:    "IDENT",
	INT:      "INTEGER",
	NUMBER:   "NUMBER",
	STRING:   "STRING",
	DATE:     "DATE",
	TIME:     "TIME",
	TIMEZONE: "TIMEZONE",
	IPADDR:   "IPADDR",
	QUERY:    "QUERY",

	// Delimiters
	LPAREN:   "LPAREN '('",
	RPAREN:   "RPAREN ')'",
	LBRACKET: "LBRACKET '['",
	RBRACKET: "RBRACKET ']'",
	DOT:      "DOT '.'",
	COLON:    "COLON ':'",
	ATSYMBOL: "ATSYMBOL '@'",
	GT:       "GT '>'",
	LT:       "LT '<'",
}

var keywords map[string]TokenType

func init() {
	keywords = make(map[string]TokenType)
	for tok := keywordBeg + 1; tok < keywordEnd; tok++ {
		keywords[strings.ToLower(Tokens[tok])] = tok
	}
}

// String returns the string representation of the token.
func (tok TokenType) String() string {
	if tok >= 0 && tok < TokenType(len(Tokens)) {
		return Tokens[tok]
	}
	return ""
}

// Lookup returns the token associated with a given string.
func Lookup(ident string) TokenType {
	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}

	return IDENT
}
