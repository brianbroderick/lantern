package ast

import (
	"testing"

	"github.com/brianbroderick/lantern/internal/sql/token"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&SelectStatement{
				Token: token.Token{Type: token.SELECT, Lit: "select"},
				Columns: &Identifier{
					Token: token.Token{Type: token.IDENT, Lit: "id"},
					Value: "id",
				},

				From: &Identifier{
					Token: token.Token{Type: token.IDENT, Lit: "users"},
					Value: "users",
				},
			},
		},
	}

	if program.String() != "select id from users;" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}
