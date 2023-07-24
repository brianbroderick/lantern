package pgLogParser

import "fmt"

var Language = &ParseTree{}

type ParseTree struct {
	Handlers map[Token]func(*Parser) (Statement, error)
	Keys     []string
}

func init() {
	// Iterate INTEGER times over rest of the command
	Language.Handle(DATE, func(p *Parser) (Statement, error) {
		return p.parseLogStatement()
	})
}

// Handle registers a handler to be invoked when seeing the given token.
func (t *ParseTree) Handle(tok Token, fn func(*Parser) (Statement, error)) {
	// Verify that there is no conflict for this token in this parse tree.
	if _, conflict := t.Handlers[tok]; conflict {
		panic(fmt.Sprintf("conflict for token %s", tok))
	}

	if t.Handlers == nil {
		t.Handlers = make(map[Token]func(*Parser) (Statement, error))
	}
	t.Handlers[tok] = fn
	t.Keys = append(t.Keys, tok.String())
}

// Parse parses a statement using the language defined in the parse tree.
func (t *ParseTree) Parse(p *Parser) (Statement, error) {
	// for { // not sure why we had this loop
	tok, pos, lit := p.ScanIgnoreWhitespace()

	if stmt := t.Handlers[tok]; stmt != nil {
		return stmt(p)
	}

	// There were no registered handlers. Return the valid tokens in the order they were added.
	return nil, newParseError(tokstr(tok, lit), t.Keys, pos)
	// }
}

// Statement represents a single command
type Statement interface {
	String() string
	KeyTok() Token
}
