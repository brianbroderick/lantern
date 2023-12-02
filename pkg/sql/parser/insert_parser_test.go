package parser

import (
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestInsertStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{
		// Select: simple
		// {"insert into users (name, email) values ('Brian', 'foo@bar.com');", "(INSERT INTO users (name, email) VALUES ('Brian', 'foo@bar.com'));"},
		{"insert into users (name, email) values ('Brian', 'foo@bar.com'), ('Bob', 'bar@foo.com');", "(INSERT INTO users (name, email) VALUES ('Brian', 'foo@bar.com'), ('Bob', 'bar@foo.com'));"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		stmt := program.Statements[0]
		assert.Equal(t, "INSERT", stmt.TokenLiteral(), "input: %s\nprogram.Statements[0] is not ast.InsertStatement. got=%T", tt.input, stmt)

		_, ok := stmt.(*ast.InsertStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.InsertStatement. got=%T", tt.input, stmt)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
