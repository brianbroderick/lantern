package evaluator

import (
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/logit"
	"github.com/brianbroderick/lantern/pkg/sql/object"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/brianbroderick/lantern/pkg/sql/resolver"
)

// TODO: extract tables from the SQL statements

func ExtractTables(input string) []string {
	maskParams := false

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()
	r := resolver.NewResolver(program)
	r.Resolve(r.Ast, env)

	if len(p.Errors()) > 0 {
		for _, msg := range p.Errors() {
			logit.Append("counter_error", msg)
		}
		return nil
	}

	for _, stmt := range program.Statements {
		output := stmt.String(maskParams)
		fmt.Println(output)
	}

	return nil
}
