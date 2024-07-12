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

		fmt.Println("Length of statements", len(program.Statements))

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
	s := &ast.LogStatement{Token: p.curToken}

	if p.curTokenIs(token.DATE) {
		s.Date = p.curToken.Lit
		p.nextToken()
	} else {
		return s, p.parseErr(1, token.DATE, p.curToken)
	}

	if p.curTokenIs(token.TIME) {
		s.Time = p.curToken.Lit
		p.nextToken()
	}

	if p.curTokenIs(token.IDENT) {
		s.Timezone = p.curToken.Lit
		p.nextToken()
	}

	if p.curTokenIs(token.COLON) {
		p.nextToken()
	}

	ip := make([]string, 0, 1)
loop:
	for {
		// scan to the end of the line and keep going until a new line starts with a date or EOF
		ip = append(ip, p.curToken.Lit)
		p.nextToken()

		switch p.curToken.Type {
		case token.DATE, token.EOF, token.LPAREN:
			break loop
		}
	}

	s.RemoteHost = strings.Join(ip, "")

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
	}

	if p.curTokenIs(token.INT) {
		i, err := p.convertToInt()
		if hasErr(err) {
			return nil, err
		}
		s.RemotePort = i
		p.nextToken()
	}

	if p.curTokenIs(token.RPAREN) {
		p.nextToken()
	}

	if p.curTokenIs(token.COLON) {
		p.nextToken()
	}

	if p.curTokenIs(token.IDENT) {
		s.User = p.curToken.Lit
		p.nextToken()
	}

	if p.curTokenIs(token.ATSYMBOL) {
		p.nextToken()
	}

	if p.curTokenIs(token.IDENT) {
		s.Database = p.curToken.Lit
		p.nextToken()
	}

	if p.curTokenIs(token.COLON) {
		p.nextToken()
	}

	if p.curTokenIs(token.LBRACKET) {
		p.nextToken()
	}

	if p.curTokenIs(token.INT) {
		i, err := p.convertToInt()
		if hasErr(err) {
			return nil, err
		}
		s.Pid = i
		p.nextToken()
	}

	if p.curTokenIs(token.RBRACKET) {
		p.nextToken()
	}

	if p.curTokenIs(token.COLON) {
		p.nextToken()
	}

	if p.curTokenIs(token.IDENT) {
		s.Severity = p.curToken.Lit
		p.nextToken()
	}

	if p.curTokenIs(token.COLON) {
		p.nextToken()
	}

	// is this a query with a duration, or something else like a parameter?
	// parameters: $1 = 'entrata_g_01_11866'",
	if p.curToken.Lit == "parameters" {
		p.nextToken()
		p.nextToken() // skip the colon

		pLines := make([]string, 0, 2)
		pLines = append(pLines, p.curToken.Lit)

		for {
			if p.peekTokenIs(token.DATE) || p.peekTokenIs(token.EOF) {
				break
			}
			// scan to the end of the line and keep going until a new line starts with a date or EOF
			p.scanQuery()
			pLines = append(pLines, p.curToken.Lit)
		}

		s.Parameters = strings.Join(pLines, " ")
		return s, nil
	} else if p.curToken.Lit == "duration" {
		p.nextToken()
		p.nextToken()
		s.DurationLit = p.curToken.Lit
		p.nextToken()

		if p.curTokenIs(token.IDENT) {
			s.DurationMeasure = p.curToken.Lit

			// Sometimes the log just ends here
			if p.peekTokenIs(token.DATE) || p.peekTokenIs(token.EOF) {
				return s, nil
			}

			p.nextToken()
		}
	}

	if p.curTokenIs(token.IDENT) {
		s.PreparedStep = p.curToken.Lit

		if p.peekTokenIs(token.IDENT) {
			p.nextToken()
			s.PreparedName = p.curToken.Lit
		}

		if p.peekTokenIs(token.COLON) {
			p.nextToken()
		}
	}

	if p.peekTokenIs(token.DATE) || p.peekTokenIs(token.EOF) {
		return s, nil
	}

	if p.curTokenIs(token.COLON) {
		p.nextToken()
	}

	// Query
	qLines := make([]string, 0, 2)
	qLines = append(qLines, p.curToken.Lit)

	for {
		if p.peekTokenIs(token.DATE) || p.peekTokenIs(token.EOF) {

			break
		}
		// scan to the end of the line and keep going until a new line starts with a date or EOF
		p.scanQuery()
		qLines = append(qLines, p.curToken.Lit)
	}

	s.Query = strings.Join(qLines, " ")

	// fmt.Println("Stmt: ", s)

	return s, nil
}

// func (p *Parser) parseLogStatement() (*ast.LogStatement, error) {
// 	stmt := &ast.LogStatement{Token: p.curToken}
// 	iter := 0
// 	eos := false // end of statement

// 	// Tokens should always be in the same order, so we can just iterate through them
// 	for !eos {
// 		var err error
// 		intLit := 0

// 		if p.curTokenIs(token.INT) {
// 			intLit, err = strconv.Atoi(p.curToken.Lit)
// 			if hasErr(err) {
// 				return nil, err
// 			}
// 		}

// 		switch iter {
// 		case 0:
// 			if !p.curTokenIs(token.DATE) {
// 				return nil, p.parseErr(iter, token.DATE, p.curToken)
// 			}
// 			stmt.Date = p.curToken.Lit

// 			// fmt.Printf("Date: %s\n", stmt.Date)
// 		case 1:
// 			if !p.curTokenIs(token.TIME) {
// 				return nil, p.parseErr(iter, token.TIME, p.curToken)
// 			}
// 			stmt.Time = p.curToken.Lit

// 			// fmt.Printf("Time: %s\n", stmt.Time)
// 		case 2:
// 			if !p.curTokenIs(token.IDENT) {
// 				return nil, p.parseErr(iter, token.IDENT, p.curToken)
// 			}
// 			stmt.Timezone = p.curToken.Lit

// 			// fmt.Printf("Timezone: %s\n", stmt.Timezone)
// 		case 3:
// 			if !p.curTokenIs(token.COLON) {
// 				return nil, p.parseErr(iter, token.COLON, p.curToken)
// 			}
// 		case 4:
// 			qToks := make([]string, 0, 1)

// 		loop:
// 			for {
// 				// scan to the end of the line and keep going until a new line starts with a date or EOF
// 				qToks = append(qToks, p.curToken.Lit)

// 				switch p.peekToken.Type {
// 				case token.DATE, token.EOF, token.LPAREN:
// 					break loop
// 				}
// 				p.nextToken()
// 			}

// 			stmt.RemoteHost = strings.Join(qToks, "")

// 			// fmt.Printf("RemoteHost: %s\n", stmt.RemoteHost)
// 		case 5:
// 			if !p.curTokenIs(token.LPAREN) {
// 				return stmt, p.parseErr(iter, token.LPAREN, p.curToken)
// 			}
// 		case 6:
// 			if !p.curTokenIs(token.INT) {
// 				return stmt, p.parseErr(iter, token.INT, p.curToken)
// 			}
// 			stmt.RemotePort = intLit

// here
// 		case 7:
// 			if !p.curTokenIs(token.RPAREN) {
// 				return stmt, p.parseErr(iter, token.RPAREN, p.curToken)
// 			}
// 		case 8:
// 			if !p.curTokenIs(token.COLON) {
// 				return stmt, p.parseErr(iter, token.COLON, p.curToken)
// 			}
// 		case 9:
// 			if !p.curTokenIs(token.IDENT) {
// 				return stmt, p.parseErr(iter, token.IDENT, p.curToken)
// 			}
// 			stmt.User = p.curToken.Lit
// 		case 10:
// 			if !p.curTokenIs(token.ATSYMBOL) {
// 				return stmt, p.parseErr(iter, token.ATSYMBOL, p.curToken)
// 			}
// 		case 11:
// 			if !p.curTokenIs(token.IDENT) {
// 				return stmt, p.parseErr(iter, token.IDENT, p.curToken)
// 			}
// 			stmt.Database = p.curToken.Lit
// 		case 12:
// 			if !p.curTokenIs(token.COLON) {
// 				return stmt, p.parseErr(iter, token.COLON, p.curToken)
// 			}
// 		case 13:
// 			if !p.curTokenIs(token.LBRACKET) {
// 				return stmt, p.parseErr(iter, token.LBRACKET, p.curToken)
// 			}
// 		case 14:
// 			if !p.curTokenIs(token.INT) {
// 				return stmt, p.parseErr(iter, token.INT, p.curToken)
// 			}
// 			stmt.Pid = intLit
// 		case 15:
// 			if !p.curTokenIs(token.RBRACKET) {
// 				return stmt, p.parseErr(iter, token.RBRACKET, p.curToken)
// 			}
// 		case 16:
// 			if !p.curTokenIs(token.COLON) {
// 				return stmt, p.parseErr(iter, token.COLON, p.curToken)
// 			}
// 		case 17:
// 			if !p.curTokenIs(token.IDENT) {
// 				return stmt, p.parseErr(iter, token.IDENT, p.curToken)
// 			}
// 			stmt.Severity = p.curToken.Lit
// 		case 18:
// 			if !p.curTokenIs(token.COLON) {
// 				return stmt, p.parseErr(iter, token.COLON, p.curToken)
// 			}
// 		case 19:
// 			// is this a query with a duration, or something else like a parameter?
// 			// parameters: $1 = 'entrata_g_01_11866'",
// 			if p.curToken.Lit == "parameters" {
// 				p.nextToken() // skip the colon
// 				p.nextToken() // skip the space

// 				pLines := make([]string, 0, 2)
// 				pLines = append(pLines, p.curToken.Lit)

// 				for {
// 					if p.peekTokenIs(token.DATE) || p.peekTokenIs(token.EOF) {
// 						eos = true
// 						break
// 					}
// 					// scan to the end of the line and keep going until a new line starts with a date or EOF
// 					p.scanQuery()
// 					pLines = append(pLines, p.curToken.Lit)
// 				}

// 				stmt.Parameters = strings.Join(pLines, " ")
// 				eos = true
// 			}

// 		case 20:
// 			if !p.curTokenIs(token.COLON) {
// 				return stmt, p.parseErr(iter, token.COLON, p.curToken)
// 			}
// 		case 21:
// 			if !p.curTokenIs(token.NUMBER) {
// 				return stmt, p.parseErr(iter, token.NUMBER, p.curToken)
// 			}
// 			stmt.DurationLit = p.curToken.Lit
// 		case 22:
// 			if !p.curTokenIs(token.IDENT) {
// 				return stmt, p.parseErr(iter, token.IDENT, p.curToken)
// 			}
// 			stmt.DurationMeasure = p.curToken.Lit

// 			// Sometimes the log just ends here
// 			if p.peekTokenIs(token.DATE) || p.peekTokenIs(token.EOF) {
// 				return stmt, nil
// 			}

// 		case 23:
// 			// bind R_42: BEGIN
// 			if !p.curTokenIs(token.IDENT) {
// 				return stmt, p.parseErr(iter, token.IDENT, p.curToken)
// 			}
// 			stmt.PreparedStep = p.curToken.Lit

// 			if p.peekTokenIs(token.IDENT) {
// 				p.nextToken()
// 				stmt.PreparedName = p.curToken.Lit

// 			}
// 		case 24:
// 			if !p.curTokenIs(token.COLON) {
// 				fmt.Println("Current token", p.curToken)
// 				return stmt, p.parseErr(iter, token.COLON, p.curToken)
// 			}
// 		default:
// 			qLines := make([]string, 0, 2)
// 			qLines = append(qLines, p.curToken.Lit)

// 			for {
// 				if p.peekTokenIs(token.DATE) || p.peekTokenIs(token.EOF) {
// 					eos = true
// 					break
// 				}
// 				// scan to the end of the line and keep going until a new line starts with a date or EOF
// 				p.scanQuery()
// 				qLines = append(qLines, p.curToken.Lit)
// 			}

// 			stmt.Query = strings.Join(qLines, " ")
// 			eos = true
// 		}

// 		// We're not done with this statement, so move to the next token
// 		if !eos {
// 			p.nextToken()
// 			iter++
// 		}
// 	}

// 	return stmt, nil
// }

func (p *Parser) parseErr(iter int, expected token.TokenType, tok token.Token) error {
	return fmt.Errorf("line %d char %d: %d: expected %s, got %s. Lit: %s", p.l.Pos.Line, p.l.Pos.Char, iter, expected, tok.Type, tok.Lit)
}

func hasErr(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
		return true
	}
	return false
}

func (p *Parser) convertToInt() (int, error) {
	return strconv.Atoi(p.curToken.Lit)
}
