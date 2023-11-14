package parser

import (
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func (p *Parser) parseAnalyzeStatement() *ast.AnalyzeStatement {
	defer untrace(trace("parseAnalyzeStatement " + p.curToken.Lit))

	stmt := &ast.AnalyzeStatement{Token: p.curToken}
	options := false
	p.nextToken()

	if p.curTokenIs(token.VERBOSE) {
		stmt.Verbose = true
		p.nextToken()
	}

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		options = true
	}

	if options {
		iter := 0
		for {
			switch p.curToken.Type {
			case token.VERBOSE:
				stmt.Verbose = true
				p.nextToken()
				if p.curTokenIsOne([]token.TokenType{token.ON, token.TRUE}) {
					p.nextToken()
				} else if p.curTokenIsOne([]token.TokenType{token.FALSE, token.IDENT}) {
					stmt.Verbose = false
					p.nextToken()
				}
			case token.IDENT:
				lit := strings.ToUpper(p.curToken.Lit)
				if lit == "BUFFER_USAGE_LIMIT" {
					p.nextToken()
					stmt.BufferUsageLimit = append(stmt.BufferUsageLimit, p.parseExpression(LOWEST))
					p.nextToken()
					if p.curTokenIs(token.IDENT) {
						stmt.BufferUsageLimit = append(stmt.BufferUsageLimit, p.parseExpression(LOWEST))
						p.nextToken()
					}
				} else if lit == "SKIP_LOCKED" {
					stmt.SkipLocked = true
					p.nextToken()
					if p.curTokenIsOne([]token.TokenType{token.ON, token.TRUE}) {
						p.nextToken()
					} else if p.curTokenIsOne([]token.TokenType{token.FALSE, token.IDENT}) {
						stmt.SkipLocked = false
						p.nextToken()
					}
				}
			}

			if p.curTokenIs(token.COMMA) {
				p.nextToken()
			}

			if p.curTokenIs(token.RPAREN) {
				p.nextToken()
				break
			}

			if iter > 5 {
				return &ast.AnalyzeStatement{Token: token.Token{Type: token.ILLEGAL, Lit: "ILLEGAL"}}
			}
			iter++
		}
	}

	if p.curTokenIs(token.VERBOSE) {
		stmt.Verbose = true
		p.nextToken()
	}

	stmt.Name = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
