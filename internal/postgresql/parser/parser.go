package parser

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/brianbroderick/lantern/internal/postgresql/ast"
	"github.com/brianbroderick/lantern/internal/postgresql/lexer"
	"github.com/brianbroderick/lantern/internal/postgresql/token"
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) scanQuery() {
	origPeek := p.peekToken
	scan := p.l.ScanQuery()
	p.peekToken = token.Token{Type: token.QUERY, Lit: fmt.Sprintf("%s %s", origPeek.Lit, scan.Lit)}
	p.nextToken()
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

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		program.Statements = append(program.Statements, stmt)

		// fmt.Println("Length of statements", len(program.Statements))

		// fmt.Printf("current token: %s %s\n", p.curToken.Type, p.curToken.Lit)
		// fmt.Printf("peek token: %s %s\n", p.peekToken.Type, p.peekToken.Lit)

		p.nextToken()
	}

	return program
}

// This function allows for the possibility of having different statement types
// in the future.
func (p *Parser) parseStatement() ast.Statement {
	stmt, err := p.parseLogStatement()
	if err != nil {
		fmt.Printf("Error parsing log statement: %s\nCurrent stmt is: %+v\n", err, stmt)
		os.Exit(1)
		// return nil
	}
	return stmt
}

func (p *Parser) parseLogStatement() (*ast.LogStatement, error) {
	stmt := &ast.LogStatement{Token: p.curToken}
	iter := 0
	eos := false // end of statement

	// Tokens should always be in the same order, so we can just iterate through them
	for !eos {
		var err error
		intLit := 0

		if p.curTokenIs(token.INT) {
			intLit, err = strconv.Atoi(p.curToken.Lit)
			if hasErr(err) {
				return nil, err
			}
		}

		switch iter {
		case 0:
			if !p.curTokenIs(token.DATE) {
				return nil, parseErr(iter, token.DATE, p.curToken.Type)
			}
			stmt.Date = p.curToken.Lit

			// fmt.Printf("Date: %s\n", stmt.Date)
		case 1:
			if !p.curTokenIs(token.TIME) {
				return nil, parseErr(iter, token.TIME, p.curToken.Type)
			}
			stmt.Time = p.curToken.Lit

			// fmt.Printf("Time: %s\n", stmt.Time)
		case 2:
			if !p.curTokenIs(token.IDENT) {
				return nil, parseErr(iter, token.IDENT, p.curToken.Type)
			}
			stmt.Timezone = p.curToken.Lit

			// fmt.Printf("Timezone: %s\n", stmt.Timezone)
		case 3:
			if !p.curTokenIs(token.COLON) {
				return nil, parseErr(iter, token.COLON, p.curToken.Type)
			}
		case 4:
			qToks := make([]string, 0, 1)

		loop:
			for {
				// scan to the end of the line and keep going until a new line starts with a date or EOF
				qToks = append(qToks, p.curToken.Lit)

				switch p.peekToken.Type {
				case token.DATE, token.EOF, token.LPAREN:
					break loop
				}
				p.nextToken()
			}

			stmt.RemoteHost = strings.Join(qToks, "")

			// fmt.Printf("RemoteHost: %s\n", stmt.RemoteHost)
		case 5:
			if !p.curTokenIs(token.LPAREN) {
				return nil, parseErr(iter, token.LPAREN, p.curToken.Type)
			}
		case 6:
			if !p.curTokenIs(token.INT) {
				return nil, parseErr(iter, token.INT, p.curToken.Type)
			}
			stmt.RemotePort = intLit
		case 7:
			if !p.curTokenIs(token.RPAREN) {
				return nil, parseErr(iter, token.RPAREN, p.curToken.Type)
			}
		case 8:
			if !p.curTokenIs(token.COLON) {
				return nil, parseErr(iter, token.COLON, p.curToken.Type)
			}
		case 9:
			if !p.curTokenIs(token.IDENT) {
				return nil, parseErr(iter, token.IDENT, p.curToken.Type)
			}
			stmt.User = p.curToken.Lit
		case 10:
			if !p.curTokenIs(token.ATSYMBOL) {
				return nil, parseErr(iter, token.ATSYMBOL, p.curToken.Type)
			}
		case 11:
			if !p.curTokenIs(token.IDENT) {
				return nil, parseErr(iter, token.IDENT, p.curToken.Type)
			}
			stmt.Database = p.curToken.Lit
		case 12:
			if !p.curTokenIs(token.COLON) {
				return nil, parseErr(iter, token.COLON, p.curToken.Type)
			}
		case 13:
			if !p.curTokenIs(token.LBRACKET) {
				return nil, parseErr(iter, token.LBRACKET, p.curToken.Type)
			}
		case 14:
			if !p.curTokenIs(token.INT) {
				return nil, parseErr(iter, token.INT, p.curToken.Type)
			}
			stmt.Pid = intLit
		case 15:
			if !p.curTokenIs(token.RBRACKET) {
				return nil, parseErr(iter, token.RBRACKET, p.curToken.Type)
			}
		case 16:
			if !p.curTokenIs(token.COLON) {
				return nil, parseErr(iter, token.COLON, p.curToken.Type)
			}
		case 17:
			if !p.curTokenIs(token.IDENT) {
				return nil, parseErr(iter, token.IDENT, p.curToken.Type)
			}
			stmt.Severity = p.curToken.Lit
		case 18:
			if !p.curTokenIs(token.COLON) {
				return nil, parseErr(iter, token.COLON, p.curToken.Type)
			}
		case 19:
			if !p.curTokenIs(token.IDENT) && p.curToken.Lit != "duration" {
				return nil, parseErr(iter, token.IDENT, p.curToken.Type)
			}
		case 20:
			if !p.curTokenIs(token.COLON) {
				return nil, parseErr(iter, token.COLON, p.curToken.Type)
			}
		case 21:
			if !p.curTokenIs(token.NUMBER) {
				return nil, parseErr(iter, token.NUMBER, p.curToken.Type)
			}
			stmt.DurationLit = p.curToken.Lit
		case 22:
			if !p.curTokenIs(token.IDENT) {
				return nil, parseErr(iter, token.IDENT, p.curToken.Type)
			}
			stmt.DurationMeasure = p.curToken.Lit

			// Sometimes the log just ends here
			if p.peekTokenIs(token.DATE) {
				return stmt, nil
			}

		case 23:
			// bind R_42: BEGIN
			if !p.curTokenIs(token.IDENT) {
				return nil, parseErr(iter, token.IDENT, p.curToken.Type)
			}
			stmt.PreparedStep = p.curToken.Lit

			if p.peekTokenIs(token.IDENT) {
				p.nextToken()
				stmt.PreparedName = p.curToken.Lit

			}
		case 24:
			if !p.curTokenIs(token.COLON) {
				fmt.Println("Current token", p.curToken)
				return stmt, parseErr(iter, token.COLON, p.curToken.Type)
			}
		default:
			qLines := make([]string, 0, 2)
			qLines = append(qLines, p.curToken.Lit)

			for {
				if p.peekTokenIs(token.DATE) || p.peekTokenIs(token.EOF) {
					eos = true
					break
				}
				// scan to the end of the line and keep going until a new line starts with a date or EOF
				p.scanQuery()
				qLines = append(qLines, p.curToken.Lit)
			}

			stmt.Query = strings.Join(qLines, " ")
			eos = true
		}

		// We're not done with this statement, so move to the next token
		if !eos {
			p.nextToken()
			iter++
		}
	}

	return stmt, nil
}

func parseErr(iter int, expected token.TokenType, tok token.TokenType) error {
	return fmt.Errorf("%d: expected %s, got %s", iter, expected, tok)
}

func hasErr(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
		return true
	}
	return false
}
