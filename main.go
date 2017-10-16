package main

import (
	"fmt"
	"os"

	pg_query "github.com/lfittl/pg_query_go"
)

func main() {
	arg := os.Args[1]
	tree, err := pg_query.Normalize(arg)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	fmt.Printf("%s\n", tree)
}

func normalizeQuery(sql string) ([]byte, error) {
	tree, err := pg_query.Normalize(sql)
	if err != nil {
		return nil, err
	}

	return []byte(tree), err
}
