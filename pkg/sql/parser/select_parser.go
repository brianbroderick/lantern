package parser

import (
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

// Clauses are parsed in the order that they appear in the SQL statement.
// Each Expression starts with the token after the keyword that starts the clause.
// To determine if a clause is done, peek to the next token and look for the next clause keyword.
// If it's found, advance the token twice to skip past the clause token (from peek, to current, to next).
// Otherwise, you'll get off by one errors and have a hard time figuring out why.

var defaultListSeparators = []token.TokenType{token.COMMA, token.WHERE, token.GROUP, token.HAVING, token.ORDER, token.LIMIT, token.OFFSET, token.FETCH, token.FOR, token.SEMICOLON}

func (p *Parser) parseSelectStatement() *ast.SelectStatement {
	defer untrace(trace("parseSelectStatement1 " + p.curToken.Lit))

	stmt := &ast.SelectStatement{Token: p.curToken}
	stmt.Expressions = []ast.Expression{}
	stmt.Expressions = append(stmt.Expressions, p.parseSelectExpression())

	for p.peekTokenIsOne([]token.TokenType{token.UNION, token.INTERSECT, token.EXCEPT}) {
		p.nextToken()
		stmt.Expressions = append(stmt.Expressions, p.parseUnionExpression())
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseSelectExpression() ast.Expression {
	defer untrace(trace("parseSelectExpression " + p.curToken.Lit))

	stmt := &ast.SelectExpression{Token: p.curToken}

	// COLUMNS
	if !p.expectPeekIsOne([]token.TokenType{token.LPAREN, token.IDENT, token.INT, token.STRING, token.ASTERISK, token.ALL, token.DISTINCT, token.CASE}) {
		return nil
	}

	// DISTINCT CLAUSE
	if p.curTokenIsOne([]token.TokenType{token.ALL, token.DISTINCT}) {
		stmt.Distinct = p.parseDistinct()
	}

	stmt.Columns = p.parseColumnList([]token.TokenType{token.COMMA, token.FROM, token.AS})

	// FROM CLAUSE: if the next token is FROM, advance the token and move on. Otherwise, return the statement.
	if p.peekTokenIs(token.FROM) {
		p.nextToken()
		p.nextToken()
	} else {
		return stmt
	}

	// fmt.Printf("parseSelectExpression001: %s %s :: %s %s == %+v\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit, stmt)

	// p.nextToken()
	stmt.Tables = p.parseTables()

	// fmt.Printf("parseSelectExpression002: %s %s :: %s %s == %+v\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit, stmt)

	// WINDOW CLAUSE
	if p.peekTokenIs(token.WINDOW) {
		p.nextToken()
		p.nextToken()
		stmt.Window = p.parseWindowList(defaultListSeparators)
	}

	// WHERE CLAUSE
	if p.peekTokenIs(token.WHERE) {
		p.nextToken()
		p.nextToken()
		stmt.Where = p.parseExpression(LOWEST)
	}

	// GROUP BY CLAUSE
	if p.peekTokenIs(token.GROUP) {
		p.nextToken()
		if !p.expectPeek(token.BY) {
			return nil
		}
		p.nextToken()
		stmt.GroupBy = p.parseColumnList(defaultListSeparators)
	}

	// HAVING CLAUSE
	if p.peekTokenIs(token.HAVING) {
		p.nextToken()
		p.nextToken()
		stmt.Having = p.parseExpression(LOWEST)
	}

	// ORDER BY CLAUSE
	if p.peekTokenIs(token.ORDER) {
		p.nextToken()
		if !p.expectPeek(token.BY) {
			return nil
		}
		p.nextToken()
		stmt.OrderBy = p.parseSortList(defaultListSeparators)
	}

	// LIMIT CLAUSE
	if p.peekTokenIs(token.LIMIT) {
		p.nextToken()
		p.nextToken()
		stmt.Limit = p.parseExpression(LOWEST)
	}

	// OFFSET CLAUSE
	if p.peekTokenIs(token.OFFSET) {
		p.nextToken()
		p.nextToken()
		stmt.Offset = p.parseExpression(LOWEST)

		if p.peekTokenIsOne([]token.TokenType{token.ROW, token.ROWS}) {
			p.nextToken()
			p.nextToken()
		}
	}

	// FETCH CLAUSE
	if p.peekTokenIs(token.FETCH) {
		p.nextToken()
		p.nextToken()
		stmt.Fetch = p.parseFetch()
	}

	// FOR UPDATE CLAUSE
	if p.peekTokenIs(token.FOR) {
		p.nextToken()
		p.nextToken()
		stmt.Lock = p.parseLock()
	}

	// fmt.Printf("parseSelectExpressionEnd: %s %s :: %s %s == %+v\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit, stmt)

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
				p.nextToken()
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

	// if p.peekTokenIsOne([]token.TokenType{token.EOF, token.SEMICOLON}) {
	// 	return tables
	// }

	tables = append(tables, p.parseFirstTable())

	for p.peekTokenIsOne([]token.TokenType{token.JOIN, token.INNER, token.LEFT, token.RIGHT, token.FULL, token.CROSS, token.LATERAL}) {
		tables = append(tables, p.parseTable())
	}

	return tables
}

// parseFirstTable will leave curToken on the last token of the table (name or alias)
func (p *Parser) parseFirstTable() ast.Expression {
	defer untrace(trace("parseFirstTable"))

	table := ast.TableExpression{Token: token.Token{Type: token.FROM}}

	// fmt.Printf("parseFirstTable1: %s :: %s == %+v\n", p.curToken.Lit, p.peekToken.Lit, table)

	table.Table = p.parseExpression(LOWEST)

	// Get the first table
	//if p.curTokenIs(token.IDENT) {
	//table.Table = &ast.Identifier{Token: token.Token{Type: token.IDENT, Lit: p.curToken.Lit}, Value: p.curToken.Lit}

	if p.peekTokenIs(token.AS) {
		p.nextToken()
	}

	// Do we have an alias
	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		table.Alias = p.curToken.Lit
	}

	// fmt.Printf("parseFirstTable2: %s :: %s == %+v\n", p.curToken.Lit, p.peekToken.Lit, table)

	return &table
}

func (p *Parser) parseTable() ast.Expression {
	defer untrace(trace("parseTable"))

	table := ast.TableExpression{Token: token.Token{Type: token.FROM}}

	// fmt.Printf("parseTable1: %s :: %s == %+v\n", p.curToken.Lit, p.peekToken.Lit, table)

	// Get the join type
	if p.peekTokenIsOne([]token.TokenType{token.INNER, token.LEFT, token.RIGHT, token.FULL, token.CROSS, token.LATERAL}) {
		p.nextToken()
		table.JoinType = strings.ToUpper(p.curToken.Lit)

		// Skip the JOIN and OUTER keywords
		for p.peekTokenIsOne([]token.TokenType{token.JOIN, token.OUTER}) {
			p.nextToken()
		}
		// fmt.Printf("parseTable compound join: %s :: %s == %+v\n", p.curToken.Lit, p.peekToken.Lit, table)
	}

	// If just using JOIN, assume INNER
	if p.peekTokenIs(token.JOIN) {
		p.nextToken()
		if table.JoinType == "" {
			table.JoinType = "INNER"
		}
		// fmt.Printf("parseTable simple join: %s :: %s == %+v\n", p.curToken.Lit, p.peekToken.Lit, table)
	}

	p.nextToken()

	// fmt.Printf("parseTable2: %s :: %s == %+v\n", p.curToken.Lit, p.peekToken.Lit, table)

	// Get the table name
	table.Table = p.parseExpression(LOWEST)

	// if p.curTokenIs(token.IDENT) {
	// 	table.Table = &ast.Identifier{Token: token.Token{Type: token.IDENT, Lit: p.curToken.Lit}, Value: p.curToken.Lit}
	// }

	// fmt.Printf("parseTable3: %s :: %s == %+v\n", p.curToken.Lit, p.peekToken.Lit, table)

	if p.peekTokenIs(token.AS) {
		p.nextToken()
	}

	// Do we have an alias?
	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		table.Alias = p.curToken.Lit
	}

	// fmt.Printf("parseTable4: %s :: %s == %+v\n", p.curToken.Lit, p.peekToken.Lit, table)

	// Get the join condition, but skip past the ON keyword
	if p.peekTokenIs(token.ON) {
		p.nextToken()
		p.nextToken()
		table.JoinCondition = p.parseExpression(LOWEST)
	}

	return &table
}

func (p *Parser) parseFetch() ast.Expression {
	defer untrace(trace("parseFetch"))

	if p.curTokenIs(token.FETCH) {
		p.nextToken()
	}

	fetch := &ast.FetchExpression{Token: p.curToken,
		Value:  &ast.IntegerLiteral{Token: token.Token{Type: token.INT, Lit: "1"}, Value: 1, ParamOffset: p.paramOffset},
		Option: token.Token{Type: token.NIL, Lit: ""}}

	if p.curTokenIsOne([]token.TokenType{token.NEXT, token.FIRST}) {
		p.nextToken()
	}

	if p.curTokenIs(token.INT) {
		fetch.Value = p.parseExpression(LOWEST)
		p.nextToken()
	} else {
		// Integer is suppressed, so we need to increment the param offset manually instead of in the IntegerLiteral
		p.paramOffset++
		fetch.Value = &ast.IntegerLiteral{Token: token.Token{Type: token.INT, Lit: "1"}, Value: 1, ParamOffset: p.paramOffset}
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

func (p *Parser) parseWindowList(end []token.TokenType) []ast.Expression {
	defer untrace(trace("parseWindowList"))

	list := []ast.Expression{}

	if p.curTokenIsOne(end) {
		return list
	}

	list = append(list, p.parseWindow(end))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseWindow(end))
	}

	return list
}

// w as (partition by c1 order by c2)
func (p *Parser) parseWindow(end []token.TokenType) ast.Expression {
	defer untrace(trace("parseWindow"))

	expression := &ast.WindowExpression{
		Token: p.curToken,
	}

	// fmt.Printf("parseWindow1: '%s' :: '%s' == %+v\n", p.curToken.Lit, p.peekToken.Lit, expression)

	if p.curTokenIs(token.IDENT) {
		expression.Alias = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}
		p.nextToken()

		if p.curTokenIs(token.AS) {
			p.nextToken()
		}

		if p.curTokenIs(token.LPAREN) {
			p.nextToken()
		}
	}

	// fmt.Printf("parseWindow2: %s :: %s == %+v\n", p.curToken.Lit, p.peekToken.Lit, expression)

	window := p.parseWindowExpression()
	expression.PartitionBy = window.(*ast.WindowExpression).PartitionBy
	expression.OrderBy = window.(*ast.WindowExpression).OrderBy

	// fmt.Printf("parseWindow3: %s :: %s == %+v\n", p.curToken.Lit, p.peekToken.Lit, expression)

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
	}

	return expression
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

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	// fmt.Printf("parseColumn: %s :: %s == %+v\n", p.curToken.Lit, p.peekToken.Lit, leftExp)

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

	if p.curTokenIs(token.IDENT) && strings.ToUpper(p.curToken.Lit) == "KEY" {
		p.nextToken()
		if p.curTokenIs(token.SHARE) {
			expression.Lock = "KEY SHARE"
		}
	}

	if p.curTokenIs(token.NO) {
		p.nextToken()
		if p.curTokenIs(token.IDENT) && strings.ToUpper(p.curToken.Lit) == "KEY" {
			p.nextToken()
			if p.curTokenIs(token.UPDATE) {
				p.nextToken()
				expression.Lock = "NO KEY UPDATE"
			}
		}
	}

	if p.curTokenIs(token.OF) {
		p.nextToken()
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

func (p *Parser) parseNotExpression(left ast.Expression) ast.Expression {
	exp := &ast.InExpression{Token: p.curToken, Operator: p.curToken.Lit, Left: left}
	p.nextToken()

	// If we're missing another NOT expression, add it here.
	if p.curTokenIsOne([]token.TokenType{token.IN, token.LIKE, token.ILIKE, token.BETWEEN}) {
		exp.Operator = "NOT " + p.curToken.Lit
		p.nextToken()
	}

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		exp.Right = p.parseExpressionList([]token.TokenType{token.RPAREN})
	}
	// fmt.Printf("parseInExpression1: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, exp)

	return exp
}

func (p *Parser) parseInExpression(left ast.Expression) ast.Expression {
	exp := &ast.InExpression{Token: p.curToken, Operator: p.curToken.Lit, Left: left}
	p.nextToken()
	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		exp.Right = p.parseExpressionList([]token.TokenType{token.RPAREN})
	}
	// fmt.Printf("parseInExpression1: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, exp)

	return exp
}

func (p *Parser) parseUnionExpression() ast.Expression {
	exp := &ast.UnionExpression{Token: p.curToken}
	p.nextToken()
	exp.Right = p.parseSelectExpression()
	// fmt.Printf("parseUnionExpression1: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, exp)

	return exp
}
