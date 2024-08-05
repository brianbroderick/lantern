package parser

import (
	"strings"
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestCreateStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{
		// Select: simple
		{"create temp table if not exists temp_my_table as ( select id from my_table );", "CREATE TEMP TABLE IF NOT EXISTS temp_my_table AS (SELECT id FROM my_table);"},
		{"create temp table temp_my_table as (select id from my_table);", "CREATE TEMP TABLE temp_my_table AS (SELECT id FROM my_table);"},
		{"create index idx_person_id ON temp_my_table( person_id );", "CREATE INDEX idx_person_id ON temp_my_table(person_id);"},
		{"create INDEX idx_person_id ON temp_my_table( person_id );", "CREATE INDEX idx_person_id ON temp_my_table(person_id);"},
		{"CREATE INDEX idx_temp_person on pg_temp.people using btree ( account_id, person_id );", "CREATE INDEX idx_temp_person ON (pg_temp.people USING btree(account_id, person_id));"},
		{"create temp table temp_my_table on commit drop as (select id from users);", "CREATE TEMP TABLE temp_my_table ON COMMIT DROP AS (SELECT id FROM users);"},
		// TODO: This is a bug, the name should be temp_my_table. Instead, it's parsing as if it's a function name.
		{"create temp table temp_my_table( like my_reports );", "CREATE TEMP TABLE temp_my_table((LIKE my_reports));"},
		{"create index idx_temp_stuff ON my_temp USING btree( id, temp_id ) WHERE ( blah_id = 1 );", "CREATE INDEX idx_temp_stuff ON (my_temp USING btree(id, temp_id)) WHERE (blah_id = 1);"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		stmt := program.Statements[0]
		assert.Equal(t, "create", strings.ToLower(stmt.TokenLiteral()), "input: %s\nprogram.Statements[0] is not ast.CreateStatement. got=%T", tt.input, stmt)

		_, ok := stmt.(*ast.CreateStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.CreateStatement. got=%T", tt.input, stmt)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}

func TestLikeExpressions(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{
		// Case Expressions
		{"like my_reports", "(LIKE my_reports)"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.ExpressionStatement. got=%T", tt.input, stmt)

		_, ok = stmt.Expression.(*ast.LikeExpression)
		assert.True(t, ok, "input: %s\nstmt is not *ast.LikeExpression. got=%T", tt.input, stmt)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
