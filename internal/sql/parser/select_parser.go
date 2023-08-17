package parser

import (
	"github.com/brianbroderick/lantern/internal/sql/ast"
	"github.com/brianbroderick/lantern/internal/sql/token"
)

func (p *Parser) parseSelectStatement() *ast.SelectStatement {
	defer untrace(trace("parseSelectStatement1 " + p.curToken.Lit))

	stmt := &ast.SelectStatement{Token: p.curToken}

	if !p.expectPeekIsOne([]token.TokenType{token.IDENT, token.INT, token.ASTERISK}) {
		return nil
	}

	stmt.Columns = p.parseColumnList(token.FROM)
	// fmt.Printf("parseSelectStatement2: %s %s\n", p.curToken.Lit, p.peekToken.Lit)

	if !p.expectPeek(token.FROM) {
		return nil
	}
	p.nextToken()

	stmt.From = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}

	if p.expectPeekIsOne([]token.TokenType{token.SEMICOLON, token.EOF}) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseColumnList(end token.TokenType) []ast.Expression {
	defer untrace(trace("parseColumnList"))

	list := []ast.Expression{}

	// fmt.Printf("parseColumnList1: %s %s\n", p.curToken.Lit, p.peekToken.Lit)

	if p.curTokenIs(end) {
		return list
	}

	list = append(list, p.parseColumn(LOWEST))

	// fmt.Printf("parseColumnList2: %s %s\n", p.curToken.Lit, p.peekToken.Lit)

	for p.peekTokenIs(token.COMMA) {
		// fmt.Printf("parseColumnList3: %s %s\n", p.curToken.Lit, p.peekToken.Lit)
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseColumn(LOWEST))
	}

	// fmt.Printf("parseColumnList4: %s %s\n", p.curToken.Lit, p.peekToken.Lit)
	return list
}

func (p *Parser) parseColumn(precedence int) ast.Expression {
	defer untrace(trace("parseColumn"))
	// fmt.Printf("parseColumn1: %s %s\n", p.curToken.Lit, p.peekToken.Lit)

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	// fmt.Printf("parseColumn2: %s %s\n", p.curToken.Lit, p.peekToken.Lit)

	for !p.peekTokenIs(token.COMMA) && !p.peekTokenIs(token.FROM) && !p.peekTokenIs(token.AS) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}
	if p.peekTokenIs(token.AS) {
		p.nextToken()
		p.nextToken()
		alias := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}
		leftExp = &ast.ColumnExpression{Token: p.curToken, Value: leftExp, Name: alias}
	}

	// fmt.Printf("parseColumn3: %s %s\n", p.curToken.Lit, p.peekToken.Lit)
	return leftExp
}
