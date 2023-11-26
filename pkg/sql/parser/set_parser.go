package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

// SET application_name = 'example';

func (p *Parser) parseSetStatement() *ast.SetStatement {
	// defer untrace(trace("parseSetStatement " + p.curToken.Lit))

	stmt := &ast.SetStatement{Token: p.curToken}
	p.nextToken()
	if p.curTokenIs(token.SESSION) {
		stmt.Session = true
		p.nextToken()
	}
	if p.curTokenIs(token.LOCAL) {
		stmt.Local = true
		p.nextToken()
	}
	if p.curTokenIs(token.IDENT) && p.curToken.Upper == "TIME" {
		stmt.TimeZone = true
		p.nextToken()
		if p.curTokenIs(token.IDENT) && p.curToken.Upper == "ZONE" {
			p.nextToken()
		}
	}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
