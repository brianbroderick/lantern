package ast

import (
	"bytes"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type ValuesExpression struct {
	Token  token.Token     `json:"token,omitempty"` // the token.VALUES token
	Tuples [][]Expression  `json:"values,omitempty"`
	Cast   Expression      `json:"cast,omitempty"`
	Branch token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *ValuesExpression) Clause() token.TokenType     { return x.Branch }
func (x *ValuesExpression) SetClause(c token.TokenType) { x.Branch = c }
func (x *ValuesExpression) Command() token.TokenType    { return x.Token.Type }
func (x *ValuesExpression) expressionNode()             {}
func (x *ValuesExpression) TokenLiteral() string        { return x.Token.Lit }
func (x *ValuesExpression) SetCast(cast Expression) {
	x.Cast = cast
}
func (x *ValuesExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString("(VALUES")
	for _, t := range x.Tuples {
		out.WriteString(" (")
		for i, e := range t {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(e.String(maskParams))
		}
		out.WriteString("),")
	}
	len := out.Len()
	// Remove the last comma, which exists if the length is greater than 7 (the length of "(VALUES")
	if len > 7 {
		out.Truncate(len - 1)
	}

	out.WriteString(")")
	return out.String()
}
