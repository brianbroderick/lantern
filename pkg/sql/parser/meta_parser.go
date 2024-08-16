package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func (p *Parser) parseShowStatement() *ast.ShowStatement {
	p.clause = token.SHOW_STATEMENT
	p.command = token.SHOW_STATEMENT

	stmt := &ast.ShowStatement{Token: token.Token{Type: token.SHOW_STATEMENT, Lit: "SHOW", Upper: "SHOW"}}
	x := &ast.ShowExpression{Token: token.Token{Type: token.SHOW_STATEMENT, Lit: "SHOW", Upper: "SHOW"}}
	p.nextToken()
	if p.curTokenIs(token.IDENT) && p.curToken.Upper == "TRANSACTION" {
		txn := &ast.TransactionExpression{Token: p.curToken}
		p.nextToken()
		if p.curTokenIs(token.IDENT) && p.curToken.Upper == "ISOLATION" {
			p.nextToken()
		}
		if p.curTokenIs(token.IDENT) && p.curToken.Upper == "LEVEL" {
			p.nextToken()
		}
		if p.curTokenIs(token.IDENT) {
			switch p.curToken.Upper {
			case "SERIALIZABLE":
				txn.IsolationLevel = "SERIALIZABLE"
			case "REPEATABLE":
				if p.peekTokenIs(token.IDENT) && p.peekToken.Upper == "READ" {
					p.nextToken()
					txn.IsolationLevel = "REPEATABLE READ"
				}
			case "READ":
				if p.peekTokenIs(token.IDENT) && p.peekToken.Upper == "UNCOMMITTED" {
					p.nextToken()
					txn.IsolationLevel = "READ UNCOMMITTED"
				} else if (p.peekTokenIs(token.IDENT) && p.peekToken.Upper == "COMMITTED") || (p.peekTokenIs(token.IDENT) && p.peekToken.Upper == "WRITE") {
					p.nextToken()
					txn.IsolationLevel = "READ COMMITTED"
				}
			}
			p.nextToken()
		}
		if p.curTokenIs(token.IDENT) && p.curToken.Upper == "READ" {
			p.nextToken()
			if p.curTokenIs(token.IDENT) && p.curToken.Upper == "ONLY" {
				txn.Rights = "READ ONLY"
			} else if p.curTokenIs(token.IDENT) && p.curToken.Upper == "WRITE" {
				txn.Rights = "READ WRITE"
			}
		}
		if p.curTokenIs(token.IDENT) && p.curToken.Upper == "NOT" {
			p.nextToken()
			if p.curTokenIs(token.IDENT) && p.curToken.Upper == "DEFERRABLE" {
				txn.Deferrable = "NOT DEFERRABLE"
			}
		}
		if p.curTokenIs(token.IDENT) && p.curToken.Upper == "DEFERRABLE" {
			txn.Deferrable = "DEFERRABLE"
		}
		x.Expression = txn
	} else {
		x.Expression = p.parseExpression(LOWEST)
	}

	stmt.Expression = x

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseSavepointStatement() *ast.SavepointStatement {
	stmt := &ast.SavepointStatement{Token: token.Token{Type: token.SAVEPOINT_STATEMENT, Lit: "SAVEPOINT", Upper: "SAVEPOINT"}}
	p.nextToken()
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
