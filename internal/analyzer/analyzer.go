package analyzer

import (
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

type Select struct {
	Columns []string
	Tables  []string
}

func ParseToJSON(sql string) string {
	json, err := pg_query.ParseToJSON(sql)
	if err != nil {
		panic(err)
	}

	return json
}

func Walk(sql string) []*Select {
	ast := Parse(sql)

	return EnumStatements(ast)
}

func Parse(sql string) *pg_query.ParseResult {
	tree, err := pg_query.Parse(sql)
	if err != nil {
		// TODO: better error handling
		panic(err)
	}

	return tree
}

func EnumStatements(ast *pg_query.ParseResult) []*Select {
	var selects = make([]*Select, 0)
	for _, stmt := range ast.Stmts {
		s := new(Select)
		s.Columns = TargetList(stmt) // fmt.Printf("stmt: %+v\n", s.Stmt)
		s.Tables = TableList(stmt)   // fmt.Printf("location: %+v\n", s.StmtLocation)
		selects = append(selects, s)
	}
	return selects
}

func TargetList(ast *pg_query.RawStmt) []string {
	cols := make([]string, 0)
	for _, col := range ast.Stmt.GetSelectStmt().GetTargetList() {
		for _, str := range col.GetResTarget().GetVal().GetColumnRef().GetFields() {
			cols = append(cols, str.GetString_().Sval)
		}
	}
	return cols
}

func TableList(ast *pg_query.RawStmt) []string {
	tables := make([]string, 0)
	for _, str := range ast.Stmt.GetSelectStmt().GetFromClause() {
		fmt.Printf("str: %+v\n", str)
		tables = append(tables, str.GetRangeVar().Relname)
	}
	return tables
}
