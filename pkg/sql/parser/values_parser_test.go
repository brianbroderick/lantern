package parser

import (
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestValuesExpressions(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{
		{"values ('UA502', 'Bananas', 105, '1971-07-13', 'Comedy', '82 minutes')", "(VALUES ('UA502', 'Bananas', 105, '1971-07-13', 'Comedy', '82 minutes'))"},
		{"values ((41, 1), (42, 2))", "(VALUES ((41, 1), (42, 2)))"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.ExpressionStatement. got=%T", tt.input, stmt)

		_, ok = stmt.Expression.(*ast.ValuesExpression)
		assert.True(t, ok, "input: %s\nstmt is not *ast.ValuesExpression. got=%T", tt.input, stmt)

		output := program.String(maskParams, nil)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
