package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookup(t *testing.T) {
	// should find and return a token
	tok := Lookup("select")
	assert.Equal(t, SELECT, tok)

	// Not found, so it's an ident
	tok = Lookup("scooby")
	assert.Equal(t, IDENT, tok)
}

// func TestSort(t *testing.T) {
// 	list := []string{
// 		"AND",
// 		"OR",
// 		"IS",
// 		"ISNULL",
// 		"NOTNULL",
// 		"IN",
// 		"LIKE",
// 		"ILIKE",
// 		"SIMILAR",
// 		"EXPONENTIATION",
// 		"BETWEEN",
// 		"NOT",
// 		"WITH",
// 		"RECURSIVE",
// 		"SELECT",
// 		"ALL",
// 		"DISTINCT",
// 		"ON",
// 		"AS",
// 		"FROM",
// 		"WHERE",
// 		"GROUP",
// 		"BY",
// 		"HAVING",
// 		"OVER",
// 		"UNION",
// 		"WINDOW",
// 		"INTERSECT",
// 		"EXCEPT",
// 		"ORDER",
// 		"ASC",
// 		"DESC",
// 		"USING",
// 		"NULLS",
// 		"FIRST",
// 		"LAST",
// 		"LIMIT",
// 		"OFFSET",
// 		"ROW",
// 		"ROWS",
// 		"FETCH",
// 		"NEXT",
// 		"ONLY",
// 		"TIES",
// 		"FOR",
// 		"UPDATE",
// 		"NO",
// 		"SHARE",
// 		"OF",
// 		"NOWAIT",
// 		"SKIP",
// 		"LOCKED",
// 		"TABLESAMPLE",
// 		"REPEATABLE",
// 		"LATERAL",
// 		"ORDINALITY",
// 		"NATURAL",
// 		"CROSS",
// 		"ROLLUP",
// 		"CUBE",
// 		"GROUPING",
// 		"SETS",
// 		"MATERIALIZED",
// 		"SEARCH",
// 		"BREADTH",
// 		"DEPTH",
// 		"SET",
// 		"TO",
// 		"DEFAULT",
// 		"TABLE",
// 		"JOIN",
// 		"INNER",
// 		"LEFT",
// 		"OUTER",
// 		"FULL",
// 		"RIGHT",
// 		"PARTITION",
// 		"RANGE",
// 		"GROUPS",
// 		"UNBOUNDED",
// 		"PRECEDING",
// 		"CURRENT",
// 		"FOLLOWING",
// 		"EXCLUDE",
// 		"OTHERS",
// 		"INSERT",
// 		"DELETE",
// 		"INTO",
// 		"VALUES",
// 		"CONFLICT",
// 		"TRUE",
// 		"FALSE",
// 		"SESSION",
// 		"LOCAL",
// 		"TIME",
// 		"ZONE",
// 	}

// 	sort.Strings(list)
// 	for _, v := range list {
// 		fmt.Printf("%s: \"%s\",\n", v, v)
// 	}
// }
