package parser

import (
	"strings"
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestDropStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{
		// Select: simple
		{"drop table if exists my_table;", "(DROP TABLE IF EXISTS my_table);"},
		{"drop table IF EXISTS my_table;", "(DROP TABLE IF EXISTS my_table);"},
		{"DROP table my_table;", "(DROP TABLE my_table);"},
		{"DROP table my_table, my_other_table;", "(DROP TABLE my_table, my_other_table);"},
		{"drop table my_table; select 1 from other;", "(DROP TABLE my_table);(SELECT 1 FROM other);"},
		{"DROP table my_table cascade;", "(DROP TABLE my_table CASCADE);"},
		{"DROP table my_table restrict;", "(DROP TABLE my_table RESTRICT);"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		stmt := program.Statements[0]
		assert.Equal(t, "drop", strings.ToLower(stmt.TokenLiteral()), "input: %s\nprogram.Statements[0] is not ast.DropStatement. got=%T", tt.input, stmt)

		_, ok := stmt.(*ast.DropStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.DropStatement. got=%T", tt.input, stmt)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
