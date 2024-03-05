package parser

import (
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestDeleteStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{
		{"delete from films;", "(DELETE FROM films);"},
		{"delete from films where id = 42;", "(DELETE FROM films WHERE (id = 42));"},
		{"delete from films using producers where producer_id = producers.id and id = 42;", "(DELETE FROM films USING producers WHERE ((producer_id = producers.id) AND (id = 42)));"},
		{"delete from films where id = 42 returning id;", "(DELETE FROM films WHERE (id = 42) RETURNING id);"},
		{"delete from films where id = 42 returning *;", "(DELETE FROM films WHERE (id = 42) RETURNING *);"},
		{"delete from films where id = 42 returning id, title;", "(DELETE FROM films WHERE (id = 42) RETURNING id, title);"},
		{"delete from films where producer_id IN (select id from producers where name = 'foo');", "(DELETE FROM films WHERE producer_id IN ((SELECT id FROM producers WHERE (name = 'foo'))));"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		stmt := program.Statements[0]
		assert.Equal(t, "DELETE", stmt.TokenLiteral(), "input: %s\nprogram.Statements[0] is not ast.DeleteStatement. got=%T", tt.input, stmt)

		_, ok := stmt.(*ast.DeleteStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.DeleteStatement. got=%T", tt.input, stmt)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
