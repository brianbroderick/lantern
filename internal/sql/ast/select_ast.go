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
	Tables  []Expression
	Where   Expression
	GroupBy []Expression
	Having  Expression
	OrderBy []Expression
	Limit   Expression
	Offset  Expression
	// Offset  *IntegerLiteral
}

func (ls *SelectStatement) statementNode()       {}
func (ls *SelectStatement) TokenLiteral() string { return ls.Token.Lit }

// String() is incomplete and only returns the most basic of select statements
func (ls *SelectStatement) String() string {
	var out bytes.Buffer

	// Columns
	out.WriteString(strings.ToUpper(ls.TokenLiteral()) + " ")
	columns := []string{}
	for _, c := range ls.Columns {
		columns = append(columns, c.String())
	}
	out.WriteString(strings.Join(columns, ", "))

	// Tables
	out.WriteString(" FROM ")
	tables := []string{}
	for _, t := range ls.Tables {
		tables = append(tables, t.String())
	}
	out.WriteString(strings.Join(tables, " "))

	// Where
	if ls.Where != nil {
		out.WriteString(" WHERE ")
		out.WriteString(ls.Where.String())
	}

	// Group By
	if len(ls.GroupBy) > 0 {
		out.WriteString(" GROUP BY ")
		groupBy := []string{}
		for _, g := range ls.GroupBy {
			groupBy = append(groupBy, g.String())
		}
		out.WriteString(strings.Join(groupBy, ", "))
	}

	// Having
	if ls.Having != nil {
		out.WriteString(" HAVING ")
		out.WriteString(ls.Having.String())
	}

	// Order By
	if len(ls.OrderBy) > 0 {
		out.WriteString(" ORDER BY ")
		orderBy := []string{}
		for _, g := range ls.OrderBy {
			orderBy = append(orderBy, g.String())
		}
		out.WriteString(strings.Join(orderBy, ", "))
	}

	// Limit
	if ls.Limit != nil {
		out.WriteString(" LIMIT ")
		out.WriteString(ls.Limit.String())
	}

	// Offset
	if ls.Offset != nil {
		out.WriteString(" OFFSET ")
		out.WriteString(ls.Offset.String())
	}

	out.WriteString(";")

	return out.String()
}

func (ls *SelectStatement) Inspect() string {
	columns := []string{}
	for _, c := range ls.Columns {
		columns = append(columns, c.String())
	}
	strColumns := strings.Join(columns, "\n\t\t")
	strTables := []string{}
	for _, t := range ls.Tables {
		strTables = append(strTables, t.String())
	}

	ins := fmt.Sprintf("\tColumns: \n\t\t%s\n\n\tTable: \n\t\t%s\n", strColumns, strTables)
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

type WildcardLiteral struct {
	Token token.Token // the token.ASTERISK token
	Value string
}

func (w *WildcardLiteral) expressionNode()      {}
func (w *WildcardLiteral) TokenLiteral() string { return w.Token.Lit }
func (w *WildcardLiteral) String() string       { return w.Value }

// JoinType  Table     Alias JoinCondition
// source    customers c
// inner     addresses a     Expression

type TableExpression struct {
	Token         token.Token // the token.JOIN token
	JoinType      string      // the type of join: source, inner, left, right, full, etc
	Schema        string      // the name of the schema
	Table         string      // the name of the table
	Alias         string      // the alias of the table
	JoinCondition Expression  // the ON clause
}

func (t *TableExpression) expressionNode()      {}
func (t *TableExpression) TokenLiteral() string { return t.Token.Lit }
func (t *TableExpression) String() string {
	var out bytes.Buffer

	if t.JoinType != "" && t.JoinType != "source" {
		out.WriteString(t.JoinType + " JOIN ")
	}

	out.WriteString(t.Table)
	if t.Alias != "" {
		out.WriteString(" " + t.Alias)
	}

	if t.JoinCondition != nil {
		out.WriteString(" ON " + t.JoinCondition.String())
	}

	return out.String()
}
