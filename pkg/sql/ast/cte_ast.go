package ast

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

// we need to specify the with clauses separate from the main query. the with tables are temporary tables with a name
// then we don't have to eff with infix vs select expressions in the case of unions

// with tmp as (select count(1) as counter from orders) select counter from tmp;

type CTEStatement struct {
	Token      token.Token `json:"token,omitempty"`
	Expression Expression  `json:"expression,omitempty"`
}

func (s *CTEStatement) Command() token.TokenType { return s.Token.Type }
func (s *CTEStatement) statementNode()           {}
func (s *CTEStatement) TokenLiteral() string     { return s.Token.Lit }
func (s *CTEStatement) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(s.Expression.String(maskParams))
	out.WriteString(";")
	return out.String()
}
func (s *CTEStatement) Inspect(maskParams bool) string {
	j, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Printf("Error loading data: %#v\n\n", err)
	}
	return string(j)
}

type CTEExpression struct {
	Token     token.Token  `json:"token,omitempty"`
	Recursive bool         `json:"recursive,omitempty"`
	Auxiliary []Expression `json:"auxiliary,omitempty"`
	Primary   Expression   `json:"primary,omitempty"`
	Cast      Expression   `json:"cast,omitempty"`
}

func (x *CTEExpression) Command() token.TokenType { return x.Token.Type }
func (x *CTEExpression) expressionNode()          {}
func (x *CTEExpression) TokenLiteral() string     { return x.Token.Lit }
func (x *CTEExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("(WITH ")

	if x.Recursive {
		out.WriteString("RECURSIVE ")
	}
	if len(x.Auxiliary) > 0 {
		for _, a := range x.Auxiliary {
			out.WriteString(a.String(maskParams))
			if a != x.Auxiliary[len(x.Auxiliary)-1] {
				out.WriteString(", ")
			}
		}
		out.WriteString(" ")
	}
	if x.Primary != nil {
		out.WriteString(x.Primary.String(maskParams))
	}
	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(x.Cast.String(maskParams))
	}
	out.WriteString(")")

	return out.String()
}
func (x *CTEExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type CTEAuxiliaryExpression struct {
	Token        token.Token `json:"token,omitempty"`
	Name         Expression  `json:"name,omitempty"`
	Materialized string      `json:"materialized,omitempty"`
	Expression   Expression  `json:"expression,omitempty"`
	Cast         Expression  `json:"cast,omitempty"`
}

func (x *CTEAuxiliaryExpression) Command() token.TokenType { return x.Token.Type }
func (x *CTEAuxiliaryExpression) expressionNode()          {}
func (x *CTEAuxiliaryExpression) TokenLiteral() string     { return x.Token.Lit }
func (x *CTEAuxiliaryExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(x.Name.String(maskParams))
	out.WriteString(" AS ")
	if x.Materialized != "" {
		out.WriteString(fmt.Sprintf("%s ", x.Materialized))
	}

	if x.Expression != nil {
		out.WriteString(x.Expression.String(maskParams))
	}

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(x.Cast.String(maskParams))
	}
	return out.String()
}
func (x *CTEAuxiliaryExpression) SetCast(cast Expression) {
	x.Cast = cast
}
