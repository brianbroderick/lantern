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
	COMMENT
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

	LT  // <
	GT  // >
	GTE // >=
	LTE // <=

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
	JSONCONCAT        // ||

	OVERLAP // &&

	LPAREN   // (
	RPAREN   // )
	LBRACE   // {
	RBRACE   // }
	LBRACKET // [
	RBRACKET // ]

	// Keywords
	keywordBeg
	ALL
	AND
	AS
	ASC
	BETWEEN
	BREADTH
	BY
	CASE
	CONFLICT
	CROSS
	CUBE
	CURRENT
	DEFAULT
	DELETE
	DEPTH
	DESC
	DISTINCT
	DROP
	ELSE
	END
	EXCEPT
	EXCLUDE
	EXPONENTIATION
	FALSE
	FETCH
	FIRST
	FOLLOWING
	FOR
	FROM
	FULL
	GROUP
	GROUPING
	GROUPS
	HAVING
	ILIKE
	IN
	INNER
	INSERT
	INTERSECT
	INTO
	IS
	ISNULL
	JOIN
	LAST
	LATERAL
	LEFT
	LIKE
	LIMIT
	LOCAL
	LOCKED
	MATERIALIZED
	NATURAL
	NEXT
	NO
	NOT
	NOTNULL
	NOWAIT
	NULLS
	OF
	OFFSET
	ON
	ONLY
	OR
	ORDER
	ORDINALITY
	OTHERS
	OUTER
	OVER
	PARTITION
	PRECEDING
	RANGE
	RECURSIVE
	REPEATABLE
	RIGHT
	ROLLUP
	ROW
	ROWS
	SEARCH
	SELECT
	SESSION
	SET
	// SETS // Not a reserverd word in PG
	SHARE
	SIMILAR
	SKIP
	TABLE
	TABLESAMPLE
	THEN
	TIES
	TIME
	TO
	TRUE
	UNBOUNDED
	UNION
	UPDATE
	USING
	VALUES
	WHEN
	WHERE
	WINDOW
	WITH
	ZONE
	keywordEnd
)

// These are how a string is mapped to the token
var Tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	WS:      "WS",
	COMMENT: "COMMENT",
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

	LT:  "LT",  // <
	GT:  "GT",  // >
	GTE: "GTE", // >=
	LTE: "LTE", // <=

	EQ:     "EQ",     // ==
	NOT_EQ: "NOT_EQ", // != or <>

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
	JSONCONCAT:        "JSONCONCAT",

	OVERLAP: "OVERLAP", // && used with PG arrays

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
	ALL:            "ALL",
	AND:            "AND",
	AS:             "AS",
	ASC:            "ASC",
	BETWEEN:        "BETWEEN",
	BREADTH:        "BREADTH",
	BY:             "BY",
	CASE:           "CASE",
	CONFLICT:       "CONFLICT",
	CROSS:          "CROSS",
	CUBE:           "CUBE",
	CURRENT:        "CURRENT",
	DEFAULT:        "DEFAULT",
	DELETE:         "DELETE",
	DEPTH:          "DEPTH",
	DESC:           "DESC",
	DISTINCT:       "DISTINCT",
	DROP:           "DROP",
	ELSE:           "ELSE",
	END:            "END",
	EXCEPT:         "EXCEPT",
	EXCLUDE:        "EXCLUDE",
	EXPONENTIATION: "EXPONENTIATION",
	FALSE:          "FALSE",
	FETCH:          "FETCH",
	FIRST:          "FIRST",
	FOLLOWING:      "FOLLOWING",
	FOR:            "FOR",
	FROM:           "FROM",
	FULL:           "FULL",
	GROUP:          "GROUP",
	GROUPING:       "GROUPING",
	GROUPS:         "GROUPS",
	HAVING:         "HAVING",
	ILIKE:          "ILIKE",
	IN:             "IN",
	INNER:          "INNER",
	INSERT:         "INSERT",
	INTERSECT:      "INTERSECT",
	INTO:           "INTO",
	IS:             "IS",
	ISNULL:         "ISNULL",
	JOIN:           "JOIN",
	LAST:           "LAST",
	LATERAL:        "LATERAL",
	LEFT:           "LEFT",
	LIKE:           "LIKE",
	LIMIT:          "LIMIT",
	LOCAL:          "LOCAL",
	LOCKED:         "LOCKED",
	MATERIALIZED:   "MATERIALIZED",
	NATURAL:        "NATURAL",
	NEXT:           "NEXT",
	NO:             "NO",
	NOT:            "NOT",
	NOTNULL:        "NOTNULL",
	NOWAIT:         "NOWAIT",
	NULLS:          "NULLS",
	OF:             "OF",
	OFFSET:         "OFFSET",
	ON:             "ON",
	ONLY:           "ONLY",
	OR:             "OR",
	ORDER:          "ORDER",
	ORDINALITY:     "ORDINALITY",
	OTHERS:         "OTHERS",
	OUTER:          "OUTER",
	OVER:           "OVER",
	PARTITION:      "PARTITION",
	PRECEDING:      "PRECEDING",
	RANGE:          "RANGE",
	RECURSIVE:      "RECURSIVE",
	REPEATABLE:     "REPEATABLE",
	RIGHT:          "RIGHT",
	ROLLUP:         "ROLLUP",
	ROW:            "ROW",
	ROWS:           "ROWS",
	SEARCH:         "SEARCH",
	SELECT:         "SELECT",
	SESSION:        "SESSION",
	SET:            "SET",
	// SETS:           "SETS", // Not a reserverd word in PG
	SHARE:       "SHARE",
	SIMILAR:     "SIMILAR",
	SKIP:        "SKIP",
	TABLE:       "TABLE",
	TABLESAMPLE: "TABLESAMPLE",
	THEN:        "THEN",
	TIES:        "TIES",
	TIME:        "TIME",
	TO:          "TO",
	TRUE:        "TRUE",
	UNBOUNDED:   "UNBOUNDED",
	UNION:       "UNION",
	UPDATE:      "UPDATE",
	USING:       "USING",
	VALUES:      "VALUES",
	WHEN:        "WHEN",
	WHERE:       "WHERE",
	WINDOW:      "WINDOW",
	WITH:        "WITH",
	ZONE:        "ZONE",
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
