package pgLogParser

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// const longForm = "2006-01-02 15:04:05 MST"

// Parser represents a parser.
type Parser struct {
	t      *tokenizer
	params map[string]interface{}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{t: newTokenizer(r)}
}

// SetParams sets the parameters that will be used for any bound parameter substitutions.
func (p *Parser) SetParams(params map[string]interface{}) {
	p.params = params
}

// ParseStatement parses a statement string and returns its AST representation.
func ParseStatement(s string) (Statement, error) {
	return NewParser(strings.NewReader(s)).ParseStatement()
}

// ParseStatement parses a string and returns a Statement AST object.
func (p *Parser) ParseStatement() (Statement, error) {
	return Language.Parse(p)
}

// ParseIdent parses an identifier.
func (p *Parser) ParseIdent() (string, error) {
	tok, pos, lit := p.ScanIgnoreWhitespace()
	if tok != IDENT {
		return "", newParseError(tokstr(tok, lit), []string{"identifier"}, pos)
	}
	return lit, nil
}

// ParseString parses a string.
func (p *Parser) ParseString() (string, error) {
	tok, pos, lit := p.ScanIgnoreWhitespace()
	if tok != STRING {
		return "", newParseError(tokstr(tok, lit), []string{"string"}, pos)
	}
	return lit, nil
}

// ParseInt parses a string representing a base 10 integer and returns the number.
// It returns an error if the parsed number is outside the range [min, max].
func (p *Parser) ParseInt(min, max int) (int, error) {
	tok, pos, lit := p.ScanIgnoreWhitespace()
	if tok != INTEGER {
		return 0, newParseError(tokstr(tok, lit), []string{"integer"}, pos)
	}

	// Convert string to int.
	n, err := strconv.Atoi(lit)
	if err != nil {
		return 0, &ParseError{Message: err.Error(), Pos: pos}
	} else if min > n || n > max {
		return 0, &ParseError{
			Message: fmt.Sprintf("invalid value %d: must be %d <= n <= %d", n, min, max),
			Pos:     pos,
		}
	}

	return n, nil
}

// ParseUInt64 parses a string and returns a 64-bit unsigned integer literal.
func (p *Parser) ParseUInt64() (uint64, error) {
	tok, pos, lit := p.ScanIgnoreWhitespace()
	if tok != INTEGER {
		return 0, newParseError(tokstr(tok, lit), []string{"integer"}, pos)
	}

	// Convert string to unsigned 64-bit integer
	n, err := strconv.ParseUint(lit, 10, 64)
	if err != nil {
		return 0, &ParseError{Message: err.Error(), Pos: pos}
	}

	return uint64(n), nil
}

// ScanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) ScanIgnoreWhitespace() (tok Token, pos Pos, lit string) {
	for {
		tok, pos, lit = p.Scan()
		if tok == WS {
			continue
		}
		return
	}
}

// UnscanIgnoreWhitespace backs up to the previous non-whitespace token.
func (p *Parser) UnscanIgnoreWhitespace() (tok Token, pos Pos, lit string) {
	return p.UnscanIgnore(WS)
}

// UnscanTo backs up to the first occurance of the Token in question.
func (p *Parser) UnscanIgnore(ignore Token) (tok Token, pos Pos, lit string) {
	for {
		tok, pos, lit = p.UnscanAndReturn()
		if tok == ignore {
			continue
		}
		return tok, pos, lit
	}
}

// UnscanTo backs up to the first occurance of the Token in question.
func (p *Parser) UnscanTo(stop Token) (tok Token, pos Pos, lit string) {
	for {
		tok, pos, lit = p.UnscanAndReturn()
		if tok != stop {
			continue
		}
		return tok, pos, lit
	}
}

// Scan returns the next token from the underlying scanner.
func (p *Parser) Scan() (tok Token, pos Pos, lit string) { return p.t.Scan() }

func (p *Parser) UnscanAndReturn() (tok Token, pos Pos, lit string) { return p.t.UnscanAndReturn() }

func (p *Parser) ScanQuery() (tok Token, pos Pos, lit string) {
	var words []string
	for p.t.n > 0 {
		_, _, lit = p.Scan()
		words = append(words, lit)
	}
	tok, pos, lit = p.t.s.scanQuery()
	// if lit != "" {
	// 	words = append(words, lit)
	// }

	return tok, pos, strings.Join(words, " ") + lit
}

// Unscan pushes the previously token back onto the underlying buffer.
func (p *Parser) Unscan() { p.t.Unscan() }

// ParseError represents an error that occurred during parsing.
type ParseError struct {
	Message  string
	Found    string
	Expected []string
	Pos      Pos
}

// newParseError returns a new instance of ParseError.
func newParseError(found string, expected []string, pos Pos) *ParseError {
	return &ParseError{Found: found, Expected: expected, Pos: pos}
}

// Error returns the string representation of the error.
func (e *ParseError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s at line %d, char %d", e.Message, e.Pos.Line+1, e.Pos.Char+1)
	}
	return fmt.Sprintf("found %s, expected %s at line %d, char %d", e.Found, strings.Join(e.Expected, ", "), e.Pos.Line+1, e.Pos.Char+1)
}
