package ast

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/brianbroderick/lantern/internal/sql/token"
)

type CTEStatement struct {
	Token       token.Token  `json:"token,omitempty"`
	Recursive   bool         `json:"recursive,omitempty"`
	Expressions []Expression `json:"expressions,omitempty"`
}

func (s *CTEStatement) statementNode()       {}
func (s *CTEStatement) TokenLiteral() string { return s.Token.Lit }
func (s *CTEStatement) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("(WITH ")

	if s.Recursive {
		out.WriteString("RECURSIVE ")
	}

	lenExpressions := len(s.Expressions) - 1
	for i, e := range s.Expressions {
		out.WriteString(e.String(maskParams))
		if i < lenExpressions-1 {
			out.WriteString(", ")
		} else if i == lenExpressions-1 {
			out.WriteString(" ")
		}
	}

	out.WriteString(");")

	return out.String()
}
func (s *CTEStatement) Inspect(maskParams bool) string {
	j, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Printf("Error loading data: %#v\n\n", err)
	}
	return string(j)
}
