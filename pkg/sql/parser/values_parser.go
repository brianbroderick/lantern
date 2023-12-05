package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func (p *Parser) parseValuesExpression() ast.Expression {
	defer p.untrace(p.trace("parseValuesExpression"))
	x := &ast.ValuesExpression{Token: p.curToken}

	if p.curTokenIs(token.VALUES) {
		if p.peekTokenIs(token.LPAREN) {
			p.nextToken()
			p.nextToken()
			values := p.parseExpressionList([]token.TokenType{token.RPAREN})
			x.Tuples = append(x.Tuples, values)
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
