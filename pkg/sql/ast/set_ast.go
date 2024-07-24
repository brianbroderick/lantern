package ast

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type SetStatement struct {
	Token      token.Token `json:"token,omitempty"` // the token.SET token
	Session    bool        `json:"session,omitempty"`
	Local      bool        `json:"local,omitempty"`
	TimeZone   bool        `json:"timeZone,omitempty"`
	Expression Expression  `json:"expression,omitempty"`
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

	out.WriteString(s.Expression.String(maskParams))

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
