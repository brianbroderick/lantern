package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

// The CTEStatement is just a thin wrapper around a CTEExpression
// because CTEExpressions can show up anywhere in a query
func (p *Parser) parseCTEStatement() *ast.CTEStatement {
	defer untrace(trace("parseCTEStatement " + p.curToken.Lit))

	x := &ast.CTEStatement{Token: p.curToken}
	// Don't advance the token. We want to keep the WITH keyword for the CTEExpression
	x.Expression = p.parseCTEExpression()

	return x
}

func (p *Parser) parseCTEExpression() ast.Expression {
	defer untrace(trace("parseCTEExpression " + p.curToken.Lit))

	x := &ast.CTEExpression{Token: p.curToken}
	p.nextToken()

	if p.curTokenIs(token.RECURSIVE) {
		x.Recursive = true
		p.nextToken()
	}

	x.Expressions = []ast.Expression{}
	x.Expressions = append(x.Expressions, p.parseCTESubExpression())

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		x.Expressions = append(x.Expressions, p.parseCTESubExpression())
	}

	if p.curTokenIs(token.RPAREN) {
		p.nextToken()
	}

	x.Expressions = append(x.Expressions, p.parseExpression(LOWEST)) // Get the main query

	if p.peekTokenIsOne([]token.TokenType{token.SEMICOLON, token.EOF}) {
		p.nextToken()
	}

	return x
}

func (p *Parser) parseCTESubExpression() ast.Expression {
	defer untrace(trace("parseCTESubExpression " + p.curToken.Lit))

	if !p.curTokenIs(token.IDENT) {
		return nil
	}
	tempTable := p.parseExpression(LOWEST)

	// fmt.Printf("parseCTESubExpression: %s %s :: %s %s\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit)

	p.expectPeek(token.AS) // expect AS and move to next token

	materialized := ""
	if p.peekTokenIs(token.NOT) {
		materialized = "NOT "
		p.nextToken()
	}

	if p.peekTokenIs(token.MATERIALIZED) {
		materialized += "MATERIALIZED"
		p.nextToken()
	}

	p.expectPeek(token.LPAREN) // expect LPAREN and move to next token
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	exp.(*ast.SelectExpression).TempTable = tempTable
	exp.(*ast.SelectExpression).WithMaterialized = materialized

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
	}

	return exp
}
