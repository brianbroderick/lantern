package ast

import (
	"testing"

	"github.com/brianbroderick/lantern/internal/sql/token"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&SelectStatement{
				Expressions: []Expression{&SelectExpression{
					Token: token.Token{Type: token.SELECT, Lit: "select"},
					Columns: []Expression{&ColumnExpression{
						Token: token.Token{Type: token.AS, Lit: "AS"},
						Name: &Identifier{
							Token: token.Token{Type: token.IDENT, Lit: "id"},
						},
						Value: &Identifier{
							Token: token.Token{Type: token.IDENT, Lit: "id"},
							Value: "id",
						},
					},
					},
					Tables: []Expression{&Identifier{
						Token: token.Token{Type: token.IDENT, Lit: "users"},
						Value: "users"},
					},
				},
				},
			},
		},
	}

	if program.String() != "SELECT id FROM users;" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}
