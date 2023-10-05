package parser

import (
	"github.com/brianbroderick/lantern/internal/sql/ast"
	"github.com/brianbroderick/lantern/internal/sql/token"
)

func (p *Parser) parseCTEStatement() *ast.CTEStatement {
	defer untrace(trace("parseCTEStatement " + p.curToken.Lit))

	stmt := &ast.CTEStatement{Token: p.curToken}
	p.nextToken()
	stmt.Expressions = []ast.Expression{}
	stmt.Expressions = append(stmt.Expressions, p.parseCTEExpression())
	// if p.peekTokenIs(token.COMMA) { // handle temp table list
	// }

	for !p.peekTokenIs(token.SEMICOLON) {
		stmt.Expressions = append(stmt.Expressions, p.parseExpression(LOWEST)) // Get the main query
	}

	p.nextToken() // We're done. Move on to the next statement

	// fmt.Println("parseCTEStatement: ", stmt)

	return stmt
}

func (p *Parser) parseCTEExpression() ast.Expression {
	defer untrace(trace("parseCTEExpression " + p.curToken.Lit))

	// fmt.Printf("parseCTEExpression000: %s %s :: %s %s\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit)

	if !p.curTokenIs(token.IDENT) {
		return nil
	}

	// fmt.Printf("parseCTEExpression001: %s %s :: %s %s\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit)

	tempTable := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}

	p.expectPeek(token.AS)     // expect AS and move to next token
	p.expectPeek(token.LPAREN) // expect LPAREN and move to next token
	p.nextToken()

	// fmt.Printf("parseCTEExpression002: %s %s :: %s %s\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit)

	exp := p.parseExpression(LOWEST)
	exp.(*ast.SelectExpression).TempTable = tempTable

	// fmt.Printf("parseCTEExpression003: %s %s :: %s %s == %+v\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit, exp)

	if p.peekTokenIs(token.RPAREN) {
		// fmt.Println("parseCTEExpression004")
		p.nextToken()
		p.nextToken()
	}

	return exp
}
