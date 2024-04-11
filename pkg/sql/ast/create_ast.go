package ast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

// Currently we're only handling CREATE TABLE AS & CREATE INDEX, but we'll need to handle CREATE TABLE, CREATE TRIGGER, etc.

type CreateStatement struct {
	Token        token.Token `json:"token,omitempty"`        // the token.CREATE token
	Scope        string      `json:"scope,omitempty"`        // GLOBAL or LOCAL
	Unique       bool        `json:"unique,omitempty"`       // UNIQUE
	Concurrently bool        `json:"concurrently,omitempty"` // CONCURRENTLY
	Temp         bool        `json:"temp,omitempty"`         // TEMP or TEMPORARY (same thing)
	Unlogged     bool        `json:"unlogged,omitempty"`     // UNLOGGED
	Object       token.Token `json:"object,omitempty"`       // TABLE, INDEX, VIEW, etc.
	Exists       bool        `json:"exists,omitempty"`       // IF NOT EXISTS
	Name         Expression  `json:"name,omitempty"`         // the name of the object
	OnCommit     string      `json:"on_commit,omitempty"`    // PRESERVE ROWS, DELETE ROWS, DROP
	Operator     string      `json:"operator,omitempty"`     // AS (for CREATE TABLE AS), ON for CREATE INDEX ON, etc.
	Expression   Expression  `json:"expression,omitempty"`   // the expression to create the object
}

func (x *CreateStatement) Clause() token.TokenType     { return x.Token.Type }
func (x *CreateStatement) SetClause(c token.TokenType) {}
func (x *CreateStatement) Command() token.TokenType    { return x.Token.Type }
func (x *CreateStatement) statementNode()              {}
func (x *CreateStatement) TokenLiteral() string        { return x.Token.Lit }
func (x *CreateStatement) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString("CREATE")
	if x.Scope != "" {
		out.WriteString(" " + x.Scope)
	}
	if x.Unique {
		out.WriteString(" UNIQUE")
	}
	if x.Temp {
		out.WriteString(" TEMP")
	}
	if x.Unlogged {
		out.WriteString(" UNLOGGED")
	}
	if x.Object.Type != token.ILLEGAL {
		out.WriteString(" " + x.Object.Upper)
	}
	if x.Concurrently {
		out.WriteString(" CONCURRENTLY")
	}
	if x.Exists {
		out.WriteString(" IF NOT EXISTS")
	}
	if x.Name != nil {
		out.WriteString(" " + x.Name.String(maskParams))
	}
	if x.OnCommit != "" {
		out.WriteString(" ON COMMIT " + strings.ToUpper(x.OnCommit))
	}
	if x.Operator != "" {
		out.WriteString(" " + strings.ToUpper(x.Operator))
	}
	if x.Expression != nil {
		out.WriteString(" " + x.Expression.String(maskParams))
	}
	out.WriteString(";")

	return out.String()
}

func (x *CreateStatement) Inspect(maskParams bool) string {
	j, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling data: %#v\n\n", err)
	}
	return string(j)
}

type LikeExpression struct {
	Token   token.Token     `json:"token,omitempty"` // the token.LIKE token
	Table   Expression      `json:"table,omitempty"`
	Options []string        `json:"options,omitempty"`
	Cast    Expression      `json:"cast,omitempty"`
	Branch  token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *LikeExpression) Clause() token.TokenType     { return x.Branch }
func (x *LikeExpression) SetClause(c token.TokenType) { x.Branch = c }
func (x *LikeExpression) Command() token.TokenType    { return x.Token.Type }
func (x *LikeExpression) expressionNode()             {}
func (x *LikeExpression) TokenLiteral() string        { return x.Token.Lit }
func (x *LikeExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString("(LIKE ")
	out.WriteString(x.Table.String(maskParams))
	if len(x.Options) > 0 {
		for _, o := range x.Options {
			out.WriteString(" ")
			out.WriteString(o)
		}
	}
	out.WriteString(")")
	return out.String()
}
func (x *LikeExpression) SetCast(cast Expression) {
	x.Cast = cast
}
