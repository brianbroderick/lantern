package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/internal/sql/token"
)

// Most of these need to be replaced by structs
type SelectStatement struct {
	Token   token.Token // the token.SELECT token
	Columns []Expression
	From    *Identifier
	Joins   []Expression
	// Joins   []*Identifier
	// Where   []*Identifier
	// GroupBy []*Identifier
	// Having  []*Identifier
	// OrderBy []*Identifier
	// Limit   *IntegerLiteral
	// Offset  *IntegerLiteral
}

func (ls *SelectStatement) statementNode()       {}
func (ls *SelectStatement) TokenLiteral() string { return ls.Token.Lit }

// String() is incomplete and only returns the most basic of select statements
func (ls *SelectStatement) String() string {
	var out bytes.Buffer

	out.WriteString(strings.ToUpper(ls.TokenLiteral()) + " ")
	columns := []string{}
	for _, c := range ls.Columns {
		columns = append(columns, c.String())
	}
	out.WriteString(strings.Join(columns, ", "))
	// columns := []string{}
	// for _, c := range ls.Columns {
	// 	columns = append(columns, c.String())
	// }
	// out.WriteString(strings.Join(columns, ", "))
	out.WriteString(" FROM ")
	out.WriteString(ls.From.String())
	out.WriteString(";")

	return out.String()
}

func (ls *SelectStatement) Inspect() string {
	columns := []string{}
	for _, c := range ls.Columns {
		columns = append(columns, c.String())
	}
	strColumns := strings.Join(columns, "\n\t\t")

	ins := fmt.Sprintf("\tColumns: \n\t\t%s\n\n\tTable: \n\t\t%s\n", strColumns, ls.From.String())
	return ins
}

type ColumnExpression struct {
	Token   token.Token   // the token.AS token
	Name    *Identifier   // the name of the column or alias
	Value   Expression    // the complete expression including all of the columns
	Columns []*Identifier // the columns that make up the expression for ease of reporting
}

func (c *ColumnExpression) expressionNode()      {}
func (c *ColumnExpression) TokenLiteral() string { return c.Token.Lit }
func (c *ColumnExpression) String() string {
	var out bytes.Buffer

	val := c.Value.String()
	out.WriteString(val)
	if c.Name.String() != val && c.Name.String() != "" {
		out.WriteString(" AS ")
		out.WriteString(c.Name.String())
	}

	return out.String()
}
