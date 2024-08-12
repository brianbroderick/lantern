package ast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type SetStatement struct {
	Token              token.Token  `json:"token,omitempty"` // the token.SET token
	Session            bool         `json:"session,omitempty"`
	HasCharacteristics bool         `json:"has_characteristics,omitempty"`
	Local              bool         `json:"local,omitempty"`
	TimeZone           bool         `json:"timeZone,omitempty"`
	Expression         Expression   `json:"expression,omitempty"`
	HasConstraints     bool         `json:"has_constraints,omitempty"`
	IsAll              bool         `json:"is_all,omitempty"`
	Constraints        []Expression `json:"constraints,omitempty"`
	ConstraintSetting  string       `json:"constraint_setting,omitempty"`
}

func (s *SetStatement) Clause() token.TokenType      { return s.Token.Type }
func (x *SetStatement) SetClause(c token.TokenType)  {}
func (s *SetStatement) Command() token.TokenType     { return s.Token.Type }
func (x *SetStatement) SetCommand(c token.TokenType) {}
func (s *SetStatement) statementNode()               {}
func (s *SetStatement) TokenLiteral() string         { return s.Token.Upper }
func (s *SetStatement) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("SET ")

	if s.Session {
		out.WriteString("SESSION ")
	} else if s.Local {
		out.WriteString("LOCAL ")
	}

	if s.TimeZone {
		out.WriteString("TIME ZONE ")
	}

	if s.HasCharacteristics {
		out.WriteString("CHARACTERISTICS AS ")
	}

	if s.Expression != nil {
		out.WriteString(s.Expression.String(maskParams))
	}

	if s.HasConstraints {
		out.WriteString("CONSTRAINTS ")
	}

	if s.IsAll {
		out.WriteString("ALL ")
	}

	if len(s.Constraints) > 0 {
		for i, c := range s.Constraints {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(c.String(maskParams))
		}
		out.WriteString(" ")
	}

	if s.ConstraintSetting != "" {
		out.WriteString(strings.ToUpper(s.ConstraintSetting))
	}

	out.WriteString(";")

	return out.String()
}
func (s *SetStatement) Inspect(maskParams bool) string {
	j, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Printf("Error loading data: %#v\n\n", err)
	}
	return string(j)
}

type TransactionExpression struct {
	Token          token.Token     `json:"token,omitempty"` // the token.SET token
	IsolationLevel string          `json:"isolationLevel,omitempty"`
	Rights         string          `json:"rights,omitempty"`
	Deferrable     string          `json:"deferrable,omitempty"`
	Cast           Expression      `json:"cast,omitempty"`
	Branch         token.TokenType `json:"clause,omitempty"`  // location in the tree representing a clause
	CommandTag     token.TokenType `json:"command,omitempty"` // location in the tree representing a command
}

func (t *TransactionExpression) Clause() token.TokenType      { return t.Token.Type }
func (x *TransactionExpression) SetClause(c token.TokenType)  {}
func (t *TransactionExpression) Command() token.TokenType     { return t.Token.Type }
func (x *TransactionExpression) SetCommand(c token.TokenType) {}
func (t *TransactionExpression) expressionNode()              {}
func (t *TransactionExpression) TokenLiteral() string         { return t.Token.Upper }
func (x *TransactionExpression) SetCast(cast Expression) {
	x.Cast = cast
}
func (t *TransactionExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("TRANSACTION ISOLATION LEVEL ")
	out.WriteString(t.IsolationLevel)

	if t.Rights != "" {
		out.WriteString(" ")
		out.WriteString(t.Rights)
	}

	if t.Deferrable != "" {
		out.WriteString(" ")
		out.WriteString(t.Deferrable)
	}

	if t.Cast != nil {
		out.WriteString(" ")
		out.WriteString(t.Cast.String(maskParams))
	}

	return out.String()
}
