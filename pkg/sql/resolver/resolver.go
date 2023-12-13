package resolver

import (
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/object"
)

type Resolver struct {
	ast    *ast.Program
	errors []string
}

func New(program *ast.Program) *Resolver {
	return &Resolver{
		ast:    program,
		errors: []string{},
	}
}

func (r *Resolver) Resolve(node ast.Node, env *object.Environment) {
	switch node := node.(type) {
	case *ast.Program:
		r.evalProgram(node, env)
	case *ast.SelectStatement:
		r.evalSelectStatement(node, env)

	// Expressions
	case *ast.SelectExpression:
		r.evalSelectExpression(node, env)
	}

	r.newError("unknown node type: %T", node)
}

func (r *Resolver) evalProgram(program *ast.Program, env *object.Environment) {
	for _, s := range program.Statements {
		r.Resolve(s, env)
	}
}

func (r *Resolver) evalSelectStatement(statement *ast.SelectStatement, env *object.Environment) {
	for _, x := range statement.Expressions {
		r.Resolve(x, env)
	}
}

func (r *Resolver) evalSelectExpression(expression *ast.SelectExpression, env *object.Environment) {

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
