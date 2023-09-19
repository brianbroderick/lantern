package parser

import (
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/internal/sql/ast"
	"github.com/brianbroderick/lantern/internal/sql/token"
)

func (p *Parser) parseSelectStatement() *ast.SelectStatement {
	defer untrace(trace("parseSelectStatement1 " + p.curToken.Lit))

	stmt := &ast.SelectStatement{Token: p.curToken}

	// COLUMNS
	if !p.expectPeekIsOne([]token.TokenType{token.IDENT, token.INT, token.ASTERISK}) {
		return nil
	}
	stmt.Columns = p.parseColumnList([]token.TokenType{token.COMMA, token.FROM, token.AS})

	fmt.Printf("from: %s :: %s\n", p.curToken.Lit, p.peekToken.Lit)

	// FROM CLAUSE
	if !p.expectPeek(token.FROM) {
		return nil
	}
	p.nextToken()
	stmt.Tables = p.parseTables()

	fmt.Printf("where: %s :: %s\n", p.curToken.Lit, p.peekToken.Lit)

	// WHERE CLAUSE
	if p.curTokenIs(token.WHERE) {
		p.nextToken()
		stmt.Where = p.parseExpression(LOWEST)
		p.nextToken()
	}

	fmt.Printf("group: %s :: %s\n", p.curToken.Lit, p.peekToken.Lit)

	// GROUP BY CLAUSE
	if p.curTokenIs(token.GROUP) {
		if !p.expectPeek(token.BY) {
			return nil
		}
		p.nextToken()
		stmt.GroupBy = p.parseColumnList([]token.TokenType{token.COMMA, token.HAVING, token.ORDER, token.LIMIT, token.OFFSET})
		p.nextToken()
	}

	// HAVING CLAUSE

	// fmt.Printf("parseSelectStatement2: %s :: %s\n", p.curToken.Lit, p.peekToken.Lit)

	if p.expectPeekIsOne([]token.TokenType{token.SEMICOLON, token.EOF}) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseTables() []ast.Expression {
	defer untrace(trace("parseTables"))

	tables := []ast.Expression{}

	if p.curTokenIsOne([]token.TokenType{token.EOF, token.SEMICOLON}) {
		return tables
	}

	tables = append(tables, p.parseFirstTable())

	for p.curTokenIsOne([]token.TokenType{token.JOIN, token.INNER, token.LEFT, token.RIGHT, token.FULL, token.CROSS, token.LATERAL}) {
		tables = append(tables, p.parseTable())
	}

	return tables
}

func (p *Parser) parseFirstTable() ast.Expression {
	defer untrace(trace("parseFirstTable"))

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
	}

	fmt.Printf("parseFirstTable: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, table)

	return &table
}

func (p *Parser) parseTable() ast.Expression {
	defer untrace(trace("parseTable"))

	table := ast.TableExpression{Token: token.Token{Type: token.FROM}}

	fmt.Printf("parseTable1: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, table)

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

	fmt.Printf("parseTable2: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, table)

	// Get the table name
	if p.curTokenIs(token.IDENT) {
		table.Table = p.curToken.Lit
		p.nextToken()
	}

	// Do we have an alias?
	if p.curTokenIs(token.IDENT) {
		table.Alias = p.curToken.Lit
		p.nextToken()
	}

	// Get the join condition
	if p.curTokenIs(token.ON) {
		p.nextToken()
		table.JoinCondition = p.parseExpression(LOWEST)
		p.nextToken()
	}
	fmt.Printf("parseTable3: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, table)

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

func (p *Parser) parseColumnList(end []token.TokenType) []ast.Expression {
	defer untrace(trace("parseColumnList"))

	list := []ast.Expression{}

	if p.curTokenIsOne(end) {
		return list
	}

	list = append(list, p.parseColumn(LOWEST, end))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseColumn(LOWEST, end))
	}

	return list
}

func (p *Parser) parseColumn(precedence int, end []token.TokenType) ast.Expression {
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
