package lexer

import (
	"bytes"
	"io"
	"strings"

	"github.com/brianbroderick/lantern/internal/postgresql/token"
)

// DOTs (i.e. periods) are allowed as part of an identity in this parser whereas the
// query parser at /pkg/sql/parser/parser.go does not allow them as

type Lexer struct {
	r       io.RuneScanner
	lastPos Pos
	Pos     Pos
	ch      rune
	eof     bool // true if reader has ever seen eof.
}

// Pos specifies the line and character position of a token.
// The Char and Line are both zero-based indexes.
type Pos struct {
	Line int
	Char int
}

// eof is a marker code to signify that the reader can't read any more.
const eof = rune(0)
const eol = '\n'

func New(input string) *Lexer {
	l := &Lexer{r: strings.NewReader(input)}
	return l
}

func (l *Lexer) Scan() (tok token.Token, pos Pos) {
	l.skipWhitespace()
	l.read()

	switch l.ch {
	// case '<':
	// 	tok = newToken(token.LT, l.ch)
	// case '>':
	// 	tok = newToken(token.GT, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case '@':
		tok = newToken(token.ATSYMBOL, l.ch)
	case '"':
		pos = l.Pos
		tok = l.scanString()
		return tok, pos
	case 0:
		tok = token.Token{Type: token.EOF, Lit: ""}
	default:
		if isLetter(l.ch) {
			l.unread()
			// get position before we scan the identity
			pos = l.Pos
			tok = l.scanIdent()
			return tok, pos
		} else if isDigit(l.ch) {
			l.unread()
			// get position before we scan the number
			pos = l.Pos
			tok = l.scanNumber()
			return tok, pos
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	return tok, pos
}

func newToken(tokenType token.TokenType, ch rune) token.Token {
	return token.Token{Type: tokenType, Lit: string(ch)}
}

func (l *Lexer) read() {
	var err error
	l.ch, _, err = l.r.ReadRune()
	if err != nil {
		l.ch = eof
	}

	l.lastPos.Char = l.Pos.Char
	l.lastPos.Line = l.Pos.Line

	// Update position
	// Only count EOF once.
	if l.ch == eol {
		l.Pos.Line++
		l.Pos.Char = 0
	} else if !l.eof {
		l.Pos.Char++
	}

	if l.ch == eof {
		l.eof = true
	}
}

func (l *Lexer) unread() {
	l.r.UnreadRune()
	l.Pos.Char = l.lastPos.Char
	l.Pos.Line = l.lastPos.Line
}

// func (l *Lexer) peek() rune {
// 	ch, _, err := l.r.ReadRune()
// 	if err != nil {
// 		ch = eof
// 	}
// 	l.r.UnreadRune()
// 	return ch
// }

func (l *Lexer) skipWhitespace() {
	for {
		l.read()
		if l.ch == eof {
			break
		} else if !isWhitespace(l.ch) {
			l.unread()
			break
		}
	}
}

func (l *Lexer) scanIdent() token.Token {
	var buf bytes.Buffer
	for {
		l.read()
		if !isIdentChar(l.ch) {
			l.unread()
			break
		} else {
			_, _ = buf.WriteRune(l.ch)
		}
	}
	lit := buf.String()
	return token.Token{Type: token.Lookup(lit), Lit: lit}
}

func (l *Lexer) scanString() token.Token {
	var buf bytes.Buffer
	for {
		l.read()
		if l.ch == '"' {
			break
		} else {
			_, _ = buf.WriteRune(l.ch)
		}
	}
	lit := buf.String()
	return token.Token{Type: token.STRING, Lit: lit}
}

func (l *Lexer) scanNumber() token.Token {
	var buf bytes.Buffer

	buf.WriteString(l.scanDigits())

	// If next code points are a full stop and digit then consume them.
	dots := 0
	colons := 0
	hyphens := 0

	for {
		l.read()
		if l.ch == eof {
			break
		} else if l.ch == '.' {
			dots++
			_, _ = buf.WriteRune(l.ch)
			_, _ = buf.WriteString(l.scanDigits())
		} else if l.ch == ':' {
			colons++
			_, _ = buf.WriteRune(l.ch)
			_, _ = buf.WriteString(l.scanDigits())
		} else if l.ch == '-' {
			hyphens++
			_, _ = buf.WriteRune(l.ch)
			_, _ = buf.WriteString(l.scanDigits())
		} else if !isDigit(l.ch) {
			l.unread()
			break
		} else {
			_, _ = buf.WriteRune(l.ch)
		}
	}

	// Date and Time
	if hyphens == 2 {
		return token.Token{Type: token.DATE, Lit: buf.String()}
	} else if colons == 2 {
		return token.Token{Type: token.TIME, Lit: buf.String()}
	}

	// If there is one dot, it's a number (aka float)
	if dots == 0 {
		return token.Token{Type: token.INT, Lit: buf.String()}
	} else if dots == 1 {
		return token.Token{Type: token.NUMBER, Lit: buf.String()}
	}

	// If there are more than one dot, it's an IP Address
	return token.Token{Type: token.IPADDR, Lit: buf.String()}
}

func (l *Lexer) scanDigits() string {
	var buf bytes.Buffer
	for {
		l.read()
		if !isDigit(l.ch) {
			l.unread()
			break
		} else {
			buf.WriteRune(l.ch)
		}
	}
	return buf.String()
}

func (l *Lexer) ScanQuery() token.Token {
	var buf bytes.Buffer
	l.skipWhitespace()

	for {
		l.read()
		if l.ch == eol || l.ch == eof {
			l.unread()
			break
		} else {
			_, _ = buf.WriteRune(l.ch)
		}
	}
	lit := buf.String()
	return token.Token{Type: token.QUERY, Lit: lit}
}

// isWhitespace returns true if the rune is a space, tab, or newline.
func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' || ch == eol || ch == '\r' }

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool {
	if isWhitespace(ch) || isDigit(ch) {
		return false
	}

	switch ch {
	case '(', ')', '[', ']', ':', '@', eof:
		return false
	default:
		return true
	}

	// return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '<' || ch == '>' || ch == '\''
}

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool { return (ch >= '0' && ch <= '9') }

// isIdentChar returns true if the rune can be used in an unquoted identifier.
func isIdentChar(ch rune) bool {
	if isDigit(ch) {
		return true
	}

	if isWhitespace(ch) {
		return false
	}

	switch ch {
	case '(', ')', '[', ']', ':', '@', eof:
		return false
	default:
		return true
	}
}
