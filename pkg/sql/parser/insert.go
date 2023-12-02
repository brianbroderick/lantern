package parser

import (
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func (p *Parser) parseInsertStatement() *ast.InsertStatement {
	defer p.untrace(p.trace("parseInsertStatement"))

	s := &ast.InsertStatement{Token: p.curToken}

	s.Expression = p.parseExpression(STATEMENT)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return s
}

func (p *Parser) parseInsertExpression() ast.Expression {
	defer p.untrace(p.trace("parseInsertExpression"))

	x := &ast.InsertExpression{Token: p.curToken}

	if !p.expectPeek(token.INTO) {
		return nil
	}
	p.nextToken()

	if p.curTokenIs(token.IDENT) {
		x.Table = p.parseIdentifier()
		p.nextToken()
	} else {
		msg := fmt.Sprintf("expected %q to be an IDENT", p.curToken.Lit)
		p.errors = append(p.errors, msg)
		return nil
	}

	if p.curTokenIs(token.AS) {
		p.nextToken()
		x.Alias = p.parseIdentifier()
		p.nextToken()
	}

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		x.Columns = p.parseExpressionList([]token.TokenType{token.RPAREN})
		p.nextToken()
	}

	if p.curTokenIs(token.DEFAULT) {
		x.Default = true
		p.nextToken()
	}

	if p.curTokenIs(token.VALUES) {
		p.nextToken()
		p.nextToken()
		values := p.parseExpressionList([]token.TokenType{token.RPAREN})
		x.Values = append(x.Values, values)
		p.nextToken()
	}

	for p.curTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		values := p.parseExpressionList([]token.TokenType{token.RPAREN})
		x.Values = append(x.Values, values)

	}

	// fmt.Printf("parseInsertExpression8: %s %s | %+v\n", p.curToken.Lit, p.peekToken.Lit, x.String(false))

	return x
}
