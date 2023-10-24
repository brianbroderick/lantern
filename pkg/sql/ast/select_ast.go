package ast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type SelectStatement struct {
	Token       token.Token  `json:"token,omitempty"`       // the token.SELECT token
	Expressions []Expression `json:"expressions,omitempty"` // a select statement may consist of multiple expressions such as the with clause in a CTE along with the primary select expression
}

func (ss *SelectStatement) statementNode()       {}
func (ss *SelectStatement) TokenLiteral() string { return ss.Token.Lit }
func (ss *SelectStatement) String(maskParams bool) string {
	var out bytes.Buffer

	for _, e := range ss.Expressions {
		out.WriteString(e.String(maskParams))
	}

	out.WriteString(";")

	return out.String()
}
func (ss *SelectStatement) Inspect(maskParams bool) string {
	j, err := json.MarshalIndent(ss, "", "  ")
	if err != nil {
		fmt.Printf("Error loading data: %#v\n\n", err)
	}
	return string(j)
}

// SelectExpression is a select inside a SELECT or WITH (Common Table Expression) statement,
// since a select statement can have multiple select expressions. i.e. WITH clause, subqueries, and the primary select expression.
type SelectExpression struct {
	Token            token.Token  `json:"token,omitempty"`      // the token.SELECT token
	TempTable        Expression   `json:"temp_table,omitempty"` // the name of the temp table named in a WITH clause (CTE)
	WithMaterialized string       `json:"with_materialized,omitempty"`
	Distinct         Expression   `json:"distinct,omitempty"` // the DISTINCT or ALL token
	Columns          []Expression `json:"columns,omitempty"`
	Tables           []Expression `json:"tables,omitempty"`
	Where            Expression   `json:"where,omitempty"`
	GroupBy          []Expression `json:"group_by,omitempty"`
	Having           Expression   `json:"having,omitempty"`
	Window           []Expression `json:"window,omitempty"`
	OrderBy          []Expression `json:"order_by,omitempty"`
	Limit            Expression   `json:"limit,omitempty"`
	Offset           Expression   `json:"offset,omitempty"`
	Fetch            Expression   `json:"fetch,omitempty"`
	Lock             Expression   `json:"lock,omitempty"`
}

func (se *SelectExpression) expressionNode()      {}
func (se *SelectExpression) TokenLiteral() string { return se.Token.Lit }

// String() is incomplete and only returns the most basic of select statements
func (se *SelectExpression) String(maskParams bool) string {
	var out bytes.Buffer

	if se.TempTable != nil {
		out.WriteString(se.TempTable.String(maskParams) + " AS ")
	}

	if se.WithMaterialized != "" {
		out.WriteString(strings.ToUpper(se.WithMaterialized) + " ")
	}

	// Subqueries need to be surrounded by parentheses. A primary query may also have parentheses, so we'll add them here to be consistent.
	out.WriteString("(")

	out.WriteString(strings.ToUpper(se.TokenLiteral()) + " ")

	// Distinct
	if se.Distinct != nil {
		out.WriteString(se.Distinct.String(maskParams) + " ")
	}

	// Columns
	columns := []string{}
	for _, c := range se.Columns {
		columns = append(columns, c.String(maskParams))
	}
	out.WriteString(strings.Join(columns, ", "))

	// Tables
	out.WriteString(" FROM ")
	tables := []string{}
	for _, t := range se.Tables {
		tables = append(tables, t.String(maskParams))
	}
	out.WriteString(strings.Join(tables, " "))

	// Window
	if len(se.Window) > 0 {
		out.WriteString(" WINDOW ")
		windows := []string{}
		for _, w := range se.Window {
			windows = append(windows, w.String(maskParams))
		}
		out.WriteString(strings.Join(windows, ", "))
	}

	// Where
	if se.Where != nil {
		out.WriteString(" WHERE ")
		out.WriteString(se.Where.String(maskParams))
	}

	// Group By
	if len(se.GroupBy) > 0 {
		out.WriteString(" GROUP BY ")
		groupBy := []string{}
		for _, g := range se.GroupBy {
			groupBy = append(groupBy, g.String(maskParams))
		}
		out.WriteString(strings.Join(groupBy, ", "))
	}

	// Having
	if se.Having != nil {
		out.WriteString(" HAVING ")
		out.WriteString(se.Having.String(maskParams))
	}

	// Order By
	if len(se.OrderBy) > 0 {
		out.WriteString(" ORDER BY ")
		orderBy := []string{}
		for _, g := range se.OrderBy {
			orderBy = append(orderBy, g.String(maskParams))
		}
		out.WriteString(strings.Join(orderBy, ", "))
	}

	// Limit
	if se.Limit != nil {
		out.WriteString(" LIMIT ")
		out.WriteString(se.Limit.String(maskParams))
	}

	// Offset
	if se.Offset != nil {
		out.WriteString(" OFFSET ")
		out.WriteString(se.Offset.String(maskParams))
	}

	// Fetch
	if se.Fetch != nil {
		out.WriteString(" FETCH NEXT ")
		out.WriteString(se.Fetch.String(maskParams))
	}

	// Lock
	if se.Lock != nil {
		out.WriteString(" FOR ")
		out.WriteString(se.Lock.String(maskParams))
	}

	out.WriteString(")")

	return out.String()
}

// func (se *SelectExpression) Inspect() string {
// 	columns := []string{}
// 	for _, c := range se.Columns {
// 		columns = append(columns, c.String(maskParams))
// 	}
// 	strColumns := strings.Join(columns, "\n\t\t")
// 	strTables := []string{}
// 	for _, t := range se.Tables {
// 		strTables = append(strTables, t.String(maskParams))
// 	}

// 	ins := fmt.Sprintf("\tColumns: \n\t\t%s\n\n\tTable: \n\t\t%s\n", strColumns, strTables)
// 	return ins
// }

type DistinctExpression struct {
	Token token.Token  `json:"token,omitempty"` // The keyword token, e.g. DISTINCT
	Right []Expression `json:"right,omitempty"` // The columns to be distinct
}

func (de *DistinctExpression) expressionNode()      {}
func (de *DistinctExpression) TokenLiteral() string { return de.Token.Lit }
func (de *DistinctExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(strings.ToUpper(de.Token.Lit))
	if len(de.Right) > 0 {
		out.WriteString(" ON ")
	}
	if de.Right != nil {
		out.WriteString("(")
		right := []string{}
		for _, r := range de.Right {
			right = append(right, r.String(maskParams))
		}
		out.WriteString(strings.Join(right, ", "))
		out.WriteString(")")
	}

	return out.String()
}

type ColumnExpression struct {
	Token token.Token `json:"token,omitempty"` // the token.AS token
	Name  *Identifier `json:"name,omitempty"`  // the name of the column or alias
	Value Expression  `json:"value,omitempty"` // the complete expression including all of the columns
	// Columns []*Identifier `json:"columns,omitempty"` // the columns that make up the expression for ease of reporting
}

func (c *ColumnExpression) expressionNode()      {}
func (c *ColumnExpression) TokenLiteral() string { return c.Token.Lit }
func (c *ColumnExpression) String(maskParams bool) string {
	var out bytes.Buffer

	val := c.Value.String(maskParams)
	out.WriteString(val)
	if c.Name != nil && c.Name.String(maskParams) != val && c.Name.String(maskParams) != "" {
		out.WriteString(" AS ")
		out.WriteString(c.Name.String(maskParams))
	}

	return out.String()
}

type SortExpression struct {
	Token     token.Token `json:"token,omitempty"`     // the token.ASC or token.DESC token
	Value     Expression  `json:"value,omitempty"`     // the column to sort on
	Direction token.Token `json:"direction,omitempty"` // the direction to sort
	Nulls     token.Token `json:"nulls,omitempty"`     // first or last
}

func (s *SortExpression) expressionNode()      {}
func (s *SortExpression) TokenLiteral() string { return s.Token.Lit }
func (s *SortExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(s.Value.String(maskParams))
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
	Token token.Token `json:"token,omitempty"` // the token.FETCH token
	// Don't need to store "first" or "next" since these are synonyms. We'll convert everything to "next" when printing in .String()
	Value Expression `json:"value,omitempty"` // the number of rows to fetch
	// We also don't need to store "row" or "rows" since these are synonyms. We'll convert everything to "rows" when printing in .String()
	Option token.Token `json:"option,omitempty"` // the token.ONLY or token.TIES token (don't need to store "with ties", just "ties" will do)
}

func (f *FetchExpression) expressionNode()      {}
func (f *FetchExpression) TokenLiteral() string { return f.Token.Lit }
func (f *FetchExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(f.Value.String(maskParams))
	out.WriteString(" ROWS")
	if f.Option.Type == token.TIES {
		out.WriteString(" WITH TIES")
	} else {
		out.WriteString(" ONLY")
	}

	return out.String()
}

type WindowExpression struct {
	Token       token.Token  `json:"token,omitempty"`        // the token.OVER token
	Alias       *Identifier  `json:"alias,omitempty"`        // the alias of the window
	PartitionBy []Expression `json:"partition_by,omitempty"` // the columns to partition by
	OrderBy     []Expression `json:"order_by,omitempty"`     // the columns to order by
}

func (w *WindowExpression) expressionNode()      {}
func (w *WindowExpression) TokenLiteral() string { return w.Token.Lit }
func (w *WindowExpression) String(maskParams bool) string {
	var out bytes.Buffer

	if w.Alias != nil {
		out.WriteString(w.Alias.String(maskParams) + " AS ")
	}

	out.WriteString("(")
	if len(w.PartitionBy) > 0 {
		out.WriteString("PARTITION BY ")
		partitionBy := []string{}
		for _, p := range w.PartitionBy {
			partitionBy = append(partitionBy, p.String(maskParams))
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
			orderBy = append(orderBy, o.String(maskParams))
		}
		out.WriteString(strings.Join(orderBy, ", "))
	}

	out.WriteString(")")
	return out.String()
}

type WildcardLiteral struct {
	Token token.Token `json:"token,omitempty"` // the token.ASTERISK token
	Value string      `json:"value,omitempty"`
}

func (w *WildcardLiteral) expressionNode()               {}
func (w *WildcardLiteral) TokenLiteral() string          { return w.Token.Lit }
func (w *WildcardLiteral) String(maskParams bool) string { return w.Value }

// JoinType  Table     Alias JoinCondition
// source    customers c
// inner     addresses a     Expression

type TableExpression struct {
	Token         token.Token `json:"token,omitempty"`          // the token.JOIN token
	JoinType      string      `json:"join_type,omitempty"`      // the type of join: source, inner, left, right, full, etc
	Schema        string      `json:"schema,omitempty"`         // the name of the schema
	Table         Expression  `json:"table,omitempty"`          // the name of the table
	Alias         string      `json:"alias,omitempty"`          // the alias of the table
	JoinCondition Expression  `json:"join_condition,omitempty"` // the ON clause
}

func (t *TableExpression) expressionNode()      {}
func (t *TableExpression) TokenLiteral() string { return t.Token.Lit }
func (t *TableExpression) String(maskParams bool) string {
	var out bytes.Buffer

	// if t.JoinType != "" && t.JoinType != "CROSS" {
	// 	out.WriteString(t.JoinType + " JOIN ")
	// }

	// if t.JoinType == "CROSS" {
	// 	out.WriteString(", ")
	// }

	if t.JoinType != "" {
		out.WriteString(t.JoinType + " ")
	}

	out.WriteString(t.Table.String(maskParams))
	if t.Alias != "" {
		out.WriteString(" " + t.Alias)
	}

	if t.JoinCondition != nil {
		out.WriteString(" ON " + t.JoinCondition.String(maskParams))
	}

	return out.String()
}

type LockExpression struct {
	Token   token.Token  `json:"token,omitempty"`   // the token.FOR token
	Lock    string       `json:"lock,omitempty"`    // the type of lock: update, share, key share, no key update
	Tables  []Expression `json:"tables,omitempty"`  // the tables to lock
	Options string       `json:"options,omitempty"` // the options: NOWAIT, SKIP LOCKED
}

func (l *LockExpression) expressionNode()      {}
func (l *LockExpression) TokenLiteral() string { return l.Token.Lit }
func (l *LockExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(l.Lock)
	if len(l.Tables) > 0 {
		tables := []string{}
		out.WriteString(" OF ")
		for _, t := range l.Tables {
			tables = append(tables, t.String(maskParams))
		}
		out.WriteString(strings.Join(tables, ", "))
	}
	if l.Options != "" {
		out.WriteString(" " + l.Options)
	}

	return out.String()
}

type InExpression struct {
	Token    token.Token  `json:"token,omitempty"` // The operator token, e.g. IN, NOT IN
	Left     Expression   `json:"left,omitempty"`
	Operator string       `json:"operator,omitempty"`
	Right    []Expression `json:"right,omitempty"`
}

func (ie *InExpression) expressionNode()      {}
func (ie *InExpression) TokenLiteral() string { return ie.Token.Lit }
func (ie *InExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(ie.Left.String(maskParams))
	out.WriteString(" " + strings.ToUpper(ie.Operator) + " ")

	out.WriteString("(")
	oneLess := len(ie.Right) - 1
	for i, e := range ie.Right {
		out.WriteString(e.String(maskParams))
		if i < oneLess {
			out.WriteString(", ")
		}
	}
	out.WriteString(")")

	return out.String()
}

type UnionExpression struct {
	Token token.Token `json:"token,omitempty"` // The operator token, e.g. UNION, INTERSECT, EXCEPT
	Right Expression  `json:"right,omitempty"`
}

func (ue *UnionExpression) expressionNode()      {}
func (ue *UnionExpression) TokenLiteral() string { return ue.Token.Lit }
func (ue *UnionExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(strings.ToUpper(" " + ue.Token.Lit + " "))
	out.WriteString(ue.Right.String(maskParams))

	return out.String()
}
