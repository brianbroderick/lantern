package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func (p *Parser) parseValuesExpression() ast.Expression {
	defer p.untrace(p.trace("parseValuesExpression"))
	x := &ast.ValuesExpression{Token: p.curToken, Branch: p.clause, CommandTag: p.command}

	if p.curTokenIs(token.VALUES) {
		if p.peekTokenIs(token.LPAREN) {
			p.nextToken()
			p.nextToken()
			values := p.parseExpressionList([]token.TokenType{token.RPAREN})
			x.Tuples = append(x.Tuples, values)
		} else {
			// A values expression is only when it's a function call
			return &ast.SimpleIdentifier{Token: p.curToken, Value: p.curToken.Lit, Branch: p.clause, CommandTag: p.command}
		}
	}

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if p.peekTokenIs(token.LPAREN) {
			p.nextToken()
			p.nextToken()
		}
		values := p.parseExpressionList([]token.TokenType{token.RPAREN})
		x.Tuples = append(x.Tuples, values)
	}

	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
	}

	return x
}
