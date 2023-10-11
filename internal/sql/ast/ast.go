package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/internal/sql/token"
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

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Lit }
func (es *ExpressionStatement) String(maskParams bool) string {
	if es.Expression != nil {
		return es.Expression.String(maskParams)
	}
	return ""
}
func (es *ExpressionStatement) Inspect(maskParams bool) string {
	return es.String(maskParams)
}

// Expressions
type Identifier struct {
	Token token.Token `json:"token,omitempty"` // the token.IDENT token
	Value string      `json:"value,omitempty"`
	Cast  string      `json:"cast,omitempty"`
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Lit }
func (i *Identifier) String(maskParams bool) string {
	if i.Cast != "" {
		return fmt.Sprintf("%s::%s", i.Value, strings.ToUpper(i.Cast))
	}
	return i.Value
}

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()               {}
func (b *Boolean) TokenLiteral() string          { return b.Token.Lit }
func (b *Boolean) String(maskParams bool) string { return b.Token.Lit }

type IntegerLiteral struct {
	Token       token.Token `json:"token,omitempty"`
	Value       int64       `json:"value,omitempty"`
	Cast        string      `json:"cast,omitempty"`
	ParamOffset int         `json:"param_offset,omitempty"`
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Lit }
func (il *IntegerLiteral) String(maskParams bool) string {
	literal := il.Token.Lit
	if maskParams {
		literal = fmt.Sprintf("$%d", il.ParamOffset)
	}
	if il.Cast != "" {
		return fmt.Sprintf("%s::%s", literal, strings.ToUpper(il.Cast))
	}
	return literal
}

type KeywordExpression struct {
	Token token.Token `json:"token,omitempty"` // The keyword token, e.g. ALL
}

func (ke *KeywordExpression) expressionNode()      {}
func (ke *KeywordExpression) TokenLiteral() string { return ke.Token.Lit }
func (ke *KeywordExpression) String(maskParams bool) string {
	var out bytes.Buffer
	out.WriteString(strings.ToUpper(ke.Token.Lit))
	return out.String()
}

type PrefixExpression struct {
	Token    token.Token `json:"token,omitempty"` // The prefix token, e.g. !
	Operator string      `json:"operator,omitempty"`
	Right    Expression  `json:"right,omitempty"`
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Lit }
func (pe *PrefixExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String(maskParams))
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token `json:"token,omitempty"` // The operator token, e.g. +
	Left     Expression  `json:"left,omitempty"`
	Operator string      `json:"operator,omitempty"`
	Right    Expression  `json:"right,omitempty"`
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Lit }
func (ie *InfixExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String(maskParams))
	out.WriteString(" " + strings.ToUpper(ie.Operator) + " ")
	if ie.Right != nil {
		out.WriteString(ie.Right.String(maskParams))
	}
	out.WriteString(")")

	return out.String()
}

type CallExpression struct {
	Token     token.Token  `json:"token,omitempty"`    // The '(' token
	Function  Expression   `json:"function,omitempty"` // Identifier or FunctionLiteral
	Arguments []Expression `json:"arguments,omitempty"`
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Lit }
func (ce *CallExpression) String(maskParams bool) string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String(maskParams))
	}

	out.WriteString(ce.Function.String(maskParams))
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type StringLiteral struct {
	Token       token.Token `json:"token,omitempty"`
	Value       string      `json:"value,omitempty"`
	Cast        string      `json:"cast,omitempty"`
	ParamOffset int         `json:"param_offset,omitempty"`
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Lit }
func (sl *StringLiteral) String(maskParams bool) string {
	literal := sl.Token.Lit
	if maskParams {
		literal = fmt.Sprintf("$%d", sl.ParamOffset)
	}
	if sl.Cast != "" {
		return fmt.Sprintf("'%s'::%s", literal, strings.ToUpper(sl.Cast))
	}
	return fmt.Sprintf("'%s'", literal)
}

type ArrayLiteral struct {
	Token    token.Token  `json:"token,omitempty"` // the '[' token
	Elements []Expression `json:"elements,omitempty"`
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Lit }
func (al *ArrayLiteral) String(maskParams bool) string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String(maskParams))
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type IndexExpression struct {
	Token token.Token `json:"token,omitempty"` // The [ token
	Left  Expression  `json:"left,omitempty"`
	Index Expression  `json:"index,omitempty"`
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Lit }
func (ie *IndexExpression) String(maskParams bool) string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String(maskParams))
	out.WriteString("[")
	out.WriteString(ie.Index.String(maskParams))
	out.WriteString("])")

	return out.String()
}

// type HashLiteral struct {
// 	Token token.Token // the '{' token
// 	Pairs map[Expression]Expression
// }

// func (hl *HashLiteral) expressionNode()      {}
// func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Lit }
// func (hl *HashLiteral) String(maskParams bool) string {
// 	var out bytes.Buffer

// 	pairs := []string{}
// 	for key, value := range hl.Pairs {
// 		pairs = append(pairs, key.String()+":"+value.String())
// 	}

// 	out.WriteString("{")
// 	out.WriteString(strings.Join(pairs, ", "))
// 	out.WriteString("}")

// 	return out.String()
// }
