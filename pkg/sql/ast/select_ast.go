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

func (s *SelectStatement) Clause() token.TokenType      { return s.Token.Type }
func (x *SelectStatement) SetClause(c token.TokenType)  {}
func (s *SelectStatement) Command() token.TokenType     { return s.Token.Type }
func (x *SelectStatement) SetCommand(c token.TokenType) {}
func (s *SelectStatement) statementNode()               {}
func (s *SelectStatement) TokenLiteral() string         { return s.Token.Upper }
func (s *SelectStatement) String(maskParams bool) string {
	var out bytes.Buffer

	for _, e := range s.Expressions {
		out.WriteString(e.String(maskParams))
	}

	out.WriteString(";")

	return out.String()
}
func (s *SelectStatement) Inspect(maskParams bool) string {
	j, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Printf("Error loading data: %#v\n\n", err)
	}
	return string(j)
}

// SelectExpression is a select inside a SELECT or WITH (Common Table Expression) statement,
// since a select statement can have multiple select expressions. i.e. WITH clause, subqueries, and the primary select expression.
type SelectExpression struct {
	Token        token.Token       `json:"token,omitempty"`    // the token.SELECT token
	Distinct     Expression        `json:"distinct,omitempty"` // the DISTINCT or ALL token
	Columns      []Expression      `json:"columns,omitempty"`
	Tables       []Expression      `json:"tables,omitempty"`
	Where        Expression        `json:"where,omitempty"`
	GroupBy      []Expression      `json:"group_by,omitempty"`
	Having       Expression        `json:"having,omitempty"`
	Window       []Expression      `json:"window,omitempty"`
	OrderBy      []Expression      `json:"order_by,omitempty"`
	Limit        Expression        `json:"limit,omitempty"`
	Offset       Expression        `json:"offset,omitempty"`
	Fetch        Expression        `json:"fetch,omitempty"`
	Lock         Expression        `json:"lock,omitempty"`
	Cast         Expression        `json:"cast,omitempty"` // probably not needed, but used for the interface
	TableAliases map[string]string `json:"-"`
	Branch       token.TokenType   `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag   token.TokenType   `json:"command,omitempty"`
}

func (x *SelectExpression) Clause() token.TokenType      { return x.Branch }
func (x *SelectExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *SelectExpression) Command() token.TokenType     { return x.CommandTag }
func (x *SelectExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *SelectExpression) expressionNode()              {}
func (x *SelectExpression) TokenLiteral() string         { return x.Token.Upper }

// String() is incomplete and only returns the most basic of select statements
func (x *SelectExpression) String(maskParams bool) string {
	var out bytes.Buffer

	// Subqueries need to be surrounded by parentheses. A primary query may also have parentheses, so we'll add them here to be consistent.
	out.WriteString("(")

	out.WriteString(x.TokenLiteral() + " ")

	// Distinct
	if x.Distinct != nil {
		out.WriteString(x.Distinct.String(maskParams) + " ")
	}

	// Columns
	columns := []string{}
	for _, c := range x.Columns {
		columns = append(columns, c.String(maskParams))
	}
	out.WriteString(strings.Join(columns, ", "))

	// Tables
	if len(x.Tables) > 0 {
		if len(x.Columns) > 0 { // add a space between columns and tables, if both exist
			out.WriteString(" ")
		}
		out.WriteString("FROM ")
		tables := []string{}
		for _, t := range x.Tables {
			tables = append(tables, t.String(maskParams))
		}
		out.WriteString(strings.Join(tables, " "))
	}

	// Window
	if len(x.Window) > 0 {
		out.WriteString(" WINDOW ")
		windows := []string{}
		for _, w := range x.Window {
			windows = append(windows, w.String(maskParams))
		}
		out.WriteString(strings.Join(windows, ", "))
	}

	// Where
	if x.Where != nil {
		out.WriteString(" WHERE ")
		out.WriteString(x.Where.String(maskParams))
	}

	// Group By
	if len(x.GroupBy) > 0 {
		out.WriteString(" GROUP BY ")
		groupBy := []string{}
		for _, g := range x.GroupBy {
			groupBy = append(groupBy, g.String(maskParams))
		}
		out.WriteString(strings.Join(groupBy, ", "))
	}

	// Having
	if x.Having != nil {
		out.WriteString(" HAVING ")
		out.WriteString(x.Having.String(maskParams))
	}

	// Order By
	if len(x.OrderBy) > 0 {
		out.WriteString(" ORDER BY ")
		orderBy := []string{}
		for _, g := range x.OrderBy {
			orderBy = append(orderBy, g.String(maskParams))
		}
		out.WriteString(strings.Join(orderBy, ", "))
	}

	// Limit
	if x.Limit != nil {
		out.WriteString(" LIMIT ")
		out.WriteString(x.Limit.String(maskParams))
	}

	// Offset
	if x.Offset != nil {
		out.WriteString(" OFFSET ")
		out.WriteString(x.Offset.String(maskParams))
	}

	// Fetch
	if x.Fetch != nil {
		out.WriteString(" FETCH NEXT ")
		out.WriteString(x.Fetch.String(maskParams))
	}

	// Lock
	if x.Lock != nil {
		out.WriteString(" FOR ")
		out.WriteString(x.Lock.String(maskParams))
	}

	out.WriteString(")")

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *SelectExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type DistinctExpression struct {
	Token      token.Token     `json:"token,omitempty"`  // The keyword token, e.g. DISTINCT
	Right      []Expression    `json:"right,omitempty"`  // The columns to be distinct
	Cast       Expression      `json:"cast,omitempty"`   // probably not needed, but used for the interface
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *DistinctExpression) Clause() token.TokenType      { return x.Branch }
func (x *DistinctExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *DistinctExpression) Command() token.TokenType     { return x.CommandTag }
func (x *DistinctExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *DistinctExpression) expressionNode()              {}
func (x *DistinctExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *DistinctExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(x.Token.Upper)
	if len(x.Right) > 0 {
		out.WriteString(" ON ")
	}
	if x.Right != nil {
		out.WriteString("(")
		right := []string{}
		for _, r := range x.Right {
			right = append(right, r.String(maskParams))
		}
		out.WriteString(strings.Join(right, ", "))
		out.WriteString(")")
	}

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *DistinctExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type ColumnExpression struct {
	Token      token.Token     `json:"token,omitempty"` // the token
	Name       Expression      `json:"name,omitempty"`  // the name of the column or alias
	Value      Expression      `json:"value,omitempty"` // the complete expression including all of the columns
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *ColumnExpression) Clause() token.TokenType      { return x.Branch }
func (x *ColumnExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *ColumnExpression) Command() token.TokenType     { return x.CommandTag }
func (x *ColumnExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *ColumnExpression) expressionNode()              {}
func (x *ColumnExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *ColumnExpression) String(maskParams bool) string {
	var out bytes.Buffer

	val := x.Value.String(maskParams)
	out.WriteString(val)
	if x.Name != nil {
		name := x.Name.String(maskParams)
		if name != val && name != "" {
			out.WriteString(" AS ")
			out.WriteString(x.Name.String(maskParams))
		}
	}

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *ColumnExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type SortExpression struct {
	Token      token.Token     `json:"token,omitempty"`     // the token.ASC or token.DESC token
	Value      Expression      `json:"value,omitempty"`     // the column to sort on
	Direction  token.Token     `json:"direction,omitempty"` // the direction to sort
	Nulls      token.Token     `json:"nulls,omitempty"`     // first or last
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *SortExpression) Clause() token.TokenType      { return x.Branch }
func (x *SortExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *SortExpression) Command() token.TokenType     { return x.CommandTag }
func (x *SortExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *SortExpression) expressionNode()              {}
func (x *SortExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *SortExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(x.Value.String(maskParams))
	if x.Direction.Type == token.DESC {
		out.WriteString(" DESC")
	}
	if x.Nulls.Type == token.FIRST {
		out.WriteString(" NULLS FIRST")
	} else if x.Nulls.Type == token.LAST {
		out.WriteString(" NULLS LAST")
	}

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *SortExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type FetchExpression struct {
	Token token.Token `json:"token,omitempty"` // the token.FETCH token
	// Don't need to store "first" or "next" since these are synonyms. We'll convert everything to "next" when printing in .String()
	Value Expression `json:"value,omitempty"` // the number of rows to fetch
	// We also don't need to store "row" or "rows" since these are synonyms. We'll convert everything to "rows" when printing in .String()
	Option     token.Token     `json:"option,omitempty"` // the token.ONLY or token.TIES token (don't need to store "with ties", just "ties" will do)
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *FetchExpression) Clause() token.TokenType      { return x.Branch }
func (x *FetchExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *FetchExpression) Command() token.TokenType     { return x.CommandTag }
func (x *FetchExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *FetchExpression) expressionNode()              {}
func (x *FetchExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *FetchExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(x.Value.String(maskParams))
	out.WriteString(" ROWS")
	if x.Option.Type == token.TIES {
		out.WriteString(" WITH TIES")
	} else {
		out.WriteString(" ONLY")
	}

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *FetchExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type AggregateExpression struct {
	Token      token.Token     `json:"token,omitempty"`
	Left       Expression      `json:"expression,omitempty"`
	Operator   string          `json:"operator,omitempty"`
	Right      []Expression    `json:"order_by,omitempty"` // the columns to order by
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *AggregateExpression) Clause() token.TokenType      { return x.Branch }
func (x *AggregateExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *AggregateExpression) Command() token.TokenType     { return x.CommandTag }
func (x *AggregateExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *AggregateExpression) expressionNode()              {}
func (x *AggregateExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *AggregateExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(x.Left.String(maskParams))
	if x.Operator != "" {
		out.WriteString(" " + strings.ToUpper(x.Operator) + " ")
	}
	// if len(x.Right) > 1 {
	// 	out.WriteString("(")
	// }

	if len(x.Right) > 0 {
		right := []string{}
		for _, r := range x.Right {
			right = append(right, r.String(maskParams))
		}
		out.WriteString(strings.Join(right, ", "))
	}

	// if len(x.Right) > 1 {
	// 	out.WriteString(")")
	// }

	return out.String()
}
func (x *AggregateExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type WindowExpression struct {
	Token       token.Token     `json:"token,omitempty"`        // the token.OVER token
	Alias       Expression      `json:"alias,omitempty"`        // the alias of the window
	PartitionBy []Expression    `json:"partition_by,omitempty"` // the columns to partition by
	OrderBy     []Expression    `json:"order_by,omitempty"`     // the columns to order by
	Cast        Expression      `json:"cast,omitempty"`
	Branch      token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag  token.TokenType `json:"command,omitempty"`
}

func (x *WindowExpression) Clause() token.TokenType      { return x.Branch }
func (x *WindowExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *WindowExpression) Command() token.TokenType     { return x.CommandTag }
func (x *WindowExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *WindowExpression) expressionNode()              {}
func (x *WindowExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *WindowExpression) String(maskParams bool) string {
	var out bytes.Buffer

	if x.Alias != nil {
		out.WriteString(x.Alias.String(maskParams) + " AS ")
	}

	out.WriteString("(")
	if len(x.PartitionBy) > 0 {
		out.WriteString("PARTITION BY ")
		partitionBy := []string{}
		for _, p := range x.PartitionBy {
			partitionBy = append(partitionBy, p.String(maskParams))
		}
		out.WriteString(strings.Join(partitionBy, ", "))
	}
	if len(x.PartitionBy) > 0 && len(x.OrderBy) > 0 {
		out.WriteString(" ")
	}
	if len(x.OrderBy) > 0 {
		out.WriteString("ORDER BY ")
		orderBy := []string{}
		for _, o := range x.OrderBy {
			orderBy = append(orderBy, o.String(maskParams))
		}
		out.WriteString(strings.Join(orderBy, ", "))
	}

	out.WriteString(")")

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *WindowExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type WildcardLiteral struct {
	Token      token.Token     `json:"token,omitempty"` // the token.ASTERISK token
	Value      string          `json:"value,omitempty"`
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *WildcardLiteral) Clause() token.TokenType      { return x.Branch }
func (x *WildcardLiteral) SetClause(c token.TokenType)  { x.Branch = c }
func (x *WildcardLiteral) Command() token.TokenType     { return x.CommandTag }
func (x *WildcardLiteral) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *WildcardLiteral) expressionNode()              {}
func (x *WildcardLiteral) TokenLiteral() string         { return x.Token.Upper }
func (x *WildcardLiteral) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(x.Value)

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *WildcardLiteral) SetCast(cast Expression) {
	x.Cast = cast
}

type TimestampExpression struct {
	Token        token.Token     `json:"token,omitempty"` // the token.TIMESTAMP token
	Value        string          `json:"value,omitempty"`
	WithTimeZone bool            `json:"with_time_zone,omitempty"`
	Cast         Expression      `json:"cast,omitempty"`
	Branch       token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag   token.TokenType `json:"command,omitempty"`
}

func (x *TimestampExpression) Clause() token.TokenType      { return x.Branch }
func (x *TimestampExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *TimestampExpression) Command() token.TokenType     { return x.CommandTag }
func (x *TimestampExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *TimestampExpression) expressionNode()              {}
func (x *TimestampExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *TimestampExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(x.Token.Upper)
	if x.Value != "" {
		out.WriteString(" '" + x.Value + "'")
	}

	if x.WithTimeZone {
		out.WriteString(" WITH TIME ZONE")
	}

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *TimestampExpression) SetCast(cast Expression) {
	x.Cast = cast
}

// JoinType  Table     Alias JoinCondition
// source    customers c
// inner     addresses a     Expression

type TableExpression struct {
	Token         token.Token     `json:"token,omitempty"`          // the token.JOIN token
	JoinType      string          `json:"join_type,omitempty"`      // the type of join: source, inner, left, right, full, etc
	Schema        string          `json:"schema,omitempty"`         // the name of the schema
	Table         Expression      `json:"table,omitempty"`          // the name of the table
	Alias         Expression      `json:"alias,omitempty"`          // the alias of the table
	JoinCondition Expression      `json:"join_condition,omitempty"` // the ON clause
	Cast          Expression      `json:"cast,omitempty"`           // :: to cast the table
	Ordinality    bool            `json:"ordinality,omitempty"`     // the WITH ORDINALITY clause
	Branch        token.TokenType `json:"clause,omitempty"`         // location in the tree representing a clause
	CommandTag    token.TokenType `json:"command,omitempty"`
}

func (x *TableExpression) Clause() token.TokenType      { return x.Branch }
func (x *TableExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *TableExpression) Command() token.TokenType     { return x.CommandTag }
func (x *TableExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *TableExpression) expressionNode()              {}
func (x *TableExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *TableExpression) String(maskParams bool) string {
	var out bytes.Buffer

	if x.JoinType != "" {
		out.WriteString(x.JoinType + " ")
	}

	out.WriteString(x.Table.String(maskParams))
	if x.Ordinality {
		out.WriteString(" WITH ORDINALITY")
	}

	if x.Alias != nil && x.Alias.String(maskParams) != "" {
		out.WriteString(" " + x.Alias.String(maskParams))
	}

	if x.JoinCondition != nil {
		out.WriteString(" ON " + x.JoinCondition.String(maskParams))
	}

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *TableExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type LockExpression struct {
	Token      token.Token     `json:"token,omitempty"`   // the token.FOR token
	Lock       string          `json:"lock,omitempty"`    // the type of lock: update, share, key share, no key update
	Tables     []Expression    `json:"tables,omitempty"`  // the tables to lock
	Options    string          `json:"options,omitempty"` // the options: NOWAIT, SKIP LOCKED
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *LockExpression) Clause() token.TokenType      { return x.Branch }
func (x *LockExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *LockExpression) Command() token.TokenType     { return x.CommandTag }
func (x *LockExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *LockExpression) expressionNode()              {}
func (x *LockExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *LockExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(x.Lock)
	if len(x.Tables) > 0 {
		tables := []string{}
		out.WriteString(" OF ")
		for _, t := range x.Tables {
			tables = append(tables, t.String(maskParams))
		}
		out.WriteString(strings.Join(tables, ", "))
	}
	if x.Options != "" {
		out.WriteString(" " + x.Options)
	}

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *LockExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type InExpression struct {
	Token      token.Token     `json:"token,omitempty"` // The operator token, e.g. IN, NOT IN
	Left       Expression      `json:"left,omitempty"`
	Not        bool            `json:"not,omitempty"`
	Operator   string          `json:"operator,omitempty"`
	Right      []Expression    `json:"right,omitempty"`
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *InExpression) Clause() token.TokenType      { return x.Branch }
func (x *InExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *InExpression) Command() token.TokenType     { return x.CommandTag }
func (x *InExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *InExpression) expressionNode()              {}
func (x *InExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *InExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(x.Left.String(maskParams))
	if x.Not {
		out.WriteString(" NOT")
	}
	out.WriteString(" " + strings.ToUpper(x.Operator) + " ")

	out.WriteString("(")
	oneLess := len(x.Right) - 1

	for i, e := range x.Right {
		out.WriteString(e.String(maskParams))
		if maskParams {
			break
		}
		if i < oneLess {
			out.WriteString(", ")
		}
	}

	out.WriteString(")")

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *InExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type CastExpression struct {
	Token      token.Token     `json:"token,omitempty"` // the token.CAST token
	Left       Expression      `json:"value,omitempty"`
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *CastExpression) Clause() token.TokenType      { return x.Branch }
func (x *CastExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *CastExpression) Command() token.TokenType     { return x.CommandTag }
func (x *CastExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *CastExpression) expressionNode()              {}
func (x *CastExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *CastExpression) String(maskParams bool) string {
	var out bytes.Buffer

	if x.Left == nil && x.Cast != nil {
		out.WriteString(x.Cast.String(maskParams))
	} else {
		out.WriteString("CAST(")
		out.WriteString(x.Left.String(maskParams))
		out.WriteString(" AS ")
		out.WriteString(x.Cast.String(maskParams))
		out.WriteString(")")
	}

	return out.String()
}
func (x *CastExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type WhereExpression struct {
	Token      token.Token     `json:"token,omitempty"` // The keyword token, e.g. WHERE
	Right      Expression      `json:"right,omitempty"`
	Cast       Expression      `json:"cast,omitempty"`   // probably not needed, but used for the interface
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *WhereExpression) Clause() token.TokenType      { return x.Branch }
func (x *WhereExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *WhereExpression) Command() token.TokenType     { return x.CommandTag }
func (x *WhereExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *WhereExpression) expressionNode()              {}
func (x *WhereExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *WhereExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString(x.Token.Upper)

	if x.Right != nil {
		out.WriteString("(")
		out.WriteString(x.Right.String(maskParams))
		out.WriteString(")")
	}

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *WhereExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type IsExpression struct {
	Token       token.Token     `json:"token,omitempty"` // the token.CAST token
	Left        Expression      `json:"left,omitempty"`
	Not         bool            `json:"not,omitempty"`
	Distinct    bool            `json:"distinct,omitempty"`
	Right       Expression      `json:"right,omitempty"`
	Cast        Expression      `json:"cast,omitempty"`
	ParamOffset int             `json:"param_offset,omitempty"`
	Branch      token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag  token.TokenType `json:"command,omitempty"`
}

func (x *IsExpression) Clause() token.TokenType      { return x.Branch }
func (x *IsExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *IsExpression) Command() token.TokenType     { return x.CommandTag }
func (x *IsExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *IsExpression) expressionNode()              {}
func (x *IsExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *IsExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(x.Left.String(maskParams))

	out.WriteString(" IS ")

	if maskParams {
		// out.WriteString("$" + fmt.Sprintf("%d", x.ParamOffset))
		out.WriteString("?")
	} else {
		if x.Not {
			out.WriteString("NOT ")
		}
		if x.Distinct {
			out.WriteString("DISTINCT FROM ")
		}
		if x.Right != nil {
			out.WriteString(x.Right.String(maskParams))
		}
	}
	out.WriteString(")")

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *IsExpression) SetCast(cast Expression) {
	x.Cast = cast
}

// trim(both 'x' from 'xTomxx') -> Tom
type TrimExpression struct {
	Token      token.Token     `json:"token,omitempty"` // the token.BOTH, token.LEADING, or token.TRAILING token
	Expression Expression      `json:"expression,omitempty"`
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *TrimExpression) Clause() token.TokenType      { return x.Branch }
func (x *TrimExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *TrimExpression) Command() token.TokenType     { return x.CommandTag }
func (x *TrimExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *TrimExpression) expressionNode()              {}
func (x *TrimExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *TrimExpression) SetCast(cast Expression) {
	x.Cast = cast
}
func (x *TrimExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(x.Token.Upper + " ")
	if x.Expression != nil {
		out.WriteString(x.Expression.String(maskParams))
	}
	return out.String()
}

// substring('Thomas' from '...$')
// substring('Thomas' from 2 for 3))
type StringFunctionExpression struct {
	Token      token.Token     `json:"token,omitempty"` // the token.STRING
	Left       Expression      `json:"left,omitempty"`
	From       Expression      `json:"from,omitempty"`
	For        Expression      `json:"for,omitempty"`
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *StringFunctionExpression) Clause() token.TokenType      { return x.Branch }
func (x *StringFunctionExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *StringFunctionExpression) Command() token.TokenType     { return x.CommandTag }
func (x *StringFunctionExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *StringFunctionExpression) expressionNode()              {}
func (x *StringFunctionExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *StringFunctionExpression) SetCast(cast Expression) {
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

type UnionExpression struct {
	Token      token.Token     `json:"token,omitempty"` // The operator token, e.g. +
	Left       Expression      `json:"left,omitempty"`
	Operator   string          `json:"operator,omitempty"` // UNION, INTERSECT, EXCEPT
	All        bool            `json:"all,omitempty"`
	Right      Expression      `json:"right,omitempty"`
	Cast       Expression      `json:"cast,omitempty"`
	Branch     token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
	CommandTag token.TokenType `json:"command,omitempty"`
}

func (x *UnionExpression) Clause() token.TokenType      { return x.Branch }
func (x *UnionExpression) SetClause(c token.TokenType)  { x.Branch = c }
func (x *UnionExpression) Command() token.TokenType     { return x.CommandTag }
func (x *UnionExpression) SetCommand(c token.TokenType) { x.CommandTag = c }
func (x *UnionExpression) expressionNode()              {}
func (x *UnionExpression) TokenLiteral() string         { return x.Token.Upper }
func (x *UnionExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(x.Left.String(maskParams))
	out.WriteString(" " + strings.ToUpper(x.Operator) + " ")
	if x.All {
		out.WriteString("ALL ")
	}
	if x.Right != nil {
		out.WriteString(x.Right.String(maskParams))
	}
	out.WriteString(")")

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *UnionExpression) SetCast(cast Expression) {
	x.Cast = cast
}
