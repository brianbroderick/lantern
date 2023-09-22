package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/internal/sql/token"
)

// Most of these need to be replaced by structs
type SelectStatement struct {
	Token    token.Token // the token.SELECT token
	Distinct Expression  // the DISTINCT or ALL token
	Columns  []Expression
	Tables   []Expression
	Where    Expression
	GroupBy  []Expression
	Having   Expression
	OrderBy  []Expression
	Limit    Expression
	Offset   Expression
}

func (ls *SelectStatement) statementNode()       {}
func (ls *SelectStatement) TokenLiteral() string { return ls.Token.Lit }

// String() is incomplete and only returns the most basic of select statements
func (ls *SelectStatement) String() string {
	var out bytes.Buffer

	out.WriteString(strings.ToUpper(ls.TokenLiteral()) + " ")

	// Distinct
	if ls.Distinct != nil {
		out.WriteString(ls.Distinct.String() + " ")
	}

	// Columns
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

type DistinctExpression struct {
	Token token.Token  // The keyword token, e.g. DISTINCT
	Right []Expression // The columns to be distinct
}

func (de *DistinctExpression) expressionNode()      {}
func (de *DistinctExpression) TokenLiteral() string { return de.Token.Lit }
func (de *DistinctExpression) String() string {
	var out bytes.Buffer

	out.WriteString(strings.ToUpper(de.Token.Lit))
	if len(de.Right) > 0 {
		out.WriteString(" ON ")
	}
	if de.Right != nil {
		out.WriteString("(")
		right := []string{}
		for _, r := range de.Right {
			right = append(right, r.String())
		}
		out.WriteString(strings.Join(right, ", "))
		out.WriteString(")")
	}

	return out.String()
}

type ColumnExpression struct {
	Token token.Token // the token.AS token
	Name  *Identifier // the name of the column or alias
	Value Expression  // the complete expression including all of the columns
	// Columns []*Identifier // the columns that make up the expression for ease of reporting
}

func (c *ColumnExpression) expressionNode()      {}
func (c *ColumnExpression) TokenLiteral() string { return c.Token.Lit }
func (c *ColumnExpression) String() string {
	var out bytes.Buffer

	val := c.Value.String()
	out.WriteString(val)
	if c.Name != nil && c.Name.String() != val && c.Name.String() != "" {
		out.WriteString(" AS ")
		out.WriteString(c.Name.String())
	}

	return out.String()
}

type SortExpression struct {
	Token     token.Token // the token.ASC or token.DESC token
	Value     Expression  // the column to sort on
	Direction token.Token // the direction to sort
	Nulls     token.Token // first or last
}

func (s *SortExpression) expressionNode()      {}
func (s *SortExpression) TokenLiteral() string { return s.Token.Lit }
func (s *SortExpression) String() string {
	var out bytes.Buffer

	out.WriteString(s.Value.String())
	if s.Direction.Type == token.DESC {
		out.WriteString(" DESC")
	}
	if s.Nulls.Type == token.FIRST {
		out.WriteString(" NULLS FIRST")
	} else if s.Nulls.Type == token.LAST {
		out.WriteString(" NULLS LAST")
	}

	return out.String()
}

type WindowExpression struct {
	Token       token.Token  // the token.OVER token
	PartitionBy []Expression // the columns to partition by
	OrderBy     []Expression // the columns to order by
}

func (w *WindowExpression) expressionNode()      {}
func (w *WindowExpression) TokenLiteral() string { return w.Token.Lit }
func (w *WindowExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	if len(w.PartitionBy) > 0 {
		out.WriteString("PARTITION BY ")
		partitionBy := []string{}
		for _, p := range w.PartitionBy {
			partitionBy = append(partitionBy, p.String())
		}
		out.WriteString(strings.Join(partitionBy, ", "))
	}
	if len(w.PartitionBy) > 0 && len(w.OrderBy) > 0 {
		out.WriteString(" ")
	}
	if len(w.OrderBy) > 0 {
		out.WriteString("ORDER BY ")
		orderBy := []string{}
		for _, o := range w.OrderBy {
			orderBy = append(orderBy, o.String())
		}
		out.WriteString(strings.Join(orderBy, ", "))
	}

	out.WriteString(")")
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
