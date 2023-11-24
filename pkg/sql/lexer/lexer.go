package lexer

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/token"
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
	pos = l.pos

	switch l.ch {
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		if l.peek() == '>' {
			l.read()
			if l.peek() == '>' {
				l.read()
				tok = token.Token{Type: token.JSONGETBYTEXT, Lit: "->>"}
			} else {
				tok = token.Token{Type: token.JSONGETBYKEY, Lit: "->"}
			}
			// Comments out the rest of the line by using --
		} else if l.peek() == '-' {
			for {
				l.read()
				if l.ch == eof {
					break
				} else if l.ch == eol {
					break
				}
			}
			tok = token.Token{Type: token.COMMENT, Lit: ""}
		} else {
			tok = newToken(token.MINUS, l.ch)
		}
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '<': // maybe a JSON operator
		switch l.peek() {
		case '=':
			l.read()
			tok = token.Token{Type: token.LTE, Lit: "<="}
		case '@':
			l.read()
			tok = token.Token{Type: token.JSONCONTAINED, Lit: "<@"}
		case '>':
			l.read()
			tok = token.Token{Type: token.NOT_EQ, Lit: "<>"}
		default:
			tok = newToken(token.LT, l.ch)
		}
	case '>':
		switch l.peek() {
		case '=':
			l.read()
			tok = token.Token{Type: token.GTE, Lit: ">="}
		default:
			tok = newToken(token.GT, l.ch)
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ':':
		if l.peek() == ':' {
			ch := l.ch
			l.read()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.DOUBLECOLON, Lit: literal}
		} else {
			tok = newToken(token.COLON, l.ch)
		}
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case '.':
		tok = newToken(token.DOT, l.ch)
	case '|':
		if l.peek() == '|' {
			ch := l.ch
			l.read()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.JSONCONCAT, Lit: literal}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '?':
		if l.peek() == '|' {
			ch := l.ch
			l.read()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.JSONHASANYKEYS, Lit: literal}
		} else if l.peek() == '&' {
			ch := l.ch
			l.read()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.JSONHASALLKEYS, Lit: literal}
		} else {
			tok = newToken(token.JSONHASKEY, l.ch)
		}
	case '=':
		if l.peek() == '=' {
			ch := l.ch
			l.read()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Lit: literal}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '!':
		if l.peek() == '=' {
			ch := l.ch
			l.read()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Lit: literal}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '/':
		// This is a comment; however, it should be skipped by skipWhitespace()
		if l.peek() == '*' {
			l.read()
			for {
				l.read()
				if l.ch == eof {
					break
				} else if l.ch == '*' {
					if l.peek() == '/' {
						l.read()
						break
					}
				}
			}
			tok = token.Token{Type: token.COMMENT, Lit: ""}
		} else {
			tok = newToken(token.SLASH, l.ch)
		}
	case '\'':
		tok = l.scanString()
		return tok, pos
	case '"': // double quotes are allowed to surround identities
		// tok = l.scanIdent()
		tok = l.scanDoubleQuoteString()
		return tok, pos
	case '#': // JSON operators
		if l.peek() == '>' {
			l.read()
			if l.peek() == '>' {
				l.read()
				tok = token.Token{Type: token.JSONGETBYPATHTEXT, Lit: "#>>"}
			} else {
				tok = token.Token{Type: token.JSONGETBYPATH, Lit: "#>"}
			}
		} else if l.peek() == '-' {
			l.read()
			tok = token.Token{Type: token.JSONDELETE, Lit: "#-"}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '@': // JSON operators
		if l.peek() == '>' {
			l.read()
			tok = token.Token{Type: token.JSONCONTAINS, Lit: "@>"}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '&': // ARRAY OPERATORS
		if l.peek() == '&' {
			l.read()
			tok = token.Token{Type: token.OVERLAP, Lit: "&&"}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case 0:
		tok = token.Token{Type: token.EOF, Lit: ""}
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
	return token.Token{Type: tokenType, Lit: string(ch)}
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
	// dot := false

	for {
		l.read()
		if isIdentChar(l.ch) {
			// dot = false
			_, _ = buf.WriteRune(l.ch)
			continue
			// We were allowing dots in identifiers to support schema.table.column, but we're now splitting that
			// } else if l.ch == '.' {
			// 	dot = true
			// 	_, _ = buf.WriteRune(l.ch)
			// 	continue
			// This was for the table.* syntax, but we're now splitting that
			// } else if l.ch == '*' && dot {
			// 	dot = false
			// 	_, _ = buf.WriteRune(l.ch)
		} else if l.ch == '"' {
			// dot = false
			// Allow double quotes in identifiers, but don't include them in the token.
			// They can be used to escape reserved words.
		} else {
			l.unread()
			break
		}

		// // Allow '.' in identifiers.
		// // Allow '*' in identifiers.
		// if !isIdentChar(l.ch) && l.ch != '.' && l.ch != '*' {
		// 	l.unread()
		// 	break
		// } else if l.ch == '"' {
		// 	// Allow double quotes in identifiers, but don't include them in the token.
		// 	// They can be used to escape reserved words.
		// } else {
		// 	_, _ = buf.WriteRune(l.ch)
		// }
	}
	lit := buf.String()
	return token.Token{Type: token.Lookup(lit), Lit: lit}
}

// SQL strings are single quoted.
func (l *Lexer) scanString() token.Token {
	var buf bytes.Buffer
	for {
		l.read()
		if l.ch == '\'' {
			l.read()
			if l.ch == '\'' {
				_, _ = buf.WriteRune(l.ch)
			} else {
				l.unread()
				break
			}
		} else {
			_, _ = buf.WriteRune(l.ch)
		}
	}
	lit := buf.String()
	return token.Token{Type: token.STRING, Lit: lit}
}

func (l *Lexer) scanDoubleQuoteString() token.Token {
	var buf bytes.Buffer
	for {
		l.read()
		if l.ch == '"' {
			l.read()
			if l.ch == '"' {
				_, _ = buf.WriteRune(l.ch)
			} else {
				l.unread()
				break
			}
		} else {
			_, _ = buf.WriteRune(l.ch)
		}
	}
	lit := fmt.Sprintf(`"%s"`, buf.String())
	return token.Token{Type: token.IDENT, Lit: lit}
}

func (l *Lexer) scanNumber() token.Token {
	numType := token.INT // default to INT
	var buf bytes.Buffer
	for {
		l.read()
		if !isDigit(l.ch) && l.ch != '.' {
			l.unread()
			break
		} else if l.ch == '.' { // Found a dot, so this is a FLOAT
			numType = token.FLOAT
			_, _ = buf.WriteRune(l.ch)
		} else {
			_, _ = buf.WriteRune(l.ch)
		}
	}
	return token.Token{Type: numType, Lit: buf.String()}
}

// isWhitespace returns true if the rune is a space, tab, or newline.
func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' || ch == eol || ch == '\r' }

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool { return (ch >= '0' && ch <= '9') }

// isIdentChar returns true if the rune can be used in an unquoted identifier.
// identities can be surrounded by double quotes and can contain any character.
func isIdentChar(ch rune) bool { return isLetter(ch) || isDigit(ch) || ch == '_' }
