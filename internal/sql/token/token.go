package token

import (
	"encoding/json"
	"strings"
)

type TokenType int

type Token struct {
	Type TokenType `json:"type,omitempty"`
	Lit  string    `json:"literal,omitempty"`
}

func (t *Token) MarshalJSON() ([]byte, error) {
	type Alias Token
	return json.Marshal(&struct {
		Type string `json:"type,omitempty"`
		*Alias
	}{
		Type:  t.Type.String(),
		Alias: (*Alias)(t),
	})
}

const (
	ILLEGAL TokenType = iota
	EOF
	WS // whitespace
	NIL

	literalBeg // Literals
	IDENT      // identity: add, foobar, x, y, my_var, ...
	INT        // 12345
	NUMBER     // 0.12345
	STRING     // "foobar"
	literalEnd

	// Operators
	ASSIGN   // =
	PLUS     // +
	MINUS    // -
	BANG     // !
	ASTERISK // *
	SLASH    // /

	LT // <
	GT // >

	EQ     // ==
	NOT_EQ // != or <>

	// Delimiters
	COMMA       // ,
	SEMICOLON   // ;
	COLON       // :
	DOUBLECOLON // ::
	DOT         // .

	JSONGETBYKEY      // ->
	JSONGETBYTEXT     // ->>
	JSONGETBYPATH     // #>
	JSONGETBYPATHTEXT // #>>
	JSONCONTAINS      // @>
	JSONCONTAINED     // <@
	JSONHASKEY        // ?
	JSONHASALLKEYS    // ?&
	JSONHASANYKEYS    // ?|
	JSONDELETE        // #-

	LPAREN   // (
	RPAREN   // )
	LBRACE   // {
	RBRACE   // }
	LBRACKET // [
	RBRACKET // ]

	// Keywords
	keywordBeg
	SYMBOLAND
	SYMBOLOR
	AND
	OR
	IS
	ISNULL
	NOTNULL
	IN
	LIKE
	ILIKE
	SIMILAR
	EXPONENTIATION
	BETWEEN
	NOT

	WITH
	RECURSIVE
	SELECT
	ALL
	DISTINCT
	ON
	AS
	FROM
	WHERE
	GROUP
	BY
	HAVING
	OVER
	UNION
	WINDOW
	INTERSECT
	EXCEPT
	ORDER
	ASC
	DESC
	USING
	NULLS
	FIRST
	LAST
	LIMIT
	OFFSET
	ROW
	ROWS
	FETCH
	NEXT
	ONLY
	TIES
	FOR
	UPDATE
	NO
	KEY
	SHARE
	OF
	NOWAIT
	SKIP
	LOCKED
	TABLESAMPLE
	REPEATABLE
	LATERAL
	ORDINALITY
	NATURAL
	CROSS
	ROLLUP
	CUBE
	GROUPING
	SETS
	MATERIALIZED
	SEARCH
	BREADTH
	DEPTH
	SET
	TO
	DEFAULT
	TABLE
	JOIN
	INNER
	LEFT
	OUTER
	FULL
	RIGHT
	PARTITION
	RANGE
	GROUPS
	UNBOUNDED
	PRECEDING
	CURRENT
	FOLLOWING
	EXCLUDE
	OTHERS
	INSERT
	DELETE
	INTO
	VALUES
	CONFLICT
	TRUE
	FALSE
	keywordEnd
)

// These are how a string is mapped to the token
var Tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	WS:      "WS",
	NIL:     "NIL",

	IDENT:  "IDENT",
	INT:    "INTEGER",
	NUMBER: "NUMBER",
	STRING: "STRING",

	ASSIGN:   "ASSIGN",   // =
	PLUS:     "PLUS",     // +
	MINUS:    "MINUS",    // -
	BANG:     "BANG",     // !
	ASTERISK: "ASTERISK", // *
	SLASH:    "SLASH",    // /

	LT: "LT", // <
	GT: "GT", // >

	EQ:     "EQ",     // ==
	NOT_EQ: "NOT_EQ", // != or <>

	AND:            "AND",
	OR:             "OR",
	SYMBOLAND:      "&&",
	SYMBOLOR:       "||",
	IS:             "IS",
	ISNULL:         "ISNULL",
	NOTNULL:        "NOTNULL",
	IN:             "IN",
	LIKE:           "LIKE",
	ILIKE:          "ILIKE",
	SIMILAR:        "SIMILAR",
	EXPONENTIATION: "EXPONENTIATION",
	BETWEEN:        "BETWEEN",

	JSONGETBYKEY:      "JSONGETBYKEY",
	JSONGETBYTEXT:     "JSONGETBYTEXT",
	JSONGETBYPATH:     "JSONGETBYPATH",
	JSONGETBYPATHTEXT: "JSONGETBYPATHTEXT",
	JSONCONTAINS:      "JSONCONTAINS",
	JSONCONTAINED:     "JSONCONTAINED",
	JSONHASKEY:        "JSONHASKEY",
	JSONHASALLKEYS:    "JSONHASALLKEYS",
	JSONHASANYKEYS:    "JSONHASANYKEYS",
	JSONDELETE:        "JSONDELETE",

	// Delimiters
	COMMA:       "COMMA",       // ,
	SEMICOLON:   "SEMICOLON",   // ;
	COLON:       "COLON",       // :
	DOUBLECOLON: "DOUBLECOLON", // ::
	DOT:         "DOT",         // .

	LPAREN:   "LPAREN",   // (
	RPAREN:   "RPAREN",   // )
	LBRACE:   "LBRACE",   // {
	RBRACE:   "RBRACE",   // }
	LBRACKET: "LBRACKET", // [
	RBRACKET: "RBRACKET", // ]

	// Keywords
	WITH:         "WITH",
	RECURSIVE:    "RECURSIVE",
	SELECT:       "SELECT",
	ALL:          "ALL",
	DISTINCT:     "DISTINCT",
	ON:           "ON",
	AS:           "AS",
	FROM:         "FROM",
	WHERE:        "WHERE",
	GROUP:        "GROUP",
	BY:           "BY",
	HAVING:       "HAVING",
	OVER:         "OVER",
	UNION:        "UNION",
	WINDOW:       "WINDOW",
	INTERSECT:    "INTERSECT",
	EXCEPT:       "EXCEPT",
	ORDER:        "ORDER",
	ASC:          "ASC",
	DESC:         "DESC",
	USING:        "USING",
	NULLS:        "NULLS",
	FIRST:        "FIRST",
	LAST:         "LAST",
	LIMIT:        "LIMIT",
	OFFSET:       "OFFSET",
	ROW:          "ROW",
	ROWS:         "ROWS",
	FETCH:        "FETCH",
	NEXT:         "NEXT",
	ONLY:         "ONLY",
	TIES:         "TIES",
	FOR:          "FOR",
	UPDATE:       "UPDATE",
	NO:           "NO",
	KEY:          "KEY",
	SHARE:        "SHARE",
	OF:           "OF",
	NOWAIT:       "NOWAIT",
	SKIP:         "SKIP",
	LOCKED:       "LOCKED",
	TABLESAMPLE:  "TABLESAMPLE",
	REPEATABLE:   "REPEATABLE",
	LATERAL:      "LATERAL",
	ORDINALITY:   "ORDINALITY",
	NATURAL:      "NATURAL",
	CROSS:        "CROSS",
	ROLLUP:       "ROLLUP",
	CUBE:         "CUBE",
	GROUPING:     "GROUPING",
	SETS:         "SETS",
	NOT:          "NOT",
	MATERIALIZED: "MATERIALIZED",
	SEARCH:       "SEARCH",
	BREADTH:      "BREADTH",
	DEPTH:        "DEPTH",
	SET:          "SET",
	TO:           "TO",
	DEFAULT:      "DEFAULT",
	TABLE:        "TABLE",
	JOIN:         "JOIN",
	INNER:        "INNER",
	LEFT:         "LEFT",
	OUTER:        "OUTER",
	FULL:         "FULL",
	RIGHT:        "RIGHT",
	PARTITION:    "PARTITION",
	RANGE:        "RANGE",
	GROUPS:       "GROUPS",
	UNBOUNDED:    "UNBOUNDED",
	PRECEDING:    "PRECEDING",
	CURRENT:      "CURRENT",
	FOLLOWING:    "FOLLOWING",
	EXCLUDE:      "EXCLUDE",
	OTHERS:       "OTHERS",
	INSERT:       "INSERT",
	DELETE:       "DELETE",
	INTO:         "INTO",
	VALUES:       "VALUES",
	CONFLICT:     "CONFLICT",
	TRUE:         "TRUE",
	FALSE:        "FALSE",
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
