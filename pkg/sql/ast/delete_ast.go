package ast

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type DeleteStatement struct {
	Token      token.Token `json:"token,omitempty"` // the token.DELETE token
	Expression Expression  `json:"expression,omitempty"`
}

func (s *DeleteStatement) Clause() token.TokenType      { return s.Token.Type }
func (x *DeleteStatement) SetClause(c token.TokenType)  {}
func (s *DeleteStatement) Command() token.TokenType     { return s.Token.Type }
func (x *DeleteStatement) SetCommand(c token.TokenType) {}
func (s *DeleteStatement) statementNode()               {}
func (s *DeleteStatement) TokenLiteral() string         { return s.Token.Upper }
func (s *DeleteStatement) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(s.Expression.String(maskParams))
	out.WriteString(";")
	return out.String()
}
func (s *DeleteStatement) Inspect(maskParams bool) string {
	j, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Printf("Error loading data: %#v\n\n", err)
	}
	return string(j)
}

type DeleteExpression struct {
	Token        token.Token       `json:"token,omitempty"` // the token.DELETE token
	Only         bool              `json:"only,omitempty"`
	Table        Expression        `json:"table,omitempty"`
	Alias        Expression        `json:"alias,omitempty"`
	Using        []Expression      `json:"using,omitempty"`
	Cursor       Expression        `json:"cursor,omitempty"`
	Where        Expression        `json:"where,omitempty"` // TODO: handle WHERE CURRENT OF cursor_name
	Returning    []Expression      `json:"returning,omitempty"`
	Cast         Expression        `json:"cast,omitempty"`
	TableAliases map[string]string `json:"-"`                // map of table aliases
	Branch       token.TokenType   `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag   token.TokenType   `json:"command,omitempty"`
}

func (x *DeleteExpression) Clause() token.TokenType      { return x.Branch }
func (x *DeleteExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *DeleteExpression) Command() token.TokenType     { return x.CommandTag }
func (x *DeleteExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *DeleteExpression) expressionNode()              {}
func (x *DeleteExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *DeleteExpression) SetCast(cast Expression) {
	x.Cast = cast
}

func (x *DeleteExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString("(DELETE FROM ")

	if x.Only {
		out.WriteString("ONLY ")
	}

	out.WriteString(x.Table.String(maskParams))

	if x.Alias != nil {
		out.WriteString(" AS ")
		out.WriteString(x.Alias.String(maskParams))
	}

	if len(x.Using) > 0 {
		out.WriteString(" USING ")
		for _, u := range x.Using {
			out.WriteString(u.String(maskParams))
		}
	}

	if x.Where != nil {
		out.WriteString(" WHERE ")
		out.WriteString(x.Where.String(maskParams))
	}

	if len(x.Returning) > 0 {
		out.WriteString(" RETURNING ")
		for i, r := range x.Returning {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(r.String(maskParams))
		}
	}

	out.WriteString(")")
	return out.String()
}
