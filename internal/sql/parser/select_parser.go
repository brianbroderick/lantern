package parser

import (
	"strings"

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

	stmt.Tables = p.parseTables()
	stmt.From = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}

	if p.expectPeekIsOne([]token.TokenType{token.SEMICOLON, token.EOF}) {
		p.nextToken()
	}

	return stmt
}

// customers inner join addresses on c.id = a.customer_id
// join_type table     alias join_condition
// source    customers c
// inner     addresses a     Expression

func (p *Parser) parseTables() []ast.Expression {
	defer untrace(trace("parseTables"))

	tables := []ast.Expression{}

	if p.curTokenIsOne([]token.TokenType{token.EOF, token.SEMICOLON}) {
		return tables
	}

	tables = append(tables, p.parseTable())

	// for p.peekTokenIs(token.COMMA) {
	// 	p.nextToken()
	// 	p.nextToken()
	// 	tables = append(tables, p.parseTable(LOWEST))
	// }

	return tables
}

func (p *Parser) parseTable() ast.Expression {
	defer untrace(trace("parseTable"))

	table := ast.TableExpression{Token: token.Token{Type: token.FROM}}

	// Get the first table
	if p.curTokenIs(token.IDENT) {
		table.Table = p.curToken.Lit
		p.nextToken()

		// Do we have an alias
		if p.curTokenIs(token.IDENT) {
			table.Alias = p.curToken.Lit
			p.nextToken()
		}

		return &table
	} // TODO: else if p.curTokenIs(token.LPAREN) { ... } // subquery

	// Get the join type
	if p.curTokenIsOne([]token.TokenType{token.INNER, token.LEFT, token.RIGHT, token.FULL, token.CROSS, token.LATERAL}) {
		table.JoinType = strings.ToUpper(p.peekToken.Lit)
		p.nextToken()

		// Skip the JOIN and OUTER keywords
		for p.curTokenIsOne([]token.TokenType{token.JOIN, token.OUTER}) {
			p.nextToken()
		}
	}

	// If just using JOIN, assume INNER
	if p.curTokenIs(token.JOIN) {
		if table.JoinType == "" {
			table.JoinType = "INNER"
		}
		p.nextToken()
	}

	// Get the table name
	if p.curTokenIs(token.IDENT) {
		table.Table = p.curToken.Lit
		p.nextToken()
	}

	// Peek forward and see if we have an alias
	if p.peekTokenIs(token.IDENT) {
		table.Alias = p.peekToken.Lit
		p.nextToken()
		p.nextToken()
	}

	// Get the join condition
	if p.curTokenIs(token.ON) {
		p.nextToken()
		table.JoinCondition = p.parseExpression(LOWEST)
	}

	return &table
}

// func (p *Parser) parseJoinList() []ast.Expression {
// 	defer untrace(trace("parseJoins"))
// 	// fmt.Printf("parseJoins1: %s %s\n", p.curToken.Lit, p.peekToken.Lit)
// 	joins := []ast.Expression{}

// 	if p.curTokenIsOne([]token.TokenType{token.EOF, token.SEMICOLON}) {
// 		return joins
// 	}

// 	// START HERE: working on getting the table and then joins loop working

// 	// Get the first table
// 	// joins = append(joins, p.parseJoin(LOWEST))
// 	table := p.parseJoin(LOWEST)

// 	for p.peekTokenIsOne([]token.TokenType{token.INNER, token.LEFT, token.RIGHT, token.FULL, token.CROSS, token.JOIN}) {
// 		p.nextToken()

// 		if p.curTokenIs(token.JOIN) {
// 			p.nextToken()
// 		}

// 		if p.peekTokenIs(token.JOIN) {
// 			p.nextToken()
// 		}

// 		p.nextToken()
// 		p.nextToken()
// 		joins = append(joins, p.parseJoin(LOWEST))
// 	}

// 	return joins
// }

// func (p *Parser) parseJoin(precedence int) ast.Expression {
// 	defer untrace(trace("parseJoin"))
// 	// fmt.Printf("parseJoin1: %s %s\n", p.curToken.Lit, p.peekToken.Lit)

// 	prefix := p.prefixParseFns[p.curToken.Type]
// 	if prefix == nil {
// 		p.noPrefixParseFnError(p.curToken.Type)
// 		return nil
// 	}
// 	leftExp := prefix()

// 	// fmt.Printf("parseJoin2: %s %s\n", p.curToken.Lit, p.peekToken.Lit)

// 	for !p.peekTokenIsOne([]token.TokenType{token.INNER, token.LEFT, token.RIGHT, token.FULL, token.CROSS, token.JOIN}) && precedence < p.peekPrecedence() {
// 		infix := p.infixParseFns[p.peekToken.Type]
// 		if infix == nil {
// 			return leftExp
// 		}

// 		p.nextToken()

// 		leftExp = infix(leftExp)
// 	}

// 	// fmt.Printf("parseJoin3: %s %s\n", p.curToken.Lit, p.peekToken.Lit)
// 	return leftExp
// }

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

	for !p.peekTokenIsOne([]token.TokenType{token.COMMA, token.FROM, token.AS}) && precedence < p.peekPrecedence() {
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
