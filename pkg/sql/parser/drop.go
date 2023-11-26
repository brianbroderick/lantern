package parser

import (
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

	if p.curTokenIs(token.IDENT) && p.curToken.Upper == "IF" {
		p.nextToken()
		if p.curTokenIs(token.IDENT) && p.curToken.Upper == "EXISTS" {
			stmt.Exists = true
			p.nextToken()
		}
	}

	if p.curTokenIs(token.IDENT) {
		stmt.Tables = p.parseDropTableList()
	}

	if p.curTokenIs(token.IDENT) && (p.curToken.Upper == "CASCADE" || p.curToken.Upper == "RESTRICT") {
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
