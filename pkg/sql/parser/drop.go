package parser

import (
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func (p *Parser) parseDropStatement() *ast.DropStatement {
	defer untrace(trace("parseDropStatement " + p.curToken.Lit))

	stmt := &ast.DropStatement{Token: p.curToken}
	p.nextToken()
	stmt.Object = p.curToken.Lit

	p.nextToken()

	if p.curTokenIs(token.IDENT) && strings.ToLower(p.curToken.Lit) == "if" {
		p.nextToken()
		if p.curTokenIs(token.IDENT) && strings.ToLower(p.curToken.Lit) == "exists" {
			stmt.Exists = true
			p.nextToken()
		}
	}

	if p.curTokenIs(token.IDENT) {
		stmt.Tables = p.parseColumnList([]token.TokenType{token.COMMA})
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
