package lexer

import (
	"testing"

	"github.com/brianbroderick/lantern/internal/postgresql/token"
	"github.com/stretchr/testify/assert"
)

func TestSimpleScan(t *testing.T) {
	input := `< > : ( ) [ ] foo 12345 `

	tests := []token.Token{
		{Type: token.LT, Literal: "<"},
		{Type: token.GT, Literal: ">"},
		{Type: token.COLON, Literal: ":"},
		{Type: token.LPAREN, Literal: "("},
		{Type: token.RPAREN, Literal: ")"},
		{Type: token.LBRACKET, Literal: "["},
		{Type: token.RBRACKET, Literal: "]"},
		{Type: token.IDENT, Literal: "foo"},
		{Type: token.INT, Literal: "12345"},
		{Type: token.EOF, Literal: ""},
	}

	l := New(input)

	for _, tt := range tests {
		tok, _ := l.Scan()

		assert.Equal(t, token.Tokens[tt.Type], token.Tokens[tok.Type])
		assert.Equal(t, tt.Literal, tok.Literal)
	}
}
