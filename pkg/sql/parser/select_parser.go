package parser

import (
	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

// Clauses are parsed in the order that they appear in the SQL statement.
// Each Expression starts with the token after the keyword that starts the clause.
// To determine if a clause is done, peek to the next token and look for the next clause keyword.
// If it's found, advance the token twice to skip past the clause token (from peek, to current, to next).
// Otherwise, you'll get off by one errors and have a hard time figuring out why.

var defaultListSeparators = []token.TokenType{token.COMMA, token.WHERE, token.GROUP_BY, token.HAVING, token.ORDER, token.LIMIT, token.OFFSET, token.FETCH, token.FOR, token.SEMICOLON}

func (p *Parser) parseSelectStatement() *ast.SelectStatement {
	defer p.untrace(p.trace("parseSelectStatement"))

	s := &ast.SelectStatement{Token: p.curToken}
	s.Expressions = []ast.Expression{}
	s.Expressions = append(s.Expressions, p.parseExpression(STATEMENT))

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return s
}

func (p *Parser) parseSelectExpression() ast.Expression {
	defer p.untrace(p.trace("parseSelectExpression"))

	x := &ast.SelectExpression{Token: p.curToken}
	p.nextToken()

	// DISTINCT CLAUSE
	if p.curTokenIsOne([]token.TokenType{token.ALL, token.DISTINCT}) {
		x.Distinct = p.parseDistinct()
	}

	x.Columns = p.parseColumnList([]token.TokenType{token.COMMA, token.FROM, token.AS})

	// Sometimes the FROM clause is already to the curToken if there are not any columns
	if p.peekTokenIs(token.FROM) {
		p.nextToken()
	}

	// FROM CLAUSE: if the next token is FROM, advance the token and move on. Otherwise, return the statement.
	if p.curTokenIs(token.FROM) {
		p.nextToken()

		// fmt.Printf("parseSelectExpression001: %s %s :: %s %s == %+v\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit, x)

		// p.nextToken()
		x.Tables, x.TableAliases = p.parseTables()
		// x.TableAliases = x.ResolveTableAliases()

		// fmt.Printf("parseSelectExpression002: %s %s :: %s %s == %+v\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit, x)

		// WINDOW CLAUSE
		if p.peekTokenIs(token.WINDOW) {
			p.nextToken()
			p.nextToken()
			x.Window = p.parseWindowList(defaultListSeparators)
		}

		// WHERE CLAUSE
		if p.peekTokenIs(token.WHERE) {
			p.nextToken()
			p.nextToken()
			x.Where = p.parseExpression(LOWEST)
		}

		// GROUP BY CLAUSE
		if p.peekTokenIs(token.GROUP_BY) {
			p.nextToken()
			p.nextToken()
			x.GroupBy = p.parseColumnList(defaultListSeparators)
		}

		// HAVING CLAUSE
		if p.peekTokenIs(token.HAVING) {
			p.nextToken()
			p.nextToken()
			x.Having = p.parseExpression(LOWEST)
		}

		// ORDER BY CLAUSE
		if p.peekTokenIs(token.ORDER) {
			p.nextToken()
			if !p.expectPeek(token.BY) {
				return nil
			}
			p.nextToken()
			x.OrderBy = p.parseSortList(defaultListSeparators)
		}

		// LIMIT and OFFSET CLAUSES can be in any order, but only one of each
		if p.peekTokenIsOne([]token.TokenType{token.LIMIT, token.OFFSET}) {
			for p.peekTokenIsOne([]token.TokenType{token.LIMIT, token.OFFSET}) {
				// LIMIT CLAUSE
				if p.peekTokenIs(token.LIMIT) {
					p.nextToken()
					p.nextToken()
					x.Limit = p.parseExpression(LOWEST)
				}

				// OFFSET CLAUSE
				if p.peekTokenIs(token.OFFSET) {
					p.nextToken()
					p.nextToken()
					x.Offset = p.parseExpression(LOWEST)

					if p.peekTokenIsOne([]token.TokenType{token.ROW, token.ROWS}) {
						p.nextToken()
						p.nextToken()
					}
				}
			}
		}

		// FETCH CLAUSE
		if p.peekTokenIs(token.FETCH) {
			p.nextToken()
			p.nextToken()
			x.Fetch = p.parseFetch()
		}

		// FOR UPDATE CLAUSE
		if p.peekTokenIs(token.FOR) {
			p.nextToken()
			p.nextToken()
			x.Lock = p.parseLock()
		}
	}

	return x
}

func (p *Parser) parseDistinct() ast.Expression {
	defer p.untrace(p.trace("parseDistinct"))

	context := p.context
	p.setContext(XDISTINCT)       // sets the context for the parseExpressionListItem function
	defer p.resetContext(context) // reset to prior context

	if p.curTokenIs(token.ALL) {
		x := &ast.DistinctExpression{Token: p.curToken}
		p.nextToken()
		return x // &ast.DistinctExpression{Token: p.curToken, On: token.Token{Type: token.NIL, Lit: ""}}
	}

	if p.curTokenIs(token.DISTINCT) {
		x := &ast.DistinctExpression{Token: p.curToken}
		p.nextToken()

		if p.curTokenIs(token.ON) {
			p.nextToken()

			if p.curTokenIs(token.LPAREN) {
				p.nextToken()
				x.Right = p.parseExpressionList([]token.TokenType{token.RPAREN})
				if p.curTokenIs(token.RPAREN) {
					p.nextToken()
				}
			}
		}

		return x
	}

	return nil
}

func (p *Parser) parseTables() ([]ast.Expression, map[string]string) {
	defer p.untrace(p.trace("parseTables"))
	aliasMap := map[string]string{}

	x := []ast.Expression{}
	firstTable, table, alias := p.parseFirstTable()
	x = append(x, firstTable)
	if alias != "" && table != "" {
		aliasMap[alias] = table
	}

	for p.peekTokenIsOne([]token.TokenType{token.JOIN, token.INNER, token.LEFT, token.RIGHT, token.FULL, token.CROSS, token.LATERAL, token.COMMA}) {
		nextTable, table, alias := p.parseTable()
		x = append(x, nextTable)
		if alias != "" && table != "" {
			aliasMap[alias] = table
		}
	}

	return x, aliasMap
}

// parseFirstTable will leave curToken on the last token of the table (name or alias)
func (p *Parser) parseFirstTable() (ast.Expression, string, string) {
	defer p.untrace(p.trace("parseFirstTable"))
	var table string
	var alias string

	x := ast.TableExpression{Token: token.Token{Type: token.FROM}}

	x.Table = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.WITH) {
		p.nextToken()
		if p.peekTokenIs(token.ORDINALITY) {
			p.nextToken()
			p.nextToken()
			x.Ordinality = true
		}
	}

	if p.peekTokenIs(token.AS) {
		p.nextToken()
	}

	// Do we have an alias
	// if p.peekTokenIsOne([]token.TokenType{token.IDENT, token.AT}) {
	if p.peekTokenIsOne([]token.TokenType{token.IDENT, token.SET}) {
		p.nextToken()
		x.Alias = p.parseExpression(LOWEST)

		switch tbl := x.Table.(type) {
		case *ast.Identifier:
			table = tbl.String(false)
		}

		switch a := x.Alias.(type) {
		case *ast.Identifier:
			alias = a.String(false)
		}
	}

	// fmt.Printf("parseFirstTable2: %s :: %s == %+v\n", p.curToken.Lit, p.peekToken.Lit, table)

	return &x, table, alias
}

func (p *Parser) parseTable() (ast.Expression, string, string) {
	defer p.untrace(p.trace("parseTable"))
	var table string
	var alias string

	x := ast.TableExpression{Token: token.Token{Type: token.FROM}}

	// Get the join type
	if p.peekTokenIsOne([]token.TokenType{token.INNER, token.LEFT, token.RIGHT, token.FULL, token.CROSS, token.LATERAL}) {
		p.nextToken()
		x.JoinType = p.curToken.Upper

		if p.peekTokenIs(token.JOIN) {
			p.nextToken()
			x.JoinType = x.JoinType + " JOIN"
		}
		// Skip the OUTER keyword
		if p.peekTokenIs(token.OUTER) {
			p.nextToken()
		}
	}

	// If just using JOIN, assume INNER
	if p.peekTokenIs(token.JOIN) {
		p.nextToken()
		if x.JoinType == "" {
			x.JoinType = "INNER JOIN"
		} else {
			x.JoinType = x.JoinType + " JOIN"
		}
	}

	if p.peekTokenIs(token.LATERAL) {
		p.nextToken()
		x.JoinType = x.JoinType + " LATERAL"
	}

	if p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if x.JoinType == "" {
			x.JoinType = ","
		}
	}

	p.nextToken()

	// Get the table name
	x.Table = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.WITH) {
		p.nextToken()
		if p.peekTokenIs(token.ORDINALITY) {
			p.nextToken()
			p.nextToken()
			x.Ordinality = true
		}
	}

	if p.peekTokenIs(token.AS) {
		p.nextToken()
	}

	// Do we have an alias?
	// if p.peekTokenIsOne([]token.TokenType{token.IDENT, token.AT}) {
	if p.peekTokenIsOne([]token.TokenType{token.IDENT, token.SET}) {
		p.nextToken()
		x.Alias = p.parseExpression(LOWEST)

		switch tbl := x.Table.(type) {
		case *ast.Identifier:
			table = tbl.String(false)
		}

		switch a := x.Alias.(type) {
		case *ast.Identifier:
			alias = a.String(false)
		}
	}

	// Get the join condition, but skip past the ON keyword
	if p.peekTokenIs(token.ON) {
		p.nextToken()
		p.nextToken()
		x.JoinCondition = p.parseExpression(LOWEST)
	}

	return &x, table, alias
}

func (p *Parser) parseFetch() ast.Expression {
	defer p.untrace(p.trace("parseFetch"))

	if p.curTokenIs(token.FETCH) {
		p.nextToken()
	}

	x := &ast.FetchExpression{Token: p.curToken,
		Value:  &ast.IntegerLiteral{Token: token.Token{Type: token.INT, Lit: "1"}, Value: 1, ParamOffset: p.paramOffset},
		Option: token.Token{Type: token.NIL, Lit: ""}}

	if p.curTokenIsOne([]token.TokenType{token.NEXT, token.FIRST}) {
		p.nextToken()
	}

	if p.curTokenIs(token.INT) {
		x.Value = p.parseExpression(LOWEST)
		p.nextToken()
	} else {
		// Integer is suppressed, so we need to increment the param offset manually instead of in the IntegerLiteral
		p.paramOffset++
		x.Value = &ast.IntegerLiteral{Token: token.Token{Type: token.INT, Lit: "1"}, Value: 1, ParamOffset: p.paramOffset}
	}

	if p.curTokenIsOne([]token.TokenType{token.ROW, token.ROWS}) {
		p.nextToken()
	}

	if p.curTokenIs(token.ONLY) {
		x.Option = p.curToken
		p.nextToken()
	} else if p.curTokenIs(token.WITH) {
		if p.peekTokenIs(token.TIES) {
			p.nextToken()
			x.Option = p.curToken
			p.nextToken()
		}
	}
	return x
}

func (p *Parser) parseSortList(end []token.TokenType) []ast.Expression {
	defer p.untrace(p.trace("parseSortList"))

	x := []ast.Expression{}

	if p.curTokenIsOne(end) {
		return x
	}

	x = append(x, p.parseSort(LOWEST, end))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		x = append(x, p.parseSort(LOWEST, end))
	}

	return x
}

func (p *Parser) parseSort(precedence int, end []token.TokenType) ast.Expression {
	defer p.untrace(p.trace("parseSort"))

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIsOne([]token.TokenType{token.COMMA}) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	x := &ast.SortExpression{Token: p.curToken, Value: leftExp, Direction: token.Token{Type: token.ASC, Lit: "ASC"}, Nulls: token.Token{Type: token.NIL, Lit: ""}}

	if p.peekTokenIsOne([]token.TokenType{token.ASC, token.DESC}) {
		p.nextToken()
		x.Direction = p.curToken
	}

	if p.peekTokenIs(token.NULLS) {
		p.nextToken()
		if p.peekTokenIsOne([]token.TokenType{token.FIRST, token.LAST}) {
			p.nextToken()
			x.Nulls = p.curToken
		}
	}

	return x
}

func (p *Parser) parseWindowList(end []token.TokenType) []ast.Expression {
	defer p.untrace(p.trace("parseWindowList"))

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
	defer p.untrace(p.trace("parseWindow"))

	x := &ast.WindowExpression{
		Token: p.curToken,
	}

	if p.curTokenIs(token.IDENT) {
		x.Alias = &ast.SimpleIdentifier{Token: p.curToken, Value: p.curToken.Lit}
		p.nextToken()

		if p.curTokenIs(token.AS) {
			p.nextToken()
		}

		if p.curTokenIs(token.LPAREN) {
			p.nextToken()
		}
	}

	window := p.parseWindowExpression()
	x.PartitionBy = window.(*ast.WindowExpression).PartitionBy
	x.OrderBy = window.(*ast.WindowExpression).OrderBy

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
	}

	return x
}

func (p *Parser) parseColumnList(end []token.TokenType) []ast.Expression {
	defer p.untrace(p.trace("parseColumnList"))

	x := []ast.Expression{}

	if p.curTokenIsOne(end) {
		return x
	}

	x = append(x, p.parseColumn(LOWEST, end))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		x = append(x, p.parseColumn(LOWEST, end))
	}

	return x
}

func (p *Parser) parseColumn(precedence int, end []token.TokenType) ast.Expression {
	defer p.untrace(p.trace("parseColumn"))

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIsOne([]token.TokenType{token.COMMA, token.FROM, token.AS, token.IDENT}) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]

		if infix == nil {
			return leftExp
		}
		// fmt.Printf("parseColumn1: %s %s :: %s %s == %+v\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit, leftExp.String(false))

		p.nextToken()
		leftExp = infix(leftExp)
	}

	x := &ast.ColumnExpression{Token: p.curToken, Value: leftExp}
	// fmt.Printf("parseColumn2: %s %s :: %s %s == %+v\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit, x.String(false))

	// AS is optional, but opens up additional keywords that can be used as an alias.
	if p.peekTokenIs(token.AS) {
		p.nextToken()
	}

	// fmt.Printf("parseColumn3: %s %s :: %s %s == %+v\n", p.curToken.Type, p.curToken.Lit, p.peekToken.Type, p.peekToken.Lit, x.String(false))

	if p.peekTokenIsOne([]token.TokenType{token.IDENT, token.VALUES}) {
		p.nextToken()

		alias := &ast.SimpleIdentifier{Token: p.curToken, Value: p.curToken.Lit}
		x = &ast.ColumnExpression{Token: p.curToken, Value: leftExp, Name: alias}
	}

	return x
}

func (p *Parser) parseWindowExpression() ast.Expression {
	defer p.untrace(p.trace("parseWindowExpression"))

	x := &ast.WindowExpression{
		Token: p.curToken,
	}

	if p.curTokenIs(token.PARTITION) {
		if p.expectPeek(token.BY) {
			p.nextToken()
			x.PartitionBy = p.parseColumnList([]token.TokenType{token.COMMA, token.ORDER, token.AS, token.RPAREN})
		}
	}
	if p.peekTokenIs(token.ORDER) {
		p.nextToken()
	}

	if p.curTokenIs(token.ORDER) {
		if p.expectPeek(token.BY) {
			p.nextToken()
			x.OrderBy = p.parseSortList([]token.TokenType{token.COMMA, token.AS, token.RPAREN})
		}
	}

	return x
}

func (p *Parser) parseLock() ast.Expression {
	defer p.untrace(p.trace("parseLock"))

	context := p.context
	p.setContext(XLOCK)           // sets the context for the parseExpressionListItem function
	defer p.resetContext(context) // reset to prior context

	x := &ast.LockExpression{
		Token: p.curToken,
	}

	switch p.curToken.Type {
	case token.UPDATE:
		x.Lock = "UPDATE"
		p.nextToken()
	case token.SHARE:
		x.Lock = "SHARE"
		p.nextToken()
	}

	if p.curTokenIs(token.IDENT) && p.curToken.Upper == "KEY" {
		p.nextToken()
		if p.curTokenIs(token.SHARE) {
			x.Lock = "KEY SHARE"
		}
	}

	if p.curTokenIs(token.NO) {
		p.nextToken()
		if p.curTokenIs(token.IDENT) && p.curToken.Upper == "KEY" {
			p.nextToken()
			if p.curTokenIs(token.UPDATE) {
				p.nextToken()
				x.Lock = "NO KEY UPDATE"
			}
		}
	}

	if p.curTokenIs(token.OF) {
		p.nextToken()
		x.Tables = p.parseExpressionList([]token.TokenType{token.NOWAIT, token.SKIP, token.SEMICOLON, token.EOF})
	}

	if p.curTokenIs(token.NOWAIT) {
		x.Options = "NOWAIT"
	} else if p.curTokenIs(token.SKIP) {
		p.nextToken()
		if p.curTokenIs(token.LOCKED) {
			x.Options = "SKIP LOCKED"
		}
	}

	p.nextToken()

	return x
}

// TODO handle other NOT expressions (NOT IN, NOT LIKE, NOT ILIKE, NOT BETWEEN)
func (p *Parser) parseNotExpression(x ast.Expression) ast.Expression {
	defer p.untrace(p.trace("parseNotExpression"))

	p.not = true // sets the next expression to be a NOT expression
	return p.determineInfix(LOWEST, x)
}

func (p *Parser) parseInExpression(left ast.Expression) ast.Expression {
	defer p.untrace(p.trace("parseInExpression"))

	context := p.context
	p.setContext(XIN)             // sets the context for the parseExpressionListItem function
	defer p.resetContext(context) // reset to prior context

	x := &ast.InExpression{Token: p.curToken, Operator: p.curToken.Lit, Left: left}
	if p.not {
		x.Not = true
		p.not = false
	}
	p.nextToken()
	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		x.Right = p.parseExpressionList([]token.TokenType{token.RPAREN})
	} else {
		x.Right = append(x.Right, p.parseExpression(LOWEST))
	}

	return x
}

func (p *Parser) parseAggregateExpression(left ast.Expression) ast.Expression {
	defer p.untrace(p.trace("parseAggregateExpression"))

	exp := &ast.AggregateExpression{Token: p.curToken, Operator: p.curToken.Lit, Left: left}

	if p.curTokenIs(token.ORDER) {
		exp.Operator = "ORDER BY"
		p.nextToken()
		if p.curTokenIs(token.BY) {
			p.nextToken()
		}
		exp.Right = p.parseSortList([]token.TokenType{token.RPAREN})
	}

	return exp
}

// select substring('Hello World!' from 2 for 4) from users;
func (p *Parser) parseStringFunctionExpression(left ast.Expression) ast.Expression {
	defer p.untrace(p.trace("parseStringFunctionExpression"))

	x := &ast.StringFunctionExpression{Token: p.curToken, Left: left}

	precedence := p.curPrecedence()
	p.nextToken()
	x.From = p.parseExpression(precedence)

	if p.peekTokenIs(token.FOR) {
		p.nextToken()
		p.nextToken()
		x.For = p.parseExpression(LOWEST)
	}

	return x
}

func (p *Parser) parseDoubleColonExpression() ast.Expression {
	defer p.untrace(p.trace("parseDoubleColonExpression"))
	x := &ast.CastExpression{Token: p.curToken}

	switch p.curToken.Type {
	case token.TIMESTAMP:
		x.Cast = p.parseTimestampExpression()
	default: // Catch IDENT and INTERVALS
		x.Cast = p.parseIdentifier()
	}

	return x
}

func (p *Parser) parseCastExpression() ast.Expression {
	defer p.untrace(p.trace("parseCastExpression"))

	x := &ast.CastExpression{Token: p.curToken}

	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		p.nextToken()
		x.Left = p.parseExpression(LOWEST)

		if p.peekTokenIs(token.AS) {
			p.nextToken()
			p.nextToken()
			x.SetCast(p.parseDoubleColonExpression())
		}

		if p.peekTokenIs(token.RPAREN) {
			p.nextToken()
		}
	}

	return x
}

func (p *Parser) parseWhereExpression() ast.Expression {
	defer p.untrace(p.trace("parseWhereExpression"))

	x := &ast.WhereExpression{Token: p.curToken}
	p.nextToken()
	x.Right = p.parseExpression(LOWEST)

	return x
}

// trim(both 'x' from 'xTomxx') -> Tom
// starts at both, leading, or trailing token
func (p *Parser) parseTrimExpression() ast.Expression {
	defer p.untrace(p.trace("parseTrimExpression"))

	x := &ast.TrimExpression{Token: p.curToken}
	p.nextToken()

	x.Expression = p.parseExpression(LOWEST)

	return x
}

func (p *Parser) parseIsExpression(left ast.Expression) ast.Expression {
	defer p.untrace(p.trace("parseIsExpression"))

	p.paramOffset++
	x := &ast.IsExpression{Token: p.curToken, Left: left, ParamOffset: p.paramOffset}
	precedence := p.curPrecedence()
	p.nextToken()

	if p.curTokenIs(token.NOT) {
		p.nextToken()
		x.Not = true
	}
	if p.curTokenIs(token.DISTINCT) {
		p.nextToken()
		if p.curTokenIs(token.FROM) {
			p.nextToken()
			x.Distinct = true
		}
	}

	x.Right = p.parseExpression(precedence)
	return x
}
