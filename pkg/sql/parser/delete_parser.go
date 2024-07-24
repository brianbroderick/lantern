package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func (p *Parser) parseDeleteStatement() *ast.DeleteStatement {
	defer p.untrace(p.trace("parseDeleteStatement"))

	s := &ast.DeleteStatement{Token: p.curToken}
	s.Expression = p.parseExpression(STATEMENT)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return s
}

func (p *Parser) parseDeleteExpression() ast.Expression {
	defer p.untrace(p.trace("parseDeleteExpression"))

	p.command = token.DELETE

	x := &ast.DeleteExpression{Token: p.curToken, Branch: p.clause, CommandTag: p.command}

	if p.peekTokenIs(token.FROM) {
		p.nextToken()
	}

	if p.peekTokenIs(token.ONLY) {
		p.nextToken()
		x.Only = true
	}

	// fmt.Printf("p.curToken: %s p.peekToken: %s\n", p.curToken.Lit, p.peekToken.Lit)

	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		x.Table = p.parseIdentifier()
	} else {
		msg := "expected IDENT for the table name"
		p.errors = append(p.errors, msg)
		return nil
	}

	if p.peekTokenIs(token.AS) {
		p.nextToken()
	}
	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		x.Alias = p.parseIdentifier()
	}

	if p.peekTokenIs(token.USING) {
		p.nextToken()
		p.nextToken()
		// x.Using = p.parseExpressionList([]token.TokenType{token.WHERE, token.RETURNING, token.SEMICOLON, token.EOF})
		x.Using, x.TableAliases = p.parseTables()
	}

	if p.peekTokenIs(token.WHERE) {
		p.nextToken()
	}
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

	if p.peekTokenIs(token.RETURNING) {
		p.nextToken()
		p.nextToken()
		x.Returning = p.parseExpressionList([]token.TokenType{token.SEMICOLON, token.EOF})
	}

	return x
}
