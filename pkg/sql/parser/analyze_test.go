package parser

import (
	"strings"
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzeStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{
		// Select: simple
		{"analyze temp_my_table;", "(ANALYZE temp_my_table);"},
		{"analyze verbose temp_my_table;", "(ANALYZE (VERBOSE) temp_my_table);"},
		{"analyze (verbose) temp_my_table;", "(ANALYZE (VERBOSE) temp_my_table);"},
		{"analyze (verbose, skip_locked) temp_my_table;", "(ANALYZE (VERBOSE, SKIP_LOCKED) temp_my_table);"},
		{"analyze (skip_locked) temp_my_table;", "(ANALYZE (SKIP_LOCKED) temp_my_table);"},
		{"analyze (verbose, skip_locked) temp_my_table;", "(ANALYZE (VERBOSE, SKIP_LOCKED) temp_my_table);"},
		{"analyze (verbose, skip_locked, buffer_usage_limit 10) temp_my_table;", "(ANALYZE (VERBOSE, SKIP_LOCKED, BUFFER_USAGE_LIMIT 10) temp_my_table);"},
		{"analyze (verbose, skip_locked, buffer_usage_limit 10 kb) temp_my_table;", "(ANALYZE (VERBOSE, SKIP_LOCKED, BUFFER_USAGE_LIMIT 10KB) temp_my_table);"},
		{"analyze (verbose on, skip_locked true) temp_my_table;", "(ANALYZE (VERBOSE, SKIP_LOCKED) temp_my_table);"},
		{"analyze (verbose off, skip_locked false) temp_my_table;", "(ANALYZE temp_my_table);"},
		{"analyze (buffer_usage_limit 20 mb) temp_my_table;", "(ANALYZE (BUFFER_USAGE_LIMIT 20MB) temp_my_table);"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		stmt := program.Statements[0]
		assert.Equal(t, "analyze", strings.ToLower(stmt.TokenLiteral()), "input: %s\nprogram.Statements[0] is not ast.AnalyzeStatement. got=%T", tt.input, stmt)

		_, ok := stmt.(*ast.AnalyzeStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.AnalyzeStatement. got=%T", tt.input, stmt)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
