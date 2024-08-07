package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

// create temp table temp_my_table( like my_reports );

func (p *Parser) parseCreateStatement() *ast.CreateStatement {
	// defer p.untrace(p.trace("parseCreateStatement"))

	stmt := &ast.CreateStatement{Token: p.curToken}
	p.nextToken()

	// Table qualifiers
	if p.curTokenIs(token.IDENT) && (p.curToken.Upper == "TEMP" || p.curToken.Upper == "TEMPORARY") {
		stmt.Temp = true
		p.nextToken()
	}
	if p.curTokenIs(token.IDENT) && p.curToken.Upper == "UNLOGGED" {
		stmt.Unlogged = true
		p.nextToken()
	}

	// Index qualifiers
	if p.curTokenIs(token.UNIQUE) {
		stmt.Unique = true
		p.nextToken()
	}

	// Create Object
	if p.curTokenIs(token.TABLE) || p.curToken.Upper == "INDEX" {
		stmt.Object = p.curToken
		p.nextToken()
	}

	if p.curTokenIs(token.CONCURRENTLY) {
		stmt.Concurrently = true
		p.nextToken()
	}

	if p.curTokenIs(token.IDENT) && p.curToken.Upper == "IF" {
		p.nextToken()
		if p.curTokenIs(token.NOT) {
			p.nextToken()
			if p.curTokenIs(token.IDENT) && p.curToken.Upper == "EXISTS" {
				stmt.Exists = true
				p.nextToken()
			}
		}
	}

	stmt.Name = p.parseExpression(LOWEST)

	p.nextToken()

	if p.curTokenIs(token.ON) {
		if p.peekTokenIs(token.COMMIT) {
			p.nextToken()
			p.nextToken()
			if p.curTokenIs(token.DELETE) {
				if p.peekTokenIs(token.ROWS) {
					p.nextToken()
					p.nextToken()
					stmt.OnCommit = "DELETE ROWS"
				}
			} else if p.curTokenIs(token.IDENT) && p.curToken.Upper == "PRESERVE" {
				if p.peekTokenIs(token.ROWS) {
					p.nextToken()
					p.nextToken()
					stmt.OnCommit = "PRESERVE ROWS"
				}
			} else if p.curTokenIs(token.DROP) {
				p.nextToken()
				stmt.OnCommit = "DROP"
			}
		}
	}

	if p.curTokenIsOne([]token.TokenType{token.AS, token.ON}) {
		stmt.Operator = p.curToken.Upper
		p.nextToken()
	}

	if !p.curTokenIs(token.SEMICOLON) {
		stmt.Expression = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.WHERE) {
		p.nextToken()
		p.nextToken()
		p.clause = token.WHERE
		stmt.Where = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseLikeExpression() ast.Expression {
	// defer p.untrace(p.trace("parseLikeExpression"))

	x := &ast.LikeExpression{Token: p.curToken, Branch: p.clause, CommandTag: p.command}

	if p.peekTokenIs(token.IDENT) {
		p.nextToken()

		context := p.context
		p.setContext(XCREATE)         // sets the context for the parseLikeExpression function
		defer p.resetContext(context) // reset to prior context

		x.Table = p.parseExpression(LOWEST)
	}

	return x
}
