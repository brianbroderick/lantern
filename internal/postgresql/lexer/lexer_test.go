package lexer

import (
	"testing"

	"github.com/brianbroderick/lantern/internal/postgresql/token"
	"github.com/stretchr/testify/assert"
)

func TestSimpleScan(t *testing.T) {
	input := `< > : ( ) [ ] foo 12345 `

	tests := []token.Token{
		{Type: token.LT, Lit: "<"},
		{Type: token.GT, Lit: ">"},
		{Type: token.COLON, Lit: ":"},
		{Type: token.LPAREN, Lit: "("},
		{Type: token.RPAREN, Lit: ")"},
		{Type: token.LBRACKET, Lit: "["},
		{Type: token.RBRACKET, Lit: "]"},
		{Type: token.IDENT, Lit: "foo"},
		{Type: token.INT, Lit: "12345"},
		{Type: token.EOF, Lit: ""},
	}

	l := New(input)

	for _, tt := range tests {
		tok, _ := l.Scan()

		assert.Equal(t, token.Tokens[tt.Type], token.Tokens[tok.Type])
		assert.Equal(t, tt.Lit, tok.Lit)
	}
}
