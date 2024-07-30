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
	l                   *lexer.Lexer
	errors              []string
	incompleteStatement map[string]*ast.LogStatement

	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:                   l,
		errors:              []string{},
		incompleteStatement: make(map[string]*ast.LogStatement),
	}

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) scanQuery() {
	origPeek := p.peekToken
	scan := p.l.ScanQuery()
	if scan.Lit != "" {
		p.peekToken = token.Token{Type: token.QUERY, Lit: fmt.Sprintf("%s %s", origPeek.Lit, scan.Lit)}
	}
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
	lenProgramStatements := 0

	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
			lenProgramStatements++

			l := lenProgramStatements
			if l%250000 == 0 {
				fmt.Printf("statements: %d, line number: %d\n", lenProgramStatements, p.l.Pos.Line)
			}
		}

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

	if stmt.Severity == "LOG" && stmt.DurationLit == "" {
		// If we don't have a duration, we need to store the statement so we can match it with the next statement.
		p.incompleteStatement[stmt.RemoteHost+stmt.User+stmt.Database+stmt.Severity] = stmt
		return nil
	} else if stmt.Query == "" {
		// If we have a duration, we need to match it with the previous statement.
		key := stmt.RemoteHost + stmt.User + stmt.Database + stmt.Severity
		if prevStmt, ok := p.incompleteStatement[key]; ok {
			prevStmt.DurationLit = stmt.DurationLit
			prevStmt.DurationMeasure = stmt.DurationMeasure

			p.incompleteStatement[key] = nil
			return prevStmt
		}
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

	// Database
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

	// Pid
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

	// Severity
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

			// Sometimes the log entries just end here because they consist of two entries that much be matched together.
			// The matching happens	in the parent function.
			if p.peekTokenIs(token.DATE) || p.peekTokenIs(token.EOF) {
				return s, nil
			}

			p.nextToken()
		}
	}

	if s.Severity != "ERROR" {
		// Options are blank, statement, execute, bind, parse.
		// We only care about parsing statement and execute right now.
		if p.curTokenIs(token.IDENT) {
			s.PreparedStep = p.curToken.Lit

			if p.peekTokenIs(token.IDENT) {
				nameToks := make([]string, 0, 1)

				for {
					p.nextToken()

					nameToks = append(nameToks, p.curToken.Lit)

					if p.peekTokenIs(token.COLON) || p.peekTokenIs(token.DATE) || p.peekTokenIs(token.EOF) {
						break
					}
				}

				s.PreparedName = strings.Join(nameToks, " ")
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

	if s.Severity == "ERROR" {
		s.Error = strings.Join(qLines, " ")
	} else {
		s.Query = strings.Join(qLines, " ")
	}

	return s, nil
}

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
