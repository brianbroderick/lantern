package parser

import (
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestShowStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input     string
		stmtCount int
		output    string
		tokenLit  string
	}{
		// Select: simple
		{"show server_version", 1, "SHOW server_version;", "SHOW"},
		{"show transaction isolation level", 1, "SHOW TRANSACTION ISOLATION LEVEL ;", "SHOW"},
		{"savepoint my_savepoint", 1, "SAVEPOINT my_savepoint;", "SAVEPOINT"},
		{"savepoint foo_bar;", 1, "SAVEPOINT foo_bar;", "SAVEPOINT"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		assert.Equal(t, tt.stmtCount, len(program.Statements), "input: %s\nprogram.Statements does not contain %d statements. got=%d\n", tt.input, tt.stmtCount, len(program.Statements))

		stmt := program.Statements[0]
		assert.Equal(t, tt.tokenLit, stmt.TokenLiteral(), "input: %s\nprogram.Statements[0] is not ast.ShowStatement. got=%T", tt.input, stmt)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
