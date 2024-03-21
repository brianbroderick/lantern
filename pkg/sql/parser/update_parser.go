package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func (p *Parser) parseUpdateStatement() *ast.UpdateStatement {
	defer p.untrace(p.trace("parseUpdateStatement"))

	s := &ast.UpdateStatement{Token: p.curToken}
	s.Expression = p.parseExpression(STATEMENT)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return s
}

func (p *Parser) parseUpdateExpression() ast.Expression {
	defer p.untrace(p.trace("parseUpdateExpression"))

	x := &ast.UpdateExpression{Token: p.curToken, Branch: p.clause}
	if p.peekTokenIs(token.ONLY) {
		p.nextToken()
		p.nextToken()
		x.Only = true
	}
	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		x.Table = p.parseIdentifier()
	} else {
		msg := "expected IDENT for the table name"
		p.errors = append(p.errors, msg)
		return nil
	}
	if p.peekTokenIs(token.ASTERISK) {
		p.nextToken()
		p.nextToken()
		x.Asterisk = true
	}
	if p.peekTokenIs(token.AS) {
		p.nextToken()
	}
	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		x.Alias = p.parseIdentifier()
	}
	if p.peekTokenIs(token.SET) {
		p.nextToken()
		p.nextToken()
		x.Set = p.parseExpressionList([]token.TokenType{token.FROM, token.WHERE, token.RETURNING, token.SEMICOLON, token.EOF})
	}
	if p.curTokenIs(token.FROM) {
		p.nextToken()
		x.From = p.parseExpressionList([]token.TokenType{token.WHERE, token.RETURNING, token.SEMICOLON, token.EOF})
	}
	// it can be a normal where expression or it could reference a cursor
	if p.curTokenIs(token.WHERE) {
		p.nextToken()
		if p.curTokenIs(token.CURRENT) && p.peekTokenIs(token.OF) {
			p.nextToken()
			p.nextToken()
			x.Cursor = p.parseIdentifier()
		} else {
			x.Where = p.parseExpression(LOWEST)
		}
	}
	// an ExpressionList returns the next token in the peek position (because of the end token), where the parseExpression returns in the curToken position
	if p.peekTokenIs(token.RETURNING) {
		p.nextToken()
	}
	if p.curTokenIs(token.RETURNING) {
		p.nextToken()
		x.Returning = p.parseExpressionList([]token.TokenType{token.SEMICOLON, token.EOF})
	}

	return x
}
