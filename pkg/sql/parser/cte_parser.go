package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

// The CTEStatement is just a thin wrapper around a CTEExpression
// because CTEExpressions can show up anywhere in a query
func (p *Parser) parseCTEStatement() *ast.CTEStatement {
	// defer p.untrace(p.trace("parseCTEStatement"))

	x := &ast.CTEStatement{Token: p.curToken}
	// Don't advance the token. We want to keep the WITH keyword for the CTEExpression
	x.Expression = p.parseCTEExpression()

	return x
}

func (p *Parser) parseCTEExpression() ast.Expression {
	// defer p.untrace(p.trace("parseCTEExpression"))

	x := &ast.CTEExpression{Token: p.curToken, Branch: p.clause}
	p.nextToken()

	if p.curTokenIs(token.RECURSIVE) {
		x.Recursive = true
		p.nextToken()
	}

	x.Auxiliary = []ast.Expression{}
	x.Auxiliary = append(x.Auxiliary, p.parseCTESubExpression())

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		x.Auxiliary = append(x.Auxiliary, p.parseCTESubExpression())
	}

	if p.curTokenIs(token.RPAREN) {
		p.nextToken()
	}

	x.Primary = p.parseExpression(STATEMENT) // Get the main query

	if p.peekTokenIsOne([]token.TokenType{token.SEMICOLON, token.EOF}) {
		p.nextToken()
	}

	return x
}

func (p *Parser) parseCTESubExpression() ast.Expression {
	// defer p.untrace(p.trace("parseCTESubExpression"))

	x := &ast.CTEAuxiliaryExpression{Token: p.curToken, Branch: p.clause}
	x.Name = p.parseExpression(LOWEST)

	p.expectPeek(token.AS) // expect AS and move to next token

	if p.peekTokenIs(token.NOT) {
		x.Materialized = "NOT "
		p.nextToken()
	}

	if p.peekTokenIs(token.MATERIALIZED) {
		x.Materialized += "MATERIALIZED"
		p.nextToken()
	}

	p.expectPeek(token.LPAREN) // expect LPAREN and move to next token
	p.nextToken()

	x.Expression = p.parseExpression(STATEMENT)

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
	}

	return x
}
