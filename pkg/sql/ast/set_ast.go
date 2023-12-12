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

func (s *SetStatement) statementNode()       {}
func (s *SetStatement) TokenLiteral() string { return s.Token.Upper }
func (s *SetStatement) String(maskParams bool, alias map[string]string) string {
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

	out.WriteString(s.Expression.String(maskParams, alias))

	out.WriteString(";")

	return out.String()
}
func (s *SetStatement) Inspect(maskParams bool, alias map[string]string) string {
	j, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Printf("Error loading data: %#v\n\n", err)
	}
	return string(j)
}
