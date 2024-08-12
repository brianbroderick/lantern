package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type BeginStatement struct {
	Token      token.Token `json:"token,omitempty"` // the token.BEGIN token
	Expression Expression  `json:"expression,omitempty"`
}

func (x *BeginStatement) Clause() token.TokenType      { return x.Expression.Clause() }
func (x *BeginStatement) SetClause(c token.TokenType)  {}
func (x *BeginStatement) Command() token.TokenType     { return x.Expression.Command() }
func (x *BeginStatement) SetCommand(c token.TokenType) {}
func (x *BeginStatement) statementNode()               {}
func (x *BeginStatement) TokenLiteral() string         { return x.Token.Lit }
func (x *BeginStatement) String(maskParams bool) string {
	var out bytes.Buffer

	if x.Expression != nil {
		out.WriteString(x.Expression.String(maskParams))
	}

	out.WriteString(";")

	return out.String()
}
func (x *BeginStatement) Inspect(maskParams bool) string {
	return x.String(maskParams)
}

type BeginExpression struct {
	Token      token.Token     `json:"token,omitempty"` // the token.BEGIN token
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *BeginExpression) Clause() token.TokenType      { return x.Branch }
func (x *BeginExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *BeginExpression) Command() token.TokenType     { return x.CommandTag }
func (x *BeginExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *BeginExpression) expressionNode()              {}
func (x *BeginExpression) TokenLiteral() string         { return x.Token.Lit }
func (x *BeginExpression) String(maskParams bool) string {
	if x.Cast != nil {
		return fmt.Sprintf("%s::%s", x.Token.Upper, strings.ToUpper(x.Cast.String(maskParams)))
	}

	return x.Token.Upper
}
func (x *BeginExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type CommitStatement struct {
	Token      token.Token `json:"token,omitempty"` // the token.COMMIT token
	Expression Expression  `json:"expression,omitempty"`
}

func (x *CommitStatement) Clause() token.TokenType      { return x.Expression.Clause() }
func (x *CommitStatement) SetClause(c token.TokenType)  {}
func (x *CommitStatement) Command() token.TokenType     { return x.Expression.Command() }
func (x *CommitStatement) SetCommand(c token.TokenType) {}
func (x *CommitStatement) statementNode()               {}
func (x *CommitStatement) TokenLiteral() string         { return x.Token.Lit }
func (x *CommitStatement) String(maskParams bool) string {
	var out bytes.Buffer

	if x.Expression != nil {
		out.WriteString(x.Expression.String(maskParams))
	}

	out.WriteString(";")

	return out.String()
}
func (x *CommitStatement) Inspect(maskParams bool) string {
	return x.String(maskParams)
}

type CommitExpression struct {
	Token      token.Token     `json:"token,omitempty"` // the token.COMMIT token
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *CommitExpression) Clause() token.TokenType      { return x.Branch }
func (x *CommitExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *CommitExpression) Command() token.TokenType     { return x.CommandTag }
func (x *CommitExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *CommitExpression) expressionNode()              {}
func (x *CommitExpression) TokenLiteral() string         { return x.Token.Lit }
func (x *CommitExpression) String(maskParams bool) string {
	if x.Cast != nil {
		return fmt.Sprintf("%s::%s", x.Token.Upper, strings.ToUpper(x.Cast.String(maskParams)))
	}

	return x.Token.Upper
}
func (x *CommitExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type RollbackStatement struct {
	Token      token.Token `json:"token,omitempty"` // the token.ROLLBACK token
	Expression Expression  `json:"expression,omitempty"`
}

func (x *RollbackStatement) Clause() token.TokenType      { return x.Expression.Clause() }
func (x *RollbackStatement) SetClause(c token.TokenType)  {}
func (x *RollbackStatement) Command() token.TokenType     { return x.Expression.Command() }
func (x *RollbackStatement) SetCommand(c token.TokenType) {}
func (x *RollbackStatement) statementNode()               {}
func (x *RollbackStatement) TokenLiteral() string         { return x.Token.Lit }
func (x *RollbackStatement) String(maskParams bool) string {
	var out bytes.Buffer

	if x.Expression != nil {
		out.WriteString(x.Expression.String(maskParams))
	}

	out.WriteString(";")

	return out.String()
}
func (x *RollbackStatement) Inspect(maskParams bool) string {
	return x.String(maskParams)
}

type RollbackExpression struct {
	Token      token.Token     `json:"token,omitempty"` // the token.ROLLBACK token
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *RollbackExpression) Clause() token.TokenType      { return x.Branch }
func (x *RollbackExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *RollbackExpression) Command() token.TokenType     { return x.CommandTag }
func (x *RollbackExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *RollbackExpression) expressionNode()              {}
func (x *RollbackExpression) TokenLiteral() string         { return x.Token.Lit }
func (x *RollbackExpression) String(maskParams bool) string {
	if x.Cast != nil {
		return fmt.Sprintf("%s::%s", x.Token.Upper, strings.ToUpper(x.Cast.String(maskParams)))
	}

	return x.Token.Upper
}
func (x *RollbackExpression) SetCast(cast Expression) {
	x.Cast = cast
}
