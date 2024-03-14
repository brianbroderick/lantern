package extractor

import (
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/object"
)

type Extractor struct {
	Ast             *ast.Program
	Tables          map[string]*Table        `json:"tables,omitempty"`
	TablesinQueries map[string]*TableInQuery `json:"tables_in_queries,omitempty"`
	TableJoins      map[string]*TableJoin    `json:"table_joins,omitempty"`
	errors          []string
}

func NewExtractor(program *ast.Program) *Extractor {
	return &Extractor{
		Ast:             program,
		Tables:          make(map[string]*Table),
		TablesinQueries: make(map[string]*TableInQuery),
		TableJoins:      make(map[string]*TableJoin),
		errors:          []string{},
	}
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
	case *ast.InsertExpression:
		r.Extract(node.Query, env)
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

	// Primitive Expressions
	case *ast.Identifier:
		r.extractIdentifier(node, env)

	// Noops
	case *ast.UpdateExpression:
		// Currently do nothing till we verify that we don't have aliases to extract

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

func (r *Extractor) extractColumnExpression(c *ast.ColumnExpression, env *object.Environment) {
	r.Extract(c.Value, env)
}

func (r *Extractor) extractIdentifier(i *ast.Identifier, env *object.Environment) {
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
