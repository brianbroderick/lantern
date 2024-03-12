package resolver

import (
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/object"
)

type Resolver struct {
	Ast    *ast.Program
	errors []string
}

func NewResolver(program *ast.Program) *Resolver {
	return &Resolver{
		Ast:    program,
		errors: []string{},
	}
}

func (r *Resolver) Resolve(node ast.Node, env *object.Environment) {
	switch node := node.(type) {
	case *ast.Program:
		r.resolveProgram(node, env)
	case *ast.SelectStatement:
		r.resolveSelectStatement(node, env)
	case *ast.CreateStatement:
		r.Resolve(node.Expression, env)
	case *ast.CTEStatement:
		r.Resolve(node.Expression, env)
	case *ast.InsertStatement:
		r.Resolve(node.Expression, env)
	case *ast.UpdateStatement:
		r.Resolve(node.Expression, env)

	// Expressions
	case *ast.CTEExpression:
		r.Resolve(node.Primary, env)
		for _, a := range node.Auxiliary {
			r.Resolve(a, env)
		}
	case *ast.InsertExpression:
		r.Resolve(node.Query, env)
	case *ast.ExpressionStatement:
		r.Resolve(node.Expression, env)
	case *ast.SelectExpression:
		envSE := object.NewEnvironment()
		setTableAliases(envSE, node.TableAliases)
		r.resolveSelectExpression(node, envSE)
	case *ast.PrefixExpression:
		r.Resolve(node.Right, env)
	case *ast.PrefixKeywordExpression:
		r.Resolve(node.Right, env)
	case *ast.InfixExpression:
		r.Resolve(node.Left, env)
		r.Resolve(node.Right, env)
	case *ast.GroupedExpression:
		for _, e := range node.Elements {
			r.Resolve(e, env)
		}
	case *ast.CallExpression:
		for _, a := range node.Arguments {
			r.Resolve(a, env)
		}
		r.Resolve(node.Distinct, env)
		r.Resolve(node.Function, env)
	case *ast.ArrayLiteral:
		for _, e := range node.Elements {
			r.Resolve(e, env)
		}
		// No need to resolve left because it'll never be a column
		// r.Resolve(node.Left, env)
	case *ast.IndexExpression:
		r.Resolve(node.Index, env)
		// shouldn't need to resolve left because it shouldn't be a column
		// r.Resolve(node.Left, env)
	case *ast.IntervalExpression:
		r.Resolve(node.Value, env)
	case *ast.CaseExpression:
		r.Resolve(node.Expression, env)
		for _, c := range node.Conditions {
			r.Resolve(c.Condition, env)
			r.Resolve(c.Consequence, env)
		}
		r.Resolve(node.Alternative, env)

	// Select Expressions
	case *ast.DistinctExpression:
		for _, e := range node.Right {
			r.Resolve(e, env)
		}
	case *ast.ColumnExpression:
		r.resolveColumnExpression(node, env)
	case *ast.SortExpression:
		r.Resolve(node.Value, env)
	case *ast.FetchExpression:
		r.Resolve(node.Value, env)
	case *ast.AggregateExpression:
		r.Resolve(node.Left, env)
		for _, a := range node.Right {
			r.Resolve(a, env)
		}
	case *ast.WindowExpression:
		r.Resolve(node.Alias, env)
		for _, p := range node.PartitionBy {
			r.Resolve(p, env)
		}
		for _, o := range node.OrderBy {
			r.Resolve(o, env)
		}
	case *ast.TableExpression:
		// eventually we'll also remove the table alias
		r.Resolve(node.JoinCondition, env)
		r.Resolve(node.Table, env)
	case *ast.LockExpression:
		for _, t := range node.Tables {
			r.Resolve(t, env)
		}
	case *ast.InExpression:
		r.Resolve(node.Left, env)
		for _, e := range node.Right {
			r.Resolve(e, env)
		}
	case *ast.CastExpression:
		r.Resolve(node.Left, env)
	case *ast.WhereExpression:
		r.Resolve(node.Right, env)
	case *ast.IsExpression:
		r.Resolve(node.Left, env)
		r.Resolve(node.Right, env)
	case *ast.TrimExpression:
		r.Resolve(node.Expression, env)
	case *ast.StringFunctionExpression:
		r.Resolve(node.Left, env)
		r.Resolve(node.From, env)
		r.Resolve(node.For, env)
	case *ast.UnionExpression:
		r.Resolve(node.Left, env)
		r.Resolve(node.Right, env)

	// Primitive Expressions
	case *ast.Identifier:
		r.resolveIdentifier(node, env)

	// Noops
	case *ast.UpdateExpression:
		// Currently do nothing till we verify that we don't have aliases to resolve

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

func (r *Resolver) resolveProgram(program *ast.Program, env *object.Environment) {
	for _, s := range program.Statements {
		r.Resolve(s, env)
	}
}

func (r *Resolver) resolveSelectStatement(s *ast.SelectStatement, env *object.Environment) {
	for _, x := range s.Expressions {
		r.Resolve(x, env)
	}
}

func (r *Resolver) resolveSelectExpression(x *ast.SelectExpression, env *object.Environment) {
	r.Resolve(x.Distinct, env)
	for _, c := range x.Columns {
		r.Resolve(c, env)
	}
	for _, t := range x.Tables {
		r.Resolve(t, env)
	}
	r.Resolve(x.Where, env)
	for _, g := range x.GroupBy {
		r.Resolve(g, env)
	}
	r.Resolve(x.Having, env)
	for _, w := range x.Window {
		r.Resolve(w, env)
	}
	for _, o := range x.OrderBy {
		r.Resolve(o, env)
	}
	r.Resolve(x.Limit, env)
	r.Resolve(x.Offset, env)
	r.Resolve(x.Fetch, env)
	r.Resolve(x.Lock, env)
}

func (r *Resolver) resolveColumnExpression(c *ast.ColumnExpression, env *object.Environment) {
	r.Resolve(c.Value, env)
}

func (r *Resolver) resolveIdentifier(i *ast.Identifier, env *object.Environment) {
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

func (r *Resolver) newError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	r.errors = append(r.errors, msg)
}

func (r *Resolver) Errors() []string {
	return r.errors
}

func (r *Resolver) PrintErrors() {
	if len(r.errors) == 0 {
		return
	}
	for _, msg := range r.errors {
		fmt.Printf("resolver error: %s\n", msg)
	}
}
