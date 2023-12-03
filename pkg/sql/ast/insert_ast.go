package ast

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type InsertStatement struct {
	Token      token.Token `json:"token,omitempty"` // the token.INSERT token
	Expression Expression  `json:"expression,omitempty"`
}

func (s *InsertStatement) statementNode()       {}
func (s *InsertStatement) TokenLiteral() string { return s.Token.Upper }
func (s *InsertStatement) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(s.Expression.String(maskParams))

	out.WriteString(";")

	return out.String()
}
func (s *InsertStatement) Inspect(maskParams bool) string {
	j, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Printf("Error loading data: %#v\n\n", err)
	}
	return string(j)
}

type InsertExpression struct {
	Token          token.Token    `json:"token,omitempty"` // the token.INSERT token
	Table          Expression     `json:"table,omitempty"`
	Alias          Expression     `json:"alias,omitempty"`
	Columns        []Expression   `json:"columns,omitempty"`
	Overriding     string         `json:"overriding,omitempty"`
	Default        bool           `json:"default,omitempty"`
	Values         [][]Expression `json:"values,omitempty"`
	Query          Expression     `json:"query,omitempty"`
	ConflictTarget []Expression   `json:"conflict_target,omitempty"`
	ConflictAction string         `json:"conflict_action,omitempty"`
	ConflictUpdate []Expression   `json:"conflict_update,omitempty"`
	Returning      []Expression   `json:"returning,omitempty"`
	Cast           Expression     `json:"cast,omitempty"`
}

func (x *InsertExpression) expressionNode()      {}
func (x *InsertExpression) TokenLiteral() string { return x.Token.Upper }
func (x *InsertExpression) SetCast(cast Expression) {
	x.Cast = cast
}

// String() is incomplete and only returns the most basic of select statements
func (x *InsertExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString("(INSERT INTO ")

	if x.Table != nil {
		out.WriteString(x.Table.String(maskParams))
	}
	if x.Alias != nil {
		out.WriteString(" AS ")
		out.WriteString(x.Alias.String(maskParams))
	}
	if len(x.Columns) > 0 {
		out.WriteString(" (")
		for i, c := range x.Columns {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(c.String(maskParams))
		}
		out.WriteString(")")
	}
	if x.Default {
		out.WriteString(" DEFAULT VALUES")
	}
	if len(x.Values) > 0 {
		out.WriteString(" VALUES ")
		for i, v := range x.Values {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString("(")
			for j, e := range v {
				if j > 0 {
					out.WriteString(", ")
				}
				out.WriteString(e.String(maskParams))
			}
			out.WriteString(")")
		}
	}
	if x.Query != nil {
		out.WriteString(" ")
		out.WriteString(x.Query.String(maskParams))
	}
	if len(x.ConflictTarget) > 0 {
		out.WriteString(" ON CONFLICT (")
		for i, c := range x.ConflictTarget {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(c.String(maskParams))
		}
		out.WriteString(")")
	}
	if x.ConflictAction != "" {
		out.WriteString(" DO ")
		out.WriteString(x.ConflictAction)
	}
	if len(x.ConflictUpdate) > 0 {
		out.WriteString(" SET ")
		for i, c := range x.ConflictUpdate {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(c.String(maskParams))
		}
	}
	if len(x.Returning) > 0 {
		out.WriteString(" RETURNING ")
		for i, c := range x.Returning {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(c.String(maskParams))
		}
	}
	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(x.Cast.String(maskParams))
	}
	out.WriteString(")")

	return out.String()
}
