package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/brianbroderick/lantern/internal/sql/ast"
	"github.com/brianbroderick/lantern/internal/sql/lexer"
	"github.com/brianbroderick/lantern/internal/sql/token"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

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
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.ASTERISK, p.parseWildcardLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

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
	switch p.curToken.Type {
	case token.SELECT:
		return p.parseSelectStatement()
	default:
		return p.parseExpressionStatement()
	}
}

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

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
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
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}
}

func (p *Parser) parseWildcardLiteral() ast.Expression {
	return &ast.WildcardLiteral{Token: p.curToken, Value: p.curToken.Lit}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Lit, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Lit)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Lit}
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

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseInfixExpression"))
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Lit,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
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
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	array.Elements = p.parseExpressionList(token.RBRACKET)

	return array
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
