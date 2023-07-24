package pgLogParser

import (
	"bufio"
	"bytes"
	"io"
)

// scanner is a wrapper for io.RuneScanner to scan a slice of runes
// It provides functions to scan the next token, peek at the next token, and backup the last token

// Scanner represents a buffered rune reader used for scanning
// It provides a fixed-length circular buffer that can be unread
type Scanner struct {
	r   io.RuneScanner
	i   int // buffer index
	n   int // buffer char count
	pos Pos // last read rune position
	buf [3]struct {
		ch  rune
		pos Pos
	}
	eof bool // true if we have read EOF
}

// eof is a marker code point to signify that the reader can't read any more
const eof = rune(0)

// eol is a marker code point to signify that the reader has read a newline
const eol = '\n'

// NewScanner returns a new instance of Scanner
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// Scan returns the next token and position from the underlying reader
func (s *Scanner) Scan() (tok Token, pos Pos, lit string) {
	// Read next code point
	ch, pos := s.read()

	if isWhitespace(ch) {
		return s.scanWhitespace()
	} else if isLetter(ch) || ch == '_' || ch == '"' {
		s.unread()
		return s.scanIdent(true)
	} else if isDigit(ch) {
		s.unread()
		return s.scanNumber()
	}

	switch ch {
	case eof:
		return EOF, pos, ""
	case '(':
		return LPAREN, pos, ""
	case ')':
		return RPAREN, pos, ""
	case '[':
		return LBRACKET, pos, ""
	case ']':
		return RBRACKET, pos, ""
	case ',':
		return COMMA, pos, ""
	case ':':
		return COLON, pos, ""
	case ';':
		return SEMI, pos, ""
	case '.':
		return DOT, pos, ""
	case '@':
		return ATSYMBOL, pos, ""
	case '>':
		return GT, pos, ""
	case '<':
		return LT, pos, ""
	}

	return ILLEGAL, pos, string(ch)
}

// curr returns the last read character and position.
func (s *Scanner) curr() (ch rune, pos Pos) {
	i := (s.i - s.n + len(s.buf)) % len(s.buf)
	buf := &s.buf[i]
	return buf.ch, buf.pos
}

// read reads the next rune from the reader.
func (s *Scanner) read() (ch rune, pos Pos) {
	// If we have unread characters then read them off the buffer first.
	if s.n > 0 {
		s.n--
		return s.curr()
	}

	// Read next rune from underlying reader.
	// Any error (including io.EOF) should return as EOF.
	ch, _, err := s.r.ReadRune()
	if err != nil {
		ch = eof
	} else if ch == '\r' {
		if ch, _, err := s.r.ReadRune(); err != nil {
			// nop
		} else if ch != eol {
			_ = s.r.UnreadRune()
		}
		ch = eol
	}

	// Save character and position to the buffer.
	s.i = (s.i + 1) % len(s.buf)
	buf := &s.buf[s.i]
	buf.ch, buf.pos = ch, s.pos

	// Update position.
	// Only count EOF once.
	if ch == eol {
		s.pos.Line++
		s.pos.Char = 0
	} else if !s.eof {
		s.pos.Char++
	}

	// Mark the reader as EOF.
	// This is used so we don't double count EOF characters.
	if ch == eof {
		s.eof = true
	}

	return s.curr()
}

// unread pushes the previously read rune back onto the buffer.
func (s *Scanner) unread() {
	s.n++
}

func (s *Scanner) scanIdent(lookup bool) (tok Token, pos Pos, lit string) {
	// Save the starting position of the identifier.
	_, pos = s.read()
	s.unread()

	var buf bytes.Buffer
	for {
		if ch, _ := s.read(); ch == eof {
			break
		} else if ch == '"' {
			s.unread()
			tok0, pos0, lit0 := s.scanString()
			if tok0 == BADSTRING || tok0 == BADESCAPE {
				return tok0, pos0, lit0
			}
			return tok0, pos, lit0
		} else if ch == '\'' {
			s.unread()
			tok0, pos0, lit0 := s.scanUnescapedString()
			if tok0 == BADSTRING || tok0 == BADESCAPE {
				return tok0, pos0, lit0
			}
			return tok0, pos, lit0
		} else if isIdentChar(ch) {
			s.unread()
			buf.WriteString(s.scanBareIdent())
		} else {
			s.unread()
			break
		}
	}

	lit = buf.String()

	// If the literal matches a keyword then return that keyword.
	if lookup {
		if tok = Lookup(lit); tok != IDENT {
			return tok, pos, ""
		}
	}
	return IDENT, pos, lit
}

// ScanBareIdent reads bare identifier from a rune reader.
func (s *Scanner) scanBareIdent() string {
	// Read every ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	var buf bytes.Buffer
	for {
		ch, _ := s.read()
		if !isIdentChar(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}
	return buf.String()
}

func (s *Scanner) scanUnescapedString() (tok Token, pos Pos, lit string) {
	_, pos = s.curr()
	ending, _ := s.read()

	var buf bytes.Buffer
	for {
		ch0, _ := s.read()

		if ch0 == ending {
			return STRING, pos, buf.String()
		} else if ch0 == eol || ch0 == eof {
			return BADSTRING, pos, buf.String()
		} else {
			_, _ = buf.WriteRune(ch0)
		}
	}
}

// scanString consumes a contiguous string of non-quote characters.
// Quote characters can be consumed if they're first escaped with a backslash.
func (s *Scanner) scanString() (tok Token, pos Pos, lit string) {
	_, pos = s.curr()
	ending, _ := s.read()

	var buf bytes.Buffer
	for {
		ch0, _ := s.read()

		if ch0 == ending {
			return STRING, pos, buf.String()
		} else if ch0 == eol || ch0 == eof {
			return BADSTRING, pos, buf.String()
		} else if ch0 == '\\' {
			// If the next character is an escape then write the escaped char.
			// If it's not a valid escape then return an error.
			ch1, _ := s.read()
			if ch1 == 'n' {
				_, _ = buf.WriteRune(eol)
			} else if ch1 == '\\' {
				_, _ = buf.WriteRune('\\')
			} else if ch1 == '"' {
				_, _ = buf.WriteRune('"')
			} else if ch1 == '\'' {
				_, _ = buf.WriteRune('\'')
			} else {
				return BADESCAPE, pos, string(ch0) + string(ch1)
			}
		} else {
			_, _ = buf.WriteRune(ch0)
		}
	}
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, pos Pos, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	ch, pos := s.curr()
	_, _ = buf.WriteRune(ch)

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		ch, _ = s.read()
		if ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	return WS, pos, buf.String()
}

// scanNumber consumes anything that looks like the start of a number.
func (s *Scanner) scanNumber() (tok Token, pos Pos, lit string) {
	var buf bytes.Buffer

	// Read as many digits as possible.
	_, _ = buf.WriteString(s.scanDigits())

	// If next code points are a full stop and digit then consume them.
	dots := 0
	colons := 0
	hyphens := 0

	for {
		ch, _ := s.read()
		if ch == eof {
			break
		} else if ch == '.' {
			dots++
			_, _ = buf.WriteRune(ch)
			_, _ = buf.WriteString(s.scanDigits())
		} else if ch == ':' {
			colons++
			_, _ = buf.WriteRune(ch)
			_, _ = buf.WriteString(s.scanDigits())
		} else if ch == '-' {
			hyphens++
			_, _ = buf.WriteRune(ch)
			_, _ = buf.WriteString(s.scanDigits())
		} else if !isDigit(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// Date and Time
	if hyphens == 2 {
		return DATE, pos, buf.String()
	} else if colons == 2 {
		return TIME, pos, buf.String()
	}

	// If there is one dot, it's a number (aka float)
	if dots == 0 {
		return INTEGER, pos, buf.String()
	} else if dots == 1 {
		return NUMBER, pos, buf.String()
	}

	// If there are more than one dot, it's an IP Address
	return IPADDR, pos, buf.String()
}

// scanDigits consumes a contiguous series of digits.
func (s *Scanner) scanDigits() string {
	var buf bytes.Buffer
	for {
		ch, _ := s.read()

		if !isDigit(ch) {
			s.unread()
			break
		}
		_, _ = buf.WriteRune(ch)
	}
	return buf.String()
}

// isWhitespace returns true if the rune is a space, tab, or newline.
func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' || ch == eol }

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool { return (ch >= '0' && ch <= '9') }

// isIdentChar returns true if the rune can be used in an unquoted identifier.
func isIdentChar(ch rune) bool { return isLetter(ch) || isDigit(ch) || ch == '_' }
