package ast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type AnalyzeStatement struct {
	Token            token.Token  `json:"token,omitempty"`              // the token.ANALYZE token
	Verbose          bool         `json:"verbose,omitempty"`            // VERBOSE
	SkipLocked       bool         `json:"skip_locked,omitempty"`        // SKIP_LOCKED
	BufferUsageLimit []Expression `json:"buffer_usage_limit,omitempty"` // BUFFER_USAGE_LIMIT <expr> [KB | MB | GB]
	Name             Expression   `json:"name,omitempty"`               // the name of the object
}

func (x *AnalyzeStatement) Clause() token.TokenType     { return x.Token.Type }
func (x *AnalyzeStatement) SetClause(c token.TokenType) {}
func (x *AnalyzeStatement) Command() token.TokenType    { return x.Token.Type }
func (x *AnalyzeStatement) statementNode()              {}
func (x *AnalyzeStatement) TokenLiteral() string        { return x.Token.Lit }
func (x *AnalyzeStatement) String(maskParams bool) string {
	var out bytes.Buffer
	count := 0
	parens := false
	if x.Verbose {
		count++
	}
	if x.SkipLocked {
		count++
	}
	if x.BufferUsageLimit != nil {
		count++
	}

	out.WriteString("ANALYZE ")
	if count > 0 {
		out.WriteString("(")
		parens = true
	}
	if x.Verbose {
		out.WriteString("VERBOSE")
		count--
		if count > 0 {
			out.WriteString(", ")
		}
	}

	if x.SkipLocked {
		out.WriteString("SKIP_LOCKED")
		count--
		if count > 0 {
			out.WriteString(", ")
		}
	}

	if len(x.BufferUsageLimit) != 0 {
		out.WriteString("BUFFER_USAGE_LIMIT ")
		for _, e := range x.BufferUsageLimit {
			out.WriteString(strings.ToUpper(e.String(maskParams)))
		}
	}
	if parens {
		out.WriteString(") ")
	}

	if x.Name != nil {
		out.WriteString(x.Name.String(maskParams))
	}
	out.WriteString(";")

	return out.String()
}

func (x *AnalyzeStatement) Inspect(maskParams bool) string {
	j, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling data: %#v\n\n", err)
	}
	return string(j)
}
