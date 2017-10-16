package main

import pg_query "github.com/lfittl/pg_query_go"

// NormalizeQuery converts "select * from users where id = 1" to "select * from users where id = ?"
func NormalizeQuery(sql string) ([]byte, error) {
	tree, err := pg_query.Normalize(sql)
	if err != nil {
		return nil, err
	}

	return []byte(tree), err
}
