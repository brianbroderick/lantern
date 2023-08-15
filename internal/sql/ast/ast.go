package ast

import (
	"bytes"
	"strings"

	"github.com/brianbroderick/lantern/internal/sql/token"
)

// Describes the Abstract Syntax Tree (AST) for the SQL language.

// The base Node interface
type Node interface {
	TokenLiteral() string
	String() string
}

// All statement nodes implement this
type Statement interface {
	Node
	statementNode()
}

// All expression nodes implement this
type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// Statements
// Most of these need to be replaced by structs
type SelectStatement struct {
	Token   token.Token // the token.SELECT token
	Columns Expression
	From    *Identifier
	// Joins   []*Identifier
	// Where   []*Identifier
	// GroupBy []*Identifier
	// Having  []*Identifier
	// OrderBy []*Identifier
	// Limit   *IntegerLiteral
	// Offset  *IntegerLiteral
}

func (ls *SelectStatement) statementNode()       {}
func (ls *SelectStatement) TokenLiteral() string { return ls.Token.Lit }

// String() is incomplete and only returns the most basic of select statements
func (ls *SelectStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Columns.String())
	// columns := []string{}
	// for _, c := range ls.Columns {
	// 	columns = append(columns, c.String())
	// }
	// out.WriteString(strings.Join(columns, ", "))
	out.WriteString(" from ")
	out.WriteString(ls.From.String())
	out.WriteString(";")

	return out.String()
}

// This is a statement without a leading token. For example: x + 10;
type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Lit }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// Expressions
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Lit }
func (i *Identifier) String() string       { return i.Value }

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Lit }
func (b *Boolean) String() string       { return b.Token.Lit }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Lit }
func (il *IntegerLiteral) String() string       { return il.Token.Lit }

type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. !
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Lit }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Lit }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Lit }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Lit }
func (sl *StringLiteral) String() string       { return sl.Token.Lit }

type ArrayLiteral struct {
	Token    token.Token // the '[' token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Lit }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type IndexExpression struct {
	Token token.Token // The [ token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Lit }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}

type ColumnExpression struct {
	Token token.Token // the token.IDENT token
	Value Expression
	Alias string
}

func (c *ColumnExpression) expressionNode()      {}
func (c *ColumnExpression) TokenLiteral() string { return c.Token.Lit }
func (c *ColumnExpression) String() string {
	var out bytes.Buffer

	out.WriteString(c.Value.String())
	if c.Alias != "" {
		out.WriteString(" as ")
		out.WriteString(c.Alias)
	}

	return out.String()
}

// type HashLiteral struct {
// 	Token token.Token // the '{' token
// 	Pairs map[Expression]Expression
// }

// func (hl *HashLiteral) expressionNode()      {}
// func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Lit }
// func (hl *HashLiteral) String() string {
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
