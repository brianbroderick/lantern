package extractor

import (
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/object"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type Extractor struct {
	Ast                 *ast.Statement
	ColumnsInQueries    map[string]*ColumnsInQueries    `json:"columns_in_queries,omitempty"`
	TablesInQueries     map[string]*TablesInQueries     `json:"tables_in_queries,omitempty"`
	TableJoinsInQueries map[string]*TableJoinsInQueries `json:"table_joins_in_queries,omitempty"`
	FunctionsInQueries  map[string]*FunctionsInQueries  `json:"functions_in_queries,omitempty"`
	Tables              map[string]*Tables              `json:"tables,omitempty"`
	MustExtract         bool
	errors              []string
}

func NewExtractor(stmt *ast.Statement, mustExtract bool) *Extractor {
	return &Extractor{
		Ast:                 stmt,
		ColumnsInQueries:    make(map[string]*ColumnsInQueries),
		TablesInQueries:     make(map[string]*TablesInQueries),
		TableJoinsInQueries: make(map[string]*TableJoinsInQueries),
		FunctionsInQueries:  make(map[string]*FunctionsInQueries),
		Tables:              make(map[string]*Tables),
		errors:              []string{},
		MustExtract:         mustExtract,
	}
}

func (r *Extractor) Execute(node ast.Node) {
	env := object.NewEnvironment()
	r.Extract(node, env)
	r.InferColumnsInTables()
}

// InferColumnsInTables will set the table if there is only one table in the query
func (r *Extractor) InferColumnsInTables() {
	// If there are more than one table in the query, there's no way to map a column directly to a table
	if len(r.TablesInQueries) != 1 {
		return
	}

	var table *TablesInQueries
	for _, t := range r.TablesInQueries {
		table = t
	}

	newColumnsInQueries := make(map[string]*ColumnsInQueries)

	for _, column := range r.ColumnsInQueries {
		column.Table = table.Name
		fqcn := fmt.Sprintf("%s|%s.%s.%s", column.Clause, column.Schema, column.Table, column.Name)

		newColumnsInQueries[fqcn] = &ColumnsInQueries{
			UID:       UuidV5(fqcn),
			Schema:    column.Schema,
			Table:     column.Table,
			TableUID:  UuidV5(fmt.Sprintf("%s.%s", column.Schema, column.Table)),
			Name:      column.Name,
			ColumnUID: UuidV5(fmt.Sprintf("%s.%s.%s", column.Schema, column.Table, column.Name)), // don't include the clause in the column UID
			Clause:    column.Clause,
		}
	}

	r.ColumnsInQueries = newColumnsInQueries
}

func (r *Extractor) Extract(node ast.Node, env *object.Environment) {
	switch node := node.(type) {
	case *ast.Program:
		r.extractProgram(node, env)
	case *ast.SelectStatement:
		r.extractSelectStatement(node, env)
	case *ast.CreateStatement:
		r.Extract(node.Expression, env)
	case *ast.CTEStatement:
		r.Extract(node.Expression, env)
	case *ast.InsertStatement:
		r.Extract(node.Expression, env)
	case *ast.UpdateStatement:
		r.Extract(node.Expression, env)
	case *ast.DeleteStatement:
		r.Extract(node.Expression, env)

	// Expressions
	case *ast.CTEExpression:
		r.Extract(node.Primary, env)
		for _, a := range node.Auxiliary {
			r.Extract(a, env)
		}
	case *ast.ExpressionStatement:
		r.Extract(node.Expression, env)
	case *ast.SelectExpression:
		envSE := object.NewEnvironment()
		setTableAliases(envSE, node.TableAliases)
		r.extractSelectExpression(node, envSE)
	case *ast.PrefixExpression:
		r.Extract(node.Right, env)
	case *ast.PrefixKeywordExpression:
		r.Extract(node.Right, env)
	case *ast.InfixExpression:
		r.Extract(node.Left, env)
		r.Extract(node.Right, env)
		// TODO: This is a hack to get the join condition, but it only works for simple cases
		// We might be better off compiling to a stack and then popping off the stack to evaluate the expression
		if node.Clause() == token.ON && r.MustExtract {
			r.extractOnExpression(*node, env)
		}
	case *ast.GroupedExpression:
		for _, e := range node.Elements {
			r.Extract(e, env)
		}
	case *ast.CallExpression:
		for _, a := range node.Arguments {
			r.Extract(a, env)
		}
		r.Extract(node.Distinct, env)
		r.Extract(node.Function, env)
	case *ast.ArrayLiteral:
		for _, e := range node.Elements {
			r.Extract(e, env)
		}
		// No need to extract left because it'll never be a column
		// r.Extract(node.Left, env)
	case *ast.IndexExpression:
		r.Extract(node.Index, env)
		// shouldn't need to extract left because it shouldn't be a column
		// r.Extract(node.Left, env)
	case *ast.IntervalExpression:
		r.Extract(node.Value, env)
	case *ast.CaseExpression:
		r.Extract(node.Expression, env)
		for _, c := range node.Conditions {
			r.Extract(c.Condition, env)
			r.Extract(c.Consequence, env)
		}
		r.Extract(node.Alternative, env)

	// Select Expressions
	case *ast.DistinctExpression:
		for _, e := range node.Right {
			r.Extract(e, env)
		}
	case *ast.ColumnExpression:
		r.extractColumnExpression(node, env)
	case *ast.SortExpression:
		r.Extract(node.Value, env)
	case *ast.FetchExpression:
		r.Extract(node.Value, env)
	case *ast.AggregateExpression:
		r.Extract(node.Left, env)
		for _, a := range node.Right {
			r.Extract(a, env)
		}
	case *ast.WindowExpression:
		r.Extract(node.Alias, env)
		for _, p := range node.PartitionBy {
			r.Extract(p, env)
		}
		for _, o := range node.OrderBy {
			r.Extract(o, env)
		}
	case *ast.TableExpression:
		// remove the table alias from the environment
		switch node.Table.(type) {
		case *ast.Identifier, *ast.SimpleIdentifier:
			node.Alias = nil
		}
		// set jointype as a variable in the env
		setJoinType(env, node.JoinType)
		r.Extract(node.JoinCondition, env)

		r.Extract(node.Table, env)
	case *ast.LockExpression:
		for _, t := range node.Tables {
			r.Extract(t, env)
		}
	case *ast.InExpression:
		r.Extract(node.Left, env)
		for _, e := range node.Right {
			r.Extract(e, env)
		}
	case *ast.CastExpression:
		r.Extract(node.Left, env)
	case *ast.WhereExpression:
		r.Extract(node.Right, env)
	case *ast.IsExpression:
		r.Extract(node.Left, env)
		r.Extract(node.Right, env)
	case *ast.TrimExpression:
		r.Extract(node.Expression, env)
	case *ast.StringFunctionExpression:
		r.Extract(node.Left, env)
		r.Extract(node.From, env)
		r.Extract(node.For, env)
	case *ast.UnionExpression:
		r.Extract(node.Left, env)
		r.Extract(node.Right, env)
	case *ast.InsertExpression:
		switch node.Table.(type) {
		// case *ast.Identifier, *ast.SimpleIdentifier:
		// we used to call out SimpleIdentifier here, but it would break the AddTablesInQueries function
		// Not sure why we were doing that, so there might be a bug here
		case *ast.Identifier:
			r.AddTablesInQueries(node.Table.(*ast.Identifier))
		}
		// The query in an insert statement is when we're inserting a select statement
		if node.Query != nil {
			r.Extract(node.Query, env)
		}
	case *ast.UpdateExpression:
		switch n := node.Table.(type) {
		case *ast.Identifier:
			r.AddTablesInQueries(n)
		}

		if node.Set != nil {
			for _, s := range node.Set {
				r.Extract(s, env)
			}
		}

		if len(node.From) > 0 {
			for _, f := range node.From {
				switch f := f.(type) {
				case *ast.Identifier:
					r.AddTablesInQueries(f)
				}
			}
		}

		if node.Where != nil {
			r.Extract(node.Where, env)
		}

	// Primitive Expressions
	case *ast.Identifier:
		r.extractIdentifier(node, env)

		// Noops
	case nil, *ast.AnalyzeStatement, *ast.DropStatement, *ast.SetStatement,
		*ast.ValuesExpression,
		*ast.WildcardLiteral, *ast.Boolean, *ast.Null,
		*ast.Unknown, *ast.Infinity, *ast.IllegalExpression,
		*ast.SimpleIdentifier, *ast.IntegerLiteral, *ast.FloatLiteral,
		*ast.StringLiteral, *ast.EscapeStringLiteral,
		*ast.TimestampExpression, *ast.KeywordExpression:
		// Do nothing

	default:
		r.newError("unknown node type: %T", node)
	}
}

func (r *Extractor) extractProgram(program *ast.Program, env *object.Environment) {
	for _, s := range program.Statements {
		r.Extract(s, env)
	}
}

func (r *Extractor) extractSelectStatement(s *ast.SelectStatement, env *object.Environment) {
	for _, x := range s.Expressions {
		r.Extract(x, env)
	}
}

func (r *Extractor) extractSelectExpression(x *ast.SelectExpression, env *object.Environment) {
	r.Extract(x.Distinct, env)
	for _, c := range x.Columns {
		r.Extract(c, env)
	}
	for _, t := range x.Tables {
		// switch table := t.(type) {
		// case *ast.TableExpression:
		// 	r.extractTableExpression(table, env)
		// }
		r.Extract(t, env)
	}
	r.Extract(x.Where, env)
	for _, g := range x.GroupBy {
		r.Extract(g, env)
	}
	r.Extract(x.Having, env)
	for _, w := range x.Window {
		r.Extract(w, env)
	}
	for _, o := range x.OrderBy {
		r.Extract(o, env)
	}
	r.Extract(x.Limit, env)
	r.Extract(x.Offset, env)
	r.Extract(x.Fetch, env)
	r.Extract(x.Lock, env)
}

// TODO: This only handles simple cases. We need to handle more complex cases
func (r *Extractor) extractOnExpression(node ast.InfixExpression, env *object.Environment) {
	var (
		left, right   *ast.Identifier
		shouldProcess int
	)
	switch node.Left.(type) {
	case *ast.Identifier:
		left = node.Left.(*ast.Identifier)
		shouldProcess++
	}
	switch node.Right.(type) {
	case *ast.Identifier:
		right = node.Right.(*ast.Identifier)
		shouldProcess++
	}

	if shouldProcess == 2 {
		r.AddJoinInQuery(left, right, node.String(false), env)
	}
}

func (r *Extractor) extractColumnExpression(c *ast.ColumnExpression, env *object.Environment) {
	r.Extract(c.Value, env)
}

func (r *Extractor) extractIdentifier(i *ast.Identifier, env *object.Environment) {
	// unalias
	aliases, ok := env.Get("table_aliases")
	if !ok {
		return
	}

	switch len(i.Value) {
	case 2:
		alias := i.Value[0].(*ast.SimpleIdentifier).Value
		if val, ok := aliases.(*object.StringHash).Value[alias]; ok {
			i.Value[0].(*ast.SimpleIdentifier).Value = val
		}
	case 3:
		alias := i.Value[1].(*ast.SimpleIdentifier).Value
		if val, ok := aliases.(*object.StringHash).Value[alias]; ok {
			i.Value[1].(*ast.SimpleIdentifier).Value = val
		}
	}

	// Extract based on the clause we're in
	if r.MustExtract {
		switch i.Clause() {
		case token.SELECT, token.WHERE, token.GROUP_BY, token.HAVING, token.ORDER: // These columns are what are selected (select id...)
			r.AddColumnsInQueries(i)
		case token.FROM: // The FROM clause will have tables
			r.AddTablesInQueries(i)
		case token.FUNCTION_CALL:
			r.AddFunctionsInQueries(i)
		}
	}
}

func (r *Extractor) newError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	r.errors = append(r.errors, msg)
}

func (r *Extractor) Errors() []string {
	return r.errors
}

func (r *Extractor) PrintErrors() {
	if len(r.errors) == 0 {
		return
	}
	for _, msg := range r.errors {
		fmt.Printf("Extractor error: %s\n", msg)
	}
}
