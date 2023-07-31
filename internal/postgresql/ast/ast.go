package ast

import (
	"bytes"
	"fmt"

	"github.com/brianbroderick/lantern/internal/postgresql/token"
)

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

// LogStatement represents a log entry
type LogStatement struct {
	Token           token.Token
	Date            string
	Time            string
	Timezone        string
	RemoteHost      string
	RemotePort      int
	User            string
	Database        string
	Pid             int
	Severity        string
	DurationLit     string
	DurationMeasure string
	PreparedStep    string
	PreparedName    string
	Query           string
	// duration        time.Duration
}

func (ls *LogStatement) statementNode()       {}
func (ls *LogStatement) TokenLiteral() string { return ls.Token.Lit }
func (ls *LogStatement) String() string {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("%s %s %s:%s(%d):%s@%s:[%d]:%s:  duration: %s %s  %s <%s>: %s",
		ls.Date, ls.Time, ls.Timezone, ls.RemoteHost, ls.RemotePort, ls.User,
		ls.Database, ls.Pid, ls.Severity, ls.DurationLit, ls.DurationMeasure,
		ls.PreparedStep, ls.PreparedName, ls.Query))

	return out.String()
}

// // Expressions
// type Identifier struct {
// 	Token token.Token // the token.IDENT token
// 	Value string
// }

// func (i *Identifier) expressionNode()      {}
// func (i *Identifier) TokenLiteral() string { return i.Token.Lit }
// func (i *Identifier) String() string       { return i.Value }

// type IntegerLiteral struct {
// 	Token token.Token
// 	Value int64
// }

// func (il *IntegerLiteral) expressionNode()      {}
// func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Lit }
// func (il *IntegerLiteral) String() string       { return il.Token.Lit }

// type StringLiteral struct {
// 	Token token.Token
// 	Value string
// }

// func (sl *StringLiteral) expressionNode()      {}
// func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Lit }
// func (sl *StringLiteral) String() string       { return sl.Token.Lit }

// type PrefixExpression struct {
// 	Token    token.Token // The prefix token, e.g. !
// 	Operator string
// 	Right    Expression
// }

// func (pe *PrefixExpression) expressionNode()      {}
// func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Lit }
// func (pe *PrefixExpression) String() string {
// 	var out bytes.Buffer

// 	out.WriteString("(")
// 	out.WriteString(pe.Operator)
// 	out.WriteString(pe.Right.String())
// 	out.WriteString(")")

// 	return out.String()
// }
