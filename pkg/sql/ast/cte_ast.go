package ast

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/token"
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

	// CTEs are a little different. We need to split the expressions into two groups. The first group is the CTEs consisting of temp tables
	// and the second group is the main query. The main query is the last statement, but that can consist of multiple
	// expressions such as SELECT. We need to split them up and put the CTEs first so that we can comma separate the query appropriately.

	tmpTables := []Expression{}
	primaryExpressions := []Expression{}

	inCTE := true

	for _, e := range s.Expressions {
		if e != nil {
			if stmt, ok := e.(*SelectExpression); ok {
				if stmt.TempTable == nil {
					inCTE = false
				}
			}
			if inCTE {
				tmpTables = append(tmpTables, e)
			} else {
				primaryExpressions = append(primaryExpressions, e)
			}
		}
	}

	lenTmpTables := len(tmpTables) - 1
	for i, e := range tmpTables {
		out.WriteString(e.String(maskParams))
		if i < lenTmpTables {
			out.WriteString(", ")
		} else if i == lenTmpTables {
			out.WriteString(" ")
		}
	}

	for _, e := range primaryExpressions {
		out.WriteString(e.String(maskParams))
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
