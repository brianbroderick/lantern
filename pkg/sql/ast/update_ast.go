package ast

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type UpdateStatement struct {
	Token      token.Token `json:"token,omitempty"` // the token.UPDATE token
	Expression Expression  `json:"expression,omitempty"`
}

func (s *UpdateStatement) Clause() token.TokenType     { return s.Token.Type }
func (x *UpdateStatement) SetClause(c token.TokenType) {}
func (s *UpdateStatement) Command() token.TokenType    { return s.Token.Type }
func (s *UpdateStatement) statementNode()              {}
func (s *UpdateStatement) TokenLiteral() string        { return s.Token.Upper }
func (s *UpdateStatement) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(s.Expression.String(maskParams))
	out.WriteString(";")
	return out.String()
}
func (s *UpdateStatement) Inspect(maskParams bool) string {
	j, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Printf("Error loading data: %#v\n\n", err)
	}
	return string(j)
}

type UpdateExpression struct {
	Token     token.Token     `json:"token,omitempty"` // the token.UPDATE token
	Only      bool            `json:"only,omitempty"`
	Table     Expression      `json:"table,omitempty"`
	Asterisk  bool            `json:"asterisk,omitempty"`
	Alias     Expression      `json:"alias,omitempty"`
	Set       []Expression    `json:"set,omitempty"`
	From      []Expression    `json:"from,omitempty"`
	Cursor    Expression      `json:"cursor,omitempty"`
	Where     Expression      `json:"where,omitempty"`
	Returning []Expression    `json:"returning,omitempty"`
	Cast      Expression      `json:"cast,omitempty"`
	Branch    token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *UpdateExpression) Clause() token.TokenType     { return x.Branch }
func (x *UpdateExpression) SetClause(c token.TokenType) { x.Branch = c }
func (x *UpdateExpression) Command() token.TokenType    { return x.Token.Type }
func (x *UpdateExpression) expressionNode()             {}
func (x *UpdateExpression) TokenLiteral() string        { return x.Token.Upper }
func (x *UpdateExpression) SetCast(cast Expression) {
	x.Cast = cast
}

func (x *UpdateExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString("(UPDATE ")

	if x.Only {
		out.WriteString("ONLY ")
	}
	if x.Table != nil {
		out.WriteString(x.Table.String(maskParams))
	}
	if x.Asterisk {
		out.WriteString(" *")
	}
	if x.Alias != nil {
		out.WriteString(" ")
		out.WriteString(x.Alias.String(maskParams))
	}
	if len(x.Set) > 0 {
		out.WriteString(" SET ")
		for i, s := range x.Set {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(s.String(maskParams))
		}
	}
	if len(x.From) > 0 {
		out.WriteString(" FROM ")
		for i, f := range x.From {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(f.String(maskParams))
		}
	}
	if x.Cursor != nil {
		out.WriteString(" WHERE CURRENT OF ")
		out.WriteString(x.Cursor.String(maskParams))
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
