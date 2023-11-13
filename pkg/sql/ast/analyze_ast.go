package ast

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type AnalyzeStatement struct {
	Token token.Token `json:"token,omitempty"` // the token.ANALYZE token
	Name  Expression  `json:"name,omitempty"`  // the name of the object
}

func (x *AnalyzeStatement) statementNode()       {}
func (x *AnalyzeStatement) TokenLiteral() string { return x.Token.Lit }
func (x *AnalyzeStatement) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString("(ANALYZE ")
	if x.Name != nil {
		out.WriteString(x.Name.String(maskParams))
	}
	out.WriteString(");")

	return out.String()
}

func (x *AnalyzeStatement) Inspect(maskParams bool) string {
	j, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling data: %#v\n\n", err)
	}
	return string(j)
}
