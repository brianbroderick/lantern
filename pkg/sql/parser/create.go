package parser

import (
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func (p *Parser) parseCreateStatement() *ast.CreateStatement {
	defer untrace(trace("parseCreateStatement " + p.curToken.Lit))

	stmt := &ast.CreateStatement{Token: p.curToken}
	p.nextToken()
	lit := strings.ToLower(p.curToken.Lit)

	// Table qualifiers
	if p.curTokenIs(token.IDENT) && (lit == "temp" || lit == "temporary") {
		stmt.Temp = true
		p.nextToken()
	}
	if p.curTokenIs(token.IDENT) && strings.ToLower(p.curToken.Lit) == "unlogged" {
		stmt.Unlogged = true
		p.nextToken()
	}

	// Index qualifiers
	if p.curTokenIs(token.UNIQUE) {
		stmt.Unique = true
		p.nextToken()
	}

	// Create Object
	lit = strings.ToLower(p.curToken.Lit)
	if p.curTokenIs(token.TABLE) || lit == "index" {
		stmt.Object = p.curToken
		p.nextToken()
	}

	if p.curTokenIs(token.CONCURRENTLY) {
		stmt.Concurrently = true
		p.nextToken()
	}

	if p.curTokenIs(token.IDENT) && strings.ToLower(p.curToken.Lit) == "if" {
		p.nextToken()
		if p.curTokenIs(token.NOT) {
			p.nextToken()
			if p.curTokenIs(token.IDENT) && strings.ToLower(p.curToken.Lit) == "exists" {
				stmt.Exists = true
				p.nextToken()
			}
		}
	}

	stmt.Name = p.parseExpression(LOWEST)
	p.nextToken()

	if p.curTokenIsOne([]token.TokenType{token.AS, token.ON}) {
		stmt.Operator = strings.ToUpper(p.curToken.Lit)
		p.nextToken()
	}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
