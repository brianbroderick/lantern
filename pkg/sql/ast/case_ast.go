package ast

import (
	"bytes"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type CaseExpression struct {
	Token       token.Token            `json:"token,omitempty"` // the token.CASE token
	Expression  Expression             `json:"expression,omitempty"`
	Conditions  []*ConditionExpression `json:"conditions,omitempty"`
	Alternative Expression             `json:"alternative,omitempty"`
	Cast        Expression             `json:"cast,omitempty"`
	Branch      token.TokenType        `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *CaseExpression) Clause() token.TokenType     { return x.Branch }
func (x *CaseExpression) SetClause(c token.TokenType) { x.Branch = c }
func (x *CaseExpression) Command() token.TokenType    { return x.Token.Type }
func (x *CaseExpression) expressionNode()             {}
func (x *CaseExpression) TokenLiteral() string        { return x.Token.Lit }
func (x *CaseExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("CASE")

	if x.Expression != nil {
		out.WriteString(" ")
		out.WriteString(x.Expression.String(maskParams))
	}

	for _, c := range x.Conditions {
		out.WriteString(c.String(maskParams))
	}

	if x.Alternative != nil {
		out.WriteString(" ELSE ")
		out.WriteString(x.Alternative.String(maskParams))
	}
	out.WriteString(" END")

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *CaseExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type ConditionExpression struct {
	Token       token.Token     `json:"token,omitempty"`
	Condition   Expression      `json:"condition,omitempty"`
	Consequence Expression      `json:"consequence,omitempty"`
	Cast        Expression      `json:"cast,omitempty"`
	Branch      token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *ConditionExpression) Clause() token.TokenType     { return x.Branch }
func (x *ConditionExpression) SetClause(c token.TokenType) { x.Branch = c }
func (x *ConditionExpression) Command() token.TokenType    { return x.Token.Type }
func (x *ConditionExpression) expressionNode()             {}
func (x *ConditionExpression) TokenLiteral() string        { return x.Token.Lit }
func (x *ConditionExpression) String(maskParams bool) string {
	var out bytes.Buffer

	if x.Condition != nil {
		out.WriteString(" WHEN ")
		out.WriteString(x.Condition.String(maskParams))
	}

	if x.Consequence != nil {
		out.WriteString(" THEN ")
		out.WriteString(x.Consequence.String(maskParams))
	}

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *ConditionExpression) SetCast(cast Expression) {
	x.Cast = cast
}
