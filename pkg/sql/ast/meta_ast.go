package ast

import (
	"bytes"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

// This file contains the AST for metadata such as SHOW, SAVEPOINT, and RESET, DISCARD, etc.

type ShowStatement struct {
	Token      token.Token `json:"token,omitempty"` // the token.SHOW token
	Expression Expression  `json:"expression,omitempty"`
}

func (x *ShowStatement) Clause() token.TokenType      { return x.Expression.Clause() }
func (x *ShowStatement) SetClause(c token.TokenType)  {}
func (x *ShowStatement) Command() token.TokenType     { return x.Expression.Command() }
func (x *ShowStatement) SetCommand(c token.TokenType) {}
func (x *ShowStatement) statementNode()               {}
func (x *ShowStatement) TokenLiteral() string         { return x.Token.Lit }
func (x *ShowStatement) String(maskParams bool) string {
	var out bytes.Buffer

	if x.Expression != nil {
		out.WriteString(x.Expression.String(maskParams))
	}

	out.WriteString(";")

	return out.String()
}

func (x *ShowStatement) Inspect(maskParams bool) string {
	return x.String(maskParams)
}

type ShowExpression struct {
	Token      token.Token     `json:"token,omitempty"` // the token.SHOW token
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
	Expression Expression      `json:"expression,omitempty"`
}

func (x *ShowExpression) Clause() token.TokenType      { return x.Branch }
func (x *ShowExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *ShowExpression) Command() token.TokenType     { return x.CommandTag }
func (x *ShowExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *ShowExpression) expressionNode()              {}
func (x *ShowExpression) TokenLiteral() string         { return x.Token.Lit }
func (x *ShowExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(x.Token.Upper)
	out.WriteString(" ")

	if x.Expression != nil {
		out.WriteString(x.Expression.String(maskParams))
	}

	return out.String()
}

func (x *ShowExpression) SetCast(cast Expression) {
	x.Cast = cast
}
