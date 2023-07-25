package pgLogParser

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

//*******
// LOG
//*******

// LogStatement represents a log entry
type LogStatement struct {
	date            string
	time            string
	timezone        string
	remoteHost      string
	remotePort      int
	user            string
	database        string
	pid             int
	severity        string
	durationLit     string
	durationMeasure string
	preparedStep    string
	preparedName    string
	statement       string
	// duration        time.Duration
}

func (s *LogStatement) String() string {
	var buf bytes.Buffer
	_, _ = buf.WriteString(fmt.Sprintf("%s %s %s:%s(%d):%s@%s:[%d]:%s:  duration: %s %s  %s <%s>: %s",
		s.date, s.time, s.timezone, s.remoteHost, s.remotePort, s.user, s.database, s.pid, s.severity, s.durationLit, s.durationMeasure,
		s.preparedStep, s.preparedName, s.statement))

	return buf.String()
}

func (s *LogStatement) KeyTok() Token {
	return DATE
}

func (s *LogStatement) ShortString() string {
	return fmt.Sprintf("%s %s %s: %s", s.date, s.time, s.timezone, s.statement)
}

// parseLogStatement parses a log entry a Statement AST object.
// a log looks like this:
// 2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>:
func (p *Parser) parseLogStatement() (*LogStatement, error) {
	p.Unscan()
	stmt := &LogStatement{}

	iter := 0
	for {
		var err error
		intLit := 0

		tok, _, lit := p.ScanIgnoreWhitespace()
		if tok == INTEGER {
			intLit, err = strconv.Atoi(lit)
			if hasErr(err) {
				return nil, err
			}
		}

		switch iter {
		case 0:
			if tok != DATE {
				return nil, parseErr(iter, DATE, tok)
			}
			stmt.date = lit
		case 1:
			if tok != TIME {
				return nil, parseErr(iter, TIME, tok)
			}
			stmt.time = lit
		case 2:
			if tok != IDENT {
				return nil, parseErr(iter, IDENT, tok)
			}
			stmt.timezone = lit
		case 3:
			if tok != COLON {
				return nil, parseErr(iter, COLON, tok)
			}
		case 4:
			if tok != IPADDR {
				return nil, parseErr(iter, IPADDR, tok)
			}
			stmt.remoteHost = lit
		case 5:
			if tok != LPAREN {
				return nil, parseErr(iter, LPAREN, tok)
			}
		case 6:
			if tok != INTEGER {
				return nil, parseErr(iter, INTEGER, tok)
			}
			stmt.remotePort = intLit
		case 7:
			if tok != RPAREN {
				return nil, parseErr(iter, RPAREN, tok)
			}
		case 8:
			if tok != COLON {
				return nil, parseErr(iter, COLON, tok)
			}
		case 9:
			if tok != IDENT {
				return nil, parseErr(iter, IDENT, tok)
			}
			stmt.user = lit
		case 10:
			if tok != ATSYMBOL {
				return nil, parseErr(iter, ATSYMBOL, tok)
			}
		case 11:
			if tok != IDENT {
				return nil, parseErr(iter, IDENT, tok)
			}
			stmt.database = lit
		case 12:
			if tok != COLON {
				return nil, parseErr(iter, COLON, tok)
			}
		case 13:
			if tok != LBRACKET {
				return nil, parseErr(iter, LBRACKET, tok)
			}
		case 14:
			if tok != INTEGER {
				return nil, parseErr(iter, INTEGER, tok)
			}
			stmt.pid = intLit
		case 15:
			if tok != RBRACKET {
				return nil, parseErr(iter, RBRACKET, tok)
			}
		case 16:
			if tok != COLON {
				return nil, parseErr(iter, COLON, tok)
			}
		case 17:
			if tok != IDENT {
				return nil, parseErr(iter, IDENT, tok)
			}
			stmt.severity = lit
		case 18:
			if tok != COLON {
				return nil, parseErr(iter, COLON, tok)
			}
		case 19:
			if tok != IDENT && lit != "duration" {
				return nil, parseErr(iter, IDENT, tok)
			}
		case 20:
			if tok != COLON {
				return nil, parseErr(iter, COLON, tok)
			}
		case 21:
			if tok != NUMBER {
				return nil, parseErr(iter, COLON, tok)
			}
			stmt.durationLit = lit
		case 22:
			if tok != IDENT {
				return nil, parseErr(iter, IDENT, tok)
			}
			stmt.durationMeasure = lit
		case 23:
			if tok != IDENT {
				return nil, parseErr(iter, IDENT, tok)
			}
			stmt.preparedStep = lit
		case 24:
			if tok != LT {
				return nil, parseErr(iter, LT, tok)
			}
		case 25:
			if tok != IDENT {
				return nil, parseErr(iter, IDENT, tok)
			}
			stmt.preparedName = lit
		case 26:
			if tok != GT {
				return nil, parseErr(iter, GT, tok)
			}
		case 27:
			if tok != COLON {
				return nil, parseErr(iter, COLON, tok)
			}
		default:
			qLines := make([]string, 0)
			qLines = append(qLines, lit)
			qtok, _, qlit := p.ScanQuery()
			if qtok == QUERY {
				qLines = append(qLines, qlit)
				for {
					peektok, _, _ := p.ScanIgnoreWhitespace()
					p.Unscan()
					if peektok == DATE || peektok == EOF {
						break
					}
					qtok, _, qlit := p.ScanQuery()
					if qtok == QUERY {
						qLines = append(qLines, qlit)
					} else {
						break
					}
				}
				stmt.statement = strings.Join(qLines, " ")
				return stmt, nil
			} else {
				return nil, parseErr(iter, QUERY, tok)
			}
		}

		iter++
	}
}

func parseErr(iter int, expected Token, tok Token) error {
	return fmt.Errorf("%d: expected %s, got %s", iter, expected, tok)
}
