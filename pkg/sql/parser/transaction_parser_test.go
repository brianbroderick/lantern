package parser

import (
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestTransactionStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input     string
		stmtCount int
		output    string
	}{
		// Select: simple
		{"commit;", 1, "COMMIT;"},
		{"rollback", 1, "ROLLBACK;"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		assert.Equal(t, tt.stmtCount, len(program.Statements), "input: %s\nprogram.Statements does not contain %d statements. got=%d\n", tt.input, tt.stmtCount, len(program.Statements))

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
