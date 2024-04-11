package parser

import (
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestRealQueries(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{

		{`SELECT 1`, "(SELECT 1);"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\n\noutput: %s\n\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
	}
}
