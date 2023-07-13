package pgLogParser

import "io"

// tokenizer.go provides functions to extract a slice of tokens.
// It uses scanner.go to get the individual token, and then appends
// it to a slice.

// tokenizer represents a wrapper for scanner to add a buffer.
// It provides a fixed-length circular buffer that can be unread.
type tokenizer struct {
	s   *Scanner
	i   int // buffer index
	n   int // buffer size
	buf [3]struct {
		tok Token
		pos Pos
		lit string
	}
}

// newTokenizer returns a new buffered scanner for a reader.
func newTokenizer(r io.Reader) *tokenizer {
	return &tokenizer{s: NewScanner(r)}
}

// Scan reads the next token from the scanner.
func (t *tokenizer) Scan() (tok Token, pos Pos, lit string) {
	// If we have unread tokens then read them off the buffer first.
	if t.n > 0 {
		t.n--
		return t.Curr()
	}

	// Move buffer position forward and save the token.
	t.i = (t.i + 1) % len(t.buf)
	buf := &t.buf[t.i]
	buf.tok, buf.pos, buf.lit = t.s.Scan()

	return t.Curr()
}

// Unscan pushes the previous token back onto the buffer.
func (t *tokenizer) Unscan() {
	t.n++
}

func (t *tokenizer) UnscanAndReturn() (tok Token, pos Pos, lit string) {
	t.Unscan()
	return t.Curr()
}

// Curr returns the last read token.
func (t *tokenizer) Curr() (tok Token, pos Pos, lit string) {
	buf := &t.buf[(t.i-t.n+len(t.buf))%len(t.buf)]
	return buf.tok, buf.pos, buf.lit
}
