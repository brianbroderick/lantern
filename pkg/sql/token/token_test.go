package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookup(t *testing.T) {
	// should find and return a token
	tok := Lookup("select")
	assert.Equal(t, SELECT, tok)

	// Not found, so it's an ident
	tok = Lookup("scooby")
	assert.Equal(t, IDENT, tok)
}
