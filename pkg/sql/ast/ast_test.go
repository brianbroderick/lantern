package ast

import (
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func TestString(t *testing.T) {
	maskParams := false

	program := &Program{
		Statements: []Statement{
			&SelectStatement{
				Expressions: []Expression{&SelectExpression{
					Token: token.Token{Type: token.SELECT, Lit: "select", Upper: "SELECT"},
					Columns: []Expression{&ColumnExpression{
						Token: token.Token{Type: token.AS, Lit: "AS", Upper: "AS"},
						Name: &SimpleIdentifier{
							Token: token.Token{Type: token.IDENT, Lit: "id"},
						},
						Value: &SimpleIdentifier{
							Token: token.Token{Type: token.IDENT, Lit: "id"},
							Value: "id",
						},
					},
					},
					Tables: []Expression{&SimpleIdentifier{
						Token: token.Token{Type: token.IDENT, Lit: "users"},
						Value: "users"},
					},
				},
				},
			},
		},
	}

	if program.String(maskParams) != "(SELECT id FROM users);" {
		t.Errorf("program.String() wrong. got=%q", program.String(maskParams))
	}
}
