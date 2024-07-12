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

// All expression nodes implement this, although we don't have any yet
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
	Parameters      string
}

func (ls *LogStatement) statementNode()       {}
func (ls *LogStatement) TokenLiteral() string { return ls.Token.Lit }
func (ls *LogStatement) String() string {
	var out bytes.Buffer

	// Prefix
	out.WriteString(fmt.Sprintf("%s %s %s:%s(%d):%s@%s:[%d]:%s:",
		ls.Date, ls.Time, ls.Timezone, ls.RemoteHost, ls.RemotePort, ls.User, ls.Database, ls.Pid, ls.Severity))

	// Duration
	if ls.DurationLit != "" {
		out.WriteString(fmt.Sprintf("  duration: %s %s", ls.DurationLit, ls.DurationMeasure))
	}

	// Prepared Statement
	if ls.PreparedStep != "" && ls.PreparedName != "" {
		out.WriteString(fmt.Sprintf("  %s %s:", ls.PreparedStep, ls.PreparedName))
	} else if ls.PreparedStep != "" {
		out.WriteString(fmt.Sprintf("  %s:", ls.PreparedStep))
	}

	// Query
	if ls.Query != "" {
		out.WriteString(fmt.Sprintf(" %s", ls.Query))
	}

	// Parameters
	if ls.Parameters != "" {
		out.WriteString(fmt.Sprintf("  parameters: %s", ls.Parameters))
	}

	return out.String()
}
