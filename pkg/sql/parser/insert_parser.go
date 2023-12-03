package parser

import (
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func (p *Parser) parseInsertStatement() *ast.InsertStatement {
	defer p.untrace(p.trace("parseInsertStatement"))

	s := &ast.InsertStatement{Token: p.curToken}

	s.Expression = p.parseExpression(STATEMENT)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return s
}

func (p *Parser) parseInsertExpression() ast.Expression {
	defer p.untrace(p.trace("parseInsertExpression"))

	x := &ast.InsertExpression{Token: p.curToken}

	if !p.expectPeek(token.INTO) {
		return nil
	}

	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		x.Table = p.parseIdentifier()
	} else {
		msg := fmt.Sprintf("expected %q to be an IDENT", p.curToken.Lit)
		p.errors = append(p.errors, msg)
		return nil
	}

	if p.peekTokenIs(token.AS) {
		p.nextToken()
		p.nextToken()
		x.Alias = p.parseIdentifier()
	}

	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		p.nextToken()
		x.Columns = p.parseExpressionList([]token.TokenType{token.RPAREN})
	}

	if p.peekTokenIs(token.DEFAULT) {
		p.nextToken()
		p.nextToken()
		x.Default = true
	}

	if p.peekTokenIs(token.VALUES) {
		p.nextToken()
		if p.peekTokenIs(token.LPAREN) {
			p.nextToken()
			p.nextToken()
			values := p.parseExpressionList([]token.TokenType{token.RPAREN})
			x.Values = append(x.Values, values)
		}
	}

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if p.peekTokenIs(token.LPAREN) {
			p.nextToken()
			p.nextToken()
		}
		values := p.parseExpressionList([]token.TokenType{token.RPAREN})
		x.Values = append(x.Values, values)
	}

	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
	}

	if p.peekTokenIs(token.SELECT) {
		p.nextToken()
		x.Query = p.parseSelectExpression()
	}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
	}

	if p.peekTokenIs(token.ON) {
		p.nextToken()
		if p.peekTokenIs(token.CONFLICT) {
			p.nextToken()
			if p.peekTokenIs(token.LPAREN) {
				p.nextToken()
				p.nextToken()
				x.ConflictTarget = p.parseExpressionList([]token.TokenType{token.RPAREN})
			}
		}
		if p.peekTokenIs(token.DO) {
			p.nextToken()
			p.nextToken()
			x.ConflictAction = p.curToken.Upper
			if p.curTokenIs(token.UPDATE) && p.peekTokenIs(token.SET) {
				p.nextToken()
				p.nextToken()
				if p.peekTokenIs(token.LPAREN) {
					p.nextToken()
					p.nextToken()
				}
				x.ConflictUpdate = p.parseExpressionList([]token.TokenType{token.RPAREN, token.RETURNING, token.SEMICOLON, token.EOF, token.WHERE})

				if p.curTokenIs(token.WHERE) {
					p.nextToken()
					x.ConflictWhere = p.parseExpression(LOWEST)
				}
			}
		}
	}

	if p.peekTokenIs(token.RETURNING) {
		p.nextToken()
		p.nextToken()
		x.Returning = p.parseExpressionList([]token.TokenType{token.SEMICOLON})
	}

	// fmt.Printf("parseInsertExpression8: %s %s | %+v\n", p.curToken.Lit, p.peekToken.Lit, x.String(false))

	return x
}
