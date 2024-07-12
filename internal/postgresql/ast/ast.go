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
	// duration        time.Duration
}

func (ls *LogStatement) statementNode()       {}
func (ls *LogStatement) TokenLiteral() string { return ls.Token.Lit }
func (ls *LogStatement) String() string {
	var out bytes.Buffer
	if ls.PreparedName != "" {
		out.WriteString(fmt.Sprintf("%s %s %s:%s(%d):%s@%s:[%d]:%s:  duration: %s %s  %s %s: %s",
			ls.Date, ls.Time, ls.Timezone, ls.RemoteHost, ls.RemotePort, ls.User,
			ls.Database, ls.Pid, ls.Severity, ls.DurationLit, ls.DurationMeasure,
			ls.PreparedStep, ls.PreparedName, ls.Query))
	} else {
		out.WriteString(fmt.Sprintf("%s %s %s:%s(%d):%s@%s:[%d]:%s:  duration: %s %s  %s: %s",
			ls.Date, ls.Time, ls.Timezone, ls.RemoteHost, ls.RemotePort, ls.User,
			ls.Database, ls.Pid, ls.Severity, ls.DurationLit, ls.DurationMeasure,
			ls.PreparedStep, ls.Query))
	}

	return out.String()
}
