package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

// CASE WHEN id = 1 THEN 'one' WHEN id = 2 THEN 'two' ELSE 'other' end

func (p *Parser) parseCaseExpression() ast.Expression {
	expression := &ast.CaseExpression{Token: p.curToken}

	for p.peekTokenIs(token.WHEN) {
		p.nextToken()
		p.nextToken()
		condition := &ast.ConditionExpression{Token: p.curToken}
		condition.Condition = p.parseExpression(LOWEST)

		// Peek must be THEN
		if !p.expectPeek(token.THEN) {
			return nil
		}
		p.nextToken()

		condition.Consequence = p.parseExpression(LOWEST)
		expression.Conditions = append(expression.Conditions, condition)
	}

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		p.nextToken()
		expression.Alternative = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.END) {
		p.nextToken()
	}

	return expression
}
