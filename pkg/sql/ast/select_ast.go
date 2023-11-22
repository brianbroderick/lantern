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
	Token                 token.Token  `json:"token,omitempty"`      // the token.SELECT token
	TempTable             Expression   `json:"temp_table,omitempty"` // the name of the temp table named in a WITH clause (CTE)
	WithMaterialized      string       `json:"with_materialized,omitempty"`
	Distinct              Expression   `json:"distinct,omitempty"` // the DISTINCT or ALL token
	Columns               []Expression `json:"columns,omitempty"`
	Tables                []Expression `json:"tables,omitempty"`
	Where                 Expression   `json:"where,omitempty"`
	GroupBy               []Expression `json:"group_by,omitempty"`
	Having                Expression   `json:"having,omitempty"`
	Window                []Expression `json:"window,omitempty"`
	OrderBy               []Expression `json:"order_by,omitempty"`
	Limit                 Expression   `json:"limit,omitempty"`
	Offset                Expression   `json:"offset,omitempty"`
	Fetch                 Expression   `json:"fetch,omitempty"`
	Lock                  Expression   `json:"lock,omitempty"`
	CompoundToken         token.Token  `json:"compound_token,omitempty"`          // the token.UNION, token.INTERSECT, or token.EXCEPT token
	CompoundTokenModifier token.Token  `json:"compound_token_modifier,omitempty"` // the token.ALL
	CompoundExpression    Expression   `json:"union,omitempty"`                   // the select expression to union with
	Cast                  string       `json:"cast,omitempty"`                    // probably not needed, but used for the interface
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

	// Compound
	if se.CompoundExpression != nil {
		out.WriteString(" " + strings.ToUpper(se.CompoundToken.Lit) + " ")
		if se.CompoundTokenModifier.Type == token.ALL {
			out.WriteString(strings.ToUpper(se.CompoundTokenModifier.Lit) + " ")
		}
		out.WriteString(se.CompoundExpression.String(maskParams))
	}

	out.WriteString(")")

	if se.Cast != "" {
		out.WriteString("::")
		out.WriteString(se.Cast)
	}

	return out.String()
}
func (se *SelectExpression) SetCast(cast string) {
	se.Cast = cast
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
	Cast  string       `json:"cast,omitempty"`  // probably not needed, but used for the interface
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

	if de.Cast != "" {
		out.WriteString("::")
		out.WriteString(de.Cast)
	}

	return out.String()
}
func (de *DistinctExpression) SetCast(cast string) {
	de.Cast = cast
}

type ColumnExpression struct {
	Token token.Token `json:"token,omitempty"` // the token.AS token
	Name  *Identifier `json:"name,omitempty"`  // the name of the column or alias
	Value Expression  `json:"value,omitempty"` // the complete expression including all of the columns
	Cast  string      `json:"cast,omitempty"`
	// Columns []*Identifier `json:"columns,omitempty"` // the columns that make up the expression for ease of reporting
}

func (x *ColumnExpression) expressionNode()      {}
func (x *ColumnExpression) TokenLiteral() string { return x.Token.Lit }
func (x *ColumnExpression) String(maskParams bool) string {
	var out bytes.Buffer

	val := x.Value.String(maskParams)
	out.WriteString(val)
	if x.Name != nil && x.Name.String(maskParams) != val && x.Name.String(maskParams) != "" {
		out.WriteString(" AS ")
		out.WriteString(x.Name.String(maskParams))
	}

	if x.Cast != "" {
		out.WriteString("::")
		out.WriteString(x.Cast)
	}

	return out.String()
}
func (x *ColumnExpression) SetCast(cast string) {
	x.Cast = cast
}

type SortExpression struct {
	Token     token.Token `json:"token,omitempty"`     // the token.ASC or token.DESC token
	Value     Expression  `json:"value,omitempty"`     // the column to sort on
	Direction token.Token `json:"direction,omitempty"` // the direction to sort
	Nulls     token.Token `json:"nulls,omitempty"`     // first or last
	Cast      string      `json:"cast,omitempty"`
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

	if s.Cast != "" {
		out.WriteString("::")
		out.WriteString(s.Cast)
	}

	return out.String()
}
func (s *SortExpression) SetCast(cast string) {
	s.Cast = cast
}

type FetchExpression struct {
	Token token.Token `json:"token,omitempty"` // the token.FETCH token
	// Don't need to store "first" or "next" since these are synonyms. We'll convert everything to "next" when printing in .String()
	Value Expression `json:"value,omitempty"` // the number of rows to fetch
	// We also don't need to store "row" or "rows" since these are synonyms. We'll convert everything to "rows" when printing in .String()
	Option token.Token `json:"option,omitempty"` // the token.ONLY or token.TIES token (don't need to store "with ties", just "ties" will do)
	Cast   string      `json:"cast,omitempty"`
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

	if f.Cast != "" {
		out.WriteString("::")
		out.WriteString(f.Cast)
	}

	return out.String()
}
func (f *FetchExpression) SetCast(cast string) {
	f.Cast = cast
}

type AggregateExpression struct {
	Token    token.Token  `json:"token,omitempty"`
	Left     Expression   `json:"expression,omitempty"`
	Operator string       `json:"operator,omitempty"`
	Right    []Expression `json:"order_by,omitempty"` // the columns to order by
	Cast     string       `json:"cast,omitempty"`
}

func (x *AggregateExpression) expressionNode()      {}
func (x *AggregateExpression) TokenLiteral() string { return x.Token.Lit }
func (x *AggregateExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(x.Left.String(maskParams))
	if x.Operator != "" {
		out.WriteString(" " + strings.ToUpper(x.Operator) + " ")
	}
	if len(x.Right) > 1 {
		out.WriteString("(")
	}

	if len(x.Right) > 0 {
		right := []string{}
		for _, r := range x.Right {
			right = append(right, r.String(maskParams))
		}
		out.WriteString(strings.Join(right, ", "))
	}

	if len(x.Right) > 1 {
		out.WriteString(")")
	}

	return out.String()
}
func (x *AggregateExpression) SetCast(cast string) {
	x.Cast = cast
}

type WindowExpression struct {
	Token       token.Token  `json:"token,omitempty"`        // the token.OVER token
	Alias       *Identifier  `json:"alias,omitempty"`        // the alias of the window
	PartitionBy []Expression `json:"partition_by,omitempty"` // the columns to partition by
	OrderBy     []Expression `json:"order_by,omitempty"`     // the columns to order by
	Cast        string       `json:"cast,omitempty"`
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

	if w.Cast != "" {
		out.WriteString("::")
		out.WriteString(w.Cast)
	}

	return out.String()
}
func (w *WindowExpression) SetCast(cast string) {
	w.Cast = cast
}

type WildcardLiteral struct {
	Token token.Token `json:"token,omitempty"` // the token.ASTERISK token
	Value string      `json:"value,omitempty"`
	Cast  string      `json:"cast,omitempty"`
}

func (w *WildcardLiteral) expressionNode()      {}
func (w *WildcardLiteral) TokenLiteral() string { return w.Token.Lit }
func (w *WildcardLiteral) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(w.Value)

	if w.Cast != "" {
		out.WriteString("::")
		out.WriteString(w.Cast)
	}

	return out.String()
}
func (w *WildcardLiteral) SetCast(cast string) {
	w.Cast = cast
}

// JoinType  Table     Alias JoinCondition
// source    customers c
// inner     addresses a     Expression

type TableExpression struct {
	Token         token.Token `json:"token,omitempty"`          // the token.JOIN token
	JoinType      string      `json:"join_type,omitempty"`      // the type of join: source, inner, left, right, full, etc
	Schema        string      `json:"schema,omitempty"`         // the name of the schema
	Table         Expression  `json:"table,omitempty"`          // the name of the table
	Alias         Expression  `json:"alias,omitempty"`          // the alias of the table
	JoinCondition Expression  `json:"join_condition,omitempty"` // the ON clause
	Cast          string      `json:"cast,omitempty"`           // :: to cast the table
	Ordinality    bool        `json:"ordinality,omitempty"`     // the WITH ORDINALITY clause
}

func (t *TableExpression) expressionNode()      {}
func (t *TableExpression) TokenLiteral() string { return t.Token.Lit }
func (t *TableExpression) String(maskParams bool) string {
	var out bytes.Buffer

	if t.JoinType != "" {
		out.WriteString(t.JoinType + " ")
	}

	out.WriteString(t.Table.String(maskParams))
	if t.Ordinality {
		out.WriteString(" WITH ORDINALITY")
	}

	if t.Alias != nil && t.Alias.String(maskParams) != "" {
		out.WriteString(" " + t.Alias.String(maskParams))
	}

	if t.JoinCondition != nil {
		out.WriteString(" ON " + t.JoinCondition.String(maskParams))
	}

	if t.Cast != "" {
		out.WriteString("::")
		out.WriteString(t.Cast)
	}

	return out.String()
}
func (t *TableExpression) SetCast(cast string) {
	t.Cast = cast
}

type LockExpression struct {
	Token   token.Token  `json:"token,omitempty"`   // the token.FOR token
	Lock    string       `json:"lock,omitempty"`    // the type of lock: update, share, key share, no key update
	Tables  []Expression `json:"tables,omitempty"`  // the tables to lock
	Options string       `json:"options,omitempty"` // the options: NOWAIT, SKIP LOCKED
	Cast    string       `json:"cast,omitempty"`
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

	if l.Cast != "" {
		out.WriteString("::")
		out.WriteString(l.Cast)
	}

	return out.String()
}
func (l *LockExpression) SetCast(cast string) {
	l.Cast = cast
}

type InExpression struct {
	Token    token.Token  `json:"token,omitempty"` // The operator token, e.g. IN, NOT IN
	Left     Expression   `json:"left,omitempty"`
	Operator string       `json:"operator,omitempty"`
	Right    []Expression `json:"right,omitempty"`
	Cast     string       `json:"cast,omitempty"`
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

	if ie.Cast != "" {
		out.WriteString("::")
		out.WriteString(ie.Cast)
	}

	return out.String()
}
func (ie *InExpression) SetCast(cast string) {
	ie.Cast = cast
}

type CastExpression struct {
	Token token.Token `json:"token,omitempty"` // the token.CAST token
	Left  Expression  `json:"value,omitempty"`
	Cast  string      `json:"cast,omitempty"`
}

func (x *CastExpression) expressionNode()      {}
func (x *CastExpression) TokenLiteral() string { return x.Token.Lit }
func (x *CastExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("CAST(")
	out.WriteString(x.Left.String(maskParams))
	out.WriteString(" AS ")
	out.WriteString(x.Cast)
	out.WriteString(")")

	return out.String()
}
func (x *CastExpression) SetCast(cast string) {
	x.Cast = cast
}

type WhereExpression struct {
	Token token.Token `json:"token,omitempty"` // The keyword token, e.g. WHERE
	Right Expression  `json:"right,omitempty"`
	Cast  string      `json:"cast,omitempty"` // probably not needed, but used for the interface
}

func (x *WhereExpression) expressionNode()      {}
func (x *WhereExpression) TokenLiteral() string { return x.Token.Lit }
func (x *WhereExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(strings.ToUpper(x.Token.Lit))

	if x.Right != nil {
		out.WriteString("(")
		out.WriteString(x.Right.String(maskParams))
		out.WriteString(")")
	}

	if x.Cast != "" {
		out.WriteString("::")
		out.WriteString(x.Cast)
	}

	return out.String()
}
func (x *WhereExpression) SetCast(cast string) {
	x.Cast = cast
}

type IsExpression struct {
	Token    token.Token `json:"token,omitempty"` // the token.CAST token
	Left     Expression  `json:"left,omitempty"`
	Not      bool        `json:"not,omitempty"`
	Distinct bool        `json:"distinct,omitempty"`
	Right    Expression  `json:"right,omitempty"`
	Cast     string      `json:"cast,omitempty"`
}

func (x *IsExpression) expressionNode()      {}
func (x *IsExpression) TokenLiteral() string { return x.Token.Lit }
func (x *IsExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(x.Left.String(maskParams))

	out.WriteString(" IS ")

	if x.Not {
		out.WriteString("NOT ")
	}
	if x.Distinct {
		out.WriteString("DISTINCT FROM ")
	}
	if x.Right != nil {
		out.WriteString(x.Right.String(maskParams))
	}
	out.WriteString(")")

	if x.Cast != "" {
		out.WriteString("::")
		out.WriteString(x.Cast)
	}

	return out.String()
}
func (x *IsExpression) SetCast(cast string) {
	x.Cast = cast
}

// trim(both 'x' from 'xTomxx') -> Tom
type TrimExpression struct {
	Token      token.Token `json:"token,omitempty"` // the token.BOTH, token.LEADING, or token.TRAILING token
	Expression Expression  `json:"expression,omitempty"`
	Cast       string      `json:"cast,omitempty"`
}

func (x *TrimExpression) expressionNode()      {}
func (x *TrimExpression) TokenLiteral() string { return x.Token.Lit }
func (x *TrimExpression) SetCast(cast string) {
	x.Cast = cast
}
func (x *TrimExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(strings.ToUpper(x.Token.Lit) + " ")
	if x.Expression != nil {
		out.WriteString(x.Expression.String(maskParams))
	}
	return out.String()
}

// substring('Thomas' from '...$')
// substring('Thomas' from 2 for 3))
type StringFunctionExpression struct {
	Token token.Token `json:"token,omitempty"` // the token.STRING
	Left  Expression  `json:"left,omitempty"`
	From  Expression  `json:"from,omitempty"`
	For   Expression  `json:"for,omitempty"`
	Cast  string      `json:"cast,omitempty"`
}

func (x *StringFunctionExpression) expressionNode()      {}
func (x *StringFunctionExpression) TokenLiteral() string { return x.Token.Lit }
func (x *StringFunctionExpression) SetCast(cast string) {
	x.Cast = cast
}
func (x *StringFunctionExpression) String(maskParams bool) string {
	var out bytes.Buffer

	if x.Left != nil {
		out.WriteString(x.Left.String(maskParams))
		if x.From != nil {
			out.WriteString(" FROM ")
			out.WriteString(x.From.String(maskParams))
		}
		if x.For != nil {
			out.WriteString(" FOR ")
			out.WriteString(x.For.String(maskParams))
		}
	}

	return out.String()
}
