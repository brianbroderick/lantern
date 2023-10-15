package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

const (
	_ int = iota
	LOWEST
	OR             // OR
	AND            // AND
	NOT            // NOT
	IS             // IS, ISNULL, NOTNULL
	EQUALS         // ==
	LESSGREATER    // > or <
	FILTER         // BETWEEN, IN, LIKE, ILIKE, SIMILAR
	WINDOW         // OVER
	SUM            // +
	PRODUCT        // *
	EXPONENTIATION // ^
	PREFIX         // -X or !X
	JSON           // ->, ->>, #>, #>>, @>, <@, ?, ?&, ?|
	CALL           // myFunction(X)
	INDEX          // array[index]
)

var precedences = map[token.TokenType]int{
	token.EQ:                EQUALS,
	token.NOT_EQ:            EQUALS,
	token.ASSIGN:            EQUALS,
	token.TO:                EQUALS,
	token.LT:                LESSGREATER,
	token.GT:                LESSGREATER,
	token.PLUS:              SUM,
	token.MINUS:             SUM,
	token.SLASH:             PRODUCT,
	token.ASTERISK:          PRODUCT,
	token.LPAREN:            CALL,
	token.LBRACKET:          INDEX,
	token.AND:               AND,
	token.OR:                OR,
	token.NOT:               NOT,
	token.IS:                IS,
	token.ISNULL:            IS,
	token.NOTNULL:           IS,
	token.OVER:              WINDOW,
	token.BETWEEN:           FILTER,
	token.IN:                FILTER,
	token.LIKE:              FILTER,
	token.ILIKE:             FILTER,
	token.SIMILAR:           FILTER,
	token.EXPONENTIATION:    EXPONENTIATION,
	token.JSONGETBYKEY:      JSON,
	token.JSONGETBYTEXT:     JSON,
	token.JSONGETBYPATH:     JSON,
	token.JSONGETBYPATHTEXT: JSON,
	token.JSONCONTAINS:      JSON,
	token.JSONCONTAINED:     JSON,
	token.JSONHASKEY:        JSON,
	token.JSONHASALLKEYS:    JSON,
	token.JSONHASANYKEYS:    JSON,
	token.JSONDELETE:        JSON,
	token.JSONCONCAT:        JSON,
	token.OVERLAP:           FILTER,
}

// https://www.postgresql.org/docs/current/sql-syntax-lexical.html#SQL-PRECEDENCE
// operators/elements, from highest to lowest precedence:
// . table/column name separator
// :: typecast
// [ ] array element selection
// + - unary plus, unary minus
// ^ exponentiation
// * / % multiplication, division, modulo
// + - addition, subtraction
// BETWEEN IN LIKE ILIKE SIMILAR range containment, set membership, string matching
// < > = <= >= <> comparison operators
// NOT logical negation
// AND logical conjunction
// OR logical disjunction

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l           *lexer.Lexer
	errors      []string
	paramOffset int

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)

	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.EXPONENTIATION, p.parsePrefixExpression)
	p.registerPrefix(token.NOT, p.parsePrefixKeywordExpression)

	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.PARTITION, p.parseWindowExpression)
	p.registerPrefix(token.ORDER, p.parseWindowExpression)
	p.registerPrefix(token.SELECT, p.parseSelectExpression)
	p.registerPrefix(token.DISTINCT, p.parseDistinct)
	p.registerPrefix(token.ALL, p.parseDistinct)

	// Some tokens don't need special parse rules and can function as an identifier
	// If this becomes a problem, we can create a generic struct for these cases
	p.registerPrefix(token.LOCAL, p.parseIdentifier)
	p.registerPrefix(token.DEFAULT, p.parseIdentifier)

	// This might be doing the same thing as parseIdentifier. TODO: check this out
	p.registerPrefix(token.ASTERISK, p.parseWildcardLiteral)
	p.registerPrefix(token.ALL, p.parseKeywordExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.IS, p.parseInfixExpression)
	p.registerInfix(token.ISNULL, p.parseInfixExpression)  // this might actually be a postfix operator
	p.registerInfix(token.NOTNULL, p.parseInfixExpression) // this might actually be a postfix operator
	p.registerInfix(token.LIKE, p.parseInfixExpression)
	p.registerInfix(token.ILIKE, p.parseInfixExpression)
	p.registerInfix(token.SIMILAR, p.parseInfixExpression)
	p.registerInfix(token.BETWEEN, p.parseInfixExpression)
	p.registerInfix(token.OVER, p.parseInfixExpression)
	p.registerInfix(token.JSONGETBYKEY, p.parseInfixExpression)
	p.registerInfix(token.JSONGETBYTEXT, p.parseInfixExpression)
	p.registerInfix(token.JSONGETBYPATH, p.parseInfixExpression)
	p.registerInfix(token.JSONGETBYPATHTEXT, p.parseInfixExpression)
	p.registerInfix(token.JSONCONTAINS, p.parseInfixExpression)
	p.registerInfix(token.JSONCONTAINED, p.parseInfixExpression)
	p.registerInfix(token.JSONHASKEY, p.parseInfixExpression)
	p.registerInfix(token.JSONHASALLKEYS, p.parseInfixExpression)
	p.registerInfix(token.JSONHASANYKEYS, p.parseInfixExpression)
	p.registerInfix(token.JSONDELETE, p.parseInfixExpression)
	p.registerInfix(token.JSONCONCAT, p.parseInfixExpression)
	p.registerInfix(token.OVERLAP, p.parseInfixExpression)
	p.registerInfix(token.TO, p.parseInfixExpression)

	p.registerInfix(token.NOT, p.parseNotExpression)

	p.registerInfix(token.IN, p.parseInExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseArrayExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken, _ = p.l.Scan() // TODO: surface the position
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) curTokenIsOne(tokens []token.TokenType) bool {
	found := false
	for _, t := range tokens {
		if p.curTokenIs(t) {
			found = true
		}
	}
	return found
}

func (p *Parser) peekTokenIsOne(tokens []token.TokenType) bool {
	found := false
	for _, t := range tokens {
		if p.peekTokenIs(t) {
			found = true
		}
	}
	return found
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) expectPeekIsOne(tokens []token.TokenType) bool {
	found := false
	for _, t := range tokens {
		if p.peekTokenIs(t) {
			found = true
		}
	}
	if found {
		p.nextToken()
		return true
	} else {
		p.peekErrorIsOne(tokens)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) PrintErrors() {
	if len(p.errors) == 0 {
		return
	}
	for _, msg := range p.errors {
		fmt.Printf("parser error: %s\n", msg)
	}
}

func (p *Parser) peekErrorIsOne(tokens []token.TokenType) {
	toks := []string{}
	for _, t := range tokens {
		toks = append(toks, t.String())
	}

	msg := fmt.Sprintf("expected next token to be one of %s, got %s: %s instead. current token is: %s: %s",
		strings.Join(toks, ", "), p.peekToken.Type, p.peekToken.Lit, p.curToken.Type, p.curToken.Lit)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s: %s instead. current token is: %s: %s",
		t, p.peekToken.Type, p.peekToken.Lit, p.curToken.Type, p.curToken.Lit)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ParseProgram() *ast.Program {
	defer untrace(trace("ParseProgram"))
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		// if stmt != nil {
		program.Statements = append(program.Statements, stmt)
		// }
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	defer untrace(trace("parseStatement"))
	// fmt.Printf("parseStatement: %s :: %s\n", p.curToken.Lit, p.peekToken.Lit)
	switch p.curToken.Type {
	case token.SELECT:
		return p.parseSelectStatement()
	case token.WITH:
		return p.parseCTEStatement()
	case token.SET:
		return p.parseSetStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	defer untrace(trace("parseExpressionStatement"))
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	defer untrace(trace("parseExpression"))

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	// fmt.Printf("parseExpression: %s :: %s == %+v\n", p.curToken.Lit, p.peekToken.Lit, leftExp)

	for !p.peekTokenIsOne([]token.TokenType{token.COMMA, token.WHERE, token.GROUP, token.HAVING, token.ORDER, token.LIMIT, token.OFFSET, token.FETCH, token.FOR, token.SEMICOLON}) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) parseIdentifier() ast.Expression {
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}

	// fmt.Printf("parseIdentifier: %s :: %s\n", p.curToken.Lit, p.peekToken.Lit)

	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		ident.Cast = p.curToken.Lit
	}
	return ident
}

func (p *Parser) parseWildcardLiteral() ast.Expression {
	return &ast.WildcardLiteral{Token: p.curToken, Value: p.curToken.Lit}
}

func (p *Parser) parseKeywordExpression() ast.Expression {
	return &ast.KeywordExpression{Token: p.curToken}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	// Incrementing the offset is to help when masking parameters in the AST
	p.paramOffset++
	lit := &ast.IntegerLiteral{Token: p.curToken, ParamOffset: p.paramOffset}

	value, err := strconv.ParseInt(p.curToken.Lit, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Lit)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		lit.Cast = p.curToken.Lit
	}

	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	p.paramOffset++
	str := &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Lit, ParamOffset: p.paramOffset}
	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		str.Cast = p.curToken.Lit
	}
	return str
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Lit,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parsePrefixKeywordExpression() ast.Expression {
	expression := &ast.PrefixKeywordExpression{
		Token:    p.curToken,
		Operator: p.curToken.Lit,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseInfixExpression"))
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Lit,
		Left:     left,
	}

	// fmt.Printf("parseInfixExpressionPrecedence: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, expression)
	precedence := p.curPrecedence()
	p.nextToken()

	// fmt.Printf("parseInfixExpression: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, expression)
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	// fmt.Printf("parseGroupedExpression1: %s :: %s\n", p.curToken.Lit, p.peekToken.Lit)
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
	}

	return exp
}

// func (p *Parser) parseFunctionParameters() []*ast.Identifier {
// 	identifiers := []*ast.Identifier{}

// 	if p.peekTokenIs(token.RPAREN) {
// 		p.nextToken()
// 		return identifiers
// 	}

// 	p.nextToken()

// 	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}
// 	identifiers = append(identifiers, ident)

// 	for p.peekTokenIs(token.COMMA) {
// 		p.nextToken()
// 		p.nextToken()
// 		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}
// 		identifiers = append(identifiers, ident)
// 	}

// 	if !p.expectPeek(token.RPAREN) {
// 		return nil
// 	}

// 	return identifiers
// }

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList([]token.TokenType{token.RPAREN})
	return exp
}

func (p *Parser) parseExpressionList(end []token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	// fmt.Printf()

	if p.peekTokenIsOne(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	// fmt.Printf("parseExpressionList: %s :: %s :: %+v\n", p.curToken.Lit, p.peekToken.Lit, list)
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.peekTokenIsOne(end) {
		return nil
	}

	p.nextToken()

	return list
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	array.Elements = p.parseExpressionList([]token.TokenType{token.RBRACKET})

	return array
}

func (p *Parser) parseArrayExpression(left ast.Expression) ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken, Left: left}

	array.Elements = p.parseExpressionList([]token.TokenType{token.RBRACKET})

	return array
}

// This would parse an index lookup such as array[0], but PG uses this form to define an array
// that looks like array[1,2,3]. For that reason, parseArrayExpression is used instead.
// func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
// 	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

// 	p.nextToken()
// 	exp.Index = p.parseExpression(LOWEST)

// 	if !p.expectPeek(token.RBRACKET) {
// 		return nil
// 	}

// 	return exp
// }

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
