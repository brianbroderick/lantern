package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func (p *Parser) parseTimestampExpression() ast.Expression {
	defer untrace(trace("parseTimestampExpression " + p.curToken.Lit))

	x := &ast.TimestampExpression{Token: p.curToken}

	// timestamp with time zone
	if p.peekTokenIs(token.WITH) {
		if p.peekTwoToken.Type == token.IDENT && p.peekTwoToken.Upper == "TIME" {
			if p.peekThreeToken.Type == token.IDENT && p.peekThreeToken.Upper == "ZONE" {
				p.nextToken()
				p.nextToken()
				p.nextToken()
				x.WithTimeZone = true
			}
		}
	}

	return x
}
