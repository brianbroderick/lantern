package analyzer

import (
	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func ParseToJSON(sql string) string {
	json, err := pg_query.ParseToJSON(sql)
	if err != nil {
		panic(err)
	}

	return json
}

func Parse(sql string) *pg_query.ParseResult {
	tree, err := pg_query.Parse(sql)
	if err != nil {
		panic(err)
	}

	// fmt.Printf("%+v\n", tree)
	return tree
}
