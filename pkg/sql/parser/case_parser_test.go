package parser

import (
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestCaseExpressions(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{
		// Case Expressions
		{"case when id = 1 then 'one' end", "CASE WHEN (id = 1) THEN 'one' END"},
		{"case when id = 1 then 'one' end;", "CASE WHEN (id = 1) THEN 'one' END"},
		{"case when id = 1 then 'one' else 'other' end", "CASE WHEN (id = 1) THEN 'one' ELSE 'other' END"},
		{"case when id = 1 then 'one' else 'other' end;", "CASE WHEN (id = 1) THEN 'one' ELSE 'other' END"},
		{"case when id = 1 then 'one' when id = 2 then 'two' else 'other' end", "CASE WHEN (id = 1) THEN 'one' WHEN (id = 2) THEN 'two' ELSE 'other' END"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.ExpressionStatement. got=%T", tt.input, stmt)

		_, ok = stmt.Expression.(*ast.CaseExpression)
		assert.True(t, ok, "input: %s\nstmt is not *ast.CaseExpression. got=%T", tt.input, stmt)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
