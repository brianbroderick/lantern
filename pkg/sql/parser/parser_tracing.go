package parser

import (
	"fmt"
	"strings"
)

const DEBUG = false
const traceIdentPlaceholder string = "  "

func (p *Parser) identLevel() string {
	return strings.Repeat(traceIdentPlaceholder, p.traceLevel-1)
}

func (p *Parser) tracePrint(fs string) {
	fmt.Printf("%s%s\n", p.identLevel(), fs)
}

func (p *Parser) incIdent() { p.traceLevel = p.traceLevel + 1 }
func (p *Parser) decIdent() { p.traceLevel = p.traceLevel - 1 }

func (p *Parser) trace(msg string) string {
	if !DEBUG {
		return ""
	}

	p.incIdent()
	p.tracePrint(fmt.Sprintf("BEGIN %s: %s %s :: %s %s", msg, p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit))
	return msg
}

func (p *Parser) untrace(msg string) {
	if !DEBUG {
		return
	}
	p.tracePrint(fmt.Sprintf("END %s: %s %s :: %s %s", msg, p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit))
	p.decIdent()
}
