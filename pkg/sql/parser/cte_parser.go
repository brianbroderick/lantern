package parser

import (
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func (p *Parser) parseCTEStatement() *ast.CTEStatement {
	defer untrace(trace("parseCTEStatement " + p.curToken.Lit))

	stmt := &ast.CTEStatement{Token: p.curToken}
	p.nextToken()

	if p.curTokenIs(token.RECURSIVE) {
		stmt.Recursive = true
		p.nextToken()
	}

	stmt.Expressions = []ast.Expression{}
	stmt.Expressions = append(stmt.Expressions, p.parseCTEExpression())

	// fmt.Printf("parseCTEStatement000: %s %s :: %s %s\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		stmt.Expressions = append(stmt.Expressions, p.parseCTEExpression())
	}

	iter := 0
	for p.peekTokenIsNot([]token.TokenType{token.SEMICOLON, token.EOF}) {
		if p.curTokenIs(token.RPAREN) {
			p.nextToken()
		}
		stmt.Expressions = append(stmt.Expressions, p.parseExpression(LOWEST)) // Get the main query

		iter++ // This is a hack to prevent an infinite loop. If we're looping 10 times, something's wrong. Bail out.
		if iter > 10 {
			fmt.Println("parseCTEStatement: Infinite loop detected")
			fmt.Printf("parseCTEStatement: %s %s :: %s %s\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit)
			return &ast.CTEStatement{Token: token.Token{Type: token.ILLEGAL, Lit: "ILLEGAL"}}
		}
	}

	p.nextToken() // We're done. Move on to the next statement

	// fmt.Println("parseCTEStatement: ", stmt)

	return stmt
}

func (p *Parser) parseCTEExpression() ast.Expression {
	defer untrace(trace("parseCTEExpression " + p.curToken.Lit))

	// fmt.Printf("parseCTEExpression: %s %s :: %s %s == %+v\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit, p.context)

	// fmt.Printf("parseCTEExpression000: %s %s :: %s %s\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit)

	if !p.curTokenIs(token.IDENT) {
		return nil
	}

	// fmt.Printf("parseCTEExpression001: %s %s :: %s %s\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit)

	// tempTable := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}
	tempTable := p.parseExpression(LOWEST)

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

	// fmt.Printf("parseCTEExpression002: %s %s :: %s %s\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit)

	exp := p.parseExpression(LOWEST)
	exp.(*ast.SelectExpression).TempTable = tempTable
	exp.(*ast.SelectExpression).WithMaterialized = materialized

	// fmt.Printf("parseCTEExpression003: %s %s :: %s %s == %+v\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit, exp)

	if p.peekTokenIs(token.RPAREN) {
		// fmt.Println("parseCTEExpression004")
		p.nextToken()
	}

	return exp
}
