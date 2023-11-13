package parser

import (
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

// This currently handles DROP TABLE, but does not handle DROP DATABASE, DROP INDEX, etc. yet.

func (p *Parser) parseDropStatement() *ast.DropStatement {
	defer untrace(trace("parseDropStatement " + p.curToken.Lit))

	stmt := &ast.DropStatement{Token: p.curToken}
	p.nextToken()
	stmt.Object = p.curToken

	p.nextToken()

	if p.curTokenIs(token.IDENT) && strings.ToLower(p.curToken.Lit) == "if" {
		p.nextToken()
		if p.curTokenIs(token.IDENT) && strings.ToLower(p.curToken.Lit) == "exists" {
			stmt.Exists = true
			p.nextToken()
		}
	}

	if p.curTokenIs(token.IDENT) {
		stmt.Tables = p.parseDropTableList()
	}

	if p.curTokenIs(token.IDENT) && (strings.ToLower(p.curToken.Lit) == "cascade" || strings.ToLower(p.curToken.Lit) == "restrict") {
		stmt.Options = p.curToken.Lit
		p.nextToken()
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseDropTableList() []ast.Expression {
	defer untrace(trace("parseDropTableList"))

	list := []ast.Expression{}

	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	p.nextToken()

	return list
}
