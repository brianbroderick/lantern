package main

import pg_query "github.com/lfittl/pg_query_go"

// normalizeQuery converts "select * from users where id = 1" to "select * from users where id = ?"
func normalizeQuery(sql string) ([]byte, error) {
	tree, err := pg_query.Normalize(sql)
	if err != nil {
		return nil, err
	}

	return []byte(tree), err
}
