package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookup(t *testing.T) {
	assert.Equal(t, Lookup("Brian"), IDENT)
}

func TestString(t *testing.T) {
	assert.Equal(t, Tokens[IDENT], "IDENT")
	assert.Equal(t, Tokens[LPAREN], "LPAREN '('")
}
