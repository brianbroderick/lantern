package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

// CASE WHEN id = 1 THEN 'one' WHEN id = 2 THEN 'two' ELSE 'other' end

func (p *Parser) parseCaseExpression() ast.Expression {
	// defer p.untrace(p.trace("parseCaseExpression"))

	x := &ast.CaseExpression{Token: p.curToken, Branch: p.clause}

	if !p.peekTokenIs(token.WHEN) {
		p.nextToken()
		x.Expression = p.parseExpression(LOWEST)
	}

	for p.peekTokenIs(token.WHEN) {
		p.nextToken()
		p.nextToken()
		condition := &ast.ConditionExpression{Token: p.curToken, Branch: p.clause}
		condition.Condition = p.parseExpression(LOWEST)

		// Peek must be THEN
		if !p.expectPeek(token.THEN) {
			return nil
		}
		p.nextToken()

		condition.Consequence = p.parseExpression(LOWEST)
		x.Conditions = append(x.Conditions, condition)
	}

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		p.nextToken()
		x.Alternative = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.END) {
		p.nextToken()
	}

	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		x.SetCast(p.parseDoubleColonExpression())
	}

	return x
}
