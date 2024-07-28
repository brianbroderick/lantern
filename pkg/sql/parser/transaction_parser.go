package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func (p *Parser) parseCommitStatement() *ast.CommitStatement {
	defer p.untrace(p.trace("parseCommitStatement"))

	s := &ast.CommitStatement{Token: p.curToken}
	s.Expression = p.parseExpression(STATEMENT)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return s
}

func (p *Parser) parseCommitExpression() ast.Expression {
	defer p.untrace(p.trace("parseCommitExpression"))

	return &ast.CommitExpression{Token: p.curToken, Branch: p.clause, CommandTag: p.command}
}

func (p *Parser) parseRollbackStatement() *ast.RollbackStatement {
	defer p.untrace(p.trace("parseRollbackStatement"))

	s := &ast.RollbackStatement{Token: p.curToken}
	s.Expression = p.parseExpression(STATEMENT)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return s
}

func (p *Parser) parseRollbackExpression() ast.Expression {
	defer p.untrace(p.trace("parseRollbackExpression"))

	return &ast.RollbackExpression{Token: p.curToken, Branch: p.clause, CommandTag: p.command}
}
