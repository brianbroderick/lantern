package parser

import (
	"fmt"

	"github.com/brianbroderick/lantern/internal/postgresql/ast"
	"github.com/brianbroderick/lantern/internal/postgresql/lexer"
	"github.com/brianbroderick/lantern/internal/postgresql/token"
)

type (
// prefixParseFn func() ast.Expression
// infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	// // aka handlers
	// prefixParseFns map[token.TokenType]prefixParseFn
	// // infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	// p.registerPrefix(token.DATE, p.parseLog)

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

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// func (p *Parser) noPrefixParseFnError(t token.TokenType) {
// 	msg := fmt.Sprintf("no prefix parse function for %s found", t)
// 	p.errors = append(p.errors, msg)
// }

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		program.Statements = append(program.Statements, stmt)

		p.nextToken()
	}

	return program
}

// This function allows for the possibility of having different statement types
// in the future.
func (p *Parser) parseStatement() ast.Statement {
	return p.parseLogStatement()
}

func (p *Parser) parseLogStatement() *ast.LogStatement {
	stmt := &ast.LogStatement{Token: p.curToken}

	if p.curTokenIs(token.DATE) {
		stmt.Date = p.curToken.Lit
	}

	fmt.Println(stmt.Date)
	// if !p.expectPeek(token.IDENT) {
	// 	return nil
	// }

	// stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}

	// if !p.expectPeek(token.ASSIGN) {
	// 	return nil
	// }

	// p.nextToken()

	// stmt.Value = p.parseExpression(LOWEST)

	// if p.peekTokenIs(token.SEMICOLON) {
	// 	p.nextToken()
	// }

	return stmt
}

// func (p *Parser) parseIdentifier() ast.Expression {
// 	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}
// }

// func (p *Parser) parseIntegerLiteral() ast.Expression {
// 	lit := &ast.IntegerLiteral{Token: p.curToken}

// 	value, err := strconv.ParseInt(p.curToken.Lit, 0, 64)
// 	if err != nil {
// 		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Lit)
// 		p.errors = append(p.errors, msg)
// 		return nil
// 	}

// 	lit.Value = value

// 	return lit
// }

// func (p *Parser) parseStringLiteral() ast.Expression {
// 	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Lit}
// }

// func (p *Parser) parsePrefixExpression() ast.Expression {
// 	expression := &ast.PrefixExpression{
// 		Token:    p.curToken,
// 		Operator: p.curToken.Lit,
// 	}

// 	p.nextToken()

// 	expression.Right = p.parseExpression(PREFIX)

// 	return expression
// }

// func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
// 	p.prefixParseFns[tokenType] = fn
// }
