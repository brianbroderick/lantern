package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestRegexMessage(t *testing.T) {
	// check standard statement
	message := "duration: 0.083 ms  statement: SET time zone 'UTC'"
	result, err := regexMessage(message)
	assert.NoError(t, err)
	assert.Equal(t, "statement", result["preparedStep"])

	// check prepared statement
	message = "duration: 0.066 ms  bind <unnamed>: select * from servers where id = 1"
	result, err = regexMessage(message)
	assert.NoError(t, err)
	assert.Equal(t, "bind", result["preparedStep"])

	// check non-greedy to colon
	message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah'"
	result, err = regexMessage(message)
	assert.NoError(t, err)
	assert.Equal(t, "bind", result["preparedStep"])
}