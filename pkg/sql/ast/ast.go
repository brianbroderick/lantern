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
	Command() token.TokenType
	Clause() token.TokenType
	SetClause(token.TokenType)
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

func (p *Program) Clause() token.TokenType     { return token.PROGRAM }
func (p *Program) SetClause(c token.TokenType) {}
func (p *Program) Command() token.TokenType    { return token.PROGRAM }
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
		out.WriteString("\n")
	}

	return out.String()
}

// Statements

// This is a statement without a leading token. For example: x + 10;
type ExpressionStatement struct {
	Token      token.Token `json:"token,omitempty"` // the first token of the expression
	Expression Expression  `json:"expression,omitempty"`
}

func (x *ExpressionStatement) Clause() token.TokenType     { return x.Expression.Clause() }
func (x *ExpressionStatement) SetClause(c token.TokenType) {}
func (x *ExpressionStatement) Command() token.TokenType    { return x.Expression.Command() }
func (x *ExpressionStatement) statementNode()              {}
func (x *ExpressionStatement) TokenLiteral() string        { return x.Token.Lit }
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
	Token  token.Token     `json:"token,omitempty"` // the token.IDENT token
	Value  string          `json:"value,omitempty"`
	Cast   Expression      `json:"cast,omitempty"`
	Branch token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *SimpleIdentifier) Clause() token.TokenType     { return x.Branch }
func (x *SimpleIdentifier) SetClause(c token.TokenType) { x.Branch = c }
func (x *SimpleIdentifier) Command() token.TokenType    { return x.Token.Type }
func (x *SimpleIdentifier) expressionNode()             {}
func (x *SimpleIdentifier) TokenLiteral() string        { return x.Token.Lit }
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
	Token  token.Token     `json:"token,omitempty"` // the token.IDENT token
	Value  []Expression    `json:"value,omitempty"` // can have multiple values, e.g. schema.table.column
	Cast   Expression      `json:"cast,omitempty"`
	Branch token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *Identifier) Clause() token.TokenType     { return x.Branch }
func (x *Identifier) SetClause(c token.TokenType) { x.Branch = c }
func (x *Identifier) Command() token.TokenType    { return x.Token.Type }
func (x *Identifier) expressionNode()             {}
func (x *Identifier) TokenLiteral() string        { return x.Token.Lit }
func (x *Identifier) String(maskParams bool) string {
	var out bytes.Buffer

	if len(x.Value) > 0 {
		val := []string{}
		for _, r := range x.Value {
			if r != nil {
				val = append(val, r.String(maskParams))
			}
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

type Boolean struct {
	Token       token.Token     `json:"token,omitempty"`
	Value       bool            `json:"value,omitempty"`
	Cast        Expression      `json:"cast,omitempty"`
	ParamOffset int             `json:"param_offset,omitempty"`
	Branch      token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *Boolean) Clause() token.TokenType     { return x.Branch }
func (x *Boolean) SetClause(c token.TokenType) { x.Branch = c }
func (x *Boolean) Command() token.TokenType    { return x.Token.Type }
func (x *Boolean) expressionNode()             {}
func (x *Boolean) TokenLiteral() string        { return x.Token.Lit }
func (x *Boolean) String(maskParams bool) string {
	literal := x.Token.Upper
	if maskParams {
		// literal = fmt.Sprintf("$%d", x.ParamOffset)
		literal = "?"
	}
	if x.Cast != nil {
		return fmt.Sprintf("%s::%s", literal, strings.ToUpper(x.Cast.String(maskParams)))
	}
	return literal
}
func (x *Boolean) SetCast(cast Expression) {
	x.Cast = cast
}

type Null struct {
	Token       token.Token     `json:"token,omitempty"`
	Cast        Expression      `json:"cast,omitempty"`
	ParamOffset int             `json:"param_offset,omitempty"`
	Branch      token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *Null) Clause() token.TokenType     { return x.Branch }
func (x *Null) SetClause(c token.TokenType) { x.Branch = c }
func (x *Null) Command() token.TokenType    { return x.Token.Type }
func (x *Null) expressionNode()             {}
func (x *Null) TokenLiteral() string        { return x.Token.Lit }
func (x *Null) String(maskParams bool) string {
	literal := x.Token.Upper
	if maskParams {
		// literal = fmt.Sprintf("$%d", x.ParamOffset)
		literal = "?"
	}
	if x.Cast != nil {
		return fmt.Sprintf("%s::%s", literal, strings.ToUpper(x.Cast.String(maskParams)))
	}
	return literal
}
func (x *Null) SetCast(cast Expression) {
	x.Cast = cast
}

type Unknown struct {
	Token       token.Token     `json:"token,omitempty"`
	Cast        Expression      `json:"cast,omitempty"`
	ParamOffset int             `json:"param_offset,omitempty"`
	Branch      token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *Unknown) Clause() token.TokenType     { return x.Branch }
func (x *Unknown) SetClause(c token.TokenType) { x.Branch = c }
func (x *Unknown) Command() token.TokenType    { return x.Token.Type }
func (x *Unknown) expressionNode()             {}
func (x *Unknown) TokenLiteral() string        { return x.Token.Lit }
func (x *Unknown) String(maskParams bool) string {
	literal := x.Token.Upper
	if maskParams {
		// literal = fmt.Sprintf("$%d", x.ParamOffset)
		literal = "?"
	}
	if x.Cast != nil {
		return fmt.Sprintf("%s::%s", literal, strings.ToUpper(x.Cast.String(maskParams)))
	}
	return literal
}
func (x *Unknown) SetCast(cast Expression) {
	x.Cast = cast
}

// Infinity is used as the token for the hidden value after the colon in array expressions such as array[1:]
// Infinity is not a true SQL type, and casts cannot be applied to it.
type Infinity struct {
	Token  token.Token
	Cast   Expression
	Branch token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *Infinity) Clause() token.TokenType     { return x.Branch }
func (x *Infinity) SetClause(c token.TokenType) { x.Branch = c }
func (x *Infinity) Command() token.TokenType    { return x.Token.Type }
func (x *Infinity) expressionNode()             {}
func (x *Infinity) TokenLiteral() string        { return "∞" }
func (x *Infinity) String(maskParams bool) string {
	return ""
}
func (x *Infinity) SetCast(cast Expression) {
	x.Cast = cast
}

type IntegerLiteral struct {
	Token       token.Token     `json:"token,omitempty"`
	Value       int64           `json:"value,omitempty"`
	Cast        Expression      `json:"cast,omitempty"`
	ParamOffset int             `json:"param_offset,omitempty"`
	Branch      token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *IntegerLiteral) Clause() token.TokenType     { return x.Branch }
func (x *IntegerLiteral) SetClause(c token.TokenType) { x.Branch = c }
func (x *IntegerLiteral) Command() token.TokenType    { return x.Token.Type }
func (x *IntegerLiteral) expressionNode()             {}
func (x *IntegerLiteral) TokenLiteral() string        { return x.Token.Lit }
func (x *IntegerLiteral) String(maskParams bool) string {
	literal := x.Token.Lit
	if maskParams {
		// literal = fmt.Sprintf("$%d", x.ParamOffset)
		literal = "?"
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
	Token       token.Token     `json:"token,omitempty"`
	Value       float64         `json:"value,omitempty"`
	Cast        Expression      `json:"cast,omitempty"`
	ParamOffset int             `json:"param_offset,omitempty"`
	Branch      token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *FloatLiteral) Clause() token.TokenType     { return x.Branch }
func (x *FloatLiteral) SetClause(c token.TokenType) { x.Branch = c }
func (x *FloatLiteral) Command() token.TokenType    { return x.Token.Type }
func (x *FloatLiteral) expressionNode()             {}
func (x *FloatLiteral) TokenLiteral() string        { return x.Token.Lit }
func (x *FloatLiteral) String(maskParams bool) string {
	literal := x.Token.Lit
	if maskParams {
		// literal = fmt.Sprintf("$%d", x.ParamOffset)
		literal = "?"
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
	Token  token.Token     `json:"token,omitempty"` // The keyword token, e.g. ALL
	Cast   Expression      `json:"cast,omitempty"`
	Branch token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *KeywordExpression) Clause() token.TokenType     { return x.Branch }
func (x *KeywordExpression) SetClause(c token.TokenType) { x.Branch = c }
func (x *KeywordExpression) Command() token.TokenType    { return x.Token.Type }
func (x *KeywordExpression) expressionNode()             {}
func (x *KeywordExpression) TokenLiteral() string        { return x.Token.Lit }
func (x *KeywordExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(x.Token.Upper)
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
	Token    token.Token     `json:"token,omitempty"` // The prefix token, e.g. !
	Operator string          `json:"operator,omitempty"`
	Right    Expression      `json:"right,omitempty"`
	Cast     Expression      `json:"cast,omitempty"`
	Branch   token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *PrefixExpression) Clause() token.TokenType     { return x.Branch }
func (x *PrefixExpression) SetClause(c token.TokenType) { x.Branch = c }
func (x *PrefixExpression) Command() token.TokenType    { return x.Token.Type }
func (x *PrefixExpression) expressionNode()             {}
func (x *PrefixExpression) TokenLiteral() string        { return x.Token.Lit }
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
	Token    token.Token     `json:"token,omitempty"` // The prefix token, e.g. !
	Operator string          `json:"operator,omitempty"`
	Right    Expression      `json:"right,omitempty"`
	Cast     Expression      `json:"cast,omitempty"`
	Branch   token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *PrefixKeywordExpression) Clause() token.TokenType     { return x.Branch }
func (x *PrefixKeywordExpression) SetClause(c token.TokenType) { x.Branch = c }
func (x *PrefixKeywordExpression) Command() token.TokenType    { return x.Token.Type }
func (x *PrefixKeywordExpression) expressionNode()             {}
func (x *PrefixKeywordExpression) TokenLiteral() string        { return x.Token.Lit }
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
	Token    token.Token     `json:"token,omitempty"` // The operator token, e.g. +
	Left     Expression      `json:"left,omitempty"`
	Not      bool            `json:"not,omitempty"` // prefix NOT to the operator
	Operator string          `json:"operator,omitempty"`
	Right    Expression      `json:"right,omitempty"`
	Cast     Expression      `json:"cast,omitempty"`
	Branch   token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *InfixExpression) Clause() token.TokenType     { return x.Branch }
func (x *InfixExpression) SetClause(c token.TokenType) { x.Branch = c }
func (x *InfixExpression) Command() token.TokenType    { return x.Token.Type }
func (x *InfixExpression) expressionNode()             {}
func (x *InfixExpression) TokenLiteral() string        { return x.Token.Lit }
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
	Token    token.Token     `json:"token,omitempty"` // The '(' token
	Elements []Expression    `json:"elements,omitempty"`
	Cast     Expression      `json:"cast,omitempty"`
	Branch   token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *GroupedExpression) Clause() token.TokenType     { return x.Branch }
func (x *GroupedExpression) SetClause(c token.TokenType) { x.Branch = c }
func (x *GroupedExpression) Command() token.TokenType    { return x.Token.Type }
func (x *GroupedExpression) expressionNode()             {}
func (x *GroupedExpression) TokenLiteral() string        { return x.Token.Lit }
func (x *GroupedExpression) String(maskParams bool) string {
	var out bytes.Buffer

	elements := []string{}
	for _, a := range x.Elements {
		elements = append(elements, a.String(maskParams))
	}

	// If there is only one element, don't wrap it in parentheses
	if len(elements) != 1 {
		out.WriteString("(")
	}
	out.WriteString(strings.Join(elements, ", "))

	if len(elements) != 1 {
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
	Token     token.Token     `json:"token,omitempty"`    // The '(' token
	Distinct  Expression      `json:"distinct,omitempty"` // the DISTINCT or ALL token
	Function  Expression      `json:"function,omitempty"` // Identifier or FunctionLiteral
	Arguments []Expression    `json:"arguments,omitempty"`
	Cast      Expression      `json:"cast,omitempty"`
	Branch    token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *CallExpression) Clause() token.TokenType     { return x.Branch }
func (x *CallExpression) SetClause(c token.TokenType) { x.Branch = c }
func (x *CallExpression) Command() token.TokenType    { return x.Token.Type }
func (x *CallExpression) expressionNode()             {}
func (x *CallExpression) TokenLiteral() string        { return x.Token.Lit }
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
	Token       token.Token     `json:"token,omitempty"`
	Value       string          `json:"value,omitempty"`
	Cast        Expression      `json:"cast,omitempty"`
	ParamOffset int             `json:"param_offset,omitempty"`
	Branch      token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *StringLiteral) Clause() token.TokenType     { return x.Branch }
func (x *StringLiteral) SetClause(c token.TokenType) { x.Branch = c }
func (x *StringLiteral) Command() token.TokenType    { return x.Token.Type }
func (x *StringLiteral) expressionNode()             {}
func (x *StringLiteral) TokenLiteral() string        { return x.Token.Lit }
func (x *StringLiteral) String(maskParams bool) string {
	literal := strings.Replace(x.Token.Lit, "'", "''", -1)
	if maskParams {
		// literal = fmt.Sprintf("$%d", x.ParamOffset)
		literal = "?"
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
	Token       token.Token     `json:"token,omitempty"`
	Value       string          `json:"value,omitempty"`
	Cast        Expression      `json:"cast,omitempty"`
	ParamOffset int             `json:"param_offset,omitempty"`
	Branch      token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *EscapeStringLiteral) Clause() token.TokenType     { return x.Branch }
func (x *EscapeStringLiteral) SetClause(c token.TokenType) { x.Branch = c }
func (x *EscapeStringLiteral) Command() token.TokenType    { return x.Token.Type }
func (x *EscapeStringLiteral) expressionNode()             {}
func (x *EscapeStringLiteral) TokenLiteral() string        { return x.Token.Lit }
func (x *EscapeStringLiteral) String(maskParams bool) string {
	literal := x.Token.Lit
	if maskParams {
		// literal = fmt.Sprintf("$%d", x.ParamOffset)
		literal = "?"
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
	Token    token.Token     `json:"token,omitempty"` // the '[' token
	Left     Expression      `json:"left,omitempty"`
	Elements []Expression    `json:"elements,omitempty"`
	Cast     Expression      `json:"cast,omitempty"`
	Branch   token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *ArrayLiteral) Clause() token.TokenType     { return x.Branch }
func (x *ArrayLiteral) SetClause(c token.TokenType) { x.Branch = c }
func (x *ArrayLiteral) Command() token.TokenType    { return x.Token.Type }
func (x *ArrayLiteral) expressionNode()             {}
func (x *ArrayLiteral) TokenLiteral() string        { return x.Token.Lit }
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
	Token  token.Token     `json:"token,omitempty"` // The [ token
	Left   Expression      `json:"left,omitempty"`
	Index  Expression      `json:"index,omitempty"`
	Cast   Expression      `json:"cast,omitempty"`
	Branch token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *IndexExpression) Clause() token.TokenType     { return x.Branch }
func (x *IndexExpression) SetClause(c token.TokenType) { x.Branch = c }
func (x *IndexExpression) Command() token.TokenType    { return x.Token.Type }
func (x *IndexExpression) expressionNode()             {}
func (x *IndexExpression) TokenLiteral() string        { return x.Token.Lit }
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
	Token  token.Token     `json:"token,omitempty"` // The interval token
	Value  Expression      `json:"value,omitempty"`
	Unit   Expression      `json:"unit,omitempty"`
	Cast   Expression      `json:"cast,omitempty"`
	Branch token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *IntervalExpression) Clause() token.TokenType     { return x.Branch }
func (x *IntervalExpression) SetClause(c token.TokenType) { x.Branch = c }
func (x *IntervalExpression) Command() token.TokenType    { return x.Token.Type }
func (x *IntervalExpression) expressionNode()             {}
func (x *IntervalExpression) TokenLiteral() string        { return x.Token.Lit }
func (x *IntervalExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("INTERVAL ")
	out.WriteString(x.Value.String(maskParams))

	if x.Unit != nil {
		out.WriteString(" ")
		out.WriteString(x.Unit.String(maskParams))
	}

	if x.Cast != nil {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast.String(maskParams)))
	}

	return out.String()
}
func (x *IntervalExpression) SetCast(cast Expression) {
	x.Cast = cast
}

// For errors in the Resolver
type IllegalExpression struct {
	Token  token.Token     `json:"token,omitempty"` // the token.ILLEGAL token
	Value  string          `json:"value,omitempty"`
	Cast   Expression      `json:"cast,omitempty"`
	Branch token.TokenType `json:"clause,omitempty"` // location in the tree representing a clause
}

func (x *IllegalExpression) Clause() token.TokenType     { return x.Branch }
func (x *IllegalExpression) SetClause(c token.TokenType) { x.Branch = c }
func (x *IllegalExpression) Command() token.TokenType    { return x.Token.Type }
func (x *IllegalExpression) expressionNode()             {}
func (x *IllegalExpression) TokenLiteral() string        { return x.Token.Upper }
func (x *IllegalExpression) String(maskParams bool) string {
	if x.Cast != nil {
		return fmt.Sprintf("%s::%s", x.Value, strings.ToUpper(x.Cast.String(maskParams)))
	}

	return x.Value
}
func (x *IllegalExpression) SetCast(cast Expression) {
	x.Cast = cast
}
