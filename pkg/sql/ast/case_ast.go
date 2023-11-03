package ast

import (
	"bytes"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type CaseExpression struct {
	Token       token.Token            `json:"token,omitempty"` // the token.CASE token
	Conditions  []*ConditionExpression `json:"conditions,omitempty"`
	Alternative Expression             `json:"alternative,omitempty"`
	Cast        string                 `json:"cast,omitempty"`
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
func (ce *CaseExpression) SetCast(cast string) {
	ce.Cast = cast
}

type ConditionExpression struct {
	Token       token.Token `json:"token,omitempty"`
	Condition   Expression  `json:"condition,omitempty"`
	Consequence Expression  `json:"consequence,omitempty"`
	Cast        string      `json:"cast,omitempty"`
}

func (ce *ConditionExpression) expressionNode()      {}
func (ce *ConditionExpression) TokenLiteral() string { return ce.Token.Lit }
func (ce *ConditionExpression) String(maskParams bool) string {
	var out bytes.Buffer

	if ce.Condition != nil {
		out.WriteString(" WHEN ")
		out.WriteString(ce.Condition.String(maskParams))
	}

	if ce.Consequence != nil {
		out.WriteString(" THEN ")
		out.WriteString(ce.Consequence.String(maskParams))
	}

	if ce.Cast != "" {
		out.WriteString("::")
		out.WriteString(ce.Cast)
	}

	return out.String()
}
func (ce *ConditionExpression) SetCast(cast string) {
	ce.Cast = cast
}
