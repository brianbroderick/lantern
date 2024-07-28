package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

// SET application_name = 'example';

func (p *Parser) parseSetStatement() *ast.SetStatement {
	// defer untrace(trace("parseSetStatement " + p.curToken.Lit))

	stmt := &ast.SetStatement{Token: p.curToken}
	p.nextToken()
	if p.curTokenIs(token.SESSION) {
		stmt.Session = true
		p.nextToken()
	}

	// Set time zone
	if p.curTokenIs(token.LOCAL) {
		stmt.Local = true
		p.nextToken()
	}
	if p.curTokenIs(token.IDENT) && p.curToken.Upper == "TIME" {
		stmt.TimeZone = true
		p.nextToken()
		if p.curTokenIs(token.IDENT) && p.curToken.Upper == "ZONE" {
			p.nextToken()
		}
	}

	// Set session characteristics as transaction isolation level read uncommitted;
	if p.curTokenIs(token.IDENT) && p.curToken.Upper == "CHARACTERISTICS" {
		stmt.HasCharacteristics = true
		p.nextToken()
		if p.curTokenIs(token.AS) {
			p.nextToken()
		}
	}

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
		stmt.Expression = txn
	}

	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
		return stmt
	}
	if p.curTokenIs(token.EOF) {
		return stmt
	}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
