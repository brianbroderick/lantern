package pgLogParser

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestTokenizer(t *testing.T) {
	type token = struct {
		tok Token
		lit string
	}

	var tests = []struct {
		s      string
		tokens []token
	}{
		{
			s: "postgres@postgres",
			tokens: []token{
				{tok: IDENT, lit: "postgres"},
				{tok: ATSYMBOL, lit: ""},
				{tok: IDENT, lit: "postgres"},
			},
		},
	}

	for _, tt := range tests {
		tokenizer := newTokenizer(strings.NewReader(tt.s))
		tok := ILLEGAL
		var lit string
		var s []token
		for tok != EOF {
			tok, _, lit = tokenizer.Scan()
			if tok != EOF {
				s = append(s, token{tok: tok, lit: lit})
			}
		}

		assert.Equal(t, tt.tokens, s)
	}

}
