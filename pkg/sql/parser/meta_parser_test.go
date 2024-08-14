package parser

import (
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestShowStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input     string
		stmtCount int
		output    string
	}{
		// Select: simple
		{"show server_version", 1, "SHOW (server_version);"},
		{"show transaction isolation level", 1, "SHOW (TRANSACTION ISOLATION LEVEL );"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		assert.Equal(t, tt.stmtCount, len(program.Statements), "input: %s\nprogram.Statements does not contain %d statements. got=%d\n", tt.input, tt.stmtCount, len(program.Statements))

		stmt := program.Statements[0]
		assert.Equal(t, "SHOW", stmt.TokenLiteral(), "input: %s\nprogram.Statements[0] is not ast.ShowStatement. got=%T", tt.input, stmt)

		_, ok := stmt.(*ast.ShowStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.ShowStatement. got=%T", tt.input, stmt)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
