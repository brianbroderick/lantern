package ast

import (
	"bytes"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type CaseExpression struct {
	Token       token.Token            `json:"token,omitempty"` // the token.CASE token
	Conditions  []*ConditionExpression `json:"conditions,omitempty"`
	Alternative Expression             `json:"alternative,omitempty"`
	Cast        Expression             `json:"cast,omitempty"`
}

func (ce *CaseExpression) expressionNode()      {}
func (ce *CaseExpression) TokenLiteral() string { return ce.Token.Lit }
func (ce *CaseExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("CASE")

	for _, c := range ce.Conditions {
		out.WriteString(c.String(maskParams))
	}

	if ce.Alternative != nil {
		out.WriteString(" ELSE ")
		out.WriteString(ce.Alternative.String(maskParams))
	}
	out.WriteString(" END")

	return out.String()
}
func (ce *CaseExpression) SetCast(cast Expression) {
	ce.Cast = cast
}

type ConditionExpression struct {
	Token       token.Token `json:"token,omitempty"`
	Condition   Expression  `json:"condition,omitempty"`
	Consequence Expression  `json:"consequence,omitempty"`
	Cast        Expression  `json:"cast,omitempty"`
}

func (x *ConditionExpression) expressionNode()      {}
func (x *ConditionExpression) TokenLiteral() string { return x.Token.Lit }
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
