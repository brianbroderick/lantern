package lexer

import (
	"bytes"
	"io"
	"strings"

	"github.com/brianbroderick/lantern/internal/postgresql/token"
)

type Lexer struct {
	r       io.RuneScanner
	lastPos Pos
	pos     Pos
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
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
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
	case '"':
		pos = l.pos
		tok = l.scanString()
		return tok, pos
	case 0:
		tok = token.Token{Type: token.EOF, Literal: ""}
	default:
		if isLetter(l.ch) {
			l.unread()
			// get position before we scan the identity
			pos = l.pos
			tok = l.scanIdent()
			return tok, pos
		} else if isDigit(l.ch) {
			l.unread()
			// get position before we scan the number
			pos = l.pos
			tok = l.scanNumber()
			return tok, pos
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	return tok, pos
}

func newToken(tokenType token.TokenType, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) read() {
	var err error
	l.ch, _, err = l.r.ReadRune()
	if err != nil {
		l.ch = eof
	}

	l.lastPos.Char = l.pos.Char
	l.lastPos.Line = l.pos.Line

	// Update position
	// Only count EOF once.
	if l.ch == eol {
		l.pos.Line++
		l.pos.Char = 0
	} else if !l.eof {
		l.pos.Char++
	}

	if l.ch == eof {
		l.eof = true
	}
}

func (l *Lexer) unread() {
	l.r.UnreadRune()
	l.pos.Char = l.lastPos.Char
	l.pos.Line = l.lastPos.Line
}

func (l *Lexer) peek() rune {
	ch, _, err := l.r.ReadRune()
	if err != nil {
		ch = eof
	}
	l.r.UnreadRune()
	return ch
}

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
	return token.Token{Type: token.Lookup(lit), Literal: lit}
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
	return token.Token{Type: token.STRING, Literal: lit}
}

func (l *Lexer) scanNumber() token.Token {
	var buf bytes.Buffer
	for {
		l.read()
		if !isDigit(l.ch) {
			l.unread()
			break
		} else {
			_, _ = buf.WriteRune(l.ch)
		}
	}
	return token.Token{Type: token.INT, Literal: buf.String()}
}

// isWhitespace returns true if the rune is a space, tab, or newline.
func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' || ch == eol || ch == '\r' }

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool { return (ch >= '0' && ch <= '9') }

// isIdentChar returns true if the rune can be used in an unquoted identifier.
func isIdentChar(ch rune) bool { return isLetter(ch) || isDigit(ch) || ch == '_' }
