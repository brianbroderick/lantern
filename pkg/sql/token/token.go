package token

import (
	"encoding/json"
	"strings"
)

type TokenType int

type Token struct {
	Type  TokenType `json:"type,omitempty"`
	Lit   string    `json:"literal,omitempty"`
	Upper string    `json:"upper,omitempty"`
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

	literalBeg   // Literals
	IDENT        // identity: add, foobar, x, y, my_var, ...
	INT          // 12345
	FLOAT        // 0.12345
	STRING       // "foobar"
	ESCAPESTRING // E'foobar'
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

	REGEXMATCH     // ~
	REGEXIMATCH    // ~*
	REGEXNOTMATCH  // !~
	REGEXNOTIMATCH // !~*

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
	ANALYZE
	AND
	ANY
	AT_TIME_ZONE // AT is not resevered in PG, but is in ANSI. We use AT all the time as table aliases, so we need to combine the phrase in the parser
	AS
	ASC
	ASYMMETRIC
	AUTHORIZATION
	BETWEEN
	BINARY
	BOTH
	BREADTH
	BY
	CASE
	CAST
	CHECK
	COLLATE
	COLLATION
	COLUMN
	COMMIT
	CONCURRENTLY
	CONFLICT
	CONSTRAINT
	CREATE
	CROSS
	CUBE
	CURRENT
	CURRENT_CATALOG
	CURRENT_DATE
	CURRENT_ROLE
	CURRENT_SCHEMA
	CURRENT_TIME
	CURRENT_TIMESTAMP
	CURRENT_USER
	DEFAULT
	DEFERRABLE
	DELETE
	DEPTH
	DESC
	DISTINCT
	DO
	DROP
	ELSE
	END
	EXCEPT
	EXCLUDE
	EXPONENTIATION
	FALSE
	FETCH
	FILTER
	FIRST
	FOLLOWING
	FOR
	FOREIGN
	FREEZE
	FROM
	FULL
	GRANT
	GROUP_BY
	GROUPING
	HAVING
	ILIKE
	IN
	INITIALLY
	INNER
	INSERT
	INTERSECT
	INTERVAL
	INTO
	IS
	ISNULL
	JOIN
	LAST
	LATERAL
	LEADING
	LEFT
	LIKE
	LIMIT
	LOCAL
	LOCALTIME
	LOCALTIMESTAMP
	LOCKED
	MATERIALIZED
	NATURAL
	NEXT
	NO
	NOT
	NOTNULL
	NOWAIT
	NULL
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
	OVERLAPS
	PARTITION
	PLACING
	PRECEDING
	PRIMARY
	// RANGE // Not a reserverd word in PG
	RECURSIVE
	REFERENCES
	REPEATABLE
	RETURNING
	RIGHT
	ROLLBACK
	ROLLUP
	ROW
	ROWS
	SEARCH
	SELECT
	SESSION
	SESSION_USER
	SET
	// SETS // Not a reserverd word in PG
	SHARE
	SIMILAR
	SKIP
	SOME
	SYMMETRIC
	SYSTEM_USER
	TABLE
	TABLESAMPLE
	THEN
	TIES
	TIMESTAMP
	TO
	TRAILING
	TRUE
	UNBOUNDED
	UNION
	UNIQUE
	UNKNOWN
	UPDATE
	USER
	USING
	VALUES
	VARIADIC
	VERBOSE
	WHEN
	WHERE
	WINDOW
	WITH
	keywordEnd
)

// These are how a string is mapped to the token
var Tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	WS:      "WS",
	COMMENT: "COMMENT",
	NIL:     "NIL",

	IDENT:        "IDENT",
	INT:          "INTEGER",
	FLOAT:        "FLOAT", // 0.12345 or 12345.12345
	STRING:       "STRING",
	ESCAPESTRING: "ESCAPESTRING", // E'foobar'

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

	REGEXMATCH:     "REGEXMATCH",     // ~
	REGEXIMATCH:    "REGEXIMATCH",    // ~*
	REGEXNOTMATCH:  "REGEXNOTMATCH",  // !~
	REGEXNOTIMATCH: "REGEXNOTIMATCH", // !~*

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
	ALL:               "ALL",
	ANALYZE:           "ANALYZE",
	AND:               "AND",
	ANY:               "ANY",
	AS:                "AS",
	ASC:               "ASC",
	ASYMMETRIC:        "ASYMMETRIC",
	AT_TIME_ZONE:      "AT_TIME_ZONE",
	AUTHORIZATION:     "AUTHORIZATION",
	BETWEEN:           "BETWEEN",
	BINARY:            "BINARY",
	BOTH:              "BOTH",
	BREADTH:           "BREADTH",
	BY:                "BY",
	CASE:              "CASE",
	CAST:              "CAST",
	CHECK:             "CHECK",
	COLLATE:           "COLLATE",
	COLLATION:         "COLLATION",
	COLUMN:            "COLUMN",
	COMMIT:            "COMMIT",
	CONCURRENTLY:      "CONCURRENTLY",
	CONFLICT:          "CONFLICT",
	CONSTRAINT:        "CONSTRAINT",
	CREATE:            "CREATE",
	CROSS:             "CROSS",
	CUBE:              "CUBE",
	CURRENT:           "CURRENT",
	CURRENT_CATALOG:   "CURRENT_CATALOG",
	CURRENT_DATE:      "CURRENT_DATE",
	CURRENT_ROLE:      "CURRENT_ROLE",
	CURRENT_SCHEMA:    "CURRENT_SCHEMA",
	CURRENT_TIME:      "CURRENT_TIME",
	CURRENT_TIMESTAMP: "CURRENT_TIMESTAMP",
	CURRENT_USER:      "CURRENT_USER",
	DEFAULT:           "DEFAULT",
	DEFERRABLE:        "DEFERRABLE",
	DELETE:            "DELETE",
	DEPTH:             "DEPTH",
	DESC:              "DESC",
	DISTINCT:          "DISTINCT",
	DO:                "DO",
	DROP:              "DROP",
	ELSE:              "ELSE",
	END:               "END",
	EXCEPT:            "EXCEPT",
	EXCLUDE:           "EXCLUDE",
	EXPONENTIATION:    "EXPONENTIATION",
	FALSE:             "FALSE",
	FETCH:             "FETCH",
	FILTER:            "FILTER",
	FIRST:             "FIRST",
	FOLLOWING:         "FOLLOWING",
	FOR:               "FOR",
	FOREIGN:           "FOREIGN",
	FREEZE:            "FREEZE",
	FROM:              "FROM",
	FULL:              "FULL",
	GRANT:             "GRANT",
	GROUP_BY:          "GROUP_BY",
	GROUPING:          "GROUPING",
	HAVING:            "HAVING",
	ILIKE:             "ILIKE",
	IN:                "IN",
	INITIALLY:         "INITIALLY",
	INNER:             "INNER",
	INSERT:            "INSERT",
	INTERSECT:         "INTERSECT",
	INTERVAL:          "INTERVAL",
	INTO:              "INTO",
	IS:                "IS",
	ISNULL:            "ISNULL",
	JOIN:              "JOIN",
	LAST:              "LAST",
	LATERAL:           "LATERAL",
	LEADING:           "LEADING",
	LEFT:              "LEFT",
	LIKE:              "LIKE",
	LIMIT:             "LIMIT",
	LOCAL:             "LOCAL",
	LOCALTIME:         "LOCALTIME",
	LOCALTIMESTAMP:    "LOCALTIMESTAMP",
	LOCKED:            "LOCKED",
	MATERIALIZED:      "MATERIALIZED",
	NATURAL:           "NATURAL",
	NEXT:              "NEXT",
	NO:                "NO",
	NOT:               "NOT",
	NOTNULL:           "NOTNULL",
	NOWAIT:            "NOWAIT",
	NULL:              "NULL",
	NULLS:             "NULLS",
	OF:                "OF",
	OFFSET:            "OFFSET",
	ON:                "ON",
	ONLY:              "ONLY",
	OR:                "OR",
	ORDER:             "ORDER",
	ORDINALITY:        "ORDINALITY",
	OTHERS:            "OTHERS",
	OUTER:             "OUTER",
	OVER:              "OVER",
	OVERLAPS:          "OVERLAPS",
	PARTITION:         "PARTITION",
	PLACING:           "PLACING",
	PRIMARY:           "PRIMARY",
	PRECEDING:         "PRECEDING",
	// RANGE:             "RANGE", // Not a reserverd word in PG
	RECURSIVE:    "RECURSIVE",
	REFERENCES:   "REFERENCES",
	REPEATABLE:   "REPEATABLE",
	RETURNING:    "RETURNING",
	RIGHT:        "RIGHT",
	ROLLBACK:     "ROLLBACK",
	ROLLUP:       "ROLLUP",
	ROW:          "ROW",
	ROWS:         "ROWS",
	SEARCH:       "SEARCH",
	SELECT:       "SELECT",
	SESSION:      "SESSION",
	SESSION_USER: "SESSION_USER",
	SET:          "SET",
	// SETS:           "SETS", // Not a reserverd word in PG
	SHARE:       "SHARE",
	SIMILAR:     "SIMILAR",
	SKIP:        "SKIP",
	SOME:        "SOME",
	SYMMETRIC:   "SYMMETRIC",
	SYSTEM_USER: "SYSTEM_USER",
	TABLE:       "TABLE",
	TABLESAMPLE: "TABLESAMPLE",
	THEN:        "THEN",
	TIES:        "TIES",
	TIMESTAMP:   "TIMESTAMP",
	TO:          "TO",
	TRAILING:    "TRAILING",
	TRUE:        "TRUE",
	UNBOUNDED:   "UNBOUNDED",
	UNION:       "UNION",
	UNIQUE:      "UNIQUE",
	UNKNOWN:     "UNKNOWN",
	UPDATE:      "UPDATE",
	USER:        "USER",
	USING:       "USING",
	VALUES:      "VALUES",
	VARIADIC:    "VARIADIC",
	VERBOSE:     "VERBOSE",
	WHEN:        "WHEN",
	WHERE:       "WHERE",
	WINDOW:      "WINDOW",
	WITH:        "WITH",
}

var keywords map[string]TokenType

func init() {
	keywords = make(map[string]TokenType)
	for tok := keywordBeg + 1; tok < keywordEnd; tok++ {
		keywords[strings.ToUpper(Tokens[tok])] = tok
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
	if tok, ok := keywords[strings.ToUpper(ident)]; ok {
		return tok
	}

	return IDENT
}

// LookupFromUpper assumes the ident passed is already upper case and returns the token associated with a given string.
func LookupFromUpper(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	return IDENT
}
