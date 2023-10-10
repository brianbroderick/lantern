package lexer

import (
	"testing"

	"github.com/brianbroderick/lantern/internal/sql/token"
	"github.com/stretchr/testify/assert"
)

func TestCodeScan(t *testing.T) {
	input := `
	select u.id, u.name from users u inner join addresses a on a.user_id = u.id where id = 42;
	`

	tests := []token.Token{
		{Type: token.SELECT, Lit: "select"},
		{Type: token.IDENT, Lit: "u.id"},
		{Type: token.COMMA, Lit: ","},
		{Type: token.IDENT, Lit: "u.name"},
		{Type: token.FROM, Lit: "from"},
		{Type: token.IDENT, Lit: "users"},
		{Type: token.IDENT, Lit: "u"},
		{Type: token.INNER, Lit: "inner"},
		{Type: token.JOIN, Lit: "join"},
		{Type: token.IDENT, Lit: "addresses"},
		{Type: token.IDENT, Lit: "a"},
		{Type: token.ON, Lit: "on"},
		{Type: token.IDENT, Lit: "a.user_id"},
		{Type: token.ASSIGN, Lit: "="},
		{Type: token.IDENT, Lit: "u.id"},
		{Type: token.WHERE, Lit: "where"},
		{Type: token.IDENT, Lit: "id"},
		{Type: token.ASSIGN, Lit: "="},
		{Type: token.INT, Lit: "42"},
		{Type: token.SEMICOLON, Lit: ";"},
		{Type: token.EOF, Lit: ""},
	}

	l := New(input)

	for _, tt := range tests {
		tok, _ := l.Scan()

		assert.Equal(t, token.Tokens[tt.Type], token.Tokens[tok.Type])
		assert.Equal(t, tt.Lit, tok.Lit)
	}
}

func TestSimpleScan(t *testing.T) {
	input := ` . + - * < >
	; , { } ( ) = == ! != / foo 12345 
	-> ->> #> #>> #-
	`

	tests := []token.Token{
		{Type: token.DOT, Lit: "."},
		{Type: token.PLUS, Lit: "+"},
		{Type: token.MINUS, Lit: "-"},
		{Type: token.ASTERISK, Lit: "*"},
		{Type: token.LT, Lit: "<"},
		{Type: token.GT, Lit: ">"},
		{Type: token.SEMICOLON, Lit: ";"},
		{Type: token.COMMA, Lit: ","},
		{Type: token.LBRACE, Lit: "{"},
		{Type: token.RBRACE, Lit: "}"},
		{Type: token.LPAREN, Lit: "("},
		{Type: token.RPAREN, Lit: ")"},
		{Type: token.ASSIGN, Lit: "="},
		{Type: token.EQ, Lit: "=="},
		{Type: token.BANG, Lit: "!"},
		{Type: token.NOT_EQ, Lit: "!="},
		{Type: token.SLASH, Lit: "/"},
		{Type: token.IDENT, Lit: "foo"},
		{Type: token.INT, Lit: "12345"},
		{Type: token.JSONGETBYKEY, Lit: "->"},
		{Type: token.JSONGETBYTEXT, Lit: "->>"},
		{Type: token.JSONGETBYPATH, Lit: "#>"},
		{Type: token.JSONGETBYPATHTEXT, Lit: "#>>"},
		{Type: token.JSONDELETE, Lit: "#-"},
		{Type: token.EOF, Lit: ""},
	}

	l := New(input)

	for _, tt := range tests {
		tok, _ := l.Scan()

		assert.Equal(t, token.Tokens[tt.Type], token.Tokens[tok.Type])
		assert.Equal(t, tt.Lit, tok.Lit)
	}
}
