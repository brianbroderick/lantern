package ast

import (
	"bytes"
	"strings"

	"github.com/brianbroderick/lantern/internal/sql/token"
)

type SelectStatement struct {
	Token       token.Token  // the token.SELECT token
	Expressions []Expression // a select statement may consist of multiple expressions such as the with clause in a CTE along with the primary select expression
}

func (ss *SelectStatement) statementNode()       {}
func (ss *SelectStatement) TokenLiteral() string { return ss.Token.Lit }
func (ss *SelectStatement) String() string {
	var out bytes.Buffer

	for _, e := range ss.Expressions {
		out.WriteString(e.String())
	}

	out.WriteString(";")

	return out.String()
}
func (ss *SelectStatement) Inspect() string {
	// columns := []string{}
	// for _, c := range ss.Columns {
	// 	columns = append(columns, c.String())
	// }
	// strColumns := strings.Join(columns, "\n\t\t")
	// strTables := []string{}
	// for _, t := range ss.Tables {
	// 	strTables = append(strTables, t.String())
	// }

	// ins := fmt.Sprintf("\tColumns: \n\t\t%s\n\n\tTable: \n\t\t%s\n", strColumns, strTables)
	// return ins
	return "Inspect method not implemented"
}

// SelectExpression is a select inside a SELECT or WITH (Common Table Expression) statement,
// since a select statement can have multiple select expressions. i.e. WITH clause, subqueries, and the primary select expression.
type SelectExpression struct {
	Token     token.Token // the token.SELECT token
	TempTable *Identifier // the name of the temp table named in a WITH clause (CTE)
	Distinct  Expression  // the DISTINCT or ALL token
	Columns   []Expression
	Tables    []Expression
	Where     Expression
	GroupBy   []Expression
	Having    Expression
	OrderBy   []Expression
	Limit     Expression
	Offset    Expression
	Fetch     Expression
	Lock      Expression
}

func (se *SelectExpression) expressionNode()      {}
func (se *SelectExpression) TokenLiteral() string { return se.Token.Lit }

// String() is incomplete and only returns the most basic of select statements
func (se *SelectExpression) String() string {
	var out bytes.Buffer

	// Subqueries need to be surrounded by parentheses. A primary query may also have parentheses, so we'll add them here to be consistent.
	out.WriteString("(")

	out.WriteString(strings.ToUpper(se.TokenLiteral()) + " ")

	// Distinct
	if se.Distinct != nil {
		out.WriteString(se.Distinct.String() + " ")
	}

	// Columns
	columns := []string{}
	for _, c := range se.Columns {
		columns = append(columns, c.String())
	}
	out.WriteString(strings.Join(columns, ", "))

	// Tables
	out.WriteString(" FROM ")
	tables := []string{}
	for _, t := range se.Tables {
		tables = append(tables, t.String())
	}
	out.WriteString(strings.Join(tables, " "))

	// Where
	if se.Where != nil {
		out.WriteString(" WHERE ")
		out.WriteString(se.Where.String())
	}

	// Group By
	if len(se.GroupBy) > 0 {
		out.WriteString(" GROUP BY ")
		groupBy := []string{}
		for _, g := range se.GroupBy {
			groupBy = append(groupBy, g.String())
		}
		out.WriteString(strings.Join(groupBy, ", "))
	}

	// Having
	if se.Having != nil {
		out.WriteString(" HAVING ")
		out.WriteString(se.Having.String())
	}

	// Order By
	if len(se.OrderBy) > 0 {
		out.WriteString(" ORDER BY ")
		orderBy := []string{}
		for _, g := range se.OrderBy {
			orderBy = append(orderBy, g.String())
		}
		out.WriteString(strings.Join(orderBy, ", "))
	}

	// Limit
	if se.Limit != nil {
		out.WriteString(" LIMIT ")
		out.WriteString(se.Limit.String())
	}

	// Offset
	if se.Offset != nil {
		out.WriteString(" OFFSET ")
		out.WriteString(se.Offset.String())
	}

	// Fetch
	if se.Fetch != nil {
		out.WriteString(" FETCH NEXT ")
		out.WriteString(se.Fetch.String())
	}

	// Lock
	if se.Lock != nil {
		out.WriteString(" FOR ")
		out.WriteString(se.Lock.String())
	}

	out.WriteString(")")

	return out.String()
}

// func (se *SelectExpression) Inspect() string {
// 	columns := []string{}
// 	for _, c := range se.Columns {
// 		columns = append(columns, c.String())
// 	}
// 	strColumns := strings.Join(columns, "\n\t\t")
// 	strTables := []string{}
// 	for _, t := range se.Tables {
// 		strTables = append(strTables, t.String())
// 	}

// 	ins := fmt.Sprintf("\tColumns: \n\t\t%s\n\n\tTable: \n\t\t%s\n", strColumns, strTables)
// 	return ins
// }

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

type FetchExpression struct {
	Token token.Token // the token.FETCH token
	// Don't need to store "first" or "next" since these are synonyms. We'll convert everything to "next" when printing in .String()
	Value Expression // the number of rows to fetch
	// We also don't need to store "row" or "rows" since these are synonyms. We'll convert everything to "rows" when printing in .String()
	Option token.Token // the token.ONLY or token.TIES token (don't need to store "with ties", just "ties" will do)
}

func (f *FetchExpression) expressionNode()      {}
func (f *FetchExpression) TokenLiteral() string { return f.Token.Lit }
func (f *FetchExpression) String() string {
	var out bytes.Buffer

	out.WriteString(f.Value.String())
	out.WriteString(" ROWS")
	if f.Option.Type == token.TIES {
		out.WriteString(" WITH TIES")
	} else {
		out.WriteString(" ONLY")
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

type LockExpression struct {
	Token   token.Token  // the token.FOR token
	Lock    string       // the type of lock: update, share, key share, no key update
	Tables  []Expression // the tables to lock
	Options string       // the options: NOWAIT, SKIP LOCKED
}

func (l *LockExpression) expressionNode()      {}
func (l *LockExpression) TokenLiteral() string { return l.Token.Lit }
func (l *LockExpression) String() string {
	var out bytes.Buffer
	out.WriteString(l.Lock)
	if len(l.Tables) > 0 {
		tables := []string{}
		out.WriteString(" OF ")
		for _, t := range l.Tables {
			tables = append(tables, t.String())
		}
		out.WriteString(strings.Join(tables, ", "))
	}
	if l.Options != "" {
		out.WriteString(" " + l.Options)
	}

	return out.String()
}

type InExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    []Expression
}

func (ie *InExpression) expressionNode()      {}
func (ie *InExpression) TokenLiteral() string { return ie.Token.Lit }
func (ie *InExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ie.Left.String())
	out.WriteString(" " + strings.ToUpper(ie.Operator) + " ")

	out.WriteString("(")
	oneLess := len(ie.Right) - 1
	for i, e := range ie.Right {
		out.WriteString(e.String())
		if i < oneLess {
			out.WriteString(", ")
		}
	}
	out.WriteString(")")

	return out.String()
}
