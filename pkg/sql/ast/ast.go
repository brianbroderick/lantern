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
	SetCast(cast string)
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
type Identifier struct {
	Token token.Token `json:"token,omitempty"` // the token.IDENT token
	Value string      `json:"value,omitempty"`
	Cast  string      `json:"cast,omitempty"`
}

func (x *Identifier) expressionNode()      {}
func (x *Identifier) TokenLiteral() string { return x.Token.Lit }
func (x *Identifier) String(maskParams bool) string {
	if x.Cast != "" {
		return fmt.Sprintf("%s::%s", x.Value, strings.ToUpper(x.Cast))
	}

	return x.Value
}
func (x *Identifier) SetCast(cast string) {
	x.Cast = cast
}

type Boolean struct {
	Token token.Token
	Value bool
	Cast  string
}

func (x *Boolean) expressionNode()               {}
func (x *Boolean) TokenLiteral() string          { return x.Token.Lit }
func (x *Boolean) String(maskParams bool) string { return x.Token.Lit }
func (x *Boolean) SetCast(cast string) {
	x.Cast = cast
}

type Null struct {
	Token token.Token
	Cast  string
}

func (x *Null) expressionNode()      {}
func (x *Null) TokenLiteral() string { return x.Token.Lit }
func (x *Null) String(maskParams bool) string {
	literal := strings.ToUpper(x.Token.Lit)
	if x.Cast != "" {
		return fmt.Sprintf("%s::%s", literal, strings.ToUpper(x.Cast))
	}
	return literal
}
func (x *Null) SetCast(cast string) {
	x.Cast = cast
}

type IntegerLiteral struct {
	Token       token.Token `json:"token,omitempty"`
	Value       int64       `json:"value,omitempty"`
	Cast        string      `json:"cast,omitempty"`
	ParamOffset int         `json:"param_offset,omitempty"`
}

func (x *IntegerLiteral) expressionNode()      {}
func (x *IntegerLiteral) TokenLiteral() string { return x.Token.Lit }
func (x *IntegerLiteral) String(maskParams bool) string {
	literal := x.Token.Lit
	if maskParams {
		literal = fmt.Sprintf("$%d", x.ParamOffset)
	}
	if x.Cast != "" {
		return fmt.Sprintf("%s::%s", literal, strings.ToUpper(x.Cast))
	}
	return literal
}
func (x *IntegerLiteral) SetCast(cast string) {
	x.Cast = cast
}

type KeywordExpression struct {
	Token token.Token `json:"token,omitempty"` // The keyword token, e.g. ALL
	Cast  string      `json:"cast,omitempty"`
}

func (x *KeywordExpression) expressionNode()      {}
func (x *KeywordExpression) TokenLiteral() string { return x.Token.Lit }
func (x *KeywordExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(strings.ToUpper(x.Token.Lit))
	if x.Cast != "" {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast))
	}
	return out.String()
}
func (x *KeywordExpression) SetCast(cast string) {
	x.Cast = cast
}

// Prefix Expressions are assumed to be unary operators such as -5 or !true
type PrefixExpression struct {
	Token    token.Token `json:"token,omitempty"` // The prefix token, e.g. !
	Operator string      `json:"operator,omitempty"`
	Right    Expression  `json:"right,omitempty"`
	Cast     string      `json:"cast,omitempty"`
}

func (x *PrefixExpression) expressionNode()      {}
func (x *PrefixExpression) TokenLiteral() string { return x.Token.Lit }
func (x *PrefixExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(x.Operator)
	out.WriteString(x.Right.String(maskParams))
	out.WriteString(")")
	if x.Cast != "" {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast))
	}

	return out.String()
}
func (x *PrefixExpression) SetCast(cast string) {
	x.Cast = cast
}

// Prefix Keyword Expressions are assumed to be separate keywords such as NOT
// some prefix keyword expressions, such as DISTINCT have special handling (i.e the keyword use of ON with DISTINCT)
// so they have their own struct.
type PrefixKeywordExpression struct {
	Token    token.Token `json:"token,omitempty"` // The prefix token, e.g. !
	Operator string      `json:"operator,omitempty"`
	Right    Expression  `json:"right,omitempty"`
	Cast     string      `json:"cast,omitempty"`
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

	if x.Cast != "" {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast))
	}

	return out.String()
}
func (x *PrefixKeywordExpression) SetCast(cast string) {
	x.Cast = cast
}

type InfixExpression struct {
	Token    token.Token `json:"token,omitempty"` // The operator token, e.g. +
	Left     Expression  `json:"left,omitempty"`
	Operator string      `json:"operator,omitempty"`
	Right    Expression  `json:"right,omitempty"`
	Cast     string      `json:"cast,omitempty"`
}

func (x *InfixExpression) expressionNode()      {}
func (x *InfixExpression) TokenLiteral() string { return x.Token.Lit }
func (x *InfixExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(x.Left.String(maskParams))
	out.WriteString(" " + strings.ToUpper(x.Operator) + " ")
	if x.Right != nil {
		out.WriteString(x.Right.String(maskParams))
	}
	out.WriteString(")")

	if x.Cast != "" {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast))
	}

	return out.String()
}
func (x *InfixExpression) SetCast(cast string) {
	x.Cast = cast
}

type GroupedExpression struct {
	Token    token.Token  `json:"token,omitempty"` // The '(' token
	Elements []Expression `json:"elements,omitempty"`
	Cast     string       `json:"cast,omitempty"`
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

	if x.Cast != "" {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast))
	}

	return out.String()
}
func (x *GroupedExpression) SetCast(cast string) {
	x.Cast = cast
}

type CallExpression struct {
	Token     token.Token  `json:"token,omitempty"`    // The '(' token
	Distinct  Expression   `json:"distinct,omitempty"` // the DISTINCT or ALL token
	Function  Expression   `json:"function,omitempty"` // Identifier or FunctionLiteral
	Arguments []Expression `json:"arguments,omitempty"`
	Cast      string       `json:"cast,omitempty"`
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

	if x.Cast != "" {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast))
	}

	return out.String()
}
func (x *CallExpression) SetCast(cast string) {
	x.Cast = cast
}

type StringLiteral struct {
	Token       token.Token `json:"token,omitempty"`
	Value       string      `json:"value,omitempty"`
	Cast        string      `json:"cast,omitempty"`
	ParamOffset int         `json:"param_offset,omitempty"`
}

func (x *StringLiteral) expressionNode()      {}
func (x *StringLiteral) TokenLiteral() string { return x.Token.Lit }
func (x *StringLiteral) String(maskParams bool) string {
	literal := strings.Replace(x.Token.Lit, "'", "''", -1)
	if maskParams {
		literal = fmt.Sprintf("$%d", x.ParamOffset)
	}
	if x.Cast != "" {
		return fmt.Sprintf("'%s'::%s", literal, strings.ToUpper(x.Cast))
	}
	return fmt.Sprintf("'%s'", literal)
}
func (x *StringLiteral) SetCast(cast string) {
	x.Cast = cast
}

type ArrayLiteral struct {
	Token    token.Token  `json:"token,omitempty"` // the '[' token
	Left     Expression   `json:"left,omitempty"`
	Elements []Expression `json:"elements,omitempty"`
	Cast     string       `json:"cast,omitempty"`
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

	if x.Cast != "" {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast))
	}

	return out.String()
}
func (x *ArrayLiteral) SetCast(cast string) {
	x.Cast = cast
}

type IndexExpression struct {
	Token token.Token `json:"token,omitempty"` // The [ token
	Left  Expression  `json:"left,omitempty"`
	Index Expression  `json:"index,omitempty"`
	Cast  string      `json:"cast,omitempty"`
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

	if x.Cast != "" {
		out.WriteString("::")
		out.WriteString(strings.ToUpper(x.Cast))
	}

	return out.String()
}
func (x *IndexExpression) SetCast(cast string) {
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
