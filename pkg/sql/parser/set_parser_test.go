package parser

import (
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestSetStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input     string
		stmtCount int
		output    string
	}{
		// Select: simple
		{"set application_name = 'example';", 1, "SET (application_name = 'example');"},
		{"set session application_name = 'example';", 1, "SET SESSION (application_name = 'example');"},
		{"set local application_name = 'example';", 1, "SET LOCAL (application_name = 'example');"},
		{"set local application_name to 'example';", 1, "SET LOCAL (application_name TO 'example');"},
		{"set local time zone 'example';", 1, "SET LOCAL TIME ZONE 'example';"},
		{"set local time zone mdt;", 1, "SET LOCAL TIME ZONE mdt;"},
		{"set local time zone local;", 1, "SET LOCAL TIME ZONE local;"},
		{"set local time zone default;", 1, "SET LOCAL TIME ZONE default;"},
		{"set session characteristics as transaction isolation level read uncommitted;", 1, "SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;"},
		{"set session characteristics as transaction isolation level serializable;", 1, "SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE;"},
		{"set session characteristics as transaction isolation level REPEATABLE read;", 1, "SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL REPEATABLE READ;"},
		{"set session characteristics as transaction isolation level read committed;", 1, "SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ COMMITTED;"},
		{"set session characteristics as transaction isolation level read UNCOMMITTED;", 1, "SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;"},
		{"set transaction isolation level repeatable read;", 1, "SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;"},
		{"set constraints my_id deferred;", 1, "SET CONSTRAINTS my_id DEFERRED;"},
		{"set constraints all immediate;", 1, "SET CONSTRAINTS ALL IMMEDIATE;"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		assert.Equal(t, tt.stmtCount, len(program.Statements), "input: %s\nprogram.Statements does not contain %d statements. got=%d\n", tt.input, tt.stmtCount, len(program.Statements))

		stmt := program.Statements[0]
		assert.Equal(t, "SET", stmt.TokenLiteral(), "input: %s\nprogram.Statements[0] is not ast.SetStatement. got=%T", tt.input, stmt)

		_, ok := stmt.(*ast.SetStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.SetStatement. got=%T", tt.input, stmt)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
