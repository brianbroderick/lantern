package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func (p *Parser) parseAnalyzeStatement() *ast.AnalyzeStatement {
	defer untrace(trace("parseAnalyzeStatement " + p.curToken.Lit))

	stmt := &ast.AnalyzeStatement{Token: p.curToken}
	p.nextToken()
	stmt.Name = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
