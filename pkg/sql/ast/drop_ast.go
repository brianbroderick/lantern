package ast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type DropStatement struct {
	Token   token.Token  `json:"token,omitempty"` // the token.DROP token
	Object  token.Token  `json:"object,omitempty"`
	Exists  bool         `json:"exists,omitempty"`
	Tables  []Expression `json:"expression,omitempty"`
	Options string       `json:"options,omitempty"`
}

func (x *DropStatement) Clause() token.TokenType  { return x.Token.Type }
func (x *DropStatement) Command() token.TokenType { return x.Token.Type }
func (x *DropStatement) statementNode()           {}
func (x *DropStatement) TokenLiteral() string     { return x.Token.Lit }
func (x *DropStatement) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("(DROP %s ", x.Object.Type.String()))
	if x.Exists {
		out.WriteString("IF EXISTS ")
	}
	if len(x.Tables) > 0 {
		for i, t := range x.Tables {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(t.String(maskParams))
		}
	}

	if x.Options != "" {
		out.WriteString(" " + strings.ToUpper(x.Options))
	}

	out.WriteString(");")

	return out.String()
}
func (x *DropStatement) Inspect(maskParams bool) string {
	j, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling data: %#v\n\n", err)
	}
	return string(j)
}
