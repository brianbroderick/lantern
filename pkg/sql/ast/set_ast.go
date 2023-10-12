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

func (ss *SetStatement) statementNode()       {}
func (ss *SetStatement) TokenLiteral() string { return ss.Token.Lit }
func (ss *SetStatement) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("SET ")

	if ss.Session {
		out.WriteString("SESSION ")
	} else if ss.Local {
		out.WriteString("LOCAL ")
	}

	if ss.TimeZone {
		out.WriteString("TIME ZONE ")
	}

	out.WriteString(ss.Expression.String(maskParams))

	out.WriteString(";")

	return out.String()
}
func (ss *SetStatement) Inspect(maskParams bool) string {
	j, err := json.MarshalIndent(ss, "", "  ")
	if err != nil {
		fmt.Printf("Error loading data: %#v\n\n", err)
	}
	return string(j)
}
