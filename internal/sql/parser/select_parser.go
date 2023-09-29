package parser

import (
	"strings"

	"github.com/brianbroderick/lantern/internal/sql/ast"
	"github.com/brianbroderick/lantern/internal/sql/token"
)

func (p *Parser) parseSelectStatement() *ast.SelectStatement {
	defer untrace(trace("parseSelectStatement1 " + p.curToken.Lit))

	stmt := &ast.SelectStatement{Token: p.curToken}
	stmt.Expressions = []ast.Expression{}
	stmt.Expressions = append(stmt.Expressions, p.parseSelectExpression())

	return stmt
}

func (p *Parser) parseSelectExpression() ast.Expression {
	defer untrace(trace("parseSelectExpression1 " + p.curToken.Lit))

	stmt := &ast.SelectExpression{Token: p.curToken}

	// COLUMNS
	if !p.expectPeekIsOne([]token.TokenType{token.IDENT, token.INT, token.ASTERISK, token.ALL, token.DISTINCT}) {
		return nil
	}

	// DISTINCT CLAUSE
	if p.curTokenIsOne([]token.TokenType{token.ALL, token.DISTINCT}) {
		stmt.Distinct = p.parseDistinct()
	}

	stmt.Columns = p.parseColumnList([]token.TokenType{token.COMMA, token.FROM, token.AS})

	// FROM CLAUSE
	if !p.expectPeek(token.FROM) {
		return nil
	}
	p.nextToken()
	stmt.Tables = p.parseTables()

	// WHERE CLAUSE
	if p.curTokenIs(token.WHERE) {
		p.nextToken()
		stmt.Where = p.parseExpression(LOWEST)
		p.nextToken()
	}

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
	if p.curTokenIs(token.HAVING) {
		p.nextToken()
		stmt.Having = p.parseExpression(LOWEST)
		p.nextToken()
	}

	// ORDER BY CLAUSE
	if p.curTokenIs(token.ORDER) {
		if !p.expectPeek(token.BY) {
			return nil
		}
		p.nextToken()
		stmt.OrderBy = p.parseSortList([]token.TokenType{token.COMMA, token.LIMIT, token.OFFSET})
		p.nextToken()
	}

	// LIMIT CLAUSE
	if p.curTokenIs(token.LIMIT) {
		p.nextToken()
		stmt.Limit = p.parseExpression(LOWEST)
		p.nextToken()
	}

	// OFFSET CLAUSE
	if p.curTokenIs(token.OFFSET) {
		p.nextToken()
		stmt.Offset = p.parseExpression(LOWEST)
		p.nextToken()
		if p.curTokenIsOne([]token.TokenType{token.ROW, token.ROWS}) {
			p.nextToken()
		}
	}

	// FETCH CLAUSE
	if p.curTokenIs(token.FETCH) {
		p.nextToken()
		stmt.Fetch = p.parseFetch()
		p.nextToken()
	}

	// FOR UPDATE CLAUSE
	if p.curTokenIs(token.FOR) {
		p.nextToken()
		stmt.Lock = p.parseLock()
		p.nextToken()
	}

	if p.curTokenIsOne([]token.TokenType{token.SEMICOLON}) {
		p.nextToken()
	}

	// fmt.Printf("parseSelectExpression2: %s :: %s\n", p.curToken.Type, p.peekToken.Type)

	// if p.peekTokenIsOne([]token.TokenType{token.SEMICOLON, token.EOF}) {
	// 	p.nextToken()
	// 	p.nextToken()
	// }

	// if p.curTokenIs(token.RPAREN) {
	// 	p.nextToken()
	// }

	return stmt
}

func (p *Parser) parseDistinct() ast.Expression {
	defer untrace(trace("parseDistinct"))

	if p.curTokenIs(token.ALL) {
		all := &ast.DistinctExpression{Token: p.curToken}
		p.nextToken()
		return all // &ast.DistinctExpression{Token: p.curToken, On: token.Token{Type: token.NIL, Lit: ""}}
	}

	if p.curTokenIs(token.DISTINCT) {
		distinct := &ast.DistinctExpression{Token: p.curToken}
		p.nextToken()

		if p.curTokenIs(token.ON) {
			p.nextToken()

			if p.curTokenIs(token.LPAREN) {
				distinct.Right = p.parseExpressionList([]token.TokenType{token.RPAREN})
				if p.curTokenIs(token.RPAREN) {
					p.nextToken()
				}
			}
		}

		return distinct
	}

	return nil
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

	return &table
}

func (p *Parser) parseTable() ast.Expression {
	defer untrace(trace("parseTable"))

	table := ast.TableExpression{Token: token.Token{Type: token.FROM}}

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

	return &table
}

func (p *Parser) parseFetch() ast.Expression {
	defer untrace(trace("parseFetch"))

	if p.curTokenIs(token.FETCH) {
		p.nextToken()
	}

	fetch := &ast.FetchExpression{Token: p.curToken,
		Value:  &ast.IntegerLiteral{Token: token.Token{Type: token.INT, Lit: "1"}, Value: 1},
		Option: token.Token{Type: token.NIL, Lit: ""}}

	if p.curTokenIsOne([]token.TokenType{token.NEXT, token.FIRST}) {
		p.nextToken()
	}

	if p.curTokenIs(token.INT) {
		fetch.Value = p.parseExpression(LOWEST)
		p.nextToken()
	}

	if p.curTokenIsOne([]token.TokenType{token.ROW, token.ROWS}) {
		p.nextToken()
	}

	if p.curTokenIs(token.ONLY) {
		fetch.Option = p.curToken
		p.nextToken()
	} else if p.curTokenIs(token.WITH) {
		if p.peekTokenIs(token.TIES) {
			p.nextToken()
			fetch.Option = p.curToken
			p.nextToken()
		}
	}
	return fetch
}

func (p *Parser) parseSortList(end []token.TokenType) []ast.Expression {
	defer untrace(trace("parseSortList"))

	list := []ast.Expression{}

	if p.curTokenIsOne(end) {
		return list
	}

	list = append(list, p.parseSort(LOWEST, end))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseSort(LOWEST, end))
	}

	return list
}

func (p *Parser) parseSort(precedence int, end []token.TokenType) ast.Expression {
	defer untrace(trace("parseSort"))
	// fmt.Printf("parseSort1: %s :: %s\n", p.curToken.Lit, p.peekToken.Lit)

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	// fmt.Printf("parseSort2: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, leftExp)

	for !p.peekTokenIsOne([]token.TokenType{token.COMMA}) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	sortExp := &ast.SortExpression{Token: p.curToken, Value: leftExp, Direction: token.Token{Type: token.ASC, Lit: "ASC"}, Nulls: token.Token{Type: token.NIL, Lit: ""}}

	if p.peekTokenIsOne([]token.TokenType{token.ASC, token.DESC}) {
		p.nextToken()
		sortExp.Direction = p.curToken
	}

	if p.peekTokenIs(token.NULLS) {
		p.nextToken()
		if p.peekTokenIsOne([]token.TokenType{token.FIRST, token.LAST}) {
			p.nextToken()
			sortExp.Nulls = p.curToken
		}
	}

	// fmt.Printf("parseSort3: %s :: %s\n", p.curToken.Lit, p.peekToken.Lit)
	return sortExp
}

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
	// fmt.Printf("parseColumn1: %s :: %s\n", p.curToken.Lit, p.peekToken.Lit)

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	// fmt.Printf("parseColumn2: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, leftExp)

	for !p.peekTokenIsOne([]token.TokenType{token.COMMA, token.FROM, token.AS}) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	colExp := &ast.ColumnExpression{Token: p.curToken, Value: leftExp}

	if p.peekTokenIs(token.AS) {
		p.nextToken()
		p.nextToken()
		alias := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}
		colExp = &ast.ColumnExpression{Token: p.curToken, Value: leftExp, Name: alias}
	}

	// fmt.Printf("parseColumn3: %s :: %s\n", p.curToken.Lit, p.peekToken.Lit)
	return colExp
}

func (p *Parser) parseWindowExpression() ast.Expression {
	expression := &ast.WindowExpression{
		Token: p.curToken,
	}

	// fmt.Printf("parseWindowExpression1: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, expression)
	if p.curTokenIs(token.PARTITION) {
		if p.expectPeek(token.BY) {
			p.nextToken()
			expression.PartitionBy = p.parseColumnList([]token.TokenType{token.COMMA, token.ORDER, token.AS, token.RPAREN})
		}
	}
	if p.peekTokenIs(token.ORDER) {
		p.nextToken()
	}

	// fmt.Printf("parseWindowExpression2: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, expression)

	if p.curTokenIs(token.ORDER) {
		if p.expectPeek(token.BY) {
			p.nextToken()
			expression.OrderBy = p.parseSortList([]token.TokenType{token.COMMA, token.AS, token.RPAREN})
		}
	}
	// fmt.Printf("parseWindowExpression3: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, expression)

	// expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseLock() ast.Expression {
	expression := &ast.LockExpression{
		Token: p.curToken,
	}

	switch p.curToken.Type {
	case token.UPDATE:
		expression.Lock = "UPDATE"
		p.nextToken()
	case token.SHARE:
		expression.Lock = "SHARE"
		p.nextToken()
	}

	if p.curTokenIs(token.KEY) {
		p.nextToken()
		if p.curTokenIs(token.SHARE) {
			expression.Lock = "KEY SHARE"
		}
	}

	if p.curTokenIs(token.NO) {
		p.nextToken()
		if p.curTokenIs(token.KEY) {
			p.nextToken()
			if p.curTokenIs(token.UPDATE) {
				p.nextToken()
				expression.Lock = "NO KEY UPDATE"
			}
		}
	}

	if p.curTokenIs(token.OF) {
		expression.Tables = p.parseExpressionList([]token.TokenType{token.NOWAIT, token.SKIP, token.SEMICOLON, token.EOF})
	}

	if p.curTokenIs(token.NOWAIT) {
		expression.Options = "NOWAIT"
	} else if p.curTokenIs(token.SKIP) {
		p.nextToken()
		if p.curTokenIs(token.LOCKED) {
			expression.Options = "SKIP LOCKED"
		}
	}

	p.nextToken()

	return expression
}

func (p *Parser) parseInExpression(left ast.Expression) ast.Expression {
	exp := &ast.InExpression{Token: p.curToken, Operator: p.curToken.Lit, Left: left}
	p.nextToken()
	if p.curTokenIs(token.LPAREN) {
		// p.nextToken()
		exp.Right = p.parseExpressionList([]token.TokenType{token.RPAREN})
	}
	// fmt.Printf("parseInExpression1: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, exp)

	return exp
}

// func (p *Parser) parseWhere(precedence int) ast.Expression {
// 	defer untrace(trace("parseExpression"))

// 	prefix := p.prefixParseFns[p.curToken.Type]
// 	if prefix == nil {
// 		p.noPrefixParseFnError(p.curToken.Type)
// 		return nil
// 	}
// 	leftExp := prefix()

// 	// if p.peekTokenIs(token.IN) {
// 	// 	p.nextToken()
// 	// 	p.nextToken()
// 	// 	infix := &ast.WhereExpression{Token: p.curToken, Left: leftExp, Right: p.parseExpressionList([]token.TokenType{token.RPAREN})}
// 	// }

// 	for !p.peekTokenIsOne([]token.TokenType{token.COMMA, token.WHERE, token.GROUP, token.HAVING, token.ORDER, token.LIMIT, token.OFFSET, token.SEMICOLON}) && precedence < p.peekPrecedence() {
// 		infix := p.infixParseFns[p.peekToken.Type]
// 		if infix == nil {
// 			return leftExp
// 		}

// 		p.nextToken()

// 		leftExp = infix(leftExp)
// 	}

// 	return leftExp
// }
