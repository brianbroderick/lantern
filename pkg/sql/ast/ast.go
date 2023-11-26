package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

// Describes the Abstract Syntax Tree (AST) for the SQL language.

// The base Node interface
type Node interface {
	TokenLiteral() string
	String(maskParams bool) string // maskParams is used to mask integers and strings in the output with a ?.
}

// All statement nodes implement this
type Statement interface {
	Node
	statementNode()
	Inspect(maskParams bool) string
}

// All expression nodes implement this
type Expression interface {
	Node
	expressionNode()
	SetCast(cast Expression)
}

type Program struct {
	Statements []Statement `json:"statements,omitempty"`
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String(maskParams bool) string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String(maskParams))
	}

	return out.String()
}

func (p *Program) Inspect(maskParams bool) string {
	var out bytes.Buffer

	for i, s := range p.Statements {
		out.WriteString(fmt.Sprintf("Statement %d:\n", i+1))
		out.WriteString(s.Inspect(maskParams))
	}

	return out.String()
}

// Statements

// This is a statement without a leading token. For example: x + 10;
type ExpressionStatement struct {
	Token      token.Token `json:"token,omitempty"` // the first token of the expression
	Expression Expression  `json:"expression,omitempty"`
}

func (x *ExpressionStatement) statementNode()       {}
func (x *ExpressionStatement) TokenLiteral() string { return x.Token.Lit }
func (x *ExpressionStatement) String(maskParams bool) string {
	if x.Expression != nil {
		return x.Expression.String(maskParams)
	}
	return ""
}
func (x *ExpressionStatement) Inspect(maskParams bool) string {
	return x.String(maskParams)
}

// Expressions
type SimpleIdentifier struct {
	Token token.Token `json:"token,omitempty"` // the token.IDENT token
	Value string      `json:"value,omitempty"`
	Cast  Expression  `json:"cast,omitempty"`
}

func (x *SimpleIdentifier) expressionNode()      {}
func (x *SimpleIdentifier) TokenLiteral() string { return x.Token.Lit }
func (x *SimpleIdentifier) String(maskParams bool) string {
	if x.Cast != nil {
		return fmt.Sprintf("%s::%s", x.Value, strings.ToUpper(x.Cast.String(maskParams)))
	}

	return x.Value
}
func (x *SimpleIdentifier) SetCast(cast Expression) {
	x.Cast = cast
}

type Identifier struct {
	Token token.Token  `json:"token,omitempty"` // the token.IDENT token
	Value []Expression `json:"value,omitempty"` // can have multiple values, e.g. schema.table.column
	Cast  Expression   `json:"cast,omitempty"`
}

func (x *Identifier) expressionNode()      {}
func (x *Identifier) TokenLiteral() string { return x.Token.Lit }
func (x *Identifier) String(maskParams bool) string {
	var out bytes.Buffer

	if len(x.Value) > 0 {
		val := []string{}
		for _, r := range x.Value {
			val = append(val, r.String(maskParams))
		}
		out.WriteString(strings.Join(val, "."))
	}

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *Identifier) SetCast(cast Expression) {
	x.Cast = cast
}

type ColumnLiteral struct {
	Token  token.Token `json:"token,omitempty"` // the token.IDENT token
	Schema Expression  `json:"schema,omitempty"`
	Table  Expression  `json:"table,omitempty"`
	Column Expression  `json:"column,omitempty"`
	Cast   Expression  `json:"cast,omitempty"`
}

func (x *ColumnLiteral) expressionNode()      {}
func (x *ColumnLiteral) TokenLiteral() string { return x.Token.Lit }
func (x *ColumnLiteral) String(maskParams bool) string {
	var out bytes.Buffer
	if x.Schema != nil {
		out.WriteString(x.Schema.String(maskParams))
		out.WriteString(".")
	}
	if x.Table != nil {
		out.WriteString(x.Table.String(maskParams))
		out.WriteString(".")
	}
	if x.Column != nil {
		out.WriteString(x.Column.String(maskParams))
	}
	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *ColumnLiteral) SetCast(cast Expression) {
	x.Cast = cast
}

type Boolean struct {
	Token token.Token
	Value bool
	Cast  Expression
}

func (x *Boolean) expressionNode()               {}
func (x *Boolean) TokenLiteral() string          { return x.Token.Lit }
func (x *Boolean) String(maskParams bool) string { return strings.ToUpper(x.Token.Lit) }
func (x *Boolean) SetCast(cast Expression) {
	x.Cast = cast
}

type Null struct {
	Token token.Token
	Cast  Expression
}

func (x *Null) expressionNode()      {}
func (x *Null) TokenLiteral() string { return x.Token.Lit }
func (x *Null) String(maskParams bool) string {
	literal := strings.ToUpper(x.Token.Lit)
	if x.Cast != nil {
		return fmt.Sprintf("%s::%s", literal, strings.ToUpper(x.Cast.String(maskParams)))
	}
	return literal
}
func (x *Null) SetCast(cast Expression) {
	x.Cast = cast
}

type Unknown struct {
	Token token.Token
	Cast  Expression
}

func (x *Unknown) expressionNode()      {}
func (x *Unknown) TokenLiteral() string { return x.Token.Lit }
func (x *Unknown) String(maskParams bool) string {
	literal := strings.ToUpper(x.Token.Lit)
	if x.Cast != nil {
		return fmt.Sprintf("%s::%s", literal, strings.ToUpper(x.Cast.String(maskParams)))
	}
	return literal
}
func (x *Unknown) SetCast(cast Expression) {
	x.Cast = cast
}

type IntegerLiteral struct {
	Token       token.Token `json:"token,omitempty"`
	Value       int64       `json:"value,omitempty"`
	Cast        Expression  `json:"cast,omitempty"`
	ParamOffset int         `json:"param_offset,omitempty"`
}

func (x *IntegerLiteral) expressionNode()      {}
func (x *IntegerLiteral) TokenLiteral() string { return x.Token.Lit }
func (x *IntegerLiteral) String(maskParams bool) string {
	literal := x.Token.Lit
	if maskParams {
		literal = fmt.Sprintf("$%d", x.ParamOffset)
	}
	if x.Cast != nil {
		return fmt.Sprintf("%s::%s", literal, strings.ToUpper(x.Cast.String(maskParams)))
	}
	return literal
}
func (x *IntegerLiteral) SetCast(cast Expression) {
	x.Cast = cast
}

type FloatLiteral struct {
	Token       token.Token `json:"token,omitempty"`
	Value       float64     `json:"value,omitempty"`
	Cast        Expression  `json:"cast,omitempty"`
	ParamOffset int         `json:"param_offset,omitempty"`
}

func (x *FloatLiteral) expressionNode()      {}
func (x *FloatLiteral) TokenLiteral() string { return x.Token.Lit }
func (x *FloatLiteral) String(maskParams bool) string {
	literal := x.Token.Lit
	if maskParams {
		literal = fmt.Sprintf("$%d", x.ParamOffset)
	}
	if x.Cast != nil {
		return fmt.Sprintf("%s::%s", literal, strings.ToUpper(x.Cast.String(maskParams)))
	}
	return literal
}
func (x *FloatLiteral) SetCast(cast Expression) {
	x.Cast = cast
}

type KeywordExpression struct {
	Token token.Token `json:"token,omitempty"` // The keyword token, e.g. ALL
	Cast  Expression  `json:"cast,omitempty"`
}

func (x *KeywordExpression) expressionNode()      {}
func (x *KeywordExpression) TokenLiteral() string { return x.Token.Lit }
func (x *KeywordExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(strings.ToUpper(x.Token.Lit))
	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}
	return out.String()
}
func (x *KeywordExpression) SetCast(cast Expression) {
	x.Cast = cast
}

// Prefix Expressions are assumed to be unary operators such as -5 or !true
type PrefixExpression struct {
	Token    token.Token `json:"token,omitempty"` // The prefix token, e.g. !
	Operator string      `json:"operator,omitempty"`
	Right    Expression  `json:"right,omitempty"`
	Cast     Expression  `json:"cast,omitempty"`
}

func (x *PrefixExpression) expressionNode()      {}
func (x *PrefixExpression) TokenLiteral() string { return x.Token.Lit }
func (x *PrefixExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(x.Operator)
	out.WriteString(x.Right.String(maskParams))
	out.WriteString(")")
	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *PrefixExpression) SetCast(cast Expression) {
	x.Cast = cast
}

// Prefix Keyword Expressions are assumed to be separate keywords such as NOT
// some prefix keyword expressions, such as DISTINCT have special handling (i.e the keyword use of ON with DISTINCT)
// so they have their own struct.
type PrefixKeywordExpression struct {
	Token    token.Token `json:"token,omitempty"` // The prefix token, e.g. !
	Operator string      `json:"operator,omitempty"`
	Right    Expression  `json:"right,omitempty"`
	Cast     Expression  `json:"cast,omitempty"`
}

func (x *PrefixKeywordExpression) expressionNode()      {}
func (x *PrefixKeywordExpression) TokenLiteral() string { return x.Token.Lit }
func (x *PrefixKeywordExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(strings.ToUpper(x.Operator))
	out.WriteString(" ")
	out.WriteString(x.Right.String(maskParams))
	out.WriteString(")")

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *PrefixKeywordExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type InfixExpression struct {
	Token    token.Token `json:"token,omitempty"` // The operator token, e.g. +
	Left     Expression  `json:"left,omitempty"`
	Not      bool        `json:"not,omitempty"` // prefix NOT to the operator
	Operator string      `json:"operator,omitempty"`
	Right    Expression  `json:"right,omitempty"`
	Cast     Expression  `json:"cast,omitempty"`
}

func (x *InfixExpression) expressionNode()      {}
func (x *InfixExpression) TokenLiteral() string { return x.Token.Lit }
func (x *InfixExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(x.Left.String(maskParams))
	if x.Not {
		out.WriteString(" NOT")
	}
	out.WriteString(" " + strings.ToUpper(x.Operator) + " ")
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
func (x *InfixExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type GroupedExpression struct {
	Token    token.Token  `json:"token,omitempty"` // The '(' token
	Elements []Expression `json:"elements,omitempty"`
	Cast     Expression   `json:"cast,omitempty"`
}

func (x *GroupedExpression) expressionNode()      {}
func (x *GroupedExpression) TokenLiteral() string { return x.Token.Lit }
func (x *GroupedExpression) String(maskParams bool) string {
	var out bytes.Buffer

	elements := []string{}
	for _, a := range x.Elements {
		elements = append(elements, a.String(maskParams))
	}

	if len(elements) > 1 {
		out.WriteString("(")
	}
	out.WriteString(strings.Join(elements, ", "))

	if len(elements) > 1 {
		out.WriteString(")")
	}

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *GroupedExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type CallExpression struct {
	Token     token.Token  `json:"token,omitempty"`    // The '(' token
	Distinct  Expression   `json:"distinct,omitempty"` // the DISTINCT or ALL token
	Function  Expression   `json:"function,omitempty"` // Identifier or FunctionLiteral
	Arguments []Expression `json:"arguments,omitempty"`
	Cast      Expression   `json:"cast,omitempty"`
}

func (x *CallExpression) expressionNode()      {}
func (x *CallExpression) TokenLiteral() string { return x.Token.Lit }
func (x *CallExpression) String(maskParams bool) string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range x.Arguments {
		args = append(args, a.String(maskParams))
	}

	out.WriteString(x.Function.String(maskParams))
	out.WriteString("(")

	// Distinct, used in aggregate functions
	if x.Distinct != nil {
		out.WriteString(x.Distinct.String(maskParams) + " ")
	}

	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *CallExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type StringLiteral struct {
	Token       token.Token `json:"token,omitempty"`
	Value       string      `json:"value,omitempty"`
	Cast        Expression  `json:"cast,omitempty"`
	ParamOffset int         `json:"param_offset,omitempty"`
}

func (x *StringLiteral) expressionNode()      {}
func (x *StringLiteral) TokenLiteral() string { return x.Token.Lit }
func (x *StringLiteral) String(maskParams bool) string {
	literal := strings.Replace(x.Token.Lit, "'", "''", -1)
	if maskParams {
		literal = fmt.Sprintf("$%d", x.ParamOffset)
	}
	if x.Cast != nil {
		return fmt.Sprintf("'%s'::%s", literal, strings.ToUpper(x.Cast.String(maskParams)))
	}
	return fmt.Sprintf("'%s'", literal)
}
func (x *StringLiteral) SetCast(cast Expression) {
	x.Cast = cast
}

type EscapeStringLiteral struct {
	Token       token.Token `json:"token,omitempty"`
	Value       string      `json:"value,omitempty"`
	Cast        Expression  `json:"cast,omitempty"`
	ParamOffset int         `json:"param_offset,omitempty"`
}

func (x *EscapeStringLiteral) expressionNode()      {}
func (x *EscapeStringLiteral) TokenLiteral() string { return x.Token.Lit }
func (x *EscapeStringLiteral) String(maskParams bool) string {
	literal := x.Token.Lit
	if maskParams {
		literal = fmt.Sprintf("$%d", x.ParamOffset)
	}
	if x.Cast != nil {
		return fmt.Sprintf("%s::%s", literal, strings.ToUpper(x.Cast.String(maskParams)))
	}
	return literal
}
func (x *EscapeStringLiteral) SetCast(cast Expression) {
	x.Cast = cast
}

type ArrayLiteral struct {
	Token    token.Token  `json:"token,omitempty"` // the '[' token
	Left     Expression   `json:"left,omitempty"`
	Elements []Expression `json:"elements,omitempty"`
	Cast     Expression   `json:"cast,omitempty"`
}

func (x *ArrayLiteral) expressionNode()      {}
func (x *ArrayLiteral) TokenLiteral() string { return x.Token.Lit }
func (x *ArrayLiteral) String(maskParams bool) string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range x.Elements {
		elements = append(elements, el.String(maskParams))
	}

	if x.Left != nil {
		out.WriteString(x.Left.String(maskParams))
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *ArrayLiteral) SetCast(cast Expression) {
	x.Cast = cast
}

type IndexExpression struct {
	Token token.Token `json:"token,omitempty"` // The [ token
	Left  Expression  `json:"left,omitempty"`
	Index Expression  `json:"index,omitempty"`
	Cast  Expression  `json:"cast,omitempty"`
}

func (x *IndexExpression) expressionNode()      {}
func (x *IndexExpression) TokenLiteral() string { return x.Token.Lit }
func (x *IndexExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(x.Left.String(maskParams))
	out.WriteString("[")
	out.WriteString(x.Index.String(maskParams))
	out.WriteString("])")

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *IndexExpression) SetCast(cast Expression) {
	x.Cast = cast
}

type IntervalExpression struct {
	Token token.Token `json:"token,omitempty"` // The interval token
	Value Expression  `json:"value,omitempty"`
	Cast  Expression  `json:"cast,omitempty"`
}

func (x *IntervalExpression) expressionNode()      {}
func (x *IntervalExpression) TokenLiteral() string { return x.Token.Lit }
func (x *IntervalExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("INTERVAL ")
	out.WriteString(x.Value.String(maskParams))

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *IntervalExpression) SetCast(cast Expression) {
	x.Cast = cast
}

// type HashLiteral struct {
// 	Token token.Token // the '{' token
// 	Pairs map[Expression]Expression
// }

// func (x *HashLiteral) expressionNode()      {}
// func (x *HashLiteral) TokenLiteral() string { return x.Token.Lit }
// func (x *HashLiteral) String(maskParams bool) string {
// 	var out bytes.Buffer

// 	pairs := []string{}
// 	for key, value := range x.Pairs {
// 		pairs = append(pairs, key.String()+":"+value.String())
// 	}

// 	out.WriteString("{")
// 	out.WriteString(strings.Join(pairs, ", "))
// 	out.WriteString("}")

// 	return out.String()
// }
