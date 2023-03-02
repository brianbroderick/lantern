package analyzer

import (
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func Analyze() {
	tree, err := pg_query.ParseToJSON("SELECT 42")
	if err != nil {
		panic(err)
	}

	// {"version":130008,"stmts":
	//  [{"stmt":{"SelectStmt":{"targetList":[{"ResTarget":{"val":{"A_Const":{"val":{"Integer":{"ival":42}},"location":7}},"location":7}}],"limitOption":"LIMIT_OPTION_DEFAULT","op":"SETOP_NONE"}}}]}

	fmt.Printf("%s\n", tree)
}
